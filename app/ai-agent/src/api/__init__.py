"""API package initialization."""

from .chat import router as chat_router
from .chat_handler import router as chat_handler_router
from .health import router as health_router

__all__ = ["chat_router", "chat_handler_router", "health_router"]
