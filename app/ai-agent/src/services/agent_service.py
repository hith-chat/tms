"""Agent service using latest OpenAI Agents SDK (openai-agents package)."""

import json
import logging
from typing import Dict, List, Any, Optional, AsyncGenerator
from datetime import datetime
from dataclasses import dataclass, field
from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession

from openai import AsyncOpenAI
from agents import Agent, Runner
from agents.tool import FunctionTool 


from ..models.chat import ChatSession, ChatMessage
from .knowledge_service import KnowledgeService
from .chat_service import ChatService
from .agent_auth_service import AgentAuthService
from .tms_api_client import TMSApiClient
from ..config import config

logger = logging.getLogger(__name__)


# Pydantic models for function tool parameters - STRICT SCHEMA COMPLIANT
class ContactInfoParams(BaseModel):
    """Parameters for contact information function."""
    class Config:
        extra = "forbid"  # Strict mode - no additional properties
    
    email: str = Field(default="", description="Email address")
    phone: str = Field(default="", description="Phone number")


class KnowledgeSearchParams(BaseModel):
    """Parameters for knowledge search function."""
    class Config:
        extra = "forbid"  # Strict mode
    
    query: str = Field(..., description="Search query for knowledge base")
    max_results: int = Field(default=5, description="Maximum number of results to return")


class TicketParams(BaseModel):
    """Parameters for ticket creation function."""
    class Config:
        extra = "forbid"  # Strict mode
    
    title: str = Field(..., description="Ticket title")
    description: str = Field(..., description="Ticket description")


class EscalationParams(BaseModel):
    """Parameters for escalation function."""
    class Config:
        extra = "forbid"  # Strict mode
    
    reason: str = Field(..., description="Reason for escalation")


@dataclass
class AgentSession:
    """In-memory session state for agent interactions."""
    session_id: str
    tenant_id: str
    project_id: str
    current_agent: str = "support"
    context: Dict = field(default_factory=dict)
    temp_data: Dict = field(default_factory=dict)
    history: List[Dict[str, str]] = field(default_factory=list)  # <---



