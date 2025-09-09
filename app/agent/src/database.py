"""Database connection and session management using async SQLAlchemy."""

from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.orm import declarative_base
from typing import AsyncGenerator
import logging

from .config import config

logger = logging.getLogger(__name__)

# SQLAlchemy base class for models
Base = declarative_base()

# Async engine
engine = create_async_engine(
    config.DATABASE_URL,
    echo=config.DEBUG,
    pool_pre_ping=True,
    pool_size=config.DB_POOL_MIN_SIZE,
    max_overflow=config.DB_POOL_MAX_SIZE - config.DB_POOL_MIN_SIZE,
)

# Session factory
AsyncSessionLocal = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
)


async def get_db_session() -> AsyncGenerator[AsyncSession, None]:
    """
    Dependency function to get database session.
    
    Yields:
        AsyncSession: Database session
    """
    async with AsyncSessionLocal() as session:
        try:
            yield session
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()


async def init_db() -> None:
    """
    Initialize database connection.
    
    Note: This does NOT create tables - that's handled by Go migrations.
    This only tests the connection.
    """
    try:
        async with engine.begin() as conn:
            # Test connection
            await conn.execute("SELECT 1")
        logger.info("Database connection established successfully")
    except Exception as e:
        logger.error(f"Failed to connect to database: {e}")
        raise


async def close_db() -> None:
    """Close database connections."""
    await engine.dispose()
    logger.info("Database connections closed")
