"""Models package initialization."""

from .base import BaseModel
from .chat import ChatSession, ChatMessage, SupportTicket
from .knowledge import (
    KnowledgeSearchRequest,
    KnowledgeSearchResponse,
    KnowledgeSearchResult
)

__all__ = [
    "BaseModel",
    "ChatSession",
    "ChatMessage", 
    "SupportTicket",
    "KnowledgeSearchRequest",
    "KnowledgeSearchResponse",
    "KnowledgeSearchResult"
]
