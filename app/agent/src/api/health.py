"""Health check API endpoints."""

import logging
from fastapi import APIRouter, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text

from ..database import get_db_session
from ..config import config

logger = logging.getLogger(__name__)

router = APIRouter(tags=["health"])


@router.get("/health")
async def health_check(db_session: AsyncSession = Depends(get_db_session)):
    """
    Health check endpoint.
    
    Returns:
        Service health status including database connectivity
    """
    try:
        # Test database connection
        await db_session.execute(text("SELECT 1"))
        
        return {
            "status": "healthy",
            "service": "agent-service",
            "database": "connected",
            "openai_configured": bool(config.OPENAI_API_KEY)
        }
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        return {
            "status": "unhealthy", 
            "service": "agent-service",
            "database": "disconnected",
            "openai_configured": bool(config.OPENAI_API_KEY),
            "error": str(e)
        }


@router.get("/ready")
async def readiness_check():
    """
    Readiness check endpoint for Kubernetes.
    
    Returns:
        Service readiness status
    """
    try:
        # Check if required configuration is available
        if not config.OPENAI_API_KEY:
            return {"status": "not_ready", "reason": "OpenAI API key not configured"}
        
        return {"status": "ready", "service": "agent-service"}
    except Exception as e:
        logger.error(f"Readiness check failed: {e}")
        return {"status": "not_ready", "reason": str(e)}
