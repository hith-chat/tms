# swarm_agent.py
# OpenAI Swarm Agent with PostgreSQL Vector Store and Streaming API

import os
import json
import asyncio
from typing import Dict, List, Optional, AsyncGenerator
from dataclasses import dataclass, field
from datetime import datetime
import asyncpg
import numpy as np
from swarm import Swarm, Agent
from pydantic import BaseModel, Field
from openai import AsyncOpenAI
import logging
from fastapi import FastAPI, HTTPException
from fastapi.responses import StreamingResponse
from contextlib import asynccontextmanager

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# ============== Configuration ==============
POSTGRES_HOST = os.getenv("POSTGRES_HOST", "postgres")
POSTGRES_PORT = os.getenv("POSTGRES_PORT", "5432")
POSTGRES_DB = os.getenv("POSTGRES_DB", "chatdb")
POSTGRES_USER = os.getenv("POSTGRES_USER", "postgres")
POSTGRES_PASSWORD = os.getenv("POSTGRES_PASSWORD", "postgres")
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")

# ============== Data Models ==============
@dataclass
class ChatSession:
    """In-memory session state"""
    session_id: str
    current_agent: str = "support"
    context: Dict = field(default_factory=dict)
    temp_data: Dict = field(default_factory=dict)  # Temporary data, not persisted

class TicketData(BaseModel):
    """Ticket structure for database"""
    title: str
    description: str
    priority: str = "medium"
    category: str = "general"
    session_id: str
    user_email: Optional[str] = None

class ContactInfo(BaseModel):
    """Contact information structure"""
    email: Optional[str] = None
    phone: Optional[str] = None
    name: Optional[str] = None

