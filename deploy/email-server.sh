#!/bin/bash

# Email Server Development Script
# Quick commands for email server development and testing

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

EMAIL_SERVICE="guerrilla-mail"
DEPLOY_DIR="/Users/sumansaurabh/Documents/bareuptime/tms/deploy"

# Ensure we're in the right directory
cd "$DEPLOY_DIR"

case "${1:-help}" in
    "build")
        print_status "🏗️  Building email server only..."
        docker-compose build $EMAIL_SERVICE
        print_success "Email server built successfully"
        ;;
    
    "start")
        print_status "🚀 Starting email server and dependencies..."
        # Start dependencies first
        docker-compose up -d postgres redis backend
        # Wait a bit for backend to be ready
        sleep 5
        # Start email server
        docker-compose up -d $EMAIL_SERVICE
        print_success "Email server started"
        docker-compose ps $EMAIL_SERVICE
        ;;
    
    "stop")
        print_status "🛑 Stopping email server..."
        docker-compose stop $EMAIL_SERVICE
        print_success "Email server stopped"
        ;;
    
    "restart")
        print_status "🔄 Restarting email server..."
        docker-compose restart $EMAIL_SERVICE
        print_success "Email server restarted"
        docker-compose ps $EMAIL_SERVICE
        ;;
    
    "logs")
        print_status "📋 Viewing email server logs..."
        docker-compose logs -f $EMAIL_SERVICE
        ;;
    
    "status")
        print_status "📊 Email server status:"
        docker-compose ps $EMAIL_SERVICE
        echo ""
        print_status "📧 Email configuration:"
        echo "   SMTP Port: 25"
        echo "   Domain: $(grep MAIL_DOMAIN .env 2>/dev/null | cut -d'=' -f2 || echo 'yourmailserver.com')"
        echo "   API Endpoint: http://backend:8080/v1/public/email-to-ticket"
        ;;
    
    "test")
        print_status "🧪 Testing email server..."
        
        # Check if container is running
        if ! docker-compose ps $EMAIL_SERVICE | grep -q "Up"; then
            print_error "Email server is not running. Start it first with: $0 start"
            exit 1
        fi
        
        # Test SMTP port
        if docker-compose exec $EMAIL_SERVICE nc -z localhost 25; then
            print_success "SMTP port 25 is accessible"
        else
            print_error "SMTP port 25 is not accessible"
        fi
        
        # Test backend connectivity
        if docker-compose exec $EMAIL_SERVICE nc -z backend 8080; then
            print_success "Backend API is reachable"
        else
            print_warning "Backend API is not reachable"
        fi
        ;;
    
    "rebuild")
        print_status "🔨 Rebuilding email server from scratch..."
        docker-compose stop $EMAIL_SERVICE
        docker-compose rm -f $EMAIL_SERVICE
        docker-compose build --no-cache $EMAIL_SERVICE
        docker-compose up -d $EMAIL_SERVICE
        print_success "Email server rebuilt and started"
        ;;
    
    "shell")
        print_status "🐚 Opening shell in email server container..."
        docker-compose exec $EMAIL_SERVICE sh
        ;;
    
    "clean")
        print_status "🧹 Cleaning email server resources..."
        docker-compose stop $EMAIL_SERVICE
        docker-compose rm -f $EMAIL_SERVICE
        # Remove email server image
        docker rmi $(docker-compose images -q $EMAIL_SERVICE) 2>/dev/null || true
        print_success "Email server resources cleaned"
        ;;
    
    "help"|*)
        echo "📧 Hith Email Server Development Script"
        echo "====================================="
        echo ""
        echo "Usage: $0 <command>"
        echo ""
        echo "Commands:"
        echo "  build      Build email server image"
        echo "  start      Start email server and dependencies"
        echo "  stop       Stop email server"
        echo "  restart    Restart email server"
        echo "  logs       View email server logs (follow mode)"
        echo "  status     Show email server status and configuration"
        echo "  test       Test email server connectivity"
        echo "  rebuild    Rebuild email server from scratch"
        echo "  shell      Open shell in email server container"
        echo "  clean      Clean email server resources"
        echo "  help       Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 start       # Start email server with dependencies"
        echo "  $0 logs        # Follow email server logs"
        echo "  $0 test        # Test email server functionality"
        echo "  $0 rebuild     # Force rebuild and restart"
        ;;
esac
