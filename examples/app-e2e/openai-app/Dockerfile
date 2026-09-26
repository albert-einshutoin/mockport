FROM python:3.13-alpine
WORKDIR /app
COPY requirements.lock ./
RUN pip install --no-cache-dir -r requirements.lock
COPY server.py ./
USER nobody
CMD ["python", "server.py"]