# ============== Database Manager ==============
class DatabaseManager:
    def __init__(self):
        self.pool = None
        self.openai_client = AsyncOpenAI(api_key=OPENAI_API_KEY)
    
    async def connect(self):
        """Create connection pool"""
        self.pool = await asyncpg.create_pool(
            host=POSTGRES_HOST,
            port=POSTGRES_PORT,
            database=POSTGRES_DB,
            user=POSTGRES_USER,
            password=POSTGRES_PASSWORD,
            min_size=10,
            max_size=20
        )
        await self.initialize_tables()
    
    async def initialize_tables(self):
        """Create necessary tables if they don't exist"""
        async with self.pool.acquire() as conn:
            # Sessions table
            await conn.execute('''
                CREATE TABLE IF NOT EXISTS chat_sessions (
                    session_id VARCHAR(255) PRIMARY KEY,
                    user_id VARCHAR(255),
                    email VARCHAR(255),
                    phone VARCHAR(50),
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            ''')
            
            # Messages table
            await conn.execute('''
                CREATE TABLE IF NOT EXISTS messages (
                    id SERIAL PRIMARY KEY,
                    session_id VARCHAR(255) REFERENCES chat_sessions(session_id),
                    role VARCHAR(50),
                    content TEXT,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            ''')
            
            # Tickets table
            await conn.execute('''
                CREATE TABLE IF NOT EXISTS tickets (
                    ticket_id SERIAL PRIMARY KEY,
                    session_id VARCHAR(255) REFERENCES chat_sessions(session_id),
                    title VARCHAR(500),
                    description TEXT,
                    priority VARCHAR(50),
                    category VARCHAR(100),
                    status VARCHAR(50) DEFAULT 'open',
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            ''')
            
            # Ensure pgvector extension
            await conn.execute('CREATE EXTENSION IF NOT EXISTS vector')
    
    async def get_or_create_session(self, session_id: str) -> Dict:
        """Get or create session in database"""
        async with self.pool.acquire() as conn:
            result = await conn.fetchrow(
                'SELECT * FROM chat_sessions WHERE session_id = $1',
                session_id
            )
            
            if not result:
                await conn.execute(
                    'INSERT INTO chat_sessions (session_id) VALUES ($1)',
                    session_id
                )
                result = await conn.fetchrow(
                    'SELECT * FROM chat_sessions WHERE session_id = $1',
                    session_id
                )
            
            return dict(result)
    
    async def save_message(self, session_id: str, role: str, content: str):
        """Save message to database"""
        async with self.pool.acquire() as conn:
            await conn.execute(
                '''INSERT INTO messages (session_id, role, content) 
                   VALUES ($1, $2, $3)''',
                session_id, role, content
            )
    
    async def get_recent_messages(self, session_id: str, limit: int = 10) -> List[Dict]:
        """Get recent messages for context"""
        async with self.pool.acquire() as conn:
            rows = await conn.fetch(
                '''SELECT role, content FROM messages 
                   WHERE session_id = $1 
                   ORDER BY created_at DESC 
                   LIMIT $2''',
                session_id, limit
            )
            return [dict(row) for row in reversed(rows)]
    
    async def search_kb(self, query: str, limit: int = 3) -> List[Dict]:
        """Search knowledge base using pgvector"""
        # Generate embedding for query
        response = await self.openai_client.embeddings.create(
            model="text-embedding-3-small",
            input=query
        )
        query_embedding = response.data[0].embedding
        
        async with self.pool.acquire() as conn:
            # Assuming kb_articles table with embedding column exists
            results = await conn.fetch(
                '''SELECT content, metadata, 
                          1 - (embedding <=> $1::vector) as similarity
                   FROM kb_articles 
                   ORDER BY embedding <=> $1::vector
                   LIMIT $2''',
                query_embedding, limit
            )
            
            return [dict(row) for row in results]
    
    async def create_ticket(self, ticket_data: TicketData) -> int:
        """Create ticket in database"""
        async with self.pool.acquire() as conn:
            ticket_id = await conn.fetchval(
                '''INSERT INTO tickets (session_id, title, description, priority, category)
                   VALUES ($1, $2, $3, $4, $5)
                   RETURNING ticket_id''',
                ticket_data.session_id, ticket_data.title, 
                ticket_data.description, ticket_data.priority, 
                ticket_data.category
            )
            return ticket_id
    
    async def update_contact_info(self, session_id: str, contact: ContactInfo):
        """Update contact information in database"""
        async with self.pool.acquire() as conn:
            updates = []
            values = [session_id]
            
            if contact.email:
                updates.append(f"email = ${len(values) + 1}")
                values.append(contact.email)
            
            if contact.phone:
                updates.append(f"phone = ${len(values) + 1}")
                values.append(contact.phone)
            
            if updates:
                query = f'''UPDATE chat_sessions 
                           SET {", ".join(updates)}, updated_at = CURRENT_TIMESTAMP
                           WHERE session_id = $1'''
                await conn.execute(query, *values)
    
    async def close(self):
        """Close connection pool"""
        if self.pool:
            await self.pool.close()

# ============== Global instances ==============
db = DatabaseManager()
sessions: Dict[str, ChatSession] = {}  # In-memory session state
client = Swarm()

# ============== Agent Functions ==============
async def search_knowledge_base(query: str, session_id: str) -> str:
    """Search KB for relevant information"""
    try:
        articles = await db.search_kb(query)
        
        if not articles:
            return None
        
        # Check relevance threshold
        if articles[0].get('similarity', 0) < 0.7:
            return None
        
        # Format response from KB articles
        response = "Based on our knowledge base:\n\n"
        for article in articles[:2]:
            metadata = article.get('metadata', {})
            response += f"**{metadata.get('title', 'Information')}**\n"
            response += f"{article['content']}\n\n"
        
        return response
    except Exception as e:
        logger.error(f"KB search error: {e}")
        return None

async def create_support_ticket(title: str, description: str, session_id: str) -> str:
    """Create a support ticket"""
    try:
        session_data = await db.get_or_create_session(session_id)
        
        ticket_data = TicketData(
            title=title,
            description=description,
            session_id=session_id,
            user_email=session_data.get('email')
        )
        
        ticket_id = await db.create_ticket(ticket_data)
        
        # Update in-memory session
        if session_id in sessions:
            sessions[session_id].context['ticket_id'] = ticket_id
        
        return f"I've created support ticket #{ticket_id}. Our team will review it shortly."
    except Exception as e:
        logger.error(f"Ticket creation error: {e}")
        return "I encountered an issue creating the ticket. Please try again."

