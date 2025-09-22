# Hith Email Server Docker Compose Setup

## Overview

The email server is now fully integrated into the Docker Compose stack and can be built and deployed automatically without manual Docker commands.

## Quick Start

### 1. Deploy Complete Stack
```bash
cd /Users/sumansaurabh/Documents/bareuptime/tms/deploy
./deploy.sh
```

### 2. Deploy Email Server Only
```bash
cd /Users/sumansaurabh/Documents/bareuptime/tms/deploy
./email-server.sh start
```

## Available Scripts

### Main Deployment Script: `deploy.sh`
- **Full deployment**: `./deploy.sh`
- **Rebuild all**: `./deploy.sh --rebuild`
- **Run in foreground**: `./deploy.sh --foreground`
- **Skip building**: `./deploy.sh --no-build`

### Email Server Script: `email-server.sh`
- **Build only**: `./email-server.sh build`
- **Start with dependencies**: `./email-server.sh start`
- **View logs**: `./email-server.sh logs`
- **Test connectivity**: `./email-server.sh test`
- **Rebuild from scratch**: `./email-server.sh rebuild`
- **Check status**: `./email-server.sh status`

## Configuration

### Environment Variables (`.env` file)
```bash
# Email Server
MAIL_DOMAIN=yourmailserver.com
MAX_MESSAGE_SIZE=1048576

# Other services...
```

### Docker Compose Service Definition
```yaml
guerrilla-mail:
  build:
    context: ../app/email-server
    dockerfile: Dockerfile
  container_name: tms-guerrilla-mail
  ports:
    - "25:25"     # SMTP port
    - "587:25"    # Alternative submission port
  environment:
    - MAIL_DOMAIN=${MAIL_DOMAIN:-yourmailserver.com}
    - TICKET_API_URL=http://backend:8080/v1/public/email-to-ticket
    - LISTEN_INTERFACE=0.0.0.0:25
    - MAX_MESSAGE_SIZE=${MAX_MESSAGE_SIZE:-1048576}
  depends_on:
    - backend
  networks:
    - tms-network
  restart: unless-stopped
```

## Service URLs

| Service | URL | Purpose |
|---------|-----|---------|
| Email Server (SMTP) | localhost:25 | Receive emails |
| Backend API | http://localhost:8080 | Main API |
| Agent Console | http://localhost:5173 | Admin interface |
| Public View | http://localhost:5174 | Customer interface |
| MailHog Web UI | http://localhost:8025 | Email testing |
| PgAdmin | http://localhost:5050 | Database admin |
| MinIO Console | http://localhost:9001 | Object storage |

## Email Processing Flow

1. **Email Reception**: SMTP server receives email on port 25
2. **Tenant Extraction**: Parse tenant from `tenant-{name}@domain.com`
3. **Content Processing**: Clean and structure email content
4. **API Integration**: POST to `backend:8080/v1/public/email-to-ticket`
5. **Ticket Creation**: Backend creates ticket in database

## Development Workflow

### 1. Make changes to email server code
```bash
# Edit files in app/email-server/
```

### 2. Rebuild and test
```bash
cd deploy
./email-server.sh rebuild
./email-server.sh test
```

### 3. View logs
```bash
./email-server.sh logs
```

### 4. Test email processing
Send test emails to: `tenant-{name}@{your-domain}`

## Production Deployment

### Prerequisites
- Docker and Docker Compose installed
- Ports 25, 587 available for SMTP
- Domain configured to point to your server
- MX record pointing to your server

### Steps
1. **Clone and configure**:
   ```bash
   git clone <repository>
   cd tms/deploy
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Deploy**:
   ```bash
   ./deploy.sh
   ```

3. **Configure DNS**:
   - Add MX record pointing to your server
   - Add A record for your mail domain

4. **Test**:
   ```bash
   ./email-server.sh test
   ```

## Monitoring

### View all service status
```bash
docker-compose ps
```

### View email server logs
```bash
./email-server.sh logs
```

### Monitor resource usage
```bash
docker-compose top
```

## Troubleshooting

### Email server won't start
```bash
./email-server.sh status
./email-server.sh logs
```

### Cannot receive emails
1. Check port 25 is accessible
2. Verify DNS MX records
3. Check firewall settings
4. Test with: `./email-server.sh test`

### Backend API connection issues
1. Ensure backend service is running
2. Check network connectivity
3. Verify API endpoint URL

### Build issues
```bash
./email-server.sh clean
./email-server.sh rebuild
```

## Security Considerations

- Email server runs as non-root user
- Resource limits configured
- Health checks enabled
- Restart policies configured
- Network isolation with Docker networks

## Advanced Configuration

### Custom environment variables
Add to `.env` file:
```bash
MAIL_DOMAIN=your-domain.com
MAX_MESSAGE_SIZE=2097152  # 2MB
```

### Scale email server
```bash
docker-compose up -d --scale guerrilla-mail=3
```

### Custom Docker build args
Edit `docker-compose.yml` build section to add build arguments if needed.
