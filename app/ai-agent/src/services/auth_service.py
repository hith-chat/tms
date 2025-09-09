"""Authentication service for communicating with Go service."""

import logging
import httpx
import os
from typing import Optional

logger = logging.getLogger(__name__)


class AuthService:
    """Service for authenticating with Go service."""
    
    def __init__(self):
        self.base_url = os.getenv("TMS_API_BASE_URL", "http://172.17.0.1:8080")
        self.agent_secret = os.getenv("AI_AGENT_LOGIN_ACCESS_KEY", "your-super-ai-access-key")
        self.ai_agent_email = os.getenv("AI_AGENT_EMAIL", "superai@acme.com")
        self.ai_agent_password = os.getenv("AI_AGENT_PASSWORD", "superai123")
        self.auth_tokens = {}  # Cache tokens per tenant/project
    
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
        
        # Check if we have a cached token
        if cache_key in self.auth_tokens:
            logger.info(f"Using cached auth token for {cache_key}")
            return self.auth_tokens[cache_key]
        
        try:
            async with httpx.AsyncClient() as client:
                url = f"{self.base_url}/v1/auth/ai-agent/tenant/{tenant_id}/project/{project_id}/login"
                response = await client.post(url,
                    json={"email": self.ai_agent_email ,"password": self.ai_agent_password},
                    headers={"Content-Type": "application/json", "X-S2S-KEY": self.agent_secret},
                    timeout=10.0
                )
                
                if response.status_code == 200:
                    data = response.json()
                    token = data.get("access_token")
                    if token:
                        # Cache the token
                        self.auth_tokens[cache_key] = token
                        logger.info(f"Successfully authenticated for {cache_key}")
                        return token
                    else:
                        logger.error(f"No token in response for {cache_key}")
                        return None
                else:
                    logger.error(f"Authentication failed for {cache_key}: {response.status_code} - {response.text}")
                    return None
                    
        except Exception as e:
            logger.error(f"Error authenticating with Go service for {cache_key}: {e}")
            return None
        
auth_service = AuthService()