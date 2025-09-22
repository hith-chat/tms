#!/bin/bash

# Hith Production Deployment Script
# This script builds and deploys the complete Hith stack including the email server

set -e  # Exit on any error

echo "🚀 Hith Production Deployment"
echo "=============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "docker-compose.yml" ]; then
    print_error "docker-compose.yml not found. Please run this script from the deploy directory."
    exit 1
fi

print_status "Checking prerequisites..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose > /dev/null 2>&1; then
    print_error "Docker Compose is not installed. Please install Docker Compose and try again."
    exit 1
fi

print_success "Prerequisites check passed"

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    print_warning ".env file not found. Creating from .env.example..."
    cp .env.example .env
    print_status "Please review and update .env file with your configuration"
fi

# Build and deploy options
BUILD_OPTION="--build"
DETACH_OPTION="-d"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-build)
            BUILD_OPTION=""
            shift
            ;;
        --foreground)
            DETACH_OPTION=""
            shift
            ;;
        --rebuild)
            BUILD_OPTION="--build --force-recreate"
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --no-build     Skip building images (use existing images)"
            echo "  --foreground   Run in foreground (don't detach)"
            echo "  --rebuild      Force rebuild and recreate all containers"
            echo "  --help         Show this help message"
            exit 0
            ;;
        *)
            print_warning "Unknown option: $1"
            shift
            ;;
    esac
done

echo ""
print_status "🏗️  Building and starting Hith services..."
print_status "Services to be deployed:"
echo "   📧 Guerrilla Mail Server (email-to-ticket processing)"
echo "   🏢 Backend API"
echo "   🖥️  Agent Console Frontend"
echo "   🌐 Public View Frontend"
echo "   🐘 PostgreSQL Database"
echo "   🔴 Redis Cache"
echo "   📦 MinIO Object Storage"
echo "   📮 MailHog (email testing)"
echo "   🔧 PgAdmin (database admin)"

echo ""
print_status "Starting deployment with options: $BUILD_OPTION $DETACH_OPTION"

# Stop existing containers if running
print_status "Stopping existing containers..."
docker-compose down --remove-orphans

# Pull latest base images
print_status "Pulling latest base images..."
docker-compose pull --ignore-pull-failures

# Build and start services
print_status "Building and starting services..."
if docker-compose up $BUILD_OPTION $DETACH_OPTION; then
    print_success "Hith deployment completed successfully!"
    
    if [ "$DETACH_OPTION" = "-d" ]; then
        echo ""
        print_status "🌐 Service URLs:"
        echo "   📧 Email Server (SMTP):     localhost:25"
        echo "   🏢 Backend API:             http://localhost:8080"
        echo "   🖥️  Agent Console:          http://localhost:5173"
        echo "   🌐 Public View:             http://localhost:5174"
        echo "   📮 MailHog Web UI:          http://localhost:8025"
        echo "   🔧 PgAdmin:                 http://localhost:5050"
        echo "   📦 MinIO Console:           http://localhost:9001"
        
        echo ""
        print_status "📧 Email Configuration:"
        echo "   Send emails to: tenant-{name}@$(grep MAIL_DOMAIN .env | cut -d'=' -f2 || echo 'yourmailserver.com')"
        echo "   Tickets will be created automatically via the backend API"
        
        echo ""
        print_status "📊 Monitoring Commands:"
        echo "   View logs:           docker-compose logs -f"
        echo "   View email logs:     docker-compose logs -f guerrilla-mail"
        echo "   Check status:        docker-compose ps"
        echo "   Stop services:       docker-compose down"
        
        echo ""
        print_status "🔄 Development Commands:"
        echo "   Rebuild all:         ./deploy.sh --rebuild"
        echo "   Run in foreground:   ./deploy.sh --foreground"
        echo "   Skip building:       ./deploy.sh --no-build"
        echo "   Email server only:   ./email-server.sh start"
        echo "   View email logs:     ./email-server.sh logs"
    fi
    
else
    print_error "Deployment failed!"
    print_status "Checking service status..."
    docker-compose ps
    print_status "Recent logs:"
    docker-compose logs --tail=20
    exit 1
fi
