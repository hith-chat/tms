# Dockerfile
FROM python:3.12-slim

WORKDIR /app

# Install system dependencies for PostgreSQL and compilation
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    postgresql-client \
    libpq-dev \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application source
COPY src/ ./src/

# Set Python path
ENV PYTHONPATH=/app

# Do not hardcode PORT into the image. Use a runtime default instead so callers
# can override at `docker run` / compose / k8s without the value being baked in.
# The shell expansion below uses a fallback of 5000 if PORT is not provided.
CMD ["sh", "-c", "uvicorn src.main:app --host 0.0.0.0 --port ${PORT:-5000}"]