"""Authentication service for communicating with Go service."""

import logging
import httpx
import os
from typing import Optional
from cachetools import TTLCache

logger = logging.getLogger(__name__)


class AuthService:
    """Service for authenticating with Go service."""

    # TTL for tokens: 8 hours (in seconds)
    _TOKEN_TTL_SECONDS = 8 * 60 * 60

    def __init__(self):
        self.base_url = os.getenv("TMS_API_BASE_URL", "http://172.17.0.1:8080")
        self.agent_secret = os.getenv("AI_AGENT_LOGIN_ACCESS_KEY", "your-super-ai-access-key")
        logger.info("Using AI_AGENT_LOGIN_ACCESS_KEY: %s", self.agent_secret)
        self.ai_agent_email = os.getenv("AI_AGENT_EMAIL", "superai@acme.com")
        self.ai_agent_password = os.getenv("AI_AGENT_PASSWORD", "superai123")
        # Use an in-memory TTL cache to automatically expire tokens after 8 hours.
        # maxsize kept reasonable; adjust if many tenants/projects will be used.
        self.auth_tokens = TTLCache(maxsize=1024, ttl=self._TOKEN_TTL_SECONDS)

    async def authenticate(self, tenant_id: str, project_id: str) -> Optional[str]:
        """
        Authenticate with Go service and get auth token.

        Args:
            tenant_id: Tenant ID
            project_id: Project ID

        Returns:
            Authentication token or None if failed
        """
        cache_key = f"{tenant_id}:{project_id}"
        logger.info("Using AI_AGENT_LOGIN_ACCESS_KEY: %s", self.agent_secret)

        # Check if we have a cached token
        token = self.auth_tokens.get(cache_key)
        if token:
            logger.info("Using cached auth token for %s", cache_key)
            return token

        try:
            async with httpx.AsyncClient() as client:
                url = f"{self.base_url}/v1/auth/ai-agent/tenant/{tenant_id}/project/{project_id}/login"
                response = await client.post(
                    url,
                    json={"email": self.ai_agent_email, "password": self.ai_agent_password},
                    headers={"Content-Type": "application/json", "X-S2S-KEY": self.agent_secret},
                    timeout=10.0,
                )

                if response.status_code == 200:
                    data = response.json()
                    token = data.get("access_token")
                    if token:
                        # Cache the token with TTLCache -> expires automatically
                        self.auth_tokens[cache_key] = token
                        logger.info("Successfully authenticated for %s", cache_key)
                        return token
                    else:
                        logger.error("No token in response for %s", cache_key)
                        return None
                else:
                    logger.error(
                        "Authentication failed for %s: %s - %s",
                        cache_key,
                        response.status_code,
                        response.text,
                    )
                    return None

        except Exception as e:
            logger.error("Error authenticating with Go service for %s: %s", cache_key, e)
            return None


auth_service = AuthService()