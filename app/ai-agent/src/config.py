"""Configuration management for the agent service."""

import os
from typing import Optional


class Config:
    """Application configuration."""
    
    # Database Configuration
    DATABASE_URL: str = os.getenv(
        "DATABASE_URL", 
        "postgresql+asyncpg://tms:tms123@localhost:5432/tms?sslmode=disable"
    )
    
    # OpenAI Configuration
    AI_API_KEY: Optional[str] = os.getenv("AI_API_KEY")
    
    # TMS API Configuration
    TMS_API_BASE_URL: str = os.getenv("TMS_API_BASE_URL", "https://api.hith.chat")
    
    # Application Configuration
    HOST: str = os.getenv("HOST", "0.0.0.0")
    PORT: int = int(os.getenv("PORT", "5000"))
    DEBUG: bool = os.getenv("DEBUG", "false").lower() in ("true", "1", "yes")
    
    # Database Pool Configuration
    DB_POOL_MIN_SIZE: int = int(os.getenv("DB_POOL_MIN_SIZE", "5"))
    DB_POOL_MAX_SIZE: int = int(os.getenv("DB_POOL_MAX_SIZE", "20"))
    
    # Knowledge Base Configuration
    KB_SIMILARITY_THRESHOLD: float = float(os.getenv("KB_SIMILARITY_THRESHOLD", "0.7"))
    KB_MAX_RESULTS: int = int(os.getenv("KB_MAX_RESULTS", "3"))
    
    @classmethod
    def validate(cls) -> None:
        """Validate required configuration."""
        if not cls.AI_API_KEY:
            raise ValueError("AI_API_KEY environment variable is required")


# Global configuration instance
config = Config()
