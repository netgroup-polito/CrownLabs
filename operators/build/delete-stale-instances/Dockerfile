FROM python:3.8-slim

RUN mkdir -p /usr/src/app


WORKDIR /usr/src/app

COPY ./cmd/delete-stale-instances/delete-stale-instances.py ./
COPY ./build/delete-stale-instances/requirements.txt ./
COPY ./deploy/delete-stale-instances/k8s-manifest-example.env ./

RUN pip install --no-cache-dir -r requirements.txt

ENTRYPOINT ["python3", "delete-stale-instances.py"]
