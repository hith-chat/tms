"""Chat-related Pydantic schemas for API requests and responses."""

from pydantic import BaseModel, Field
from typing import Optional, Dict, Any, List
from datetime import datetime


class ChatRequest(BaseModel):
    """Chat request schema."""
    
    session_id: str = Field(..., description="Unique session identifier")
    message: str = Field(..., min_length=1, description="User message")


class ChatResponse(BaseModel):
    """Chat response schema."""
    
    content: Optional[str] = Field(None, description="Response content")
    metadata: Optional[Dict[str, Any]] = Field(None, description="Response metadata")
    error: Optional[str] = Field(None, description="Error message if any")
    done: Optional[bool] = Field(None, description="Whether streaming is complete")


class ChatSessionCreate(BaseModel):
    """Schema for creating a chat session."""
    
    session_id: str
    user_id: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    name: Optional[str] = None


class ChatSessionResponse(BaseModel):
    """Chat session response schema."""
    
    session_id: str
    user_id: Optional[str]
    email: Optional[str] 
    phone: Optional[str]
    name: Optional[str]
    created_at: datetime
    updated_at: Optional[datetime]
    
    class Config:
        from_attributes = True


class ChatMessageResponse(BaseModel):
    """Chat message response schema."""
    
    id: int
    session_id: str
    role: str
    content: str
    created_at: datetime
    
    class Config:
        from_attributes = True


class ContactInfoUpdate(BaseModel):
    """Schema for updating contact information."""
    
    email: Optional[str] = Field(None, description="Email address")
    phone: Optional[str] = Field(None, description="Phone number")
    name: Optional[str] = Field(None, description="Full name")
