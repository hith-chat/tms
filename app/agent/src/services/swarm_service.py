"""Swarm agent service for orchestrating AI agents."""

import json
import logging
from typing import Dict, List, Optional, AsyncGenerator, Any
from dataclasses import dataclass, field
from sqlalchemy.ext.asyncio import AsyncSession
from swarm import Swarm, Agent

from ..models.chat import ChatSession, ChatMessage
from ..schemas.chat import ContactInfoUpdate
from .knowledge_service import KnowledgeService
from .chat_service import ChatService
from .agent_auth_service import AgentAuthService
from .tms_api_client import TMSApiClient

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
        self.chat_service = ChatService()
        self.auth_service = AgentAuthService()
        self.api_client = TMSApiClient(self.auth_service)
        self.sessions: Dict[str, AgentSession] = {}
        
        # Initialize agents
        self._setup_agents()
    
    def _setup_agents(self):
        """Setup Swarm agents with their functions."""
        
        # Support agent - main entry point
        self.support_agent = Agent(
            name="Support Agent",
            instructions="""You are a helpful customer support agent with the ability to perform real actions.

Your capabilities:
1. Search the knowledge base for answers using search_knowledge_base
2. Create actual support tickets via create_support_ticket (this will create a real ticket in the system)
3. Escalate cases to human agents via escalate_to_human (this will notify human agents)
4. Be professional, empathetic, and concise

You can perform actual actions - when you create tickets or escalate, these actions happen immediately in the system.""",
            functions=[
                self._search_knowledge_base,
                self._create_support_ticket,
                self._escalate_to_human
            ]
        )
        
        # Contact agent for collecting user information
        self.contact_agent = Agent(
            name="Contact Agent", 
            instructions="""You collect user contact information and can update it in the system.
            
Your approach:
1. Ask for email first (required), then phone (optional)
2. Explain this helps with follow-up
3. Use save_contact_info to actually update the contact information in the system
4. You can perform real updates - the information will be saved immediately""",
            functions=[
                self._collect_contact_information,
                self._save_contact_info
            ]
        )
        
        # Ticket agent for creating support tickets
        self.ticket_agent = Agent(
            name="Ticket Agent",
            instructions="""You create actual support tickets in the system.
            
Your approach:
1. Gather clear issue description, error messages, what user tried
2. Use create_support_ticket to create a real ticket in the system
3. After ticket creation, hand off to contact_agent for contact details
4. You create real tickets - they will appear in the system immediately""",
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
        """Create a real support ticket via TMS API."""
        try:
            # Get session info to extract tenant_id and project_id
            chat_session = await self.chat_service.get_session_info(
                db_session, session_id
            )
            
            if not chat_session:
                return "I couldn't find your session information. Please refresh and try again."
            
            # Create ticket via API
            ticket_result = await self.api_client.create_ticket(
                tenant_id=str(chat_session.tenant_id),
                project_id=str(chat_session.project_id),
                title=title,
                description=description,
                customer_email=chat_session.email,
                priority="medium",
                category="general"
            )
            
            if ticket_result:
                ticket_id = ticket_result.get('id')
                
                # Update in-memory session context
                if session_id in self.sessions:
                    self.sessions[session_id].context['ticket_id'] = ticket_id
                
                return f"✅ I've successfully created support ticket #{ticket_id}. Our team will review it shortly and get back to you."
            else:
                return "❌ I encountered an issue creating your support ticket. Please try again or contact support directly."
            
        except Exception as e:
            logger.error(f"Ticket creation error: {e}")
            return "❌ I encountered an issue creating your support ticket. Please try again or contact support directly."
    
    async def _escalate_to_human(self, reason: str, session_id: str, db_session: AsyncSession) -> str:
        """Escalate session to human agents via TMS API."""
        try:
            # Get session info
            chat_session = await self.chat_service.get_session_info(
                db_session, session_id
            )
            
            if not chat_session:
                return "I couldn't find your session information. Please refresh and try again."
            
            # Try to escalate via API
            escalation_result = await self.api_client.escalate_session(
                tenant_id=str(chat_session.tenant_id),
                project_id=str(chat_session.project_id),
                session_id=session_id,
                reason=reason,
                priority="high"
            )
            
            if escalation_result:
                # Update in-memory session
                if session_id in self.sessions:
                    self.sessions[session_id].context['escalated'] = True
                    self.sessions[session_id].context['escalation_reason'] = reason
                    if 'id' in escalation_result:  # If it created a ticket as fallback
                        self.sessions[session_id].context['escalation_ticket_id'] = escalation_result['id']
                
                if 'id' in escalation_result:  # Ticket was created as fallback
                    return f"🚨 I've escalated your case and created priority ticket #{escalation_result['id']}. A human agent will assist you shortly."
                else:
                    return f"🚨 I've escalated your case for immediate human attention. Reason: {reason}. A human agent will join this conversation shortly."
            else:
                return "❌ I encountered an issue escalating your case. Let me create a priority support ticket instead."
                
        except Exception as e:
            logger.error(f"Escalation error: {e}")
            return "❌ I encountered an issue escalating your case. Let me create a priority support ticket for immediate attention."
    
    async def _collect_contact_information(self, session_id: str, db_session: AsyncSession) -> str:
        """Guide contact information collection (READ-ONLY)."""
        session_data = await self.chat_service.get_session_info(
            db_session, session_id
        )
        
        if not session_data or not session_data.email:
            return "To better assist you, could you please provide your email address?"
        elif not session_data.phone:
            return "Thank you! Could you also provide a phone number where we can reach you if needed?"
        else:
            return "We have your contact information and will reach out soon."
    
    async def _save_contact_info(self, contact_info: dict, session_id: str, db_session: AsyncSession) -> str:
        """Save contact information via TMS API."""
        try:
            # Get session info
            chat_session = await self.chat_service.get_session_info(
                db_session, session_id
            )
            
            if not chat_session:
                return "I couldn't find your session information. Please refresh and try again."
            
            # Call update contact info API
            result = await self.api_client.update_contact_info(
                tenant_id=str(chat_session.tenant_id),
                project_id=str(chat_session.project_id),
                session_id=session_id,
                contact_info=contact_info
            )
            
            if result:
                # Update in-memory session
                if session_id in self.sessions:
                    self.sessions[session_id].context['contact_info'] = contact_info
                
                contact_type = "email" if "email" in contact_info else "phone" if "phone" in contact_info else "contact"
                return f"✅ Thank you! I've saved your {contact_type} information securely. This will help us provide better support."
            else:
                return "❌ I encountered an issue saving your contact information. Please try again or continue with your inquiry."
                
        except Exception as e:
            logger.error(f"Contact info save error: {e}")
            # Store in session as fallback
            if session_id in self.sessions:
                self.sessions[session_id].context['contact_info'] = contact_info
            return "✅ I've noted your contact information. There was a temporary issue with our system, but I've recorded it for this session."
    
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
            
            # Get recent messages for context (READ-ONLY)
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
            
            # Send final metadata with pending actions
            metadata = {
                "session_id": session_id,
                "pending_actions": {
                    "ticket": agent_session.context.get('pending_ticket'),
                    "escalation": agent_session.context.get('escalation_request'),
                    "contact_update": agent_session.context.get('contact_update_request')
                },
                "context": {
                    "escalated": agent_session.context.get('escalated', False),
                    "has_email": agent_session.context.get('has_email', False),
                    "has_phone": agent_session.context.get('has_phone', False)
                }
            }
            yield f"data: {json.dumps({'metadata': metadata, 'done': True})}\n\n"
            
        except Exception as e:
            logger.error(f"Stream processing error: {e}")
            yield f"data: {json.dumps({'error': str(e)})}\n\n"