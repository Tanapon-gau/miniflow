FROM python:3.12-slim
WORKDIR /app

COPY pyproject.toml .
RUN pip install --no-cache-dir redis "psycopg[binary]>=3.1"

COPY worker/ ./worker/

CMD ["python", "-m", "worker.main"]
