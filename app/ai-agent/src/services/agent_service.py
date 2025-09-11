"""Agent service using latest OpenAI Agents SDK (openai-agents package)."""

import json
import logging
from typing import Dict, List, Any, Optional, AsyncGenerator
from datetime import datetime
from dataclasses import dataclass, field
from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession
import traceback
from agents import Agent, Runner
from agents.tool import FunctionTool 


from .knowledge_service import KnowledgeService
from .agent_auth_service import AgentAuthService
from .tms_api_client import TMSApiClient, tms_api_client
from ..config import config
from ..schemas.knowledge import (
    KnowledgeSearchResponse,
)

logger = logging.getLogger(__name__)


# Pydantic models for function tool parameters - STRICT SCHEMA COMPLIANT
class ContactInfoParams(BaseModel):
    """Parameters for contact information function."""
    class Config:
        extra = "forbid"  # Strict mode - no additional properties
    
    name: Optional[str] = Field(default="", description="Full name")
    email: Optional[str] = Field(default="", description="Email address")
    phone: Optional[str] = Field(default="", description="Phone number")


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
    priority: Optional[str] = Field(default="normal", description="low normal high urgent - priority level based on urgency")
    category: Optional[str] = Field(default="problem", description="question incident problem task - Ticket category")
    name:  Optional[str] = Field(default="", description="Full name of the reporter")
    email: Optional[str] = Field(default="", description="Email address of the reporter") 


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
    
    def __init__(self):
        # Initialize OpenAI client with API key
        if not config.AI_API_KEY:
            raise ValueError("AI_API_KEY is required for agent service")
                    
        # Initialize services
        self.knowledge_service = KnowledgeService()
        self.api_client: TMSApiClient = tms_api_client
        self.sessions: Dict[str, AgentSession] = {}
        
        # Agents will be initialized later when we have tenant/project info
        self.support_agent = None
        self.contact_agent = None  
        self.ticket_agent = None
    
    async def _setup_agents_for_session(self, tenant_id: str, project_id: str):
        """Setup agents with personalized instructions based on project settings."""
        # Only setup if not already done for this project
        if self.support_agent is not None:
            return
            
        await self._setup_agents(tenant_id, project_id)
    
    async def _setup_agents(self, tenant_id: str, project_id: str):
        """Setup all agents and their tools with personalized instructions."""
        
        # Search knowledge tool
        async def search_knowledge_impl(tool_context, params: str) -> str:
            """Search the knowledge base for relevant information."""
            try:
                session_id = getattr(self, '_current_session_id', None)
                params: KnowledgeSearchParams = KnowledgeSearchParams.model_validate_json(params)
                
                # Search knowledge base
                session_id = getattr(self, '_current_session_id', None)
                agent_session = self.sessions.get(session_id) if session_id else None
                tenant_id = agent_session.tenant_id if agent_session else None
                project_id = agent_session.project_id if agent_session else None
                
                results: KnowledgeSearchResponse = await self.knowledge_service.search_knowledge_base(
                    params.query, tenant_id=tenant_id, 
                    project_id=project_id, limit=params.max_results
                )
                
                if results:
                    response = "I found the following relevant information:\n\n"
                    for idx, result in enumerate(results.results):
                        title = result.title or ""
                        content = result.content
                        source = result.source
                        score = result.score
                        response += f"**{idx}. Title{title}**\nContent:\n{content}\n\n Source: {source} (Score: {score:.2f})\n\n"
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
        async def create_ticket_impl(tool_context, params: str) -> str:
            """Create a support ticket in the system."""
            try:
                session_id = getattr(self, '_current_session_id', None)
                params: TicketParams = TicketParams.model_validate_json(params)
                
                # Get session context from agent session instead of auth service
                agent_session = self.sessions.get(session_id) if session_id else None
                
                if not agent_session:
                    return "I couldn't access session information to create a ticket."
                
                tenant_id = agent_session.tenant_id
                project_id = agent_session.project_id
                
                if not tenant_id or not project_id:
                    return "I couldn't find your tenant and project information. Please refresh and try again."
                # ---- HARD GATE: require name + email before actual ticket creation ----
                contact_info = params
                name = contact_info.name
                email = contact_info.email
                if not name or not email:
                    missing = []
                    if not name: missing.append("full name")
                    if not email: missing.append("email")
                    need = " and ".join(missing)
                    return (
                        "ℹ️ Before I create your ticket, I need your "
                        f"{need}. Please reply with:\n"
                        "- Full name:\n"
                        "- Email:\n\n"
                        "I’ll save these and proceed immediately."
                    )
                # Create ticket via API (include reporter details inside description to preserve external API shape)
                enriched_description = (
                    f"{params.description}\n\n"
                    f"— Reporter: {name} <{email}>"
                )
                # Create ticket via API
                
                ticket_result = await self.api_client.create_ticket(
                    tenant_id=tenant_id,
                    project_id=project_id,
                    title=params.title,
                    description=enriched_description,
                    customer_email=email,
                    customer_name=name,
                    source="chat",
                    priority=params.priority or "normal",
                    category=params.category or "problem"
                )
                
                if ticket_result and 'id' in ticket_result:
                    ticket_id = ticket_result['id']
                    
                    # Update session context
                    if session_id in self.sessions:
                        self.sessions[session_id].context['ticket_id'] = ticket_id
                    
                    return (
                        f"✅ I've created support ticket #{ticket_id}. "
                        f"We’ll reach out to {name} at {email} if we need anything else."
                    )
               
                else:
                    return "❌ I encountered an issue creating your support ticket. Please try again or contact support directly."
                    
            except Exception as e:
                traceback.print_exc()
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

                params = EscalationParams.model_validate_json(params)
                
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
                traceback.print_exc()
                logger.error(f"Escalation error: {e}")
                return "❌ I encountered an issue escalating your case. Let me create a priority support ticket for immediate attention."
        
        self.escalate_tool = FunctionTool(
            name="escalate_to_human",
            description="Escalate the case to a human agent",
            params_json_schema=EscalationParams.model_json_schema(),
            on_invoke_tool=escalate_impl
        )
        
        # Save contact info tool
        async def save_contact_impl(tool_context, params: str) -> str:
            """Save contact information via TMS API."""
            try:
                contact_info = ContactInfoParams.model_validate_json(params)
                logger.info(f"Saving contact info: {params}")
                if not contact_info:
                    return "Please provide either an email address or phone number."
                
                session_id = getattr(self, '_current_session_id', None)
                
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
                    email = contact_info.email,
                    phone = contact_info.phone,
                    name = contact_info.name
                )
                
                if result:
                    # Update session context - convert Pydantic model to dict to avoid JSON serialization issues
                    if session_id in self.sessions:
                        self.sessions[session_id].context['contact_info'] = contact_info.model_dump()
                    
                    contact_type = "email" if contact_info.email else "phone" if contact_info.phone else "contact"
                    # Prefer granular acknowledgment when both present
                    if contact_info.email and contact_info.name:
                        return "✅ Thanks! I’ve saved your full name and email for follow-up."
                    contact_type = "email" if contact_info.email else "phone" if contact_info.phone else "contact"
                    return f"✅ Thank you! I've saved your {contact_type} information securely."
                else:
                    return "❌ I encountered an issue saving your contact information. Please try again or continue with your inquiry."
                    
            except Exception as e:
                traceback.print_exc()
                logger.error(f"Contact info save error: {e}")
                # Store in session as fallback
                return "✅ I've noted your contact information. There was a temporary issue with our system, but I've recorded it for this session."
        
        self.save_contact_tool = FunctionTool(
            name="save_contact_info",
            description="Save customer contact information",
            on_invoke_tool=save_contact_impl,
            params_json_schema=ContactInfoParams.model_json_schema()
        )
        
        # Create agents with tools
        support_instructions = await self._get_support_instructions(tenant_id, project_id)
        
        self.support_agent = Agent(
            name="Support Agent",
            instructions=support_instructions,
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
    
    async def _get_support_instructions(self, tenant_id: str, project_id: str) -> str:
        """Get instructions for the support agent with personalized about me information."""
        
        # Fetch about me settings from the API
        about_me_content = await self.api_client.get_about_me_settings(tenant_id, project_id)
        
        base_instructions = """You are a helpful customer support agent with the ability to perform real actions.

Your capabilities:
1. Search the knowledge base for answers using search_knowledge
2. Create actual support tickets via create_ticket (this will create a real ticket in the system)
3. Escalate cases to human agents via escalate_to_human (this will notify human agents)
4. Save contact information via save_contact_info

Guidelines:
- Be professional, empathetic, and concise
- Search the knowledge base first for common questions
- **Before creating any ticket, collect and save the user's full name and email using save_contact_info. Do not call create_ticket until both are present in context.**
- Create tickets for issues that need follow-up
- Escalate complex or urgent issues to human agents
- You can perform actual actions - when you create tickets or escalate, these actions happen immediately in the system"""

        # If we have about me content, add it to the instructions
        if about_me_content and about_me_content.strip():
            personalized_instructions = f"""{base_instructions}

IMPORTANT CONTEXT ABOUT THE SUPPORT ORGANIZATION:
{about_me_content}

Use this context to provide more personalized and relevant support. When greeting users or handling basic requests, reference this information appropriately to create a more tailored experience."""
            return personalized_instructions
        
        return base_instructions
    
    def _get_contact_instructions(self) -> str:
        """Get instructions for the contact agent."""
        return """You collect user contact information and can update it in the system.

Your approach:
1. Ask for full name and email (both required), then phone (optional)
2. Explain this helps with follow-up
3. Use save_contact_info to actually update the contact information in the system
4. You can perform real updates - the information will be saved immediately
5. Be polite and explain why contact info is helpful"""
    
    def _get_ticket_instructions(self) -> str:
        """Get instructions for the ticket agent."""
        return """You create actual support tickets in the system.

Your approach:
1. Gather clear issue description, error messages, what user tried
2. **Collect and save the user's full name and email first (use save_contact_info). Do not create a ticket before these are present.**
3. Use create_ticket to create a real ticket in the system
4. Confirm back the saved contact details in your acknowledgement
5. Be thorough in gathering information before creating tickets"""
    
    async def process_message_stream(
        self,
        message: str,
        session_id: str,
        auth_token: str = None,
        tenant_id: str = None,
        project_id: str = None,
        metadata: dict = None,
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
            
            # Setup agents with personalized instructions if not already done
            await self._setup_agents_for_session(tenant_id, project_id)
            
            # Set current context for function tools
            self._current_session_id = session_id
            self._current_auth_token = auth_token
            
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