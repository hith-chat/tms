"""Ticket-related Pydantic schemas for API requests and responses."""

from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime


class TicketCreate(BaseModel):
    """Schema for creating a support ticket."""
    
    title: str = Field(..., min_length=1, max_length=500, description="Ticket title")
    description: str = Field(..., min_length=1, description="Ticket description")
    priority: str = Field("medium", description="Ticket priority (low, medium, high, urgent)")
    category: str = Field("general", description="Ticket category")
    session_id: str = Field(..., description="Associated chat session ID")
    user_email: Optional[str] = Field(None, description="User email address")


class TicketResponse(BaseModel):
    """Support ticket response schema."""
    
    ticket_id: int
    session_id: str
    title: str
    description: str
    priority: str
    category: str
    status: str
    created_at: datetime
    updated_at: Optional[datetime]
    
    class Config:
        from_attributes = True
