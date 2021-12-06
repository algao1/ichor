from concurrent import futures

import math
from datetime import datetime, timedelta
from google.protobuf.timestamp_pb2 import Timestamp

import lightgbm as lgb
import numpy as np
import pandas as pd

import grpc
import glucose_pb2
import glucose_pb2_grpc

def transformTimeSeries(request):
  tts = []
  dt = datetime.fromtimestamp(request.time.seconds)

  # Hour sin/cos.
  tts.append(math.sin(dt.hour * 2 * math.pi / 24))
  tts.append(math.cos(dt.hour * 2 * math.pi / 24))

  # Day sin/cos.
  tts.append(math.sin(dt.day * 2 * math.pi / 7))
  tts.append(math.cos(dt.day * 2 * math.pi / 7))

  # Value change pos/neg, max pos/neg.
  s = pd.Series(np.array(request.values))
  sc = s.diff().dropna()

  tts.append(sc.max())
  tts.append(sc.min())
  tts.append(sc[sc > 0].sum())
  tts.append(sc[sc < 0].sum())

  # Difference at 15, 30, 60 minutes.
  tts.append(s.diff(3).iloc[-1])
  tts.append(s.diff(6).iloc[-1])
  tts.append(s.diff(12).iloc[-1])

  # Standard deviation at 1 hour and 2 hour.
  tts.append(s.rolling(12).std().iloc[-1])
  tts.append(s.rolling(24).std().iloc[-1])

  tts.extend(reversed(request.values))

  return np.array(tts).reshape((1, -1))

class GlucoseServicer(glucose_pb2_grpc.GlucoseServicer):

  def __init__(self):
    self.model = lgb.Booster(model_file="models/lgbm_regression.txt")
    print("Model loaded.")

  def Predict(self, request, context):
    if len(request.values) < 24:
      return glucose_pb2.Label(value=4.2)
    else:
      pred = self.model.predict(transformTimeSeries(request))
      predTime = datetime.fromtimestamp(request.time.seconds) + timedelta(hours=2)
      ts = Timestamp()
      ts.FromDatetime(predTime)
      return glucose_pb2.Label(value=math.exp(pred), time=ts)

def serve():
  server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
  glucose_pb2_grpc.add_GlucoseServicer_to_server(GlucoseServicer(), server)
  server.add_insecure_port("[::]:50051")
  server.start()
  server.wait_for_termination()

if __name__ == "__main__":
  serve()