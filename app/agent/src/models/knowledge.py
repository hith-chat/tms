"""Knowledge base related SQLAlchemy models based on existing migrations."""

from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy import String, Text, Integer, Boolean, Float, DateTime, ForeignKey, BIGINT
from sqlalchemy.dialects.postgresql import UUID, JSONB
from pgvector.sqlalchemy import Vector
from uuid import uuid4
from datetime import datetime
from typing import Optional, Dict, Any, List

from .base import BaseModel


class KnowledgeDocument(BaseModel):
    """Knowledge document model from migration 022."""
    
    __tablename__ = "knowledge_documents"
    
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
    filename: Mapped[str] = mapped_column(String(255), nullable=False)
    content_type: Mapped[str] = mapped_column(String(100), nullable=False)
    file_size: Mapped[int] = mapped_column(BIGINT, nullable=False)
    file_path: Mapped[str] = mapped_column(Text, nullable=False)
    original_content: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    processed_content: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    status: Mapped[str] = mapped_column(String(50), nullable=False, default="processing")
    error_message: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    metadata: Mapped[Dict[str, Any]] = mapped_column(JSONB, default=dict)
    
    # Relationships
    chunks: Mapped[List["KnowledgeChunk"]] = relationship(
        "KnowledgeChunk",
        back_populates="document",
        cascade="all, delete-orphan"
    )


class KnowledgeChunk(BaseModel):
    """Knowledge chunk model for embeddings from migration 022."""
    
    __tablename__ = "knowledge_chunks"
    
    id: Mapped[UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    document_id: Mapped[UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("knowledge_documents.id", ondelete="CASCADE"),
        nullable=False
    )
    chunk_index: Mapped[int] = mapped_column(Integer, nullable=False)
    content: Mapped[str] = mapped_column(Text, nullable=False)
    token_count: Mapped[int] = mapped_column(Integer, nullable=False)
    embedding: Mapped[Optional[List[float]]] = mapped_column(Vector(1536), nullable=True)  # Made nullable in migration 023
    metadata: Mapped[Dict[str, Any]] = mapped_column(JSONB, default=dict)
    
    # Relationships
    document: Mapped["KnowledgeDocument"] = relationship("KnowledgeDocument", back_populates="chunks")


class KnowledgeScrapingJob(BaseModel):
    """Web scraping job model from migration 022."""
    
    __tablename__ = "knowledge_scraping_jobs"
    
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
    url: Mapped[str] = mapped_column(Text, nullable=False)
    max_depth: Mapped[int] = mapped_column(Integer, nullable=False, default=5)
    status: Mapped[str] = mapped_column(String(50), nullable=False, default="pending")
    pages_scraped: Mapped[int] = mapped_column(Integer, default=0)
    total_pages: Mapped[int] = mapped_column(Integer, default=0)
    error_message: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    started_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)
    completed_at: Mapped[Optional[datetime]] = mapped_column(DateTime(timezone=True), nullable=True)
    
    # Relationships
    pages: Mapped[List["KnowledgeScrapedPage"]] = relationship(
        "KnowledgeScrapedPage",
        back_populates="job",
        cascade="all, delete-orphan"
    )


class KnowledgeScrapedPage(BaseModel):
    """Scraped page model from migration 022."""
    
    __tablename__ = "knowledge_scraped_pages"
    
    id: Mapped[UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid4)
    job_id: Mapped[UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("knowledge_scraping_jobs.id", ondelete="CASCADE"),
        nullable=False
    )
    url: Mapped[str] = mapped_column(Text, nullable=False)
    title: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    content: Mapped[str] = mapped_column(Text, nullable=False)
    token_count: Mapped[int] = mapped_column(Integer, nullable=False)
    scraped_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=lambda: datetime.now())
    embedding: Mapped[Optional[List[float]]] = mapped_column(Vector(1536), nullable=True)  # Made nullable in migration 023
    metadata: Mapped[Dict[str, Any]] = mapped_column(JSONB, default=dict)
    content_hash: Mapped[Optional[str]] = mapped_column(String(64), nullable=True)  # Added in migration 024/026
    page_id: Mapped[Optional[UUID]] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("knowledge_pages.id", ondelete="SET NULL"),
        nullable=True
    )  # Added in migration 025
    
    # Relationships
    job: Mapped["KnowledgeScrapingJob"] = relationship("KnowledgeScrapingJob", back_populates="pages")
    page: Mapped[Optional["KnowledgePage"]] = relationship("KnowledgePage", back_populates="scraped_versions")


class KnowledgePage(BaseModel):
    """Deduplicated knowledge page model from migration 025."""
    
    __tablename__ = "knowledge_pages"
    
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
    url: Mapped[str] = mapped_column(Text, nullable=False)
    content_hash: Mapped[str] = mapped_column(String(64), nullable=False)
    title: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    content: Mapped[str] = mapped_column(Text, nullable=False)
    token_count: Mapped[int] = mapped_column(Integer, nullable=False)
    embedding: Mapped[Optional[List[float]]] = mapped_column(Vector(1536), nullable=True)
    metadata: Mapped[Dict[str, Any]] = mapped_column(JSONB, default=dict)
    first_scraped_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=lambda: datetime.now())
    last_scraped_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=lambda: datetime.now())
    scrape_count: Mapped[int] = mapped_column(Integer, default=1)
    
    # Relationships
    scraped_versions: Mapped[List["KnowledgeScrapedPage"]] = relationship(
        "KnowledgeScrapedPage",
        back_populates="page"
    )


class KnowledgeSettings(BaseModel):
    """Knowledge base settings per project from migration 022."""
    
    __tablename__ = "knowledge_settings"
    
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
    enabled: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    embedding_model: Mapped[str] = mapped_column(String(100), nullable=False, default="text-embedding-ada-002")
    chunk_size: Mapped[int] = mapped_column(Integer, nullable=False, default=1000)
    chunk_overlap: Mapped[int] = mapped_column(Integer, nullable=False, default=200)
    max_context_chunks: Mapped[int] = mapped_column(Integer, nullable=False, default=5)
    similarity_threshold: Mapped[float] = mapped_column(Float, nullable=False, default=0.7)
