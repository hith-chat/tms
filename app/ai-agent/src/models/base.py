"""SQLAlchemy base model definitions."""

from sqlalchemy.ext.declarative import declared_attr
from sqlalchemy.orm import Mapped, mapped_column
from sqlalchemy import DateTime, func
from sqlalchemy.ext.asyncio import AsyncAttrs
from datetime import datetime
from typing import Optional

from ..database import Base


class TimestampMixin:
    """Mixin for created_at and updated_at timestamps."""
    
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        default=func.now(),
        nullable=False
    )
    updated_at: Mapped[Optional[datetime]] = mapped_column(
        DateTime(timezone=True),
        default=func.now(),
        onupdate=func.now(),
        nullable=True
    )


class BaseModel(AsyncAttrs, Base, TimestampMixin):
    """Base model class with common functionality."""
    
    __abstract__ = True
    
    @declared_attr
    def __tablename__(cls) -> str:
        """Generate table name from class name."""
        return cls.__name__.lower()
