"""Swarm agent service for orchestrating AI agents."""

import json
import logging
from typing import Dict, List, Optional, AsyncGenerator, Any
from dataclasses import dataclass, field
from sqlalchemy.ext.asyncio import AsyncSession
from swarm import Swarm, Agent

from ..models.chat import ChatSession, ChatMessage
from ..schemas.chat import ContactInfoUpdate
from ..schemas.ticket import TicketCreate
from .knowledge_service import KnowledgeService
from .ticket_service import TicketService
from .chat_service import ChatService

logger = logging.getLogger(__name__)


@dataclass
class AgentSession:
    """In-memory session state for agent interactions."""
    session_id: str
    current_agent: str = "support"
    context: Dict = field(default_factory=dict)
    temp_data: Dict = field(default_factory=dict)


class SwarmService:
    """Service for managing Swarm AI agents."""
    
    def __init__(self):
        self.client = Swarm()
        self.knowledge_service = KnowledgeService()
        self.ticket_service = TicketService()
        self.chat_service = ChatService()
        self.sessions: Dict[str, AgentSession] = {}
        
        # Initialize agents
        self._setup_agents()
    
    def _setup_agents(self):
        """Setup Swarm agents with their functions."""
        
        # Support agent - main entry point
        self.support_agent = Agent(
            name="Support Agent",
            instructions="""You are a helpful customer support agent. Your approach:
            1. First search the knowledge base for answers
            2. If no relevant KB article, offer to create a support ticket
            3. Be professional, empathetic, and concise
            4. Collect contact information when creating tickets
            
            Always use search_knowledge_base first for support questions.
            If the KB doesn't have the answer, use create_support_ticket.
            For complex issues, use escalate_to_human.""",
            functions=[
                self._search_knowledge_base,
                self._create_support_ticket,
                self._escalate_to_human
            ]
        )
        
        # Contact agent for collecting user information
        self.contact_agent = Agent(
            name="Contact Agent", 
            instructions="""Collect user contact information politely.
            Ask for email first (required), then phone (optional).
            Explain this helps us follow up on their issue.
            Use save_contact_info to store the information.""",
            functions=[
                self._collect_contact_information,
                self._save_contact_info
            ]
        )
        
        # Ticket agent for creating support tickets
        self.ticket_agent = Agent(
            name="Ticket Agent",
            instructions="""Create support tickets for unresolved issues.
            Gather: clear issue description, any error messages, what user tried.
            Use create_support_ticket to create the ticket.
            After ticket creation, switch to contact_agent for contact details.""",
            functions=[self._create_support_ticket],
            handoff_to=[self.contact_agent]
        )
    
    async def _search_knowledge_base(self, query: str, session_id: str, db_session: AsyncSession) -> str:
        """Search KB for relevant information."""
        try:
            articles = await self.knowledge_service.search_knowledge_base(
                db_session, query
            )
            
            response = await self.knowledge_service.format_kb_response(articles)
            return response or "I couldn't find relevant information in our knowledge base."
            
        except Exception as e:
            logger.error(f"KB search error: {e}")
            return "I encountered an issue searching our knowledge base. Let me create a ticket for you."
    
    async def _create_support_ticket(
        self,
        title: str,
        description: str,
        session_id: str,
        db_session: AsyncSession
    ) -> str:
        """Create a support ticket."""
        try:
            # Get or create chat session
            chat_session = await self.chat_service.get_or_create_session(
                db_session, session_id
            )
            
            ticket_data = TicketCreate(
                title=title,
                description=description,
                session_id=session_id,
                user_email=chat_session.email
            )
            
            ticket_id = await self.ticket_service.create_ticket(db_session, ticket_data)
            
            # Update in-memory session
            if session_id in self.sessions:
                self.sessions[session_id].context['ticket_id'] = ticket_id
            
            return f"I've created support ticket #{ticket_id}. Our team will review it shortly."
            
        except Exception as e:
            logger.error(f"Ticket creation error: {e}")
            return "I encountered an issue creating the ticket. Please try again."
    
    def _escalate_to_human(self, reason: str, session_id: str) -> str:
        """Mark session for human escalation."""
        if session_id in self.sessions:
            self.sessions[session_id].context['escalated'] = True
            self.sessions[session_id].context['escalation_reason'] = reason
        
        return "I'll connect you with a human agent who can better assist you with this issue."
    
    async def _collect_contact_information(self, session_id: str, db_session: AsyncSession) -> str:
        """Guide contact information collection."""
        session_data = await self.chat_service.get_or_create_session(
            db_session, session_id
        )
        
        if not session_data.email:
            return "To better assist you, could you please provide your email address?"
        elif not session_data.phone:
            return "Thank you! Could you also provide a phone number where we can reach you if needed?"
        else:
            return "We have your contact information and will reach out soon."
    
    async def _save_contact_info(
        self,
        session_id: str,
        email: str = None,
        phone: str = None,
        name: str = None,
        db_session: AsyncSession = None
    ) -> str:
        """Save user's contact information."""
        try:
            contact_info = ContactInfoUpdate(email=email, phone=phone, name=name)
            await self.chat_service.update_contact_info(
                db_session, session_id, contact_info
            )
            
            # Update in-memory session
            if session_id in self.sessions:
                if email:
                    self.sessions[session_id].context['has_email'] = True
                if phone:
                    self.sessions[session_id].context['has_phone'] = True
            
            return "Thank you! I've saved your contact information."
            
        except Exception as e:
            logger.error(f"Contact save error: {e}")
            return "I had trouble saving your information. Please try again."
    
    async def process_message_stream(
        self,
        db_session: AsyncSession,
        session_id: str,
        message: str
    ) -> AsyncGenerator[str, None]:
        """
        Process message and stream response.
        
        Args:
            db_session: Database session
            session_id: Chat session ID
            message: User message
            
        Yields:
            Formatted SSE chunks with response data
        """
        try:
            # Get or create agent session
            if session_id not in self.sessions:
                self.sessions[session_id] = AgentSession(session_id=session_id)
            
            agent_session = self.sessions[session_id]
            
            # Save user message
            await self.chat_service.save_message(
                db_session, session_id, "user", message
            )
            
            # Get recent messages for context
            recent_messages = await self.chat_service.get_recent_messages(
                db_session, session_id
            )
            
            # Prepare context variables for Swarm
            context_variables = {
                "session_id": session_id,
                "db_session": db_session,
                "agent_session": agent_session
            }
            
            # Run Swarm agent
            response = self.client.run(
                agent=self.support_agent,
                messages=recent_messages + [{"role": "user", "content": message}],
                context_variables=context_variables,
                stream=True
            )
            
            # Stream response chunks
            full_response = ""
            for chunk in response:
                if chunk.get("content"):
                    content = chunk["content"]
                    full_response += content
                    yield f"data: {json.dumps({'content': content})}\n\n"
            
            # Save assistant response
            await self.chat_service.save_message(
                db_session, session_id, "assistant", full_response
            )
            
            # Send final metadata
            metadata = {
                "ticket_id": agent_session.context.get('ticket_id'),
                "escalated": agent_session.context.get('escalated', False),
                "has_email": agent_session.context.get('has_email', False)
            }
            yield f"data: {json.dumps({'metadata': metadata, 'done': True})}\n\n"
            
        except Exception as e:
            logger.error(f"Stream processing error: {e}")
            yield f"data: {json.dumps({'error': str(e)})}\n\n"
    
    def get_session_context(self, session_id: str) -> Dict[str, Any]:
        """Get session context for external use."""
        if session_id in self.sessions:
            return self.sessions[session_id].context.copy()
        return {}
