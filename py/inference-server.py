from concurrent import futures

import logging
import math
import os

from datetime import datetime, timedelta
from google.protobuf.timestamp_pb2 import Timestamp

import lightgbm as lgb
import numpy as np
import pandas as pd
import tensorflow as tf
from tensorflow import keras

import grpc
import glucose_pb2
import glucose_pb2_grpc

os.environ["CUDA_VISIBLE_DEVICES"] = "-1"

class GlucoseServicer(glucose_pb2_grpc.GlucoseServicer):

  def __init__(self):
    self.model = keras.models.load_model("models/dnn_model")
    logging.info("Models loaded.")
  
  def Predict(self, request, context):
    logging.warning(request.values)

    if len(request.values) < 24:
      context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
      context.set_details("Not enough points.")
      return glucose_pb2.Labels()
    else:
      # Processing.
      data = np.array(request.values).reshape([-1, 24])
      mean, std = data.mean(), data.std()

      # Unnormalize data.
      data = (data - mean) / std
      preds = self.model.predict(data)
      preds = preds * std + mean

      labels = []
      for i, pred in enumerate(preds.reshape(12)):
        pred_time = datetime.fromtimestamp(request.time.seconds) + timedelta(minutes=(i+1)*5)
        ts = Timestamp()
        ts.FromDatetime(pred_time)
        labels.append(glucose_pb2.Label(value=np.float64(pred), time=ts))

      return glucose_pb2.Labels(labels=labels)

def serve():
  server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
  glucose_pb2_grpc.add_GlucoseServicer_to_server(GlucoseServicer(), server)
  server.add_insecure_port("[::]:50051")
  server.start()
  server.wait_for_termination()

if __name__ == "__main__":
  serve()