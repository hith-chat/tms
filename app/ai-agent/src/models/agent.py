"""Agent-related SQLAlchemy models based on existing migrations."""

from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy import String, Integer, DateTime, Boolean, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from uuid import uuid4
from datetime import datetime
from typing import Optional, List

from .base import BaseModel


class Agent(BaseModel):
    """Agent model - references existing agents table."""
    
    __tablename__ = "agents"
    
    id: Mapped[UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    tenant_id: Mapped[UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("tenants.id", ondelete="CASCADE"),
        nullable=False
    )
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    email: Mapped[str] = mapped_column(String(255), nullable=False)
    status: Mapped[str] = mapped_column(String(50), nullable=False, default="active")
    max_chats: Mapped[int] = mapped_column(Integer, default=5)  # Added in migration 028
    last_activity_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)  # Added in migration 028
    last_assignment_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)  # Added in migration 028
    
    # Relationships
    skills: Mapped[List["AgentSkill"]] = relationship(
        "AgentSkill",
        back_populates="agent",
        cascade="all, delete-orphan"
    )
    acknowledgments: Mapped[List["AlarmAcknowledgment"]] = relationship(
        "AlarmAcknowledgment",
        back_populates="agent"
    )


class AgentSkill(BaseModel):
    """Agent skills model from migration 028."""
    
    __tablename__ = "agent_skills"
    
    agent_id: Mapped[UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("agents.id", ondelete="CASCADE"),
        primary_key=True
    )
    tenant_id: Mapped[UUID] = mapped_column(UUID(as_uuid=True), nullable=False)
    skill: Mapped[str] = mapped_column(String(50), primary_key=True)
    
    # Relationships
    agent: Mapped["Agent"] = relationship("Agent", back_populates="skills")


class Alarm(BaseModel):
    """Alarm model from migration 027."""
    
    __tablename__ = "alarms"
    
    id: Mapped[UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    tenant_id: Mapped[UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("tenants.id", ondelete="CASCADE"),
        nullable=False
    )
    project_id: Mapped[UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("projects.id", ondelete="CASCADE"),
        nullable=False
    )
    assignment_id: Mapped[Optional[UUID]] = mapped_column(UUID(as_uuid=True), nullable=True)  # Logical reference
    agent_id: Mapped[Optional[UUID]] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("agents.id", ondelete="SET NULL"),
        nullable=True
    )
    title: Mapped[str] = mapped_column(String, nullable=False)
    message: Mapped[str] = mapped_column(String, nullable=False)
    priority: Mapped[str] = mapped_column(String, nullable=False, default="normal")  # notification_priority enum
    current_level: Mapped[str] = mapped_column(String, nullable=False, default="soft")
    start_time: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False, default=lambda: datetime.now())
    last_escalation: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False, default=lambda: datetime.now())
    escalation_count: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    is_acknowledged: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    acknowledged_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)
    acknowledged_by: Mapped[Optional[UUID]] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("agents.id", ondelete="SET NULL"),
        nullable=True
    )
    
    # Relationships
    acknowledgments: Mapped[List["AlarmAcknowledgment"]] = relationship(
        "AlarmAcknowledgment",
        back_populates="alarm",
        cascade="all, delete-orphan"
    )


class AlarmAcknowledgment(BaseModel):
    """Alarm acknowledgment model from migration 027."""
    
    __tablename__ = "alarm_acknowledgments"
    
    id: Mapped[UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    alarm_id: Mapped[UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("alarms.id", ondelete="CASCADE"),
        nullable=False
    )
    agent_id: Mapped[UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("agents.id", ondelete="CASCADE"),
        nullable=False
    )
    response: Mapped[str] = mapped_column(String, nullable=False, default="")
    acknowledged_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False, default=lambda: datetime.now())
    
    # Relationships
    alarm: Mapped["Alarm"] = relationship("Alarm", back_populates="acknowledgments")
    agent: Mapped["Agent"] = relationship("Agent", back_populates="acknowledgments")
