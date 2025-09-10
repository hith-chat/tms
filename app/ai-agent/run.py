#!/usr/bin/env python3
"""Entry point for the agent service."""

import uvicorn
from src.config import config

if __name__ == "__main__":
    uvicorn.run(
        "src.main:app",
        host=config.HOST,
        port=config.PORT,
        reload=config.DEBUG,
        log_level="info"
    )
