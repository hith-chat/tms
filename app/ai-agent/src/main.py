"""Main FastAPI application."""

import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .config import config
from .database import init_db, close_db
from .api import chat_router, chat_handler_router, health_router

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events."""
    # Startup
    try:
        config.validate()
        await init_db()
        logger.info("Agent service started successfully")
    except Exception as e:
        logger.error(f"Failed to start agent service: {e}")
        raise
    
    yield
    
    # Shutdown
    try:
        await close_db()
        logger.info("Agent service stopped")
    except Exception as e:
        logger.error(f"Error during shutdown: {e}")


# Create FastAPI application
app = FastAPI(
    title="TMS Agent Service",
    description="AI-powered customer support agent service using OpenAI Agents SDK",
    version="1.0.0",
    lifespan=lifespan
)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(chat_router)
app.include_router(chat_handler_router)
app.include_router(health_router)


# Root endpoint
@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "service": "tms-agent-service",
        "status": "running",
        "docs": "/docs",
        "health": "/health"
    }
