"""Chat session and message management service using async SQLAlchemy."""

import logging
from typing import List, Dict, Optional
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, desc

from ..models.chat import ChatSession, ChatMessage
from ..schemas.chat import ContactInfoUpdate, ChatSessionCreate

logger = logging.getLogger(__name__)


class ChatService:
    """Service for chat operations."""
    
    async def get_session_info(
        self,
        session: AsyncSession,
        session_id: str
    ) -> Optional[ChatSession]:
        """
        Get existing session information (READ-ONLY).
        
        Args:
            session: Database session
            session_id: Unique session identifier
            
        Returns:
            ChatSession object if exists, None otherwise
        """
        try:
            query = select(ChatSession).where(ChatSession.session_id == session_id)  # session_id maps to client_session_id
            result = await session.execute(query)
            chat_session = result.scalar_one_or_none()
            return chat_session
            
        except Exception as e:
            logger.error(f"Error getting session {session_id}: {e}")
            return None
    
    # REMOVED: save_message, update_contact_info methods 
    # Agent is READ-ONLY and doesn't perform database writes
    
    async def get_recent_messages(
        self,
        session: AsyncSession,
        session_id: str,
        limit: int = 10
    ) -> List[Dict[str, str]]:
        """
        Get recent messages for a session.
        
        Args:
            session: Database session
            session_id: Chat session ID
            limit: Maximum number of messages to return
            
        Returns:
            List of message dictionaries with role and content
        """
        try:
            query = select(ChatMessage.role, ChatMessage.content).where(
                ChatMessage.session_id == session_id
            ).order_by(desc(ChatMessage.created_at)).limit(limit)
            
            result = await session.execute(query)
            messages = result.fetchall()
            
            # Return in chronological order (oldest first)
            return [
                {"role": msg.role, "content": msg.content}
                for msg in reversed(messages)
            ]
            
        except Exception as e:
            logger.error(f"Error fetching messages for session {session_id}: {e}")
            return []
    

    
    async def get_session_messages(
        self,
        session: AsyncSession,
        session_id: str,
        offset: int = 0,
        limit: int = 50
    ) -> List[ChatMessage]:
        """
        Get paginated messages for a session.
        
        Args:
            session: Database session
            session_id: Chat session ID
            offset: Number of messages to skip
            limit: Maximum number of messages to return
            
        Returns:
            List of ChatMessage objects
        """
        try:
            query = select(ChatMessage).where(
                ChatMessage.session_id == session_id
            ).order_by(ChatMessage.created_at).offset(offset).limit(limit)
            
            result = await session.execute(query)
            return result.scalars().all()
            
        except Exception as e:
            logger.error(f"Error fetching session messages for {session_id}: {e}")
            return []
