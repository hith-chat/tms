"""Chat API endpoints."""

import logging
from fastapi import APIRouter, Depends, HTTPException
from fastapi.responses import StreamingResponse
from sqlalchemy.ext.asyncio import AsyncSession

from ..database import get_db_session
from ..schemas.chat import ChatRequest, ChatSessionResponse
from ..services.swarm_service import SwarmService
from ..services.chat_service import ChatService

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/chat", tags=["chat"])

# Initialize services
swarm_service = SwarmService()
chat_service = ChatService()


@router.post("/stream")
async def chat_stream(
    request: ChatRequest,
    db_session: AsyncSession = Depends(get_db_session)
):
    """
    Streaming chat endpoint for real-time AI agent responses.
    
    Args:
        request: Chat request with session_id and message
        db_session: Database session dependency
        
    Returns:
        StreamingResponse with Server-Sent Events
    """
    try:
        return StreamingResponse(
            swarm_service.process_message_stream(
                db_session, request.session_id, request.message
            ),
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
            }
        )
    except Exception as e:
        logger.error(f"Chat stream error: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")

