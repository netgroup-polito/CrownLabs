FROM python:3.8-slim

WORKDIR /usr/share/app

COPY list-crownlabs-images.py requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

ENTRYPOINT [ "python", "./list-crownlabs-images.py" ]
