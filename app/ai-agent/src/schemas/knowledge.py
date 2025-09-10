"""Pydantic request/response models for knowledge search operations.

These are plain Pydantic models used for request validation and response
serialization (not SQLAlchemy/DB models).
"""

from typing import Optional, Dict, Any, List
from uuid import uuid4, UUID

from pydantic import BaseModel, Field


class KnowledgeSearchRequest(BaseModel):
    """Model for knowledge search requests."""

    query: str = Field(..., description="Search query string")
    max_results: int = Field(5, ge=1, description="Maximum number of results to return")
    similarity_score: float = Field(
        0.75, ge=0.0, le=1.0, description="Minimum similarity score filter (0.0-1.0)"
    )
    include_documents: bool = Field(True, description="Whether to include document results")
    include_pages: bool = Field(True, description="Whether to include page/webpage results")


class KnowledgeSearchResult(BaseModel):
    """Model for an individual knowledge search result."""

    id: UUID = Field(default_factory=uuid4, description="Unique identifier for the result")
    type: str = Field(..., description='Result type, e.g. "document" or "webpage"')
    content: str = Field(..., description="Extracted content of the result")
    score: float = Field(..., description="Similarity or relevance score")
    source: str = Field(..., description="Source filename or URL")
    title: Optional[str] = Field(None, description="Optional title for the result")
    meta: Dict[str, Any] = Field(default_factory=dict, description="Arbitrary metadata")


class KnowledgeSearchResponse(BaseModel):
    """Model for knowledge search responses."""

    results: List[KnowledgeSearchResult] = Field(default_factory=list)
    total_count: int = Field(..., ge=0)
    query: str = Field(..., description="Echo of the original query")
    processed_in: str = Field(..., description="Human-readable processing time, e.g. '123ms'")
