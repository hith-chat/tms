"""Ticket management service using async SQLAlchemy."""

import logging
from typing import Optional
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select

from ..models.chat import SupportTicket, ChatSession
from ..schemas.ticket import TicketCreate

logger = logging.getLogger(__name__)


class TicketService:
    """Service for ticket operations."""
    
    async def create_ticket(
        self,
        session: AsyncSession,
        ticket_data: TicketCreate
    ) -> int:
        """
        Create a support ticket.
        
        Args:
            session: Database session
            ticket_data: Ticket creation data
            
        Returns:
            Created ticket ID
            
        Raises:
            ValueError: If session doesn't exist
            Exception: For database errors
        """
        try:
            # Verify session exists
            session_query = select(ChatSession).where(
                ChatSession.session_id == ticket_data.session_id
            )
            session_result = await session.execute(session_query)
            chat_session = session_result.scalar_one_or_none()
            
            if not chat_session:
                raise ValueError(f"Chat session {ticket_data.session_id} not found")
            
            # Create ticket
            ticket = SupportTicket(
                session_id=ticket_data.session_id,
                title=ticket_data.title,
                description=ticket_data.description,
                priority=ticket_data.priority,
                category=ticket_data.category,
                status="open"
            )
            
            session.add(ticket)
            await session.commit()
            await session.refresh(ticket)
            
            logger.info(f"Created ticket #{ticket.ticket_id} for session {ticket_data.session_id}")
            return ticket.ticket_id
            
        except Exception as e:
            await session.rollback()
            logger.error(f"Ticket creation error: {e}")
            raise
    
    async def get_ticket(
        self,
        session: AsyncSession,
        ticket_id: int
    ) -> Optional[SupportTicket]:
        """
        Get a ticket by ID.
        
        Args:
            session: Database session
            ticket_id: Ticket ID
            
        Returns:
            Ticket if found, None otherwise
        """
        try:
            query = select(SupportTicket).where(SupportTicket.ticket_id == ticket_id)
            result = await session.execute(query)
            return result.scalar_one_or_none()
        except Exception as e:
            logger.error(f"Error fetching ticket {ticket_id}: {e}")
            return None
    
    async def update_ticket_status(
        self,
        session: AsyncSession,
        ticket_id: int,
        status: str
    ) -> bool:
        """
        Update ticket status.
        
        Args:
            session: Database session
            ticket_id: Ticket ID
            status: New status
            
        Returns:
            True if updated successfully, False otherwise
        """
        try:
            query = select(SupportTicket).where(SupportTicket.ticket_id == ticket_id)
            result = await session.execute(query)
            ticket = result.scalar_one_or_none()
            
            if ticket:
                ticket.status = status
                await session.commit()
                logger.info(f"Updated ticket #{ticket_id} status to {status}")
                return True
            else:
                logger.warning(f"Ticket #{ticket_id} not found for status update")
                return False
                
        except Exception as e:
            await session.rollback()
            logger.error(f"Error updating ticket {ticket_id} status: {e}")
            return False
