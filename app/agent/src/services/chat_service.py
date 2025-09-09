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
    
    async def get_or_create_session(
        self,
        session: AsyncSession,
        session_id: str
    ) -> ChatSession:
        """
        Get existing session or create a new one.
        
        Args:
            session: Database session
            session_id: Unique session identifier
            
        Returns:
            ChatSession object
        """
        try:
            # Try to get existing session
            query = select(ChatSession).where(ChatSession.session_id == session_id)
            result = await session.execute(query)
            chat_session = result.scalar_one_or_none()
            
            if not chat_session:
                # Create new session
                chat_session = ChatSession(session_id=session_id)
                session.add(chat_session)
                await session.commit()
                await session.refresh(chat_session)
                logger.info(f"Created new chat session: {session_id}")
            
            return chat_session
            
        except Exception as e:
            await session.rollback()
            logger.error(f"Error getting/creating session {session_id}: {e}")
            raise
    
    async def save_message(
        self,
        session: AsyncSession,
        session_id: str,
        role: str,
        content: str
    ) -> ChatMessage:
        """
        Save a message to the database.
        
        Args:
            session: Database session
            session_id: Chat session ID
            role: Message role (user, assistant, system)
            content: Message content
            
        Returns:
            Created ChatMessage object
        """
        try:
            # Ensure session exists
            await self.get_or_create_session(session, session_id)
            
            # Create message
            message = ChatMessage(
                session_id=session_id,
                role=role,
                content=content
            )
            
            session.add(message)
            await session.commit()
            await session.refresh(message)
            
            return message
            
        except Exception as e:
            await session.rollback()
            logger.error(f"Error saving message for session {session_id}: {e}")
            raise
    
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
    
    async def update_contact_info(
        self,
        session: AsyncSession,
        session_id: str,
        contact_info: ContactInfoUpdate
    ) -> bool:
        """
        Update contact information for a session.
        
        Args:
            session: Database session
            session_id: Chat session ID
            contact_info: Contact information to update
            
        Returns:
            True if updated successfully, False otherwise
        """
        try:
            # Get session
            query = select(ChatSession).where(ChatSession.session_id == session_id)
            result = await session.execute(query)
            chat_session = result.scalar_one_or_none()
            
            if not chat_session:
                logger.warning(f"Session {session_id} not found for contact update")
                return False
            
            # Update fields that are provided
            if contact_info.email is not None:
                chat_session.email = contact_info.email
            if contact_info.phone is not None:
                chat_session.phone = contact_info.phone
            if contact_info.name is not None:
                chat_session.name = contact_info.name
            
            await session.commit()
            logger.info(f"Updated contact info for session {session_id}")
            return True
            
        except Exception as e:
            await session.rollback()
            logger.error(f"Error updating contact info for session {session_id}: {e}")
            return False
    
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
