"""Chat-related SQLAlchemy models."""

from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy import String, Text, DateTime, UUID, Integer, ForeignKey
from sqlalchemy.dialects.postgresql import UUID as pg_UUID
from uuid import uuid4, UUID as PyUUID
from datetime import datetime
from typing import Optional, List

from .base import BaseModel


class ChatSession(BaseModel):
    """Chat session model."""
    
    __tablename__ = "chat_sessions"
    
    id: Mapped[PyUUID] = mapped_column(pg_UUID(as_uuid=True), primary_key=True, default=uuid4)
    session_id: Mapped[str] = mapped_column(String(255), unique=True, nullable=False)  # client_session_id
    tenant_id: Mapped[PyUUID] = mapped_column(pg_UUID(as_uuid=True), nullable=False)
    project_id: Mapped[PyUUID] = mapped_column(pg_UUID(as_uuid=True), nullable=False)
    widget_id: Mapped[PyUUID] = mapped_column(pg_UUID(as_uuid=True), nullable=False)
    
    # Customer information
    customer_id: Mapped[Optional[PyUUID]] = mapped_column(pg_UUID(as_uuid=True), nullable=True)
    user_id: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    email: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    phone: Mapped[Optional[str]] = mapped_column(String(50), nullable=True)
    name: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    
    # Session status
    status: Mapped[str] = mapped_column(String(50), default="active")
    
    # Agent assignment
    assigned_agent_id: Mapped[Optional[PyUUID]] = mapped_column(pg_UUID(as_uuid=True), nullable=True)
    
    # Relationships
    messages: Mapped[List["ChatMessage"]] = relationship(
        "ChatMessage", 
        back_populates="session",
        cascade="all, delete-orphan"
    )
    tickets: Mapped[List["SupportTicket"]] = relationship(
        "SupportTicket", 
        back_populates="session",
        cascade="all, delete-orphan"
    )


class ChatMessage(BaseModel):
    """Chat message model."""
    
    __tablename__ = "messages"
    
    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    session_id: Mapped[str] = mapped_column(
        String(255), 
        ForeignKey("chat_sessions.session_id", ondelete="CASCADE"),
        nullable=False
    )
    role: Mapped[str] = mapped_column(String(50), nullable=False)  # user, assistant, system
    content: Mapped[str] = mapped_column(Text, nullable=False)
    
    # Relationships
    session: Mapped["ChatSession"] = relationship("ChatSession", back_populates="messages")


class SupportTicket(BaseModel):
    """Support ticket model."""
    
    __tablename__ = "tickets"
    
    ticket_id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    session_id: Mapped[str] = mapped_column(
        String(255),
        ForeignKey("chat_sessions.session_id", ondelete="CASCADE"),
        nullable=False
    )
    title: Mapped[str] = mapped_column(String(500), nullable=False)
    description: Mapped[str] = mapped_column(Text, nullable=False)
    priority: Mapped[str] = mapped_column(String(50), nullable=False, default="medium")
    category: Mapped[str] = mapped_column(String(100), nullable=False, default="general")
    status: Mapped[str] = mapped_column(String(50), nullable=False, default="open")
    
    # Relationships
    session: Mapped["ChatSession"] = relationship("ChatSession", back_populates="tickets")
