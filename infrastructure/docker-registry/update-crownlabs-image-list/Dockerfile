FROM python:3.8-slim

WORKDIR /usr/share/app

COPY update-crownlabs-image-list.py requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

ENTRYPOINT [ "python", "update-crownlabs-image-list.py" ]