class AgentService:
    """Service for managing OpenAI Agents SDK-based AI agents."""

    _current_session_id = None
    _current_auth_token = None
    _current_db_session = None
    
    def __init__(self):
        # Initialize OpenAI client with API key
        if not config.AI_API_KEY:
            raise ValueError("AI_API_KEY is required for agent service")
                    
        # Initialize services
        self.knowledge_service = KnowledgeService()
        self.chat_service = ChatService()
        self.auth_service = AgentAuthService()
        self.api_client = TMSApiClient(self.auth_service)
        self.sessions: Dict[str, AgentSession] = {}
        
        # Initialize agents
        self._setup_agents()
    
    def _setup_agents(self):
        """Setup all agents and their tools."""
        
        # Search knowledge tool
        async def search_knowledge_impl(params: KnowledgeSearchParams) -> str:
            """Search the knowledge base for relevant information."""
            try:
                session_id = getattr(self, '_current_session_id', None)
                db_session = getattr(self, '_current_db_session', None)
                
                if not db_session:
                    return "I couldn't access the knowledge base at this moment."
                
                # Search knowledge base
                session_id = getattr(self, '_current_session_id', None)
                agent_session = self.sessions.get(session_id) if session_id else None
                tenant_id = agent_session.tenant_id if agent_session else None
                project_id = agent_session.project_id if agent_session else None
                
                results = await self.knowledge_service.search_knowledge_base(
                    db_session, params.query, tenant_id=tenant_id, 
                    project_id=project_id, limit=params.max_results
                )
                
                if results:
                    response = "I found the following relevant information:\n\n"
                    for idx, result in enumerate(results[:3], 1):
                        title = result.get('title', 'Information')
                        content = result.get('content', '')
                        response += f"**{idx}. {title}**\n{content}\n\n"
                    return response
                else:
                    return "I couldn't find specific information about that in our knowledge base."
                    
            except Exception as e:
                logger.error(f"Knowledge search error: {e}")
                return "I encountered an issue searching the knowledge base."
        
        self.search_knowledge_tool = FunctionTool(
            name="search_knowledge",
            description="Search the knowledge base for relevant information",
            params_json_schema=KnowledgeSearchParams.model_json_schema(),
            on_invoke_tool=search_knowledge_impl
        )
        
        # Create ticket tool
        async def create_ticket_impl(tool_context, params: TicketParams) -> str:
            """Create a support ticket in the system."""
            try:
                session_id = getattr(self, '_current_session_id', None)
                db_session = getattr(self, '_current_db_session', None)
                
                if not session_id or not db_session:
                    return "I couldn't access session information to create a ticket."
                
                # Get session context from agent session instead of auth service
                agent_session = self.sessions.get(session_id) if session_id else None
                
                if not agent_session:
                    return "I couldn't access session information to create a ticket."
                
                tenant_id = agent_session.tenant_id
                project_id = agent_session.project_id
                
                if not tenant_id or not project_id:
                    return "I couldn't find your tenant and project information. Please refresh and try again."
                
                # Create ticket via API
                ticket_result = await self.api_client.create_ticket(
                    tenant_id=tenant_id,
                    project_id=project_id,
                    title=params.title,
                    description=params.description,
                    priority="medium",
                    category="general"
                )
                
                if ticket_result and 'id' in ticket_result:
                    ticket_id = ticket_result['id']
                    
                    # Update session context
                    if session_id in self.sessions:
                        self.sessions[session_id].context['ticket_id'] = ticket_id
                    
                    return f"✅ I've successfully created support ticket #{ticket_id}. Our team will review it and get back to you shortly."
                else:
                    return "❌ I encountered an issue creating your support ticket. Please try again or contact support directly."
                    
            except Exception as e:
                logger.error(f"Ticket creation error: {e}")
                return "❌ I encountered an issue creating your support ticket. Please try again or contact support directly."
        
        self.create_ticket_tool = FunctionTool(
            name="create_ticket",
            description="Create a support ticket in the system",
            params_json_schema=TicketParams.model_json_schema(),
            on_invoke_tool =create_ticket_impl
        )
        
        # Escalate to human tool
        async def escalate_impl(tool_context, params: EscalationParams) -> str:
            """Escalate session to human agents via TMS API."""
            try:
                session_id = getattr(self, '_current_session_id', None)
                db_session = getattr(self, '_current_db_session', None)
                
                if not session_id or not db_session:
                    return "I couldn't access session information for escalation."
                
                # Get session context from agent session instead of auth service
                agent_session = self.sessions.get(session_id) if session_id else None
                
                if not agent_session:
                    return "I couldn't access session information for escalation."
                
                tenant_id = agent_session.tenant_id
                project_id = agent_session.project_id
                
                if not tenant_id or not project_id:
                    return "I couldn't find your tenant and project information. Please refresh and try again."
                
                # Escalate via API
                escalation_result = await self.api_client.escalate_session(
                    tenant_id=tenant_id,
                    project_id=project_id,
                    session_id=session_id,
                    reason=params.reason,
                    priority="high"
                )
                
                if escalation_result:
                    # Update session context
                    if session_id in self.sessions:
                        self.sessions[session_id].context['escalated'] = True
                        self.sessions[session_id].context['escalation_reason'] = params.reason
                    
                    if 'id' in escalation_result:
                        return f"🚨 I've escalated your case and created priority ticket #{escalation_result['id']}. A human agent will assist you shortly."
                    else:
                        return f"🚨 I've escalated your case for immediate human attention. A human agent will join this conversation shortly."
                else:
                    return "❌ I encountered an issue escalating your case. Let me create a priority support ticket instead."
                    
            except Exception as e:
                logger.error(f"Escalation error: {e}")
                return "❌ I encountered an issue escalating your case. Let me create a priority support ticket for immediate attention."
        
        self.escalate_tool = FunctionTool(
            name="escalate_to_human",
            description="Escalate the case to a human agent",
            params_json_schema=EscalationParams.model_json_schema(),
            on_invoke_tool=escalate_impl
        )
        
        # Save contact info tool
        async def save_contact_impl(tool_context, params: ContactInfoParams) -> str:
            """Save contact information via TMS API."""
            try:
                contact_info = {}
                if params.email:
                    contact_info["email"] = params.email
                if params.phone:
                    contact_info["phone"] = params.phone
                
                if not contact_info:
                    return "Please provide either an email address or phone number."
                
                session_id = getattr(self, '_current_session_id', None)
                db_session = getattr(self, '_current_db_session', None)
                
                if not session_id or not db_session:
                    return "I couldn't access session information to save contact info."
                
                # Get session context from agent session instead of auth service
                agent_session = self.sessions.get(session_id) if session_id else None
                
                if not agent_session:
                    return "I couldn't access session information to save contact info."
                
                tenant_id = agent_session.tenant_id
                project_id = agent_session.project_id
                
                if not tenant_id or not project_id:
                    return "I couldn't find your tenant and project information. Please refresh and try again."
                
                # Update contact info via API
                result = await self.api_client.update_contact_info(
                    tenant_id=tenant_id,
                    project_id=project_id,
                    session_id=session_id,
                    contact_info=contact_info
                )
                
                if result:
                    # Update session context
                    if session_id in self.sessions:
                        self.sessions[session_id].context['contact_info'] = contact_info
                    
                    contact_type = "email" if "email" in contact_info else "phone" if "phone" in contact_info else "contact"
                    return f"✅ Thank you! I've saved your {contact_type} information securely. This will help us provide better support."
                else:
                    return "❌ I encountered an issue saving your contact information. Please try again or continue with your inquiry."
                    
            except Exception as e:
                logger.error(f"Contact info save error: {e}")
                # Store in session as fallback
                session_id = getattr(self, '_current_session_id', None)
                if session_id and session_id in self.sessions:
                    contact_info = {}
                    if params.email:
                        contact_info["email"] = params.email
                    if params.phone:
                        contact_info["phone"] = params.phone
                    self.sessions[session_id].context['contact_info'] = contact_info
                return "✅ I've noted your contact information. There was a temporary issue with our system, but I've recorded it for this session."
        
        self.save_contact_tool = FunctionTool(
            name="save_contact_info",
            description="Save customer contact information",
            on_invoke_tool=save_contact_impl,
            params_json_schema=ContactInfoParams.model_json_schema()
        )
        
        # Create agents with tools
        self.support_agent = Agent(
            name="Support Agent",
            instructions=self._get_support_instructions(),
            tools=[
                self.search_knowledge_tool,
                self.create_ticket_tool,
                self.escalate_tool,
                self.save_contact_tool
            ],
            model="gpt-4o"
        )
        
        self.contact_agent = Agent(
            name="Contact Agent",
            instructions=self._get_contact_instructions(),
            tools=[self.save_contact_tool],
            model="gpt-4o"
        )
        
        self.ticket_agent = Agent(
            name="Ticket Agent",
            instructions=self._get_ticket_instructions(),
            tools=[self.create_ticket_tool],
            handoffs=[self.contact_agent]  # Can handoff to contact agent
        )
    
    def _get_support_instructions(self) -> str:
        """Get instructions for the support agent."""
        return """You are a helpful customer support agent with the ability to perform real actions.

Your capabilities:
1. Search the knowledge base for answers using search_knowledge
2. Create actual support tickets via create_ticket (this will create a real ticket in the system)
3. Escalate cases to human agents via escalate_to_human (this will notify human agents)
4. Save contact information via save_contact_info

Guidelines:
- Be professional, empathetic, and concise
- Search the knowledge base first for common questions
- Create tickets for issues that need follow-up
- Escalate complex or urgent issues to human agents
- You can perform actual actions - when you create tickets or escalate, these actions happen immediately in the system"""
    
    def _get_contact_instructions(self) -> str:
        """Get instructions for the contact agent."""
        return """You collect user contact information and can update it in the system.

Your approach:
1. Ask for email first (required), then phone (optional) 
2. Explain this helps with follow-up
3. Use save_contact_info to actually update the contact information in the system
4. You can perform real updates - the information will be saved immediately
5. Be polite and explain why contact info is helpful"""
    
    def _get_ticket_instructions(self) -> str:
        """Get instructions for the ticket agent."""
        return """You create actual support tickets in the system.

Your approach:
1. Gather clear issue description, error messages, what user tried
2. Use create_ticket to create a real ticket in the system
3. After ticket creation, suggest collecting contact details for follow-up
4. You create real tickets - they will appear in the system immediately
5. Be thorough in gathering information before creating tickets"""
    
    async def process_message_stream(
        self,
        message: str,
        session_id: str,
        auth_token: str = None,
        tenant_id: str = None,
        project_id: str = None,
        metadata: dict = None,
        db_session: AsyncSession = None
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """
        Process message and stream response using OpenAI Agents SDK.
        
        Args:
            message: User message
            session_id: Chat session ID  
            auth_token: Authentication token from Go service
            tenant_id: Tenant ID
            project_id: Project ID
            metadata: Additional metadata
            
        Yields:
            Response dictionaries for SSE formatting
        """
        try:
            # Get or create agent session
            if session_id not in self.sessions:
                self.sessions[session_id] = AgentSession(
                    session_id=session_id,
                    tenant_id=tenant_id,
                    project_id=project_id
                )
            
            agent_session = self.sessions[session_id]
            
            # Set current context for function tools
            self._current_session_id = session_id
            self._current_auth_token = auth_token
            self._current_db_session = db_session
            
            # Prepare messages for agent (simplified for SSE approach)
            agent_session.history.append({"role": "user", "content": message})

            
            try:
                # Use Runner with the AsyncOpenAI client configured with API key
                result = await Runner.run(
                    starting_agent=self.support_agent,
                    input=agent_session.history,
                    context=agent_session
                )
                
                # response_content = ""
                
                # Get response text
                response_content = result.final_output or ""

                print(f"Agent response: {result.final_output}")
                agent_session.history.append({"role": "assistant", "content": str(result.final_output or "")})


                # Stream out in chunks (for SSE)
                yield {
                    "type": "message",
                    "content": response_content,
                    "metadata": {"session_id": session_id}
                }

                # Send final metadata
                yield {
                    "type": "metadata",
                    "content": "Processing complete",
                    "metadata": {
                        "session_id": session_id,
                        "context": {
                            "escalated": agent_session.context.get('escalated', False),
                            "has_ticket": 'ticket_id' in agent_session.context,
                            "has_contact_info": 'contact_info' in agent_session.context
                        }
                    }
                }
                
            except Exception as agent_error:
                logger.error(f"Agent execution error: {agent_error}")
                # Fallback to direct OpenAI if Agents SDK fails
                
                yield {
                    "type": "metadata",
                    "content": "Processing complete (fallback)",
                    "metadata": {
                        "session_id": session_id,
                        "context": agent_session.context
                    }
                }
            
        except Exception as e:
            logger.error(f"Stream processing error: {e}")
            yield {
                "type": "error",
                "content": f"Error processing message: {str(e)}",
                "metadata": {"session_id": session_id}
            }
            yield f"data: {json.dumps({'error': str(e)})}\n\n"
        finally:
            # Clean up context
            if hasattr(self, '_current_session_id'):
                delattr(self, '_current_session_id')
            if hasattr(self, '_current_db_session'):
                delattr(self, '_current_db_session')
