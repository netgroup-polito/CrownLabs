FROM python:3.8-slim

RUN mkdir -p /usr/src/app

WORKDIR /usr/src/app

COPY ./cmd/delete-stale-instances/delete_stale_instances.py ./
COPY ./build/delete-stale-instances/requirements.txt ./

RUN pip install --no-cache-dir -r requirements.txt

ENTRYPOINT ["python3", "delete_stale_instances.py"]
