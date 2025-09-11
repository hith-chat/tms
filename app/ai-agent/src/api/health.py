"""Health check API endpoints."""

import logging
from fastapi import APIRouter, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text

from ..config import config

logger = logging.getLogger(__name__)

router = APIRouter(tags=["health"])


@router.get("/api/health")
async def health_check():
    """
    Health check endpoint.
    
    Returns:
        Service health status including database connectivity
    """
    try:        
        return {
            "status": "healthy",
            "service": "agent-service",
            "database": "connected",
            "openai_configured": bool(config.AI_API_KEY)
        }
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        return {
            "status": "unhealthy", 
            "service": "agent-service",
            "database": "disconnected",
            "openai_configured": bool(config.AI_API_KEY),
            "error": str(e)
        }
