# Dockerfile
FROM python:3.11-slim

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

# Run the application
CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "5000"]