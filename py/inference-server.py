from concurrent import futures
from datetime import datetime, timedelta
import logging
import math
import os

import glucose_pb2
import glucose_pb2_grpc
from google.protobuf.timestamp_pb2 import Timestamp
import grpc

# import lightgbm as lgb
import numpy as np
# import pandas as pd
# import tensorflow as tf
from tensorflow import keras

# Some constants, will eventually be loaded in
# from a .yaml config file.
LOOK_AHEAD = 6
LOOK_BEHIND = 48

DATA_MEAN = [
  0.512271, # Carbs.
  0.175458, # Insulin.
  8.034759, # Glucose.
]
DATA_STD = [
  5.60357322, # Carbs.
  1.93975756, # Insulin.
  2.65790678, # Glucose.
]

# Similar thing with logging level.
logging.basicConfig(level=logging.DEBUG)
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"

class GlucoseServicer(glucose_pb2_grpc.GlucoseServicer):

  def __init__(self):
    self.model = keras.models.load_model("models/gic_model")
    logging.info("model loaded")

  def Predict(self, request, context):
    if len(request.features) < LOOK_BEHIND:
      context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
      context.set_details("not enough points")
      return glucose_pb2.Labels()
    else:
      # Processing.
      data = []
      for feat in request.features:
        data.append([feat.carbs, feat.insulin, feat.glucose])
      data = np.array(data)
      data = (data - DATA_MEAN) / DATA_STD

      # Make predictions.
      preds = self.repeat_predict(data)
      preds = preds * DATA_STD[-1] + DATA_MEAN[-1]

      last_time = request.features[-1].time.seconds

      labels = []
      for i, pred in enumerate(preds):
        pred_time = datetime.fromtimestamp(last_time) + timedelta(minutes=(i+1)*5)
        ts = Timestamp()
        ts.FromDatetime(pred_time)
        labels.append(glucose_pb2.Label(value=np.float64(pred), time=ts))

      return glucose_pb2.Labels(labels=labels)

  def repeat_predict(self, feats):
    vals = feats
    future = 12

    for i in range(future):
      pred = self.model.predict(vals[i * LOOK_AHEAD:].reshape((1, LOOK_BEHIND, 3)))
      # Carbs | Rapid Insulin | Value.
      pred = np.pad(
        pred.reshape(LOOK_AHEAD, 1), ((0, 0), (1, 0)),
        constant_values=(-DATA_MEAN[0] / DATA_STD[0], 0),
      )
      pred = np.pad(
        pred, ((0, 0), (1, 0)),
        constant_values=(-DATA_MEAN[1] / DATA_STD[1], 0),
      )
      vals = np.append(vals, pred, 0)

    return vals[-future * LOOK_AHEAD:,-1].reshape(72)

def serve():
  server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
  glucose_pb2_grpc.add_GlucoseServicer_to_server(GlucoseServicer(), server)
  server.add_insecure_port("[::]:50051")
  server.start()
  server.wait_for_termination()

if __name__ == "__main__":
  serve()