FROM python:latest AS build_base

WORKDIR /model-inference

COPY ./py/requirements.txt requirements.txt
RUN pip3 install --upgrade pip
RUN pip3 install -r requirements.txt

COPY ./py .

EXPOSE 50051

CMD python3 inference-server.py