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


class AgentService:
    """Service for managing OpenAI Agents SDK-based AI agents."""
    
    def __init__(self):
        # In the latest SDK, AsyncOpenAI IS the client - no wrapper needed!
        self.agents_client = AsyncOpenAI(api_key=config.AI_API_KEY)
        
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
                results = await self.knowledge_service.search(
                    db_session, params.query, params.max_results
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
        async def create_ticket_impl(params: TicketParams) -> str:
            """Create a support ticket in the system."""
            try:
                session_id = getattr(self, '_current_session_id', None)
                db_session = getattr(self, '_current_db_session', None)
                
                if not session_id or not db_session:
                    return "I couldn't access session information to create a ticket."
                
                # Get session auth context
                auth_context = await self.auth_service.get_session_auth_context(
                    db_session, session_id
                )
                
                if not auth_context:
                    return "I couldn't find your session information. Please refresh and try again."
                
                # Create ticket via API
                ticket_result = await self.api_client.create_ticket(
                    tenant_id=auth_context["tenant_id"],
                    project_id=auth_context["project_id"], 
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
        async def escalate_impl(params: EscalationParams) -> str:
            """Escalate session to human agents via TMS API."""
            try:
                session_id = getattr(self, '_current_session_id', None)
                db_session = getattr(self, '_current_db_session', None)
                
                if not session_id or not db_session:
                    return "I couldn't access session information for escalation."
                
                # Get session auth context
                auth_context = await self.auth_service.get_session_auth_context(
                    db_session, session_id
                )
                
                if not auth_context:
                    return "I couldn't find your session information. Please refresh and try again."
                
                # Escalate via API
                escalation_result = await self.api_client.escalate_session(
                    tenant_id=auth_context["tenant_id"],
                    project_id=auth_context["project_id"],
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
        async def save_contact_impl(params: ContactInfoParams) -> str:
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
                
                # Get session auth context
                auth_context = await self.auth_service.get_session_auth_context(
                    db_session, session_id
                )
                
                if not auth_context:
                    return "I couldn't find your session information. Please refresh and try again."
                
                # Update contact info via API
                result = await self.api_client.update_contact_info(
                    tenant_id=auth_context["tenant_id"],
                    project_id=auth_context["project_id"],
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
        db_session: AsyncSession,
        session_id: str,
        message: str
    ) -> AsyncGenerator[str, None]:
        """
        Process message and stream response using OpenAI Agents SDK.
        
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
                # Get auth context to get tenant and project IDs
                auth_context = await self.auth_service.get_session_auth_context(
                    db_session, session_id
                )
                if auth_context:
                    self.sessions[session_id] = AgentSession(
                        session_id=session_id,
                        tenant_id=auth_context.get("tenant_id", ""),
                        project_id=auth_context.get("project_id", "")
                    )
                else:
                    # Fallback if no auth context
                    self.sessions[session_id] = AgentSession(
                        session_id=session_id,
                        tenant_id="",
                        project_id=""
                    )
            
            agent_session = self.sessions[session_id]
            
            # Set current context for function tools
            self._current_session_id = session_id
            self._current_db_session = db_session
            
            # Get recent messages for context
            recent_messages = await self.chat_service.get_recent_messages(
                db_session, session_id, limit=10
            )
            
            # Convert messages to format expected by Agents SDK
            messages = []
            for msg in recent_messages:
                role = "user" if msg.get("role") == "user" else "assistant"
                messages.append({"role": role, "content": msg.get("content", "")})
            
            # Add current message
            messages.append({"role": "user", "content": message})
            
            try:
                # Use Runner with the AsyncOpenAI client directly
                runner = Runner(client=self.agents_client)
                
                # Run the agent and stream response
                # The latest SDK supports streaming natively
                result = await runner.run(
                    agent=self.support_agent,
                    messages=messages,
                    stream=True  # Enable streaming
                )
                
                response_content = ""
                
                # Stream the response
                if hasattr(result, '__aiter__'):  # Check if result is async iterable
                    async for chunk in result:
                        if hasattr(chunk, 'content') and chunk.content:
                            response_content += chunk.content
                            yield f"data: {json.dumps({'content': chunk.content})}\n\n"
                else:
                    # Non-streaming fallback
                    response_content = result.final_output if hasattr(result, 'final_output') else str(result)
                    # Stream in chunks
                    chunk_size = 20
                    for i in range(0, len(response_content), chunk_size):
                        chunk = response_content[i:i + chunk_size]
                        yield f"data: {json.dumps({'content': chunk})}\n\n"
                
                # Save the message and response
                await self.chat_service.save_message(
                    db_session, session_id, "user", message
                )
                await self.chat_service.save_message(
                    db_session, session_id, "assistant", response_content
                )
                
                # Send final metadata
                metadata = {
                    "session_id": session_id,
                    "context": {
                        "escalated": agent_session.context.get('escalated', False),
                        "has_ticket": 'ticket_id' in agent_session.context,
                        "has_contact_info": 'contact_info' in agent_session.context
                    }
                }
                yield f"data: {json.dumps({'metadata': metadata, 'done': True})}\n\n"
                
            except Exception as agent_error:
                logger.error(f"Agent execution error: {agent_error}")
                # Fallback to direct OpenAI if Agents SDK fails
                response_content = await self._generate_agent_response(
                    agent_session, message, messages
                )
                # Stream the fallback response
                chunk_size = 20
                for i in range(0, len(response_content), chunk_size):
                    chunk = response_content[i:i + chunk_size]
                    yield f"data: {json.dumps({'content': chunk})}\n\n"
                    
                await self.chat_service.save_message(
                    db_session, session_id, "user", message
                )
                await self.chat_service.save_message(
                    db_session, session_id, "assistant", response_content
                )
                
                metadata = {
                    "session_id": session_id,
                    "context": agent_session.context
                }
                yield f"data: {json.dumps({'metadata': metadata, 'done': True})}\n\n"
            
        except Exception as e:
            logger.error(f"Stream processing error: {e}")
            yield f"data: {json.dumps({'error': str(e)})}\n\n"
        finally:
            # Clean up context
            if hasattr(self, '_current_session_id'):
                delattr(self, '_current_session_id')
            if hasattr(self, '_current_db_session'):
                delattr(self, '_current_db_session')
    
    async def _generate_agent_response(
        self, 
        agent_session: AgentSession,
        message: str,
        messages: List[Dict]
    ) -> str:
        """
        Fallback: Generate agent response using direct OpenAI API.
        This is used when the Agents SDK has issues.
        """
        try:
            # Define available functions for OpenAI function calling
            functions = [
                {
                    "name": "search_knowledge",
                    "description": "Search the knowledge base for relevant information",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "query": {
                                "type": "string", 
                                "description": "Search query for knowledge base"
                            },
                            "max_results": {
                                "type": "integer",
                                "description": "Maximum number of results to return",
                                "default": 5
                            }
                        },
                        "required": ["query"],
                        "additionalProperties": False
                    }
                },
                {
                    "name": "create_ticket",
                    "description": "Create a support ticket in the system",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "title": {
                                "type": "string",
                                "description": "Ticket title"
                            },
                            "description": {
                                "type": "string",
                                "description": "Ticket description"
                            }
                        },
                        "required": ["title", "description"],
                        "additionalProperties": False
                    }
                },
                {
                    "name": "escalate_to_human",
                    "description": "Escalate the case to a human agent",
                    "parameters": {
                        "type": "object", 
                        "properties": {
                            "reason": {
                                "type": "string",
                                "description": "Reason for escalation"
                            }
                        },
                        "required": ["reason"],
                        "additionalProperties": False
                    }
                },
                {
                    "name": "save_contact_info",
                    "description": "Save customer contact information",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "email": {
                                "type": "string",
                                "description": "Email address",
                                "default": ""
                            },
                            "phone": {
                                "type": "string", 
                                "description": "Phone number",
                                "default": ""
                            }
                        },
                        "additionalProperties": False
                    }
                }
            ]
            
            # Get response from OpenAI with function calling
            response = await self.agents_client.chat.completions.create(
                model="gpt-4o",
                messages=[
                    {
                        "role": "system",
                        "content": self._get_support_instructions()
                    }
                ] + messages,
                functions=functions,
                function_call="auto",
                temperature=0.7,
                max_tokens=1000
            )
            
            message_response = response.choices[0].message
            
            # Handle function calls
            if message_response.function_call:
                function_name = message_response.function_call.name
                function_args = json.loads(message_response.function_call.arguments)
                
                # Execute the function
                if function_name == "search_knowledge":
                    params = KnowledgeSearchParams(**function_args)
                    function_response = await self.search_knowledge_tool.function(params)
                elif function_name == "create_ticket":
                    params = TicketParams(**function_args)
                    function_response = await self.create_ticket_tool.function(params)
                elif function_name == "escalate_to_human":
                    params = EscalationParams(**function_args)
                    function_response = await self.escalate_tool.function(params)
                elif function_name == "save_contact_info":
                    params = ContactInfoParams(**function_args)
                    function_response = await self.save_contact_tool.function(params)
                else:
                    function_response = f"Unknown function: {function_name}"
                
                # Get follow-up response from OpenAI
                follow_up_response = await self.agents_client.chat.completions.create(
                    model="gpt-4o",
                    messages=[
                        {
                            "role": "system", 
                            "content": self._get_support_instructions()
                        }
                    ] + messages + [
                        message_response.model_dump(),
                        {
                            "role": "function",
                            "name": function_name,
                            "content": function_response
                        }
                    ],
                    temperature=0.7,
                    max_tokens=1000
                )
                
                return follow_up_response.choices[0].message.content
            else:
                return message_response.content
                
        except Exception as e:
            logger.error(f"Response generation error: {e}")
            return "I apologize, but I encountered an issue processing your request. Please try again or contact support directly."