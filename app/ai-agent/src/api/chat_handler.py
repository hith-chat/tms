"""Chat handler for receiving and processing chat messages with SSE response."""

import logging
import asyncio
from typing import AsyncGenerator
from fastapi import APIRouter, HTTPException, Depends
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
import json

from ..services.auth_service import auth_service
from ..services.agent_service import AgentService
from sqlalchemy.ext.asyncio import AsyncSession

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/chat", tags=["chat"])


class ChatRequest(BaseModel):
    message: str
    tenant_id: str
    project_id: str
    session_id: str
    user_id: str = None
    metadata: dict = None


class ChatResponse(BaseModel):
    type: str  # "message", "thinking", "error", "done"
    content: str = None
    metadata: dict = None


@router.post("/process")
async def process_chat_message(
    request: ChatRequest 
):
    """
    Process chat message and return streaming SSE response.
    
    This endpoint:
    1. Authenticates with Go service using tenant/project
    2. Processes the message using AI agent
    3. Returns streaming response via SSE
    """
    try:
        logger.info(f"Processing chat message for session {request.session_id}")
        
        # Authenticate with Go service
        auth_token = await auth_service.authenticate(request.tenant_id, request.project_id)
        
        if not auth_token:
            raise HTTPException(status_code=401, detail="Failed to authenticate with Go service")
        
        # Process message with streaming response
        return StreamingResponse(
            _process_message_stream(request, auth_token),
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
                "X-Accel-Buffering": "no",  # Disable nginx buffering
            }
        )
        
    except Exception as e:
        logger.error(f"Error processing chat message: {e}")
        raise HTTPException(status_code=500, detail=str(e))


async def _process_message_stream(request: ChatRequest, auth_token: str) -> AsyncGenerator[str, None]:
    """Generate SSE stream for chat message processing."""
    try:
        # Send initial thinking message
        yield _format_sse_message(ChatResponse(
            type="thinking",
            content="Processing your message...",
            metadata={"session_id": request.session_id}
        ))
        
        # Initialize agent service
        agent_service = AgentService()
        
        # Process message with streaming (now with database session)
        async for response in agent_service.process_message_stream(
            message=request.message,
            session_id=request.session_id,
            auth_token=auth_token,
            tenant_id=request.tenant_id,
            project_id=request.project_id,
            metadata=request.metadata
        ):
            yield _format_sse_message(response)
        
        # Send completion message
        yield _format_sse_message(ChatResponse(
            type="done",
            content="Message processing complete",
            metadata={"session_id": request.session_id}
        ))
        
    except Exception as e:
        logger.error(f"Error in message stream: {e}")
        yield _format_sse_message(ChatResponse(
            type="error",
            content=f"Error processing message: {str(e)}",
            metadata={"session_id": request.session_id}
        ))


def _format_sse_message(response: ChatResponse) -> str:
    """Format response as SSE message."""
    logger.info(f"Formatting SSE message: {response}")
    if isinstance(response, dict):
        data = response
    else:
        data = response.model_dump()
    return f"data: {json.dumps(data)}\n\n"


# Health check endpoint
@router.get("/health")
async def health_check():
    """Health check endpoint."""
    return {"status": "healthy", "service": "agent-chat-handler"}