async def collect_contact_information(session_id: str) -> str:
    """Guide contact information collection"""
    session_data = await db.get_or_create_session(session_id)
    
    if not session_data.get('email'):
        return "To better assist you, could you please provide your email address?"
    elif not session_data.get('phone'):
        return "Thank you! Could you also provide a phone number where we can reach you if needed?"
    else:
        return "We have your contact information and will reach out soon."

async def save_contact_info(email: str = None, phone: str = None, name: str = None, session_id: str = None) -> str:
    """Save user's contact information"""
    try:
        contact = ContactInfo(email=email, phone=phone, name=name)
        await db.update_contact_info(session_id, contact)
        
        # Update in-memory session
        if session_id in sessions:
            if email:
                sessions[session_id].context['has_email'] = True
            if phone:
                sessions[session_id].context['has_phone'] = True
        
        return "Thank you! I've saved your contact information."
    except Exception as e:
        logger.error(f"Contact save error: {e}")
        return "I had trouble saving your information. Please try again."

def escalate_to_human(reason: str, session_id: str) -> str:
    """Mark session for human escalation"""
    if session_id in sessions:
        sessions[session_id].context['escalated'] = True
        sessions[session_id].context['escalation_reason'] = reason
    
    return "I'll connect you with a human agent who can better assist you with this issue."

# ============== Swarm Agents ==============
support_agent = Agent(
    name="Support Agent",
    instructions="""You are a helpful customer support agent. Your approach:
    1. First search the knowledge base for answers
    2. If no relevant KB article, offer to create a support ticket
    3. Be professional, empathetic, and concise
    4. Collect contact information when creating tickets
    
    Always use search_knowledge_base first for support questions.
    If the KB doesn't have the answer, use create_support_ticket.
    For complex issues, use escalate_to_human.""",
    functions=[search_knowledge_base, create_support_ticket, escalate_to_human]
)

contact_agent = Agent(
    name="Contact Agent",
    instructions="""Collect user contact information politely.
    Ask for email first (required), then phone (optional).
    Explain this helps us follow up on their issue.
    Use save_contact_info to store the information.""",
    functions=[collect_contact_information, save_contact_info]
)

ticket_agent = Agent(
    name="Ticket Agent",
    instructions="""Create support tickets for unresolved issues.
    Gather: clear issue description, any error messages, what user tried.
    Use create_support_ticket to create the ticket.
    After ticket creation, switch to contact_agent for contact details.""",
    functions=[create_support_ticket],
    handoff_to=[contact_agent]
)

# ============== Streaming API ==============
@asynccontextmanager
async def lifespan(app: FastAPI):
    """Startup and shutdown events"""
    await db.connect()
    yield
    await db.close()

app = FastAPI(lifespan=lifespan)

async def process_message_stream(session_id: str, message: str) -> AsyncGenerator[str, None]:
    """Process message and stream response"""
    try:
        # Get or create session
        if session_id not in sessions:
            sessions[session_id] = ChatSession(session_id=session_id)
        
        session = sessions[session_id]
        
        # Save user message
        await db.save_message(session_id, "user", message)
        
        # Get recent messages for context
        recent_messages = await db.get_recent_messages(session_id)
        
        # Prepare context
        context_variables = {
            "session_id": session_id,
            "session": session
        }
        
        # Run Swarm agent
        response = client.run(
            agent=support_agent,
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
        await db.save_message(session_id, "assistant", full_response)
        
        # Send final metadata
        metadata = {
            "ticket_id": session.context.get('ticket_id'),
            "escalated": session.context.get('escalated', False),
            "has_email": session.context.get('has_email', False)
        }
        yield f"data: {json.dumps({'metadata': metadata, 'done': True})}\n\n"
        
    except Exception as e:
        logger.error(f"Stream processing error: {e}")
        yield f"data: {json.dumps({'error': str(e)})}\n\n"

@app.post("/chat/stream")
async def chat_stream(request: dict):
    """Streaming chat endpoint"""
    session_id = request.get("session_id")
    message = request.get("message")
    
    if not session_id or not message:
        raise HTTPException(status_code=400, detail="Missing session_id or message")
    
    return StreamingResponse(
        process_message_stream(session_id, message),
        media_type="text/event-stream"
    )

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy", "service": "swarm-agent"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=5000)