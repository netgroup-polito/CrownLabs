FROM python:3.8-slim

WORKDIR /usr/share/app

COPY ./cmd/crownlabs-image-list/update-crownlabs-image-list.py \
     ./build/crownlabs-image-list/requirements.txt \
     ./
RUN pip install --no-cache-dir -r requirements.txt

ENTRYPOINT [ "python", "update-crownlabs-image-list.py" ]
