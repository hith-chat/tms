"""Agent authentication service for TMS API access."""

import asyncio
import logging
from typing import Dict, Optional, Tuple
from datetime import datetime, timedelta
import httpx
from dataclasses import dataclass

from ..config import config

logger = logging.getLogger(__name__)


@dataclass
class AgentToken:
    """Agent authentication token with expiration."""
    access_token: str
    refresh_token: str
    expires_at: datetime
    tenant_id: str
    agent_id: str


@dataclass
class AgentCredentials:
    """Agent login credentials."""
    email: str
    password: str
    tenant_id: str


class AgentAuthService:
    """Service for managing agent authentication with TMS APIs."""
    
    def __init__(self):
        self.tokens: Dict[str, AgentToken] = {}  # tenant_id -> token
        self.credentials: Dict[str, AgentCredentials] = {}  # tenant_id -> credentials
        self._load_agent_credentials()
    
    def _load_agent_credentials(self):
        """Load agent credentials from configuration."""
        # TODO: Load from environment variables or config file
        # For now, using example credentials structure
        self.credentials = {
            # Format: tenant_id -> credentials
            "550e8400-e29b-41d4-a716-446655440000": AgentCredentials(
                email="admin@acme.com",
                password="password",
                tenant_id="550e8400-e29b-41d4-a716-446655440000"
            )
        }
    
    async def get_authenticated_token(self, tenant_id: str) -> Optional[str]:
        """
        Get valid authentication token for a tenant.
        
        Args:
            tenant_id: Tenant ID to authenticate for
            
        Returns:
            Valid access token or None if authentication fails
        """
        # Check if we have a valid token
        if tenant_id in self.tokens:
            token = self.tokens[tenant_id]
            if datetime.now() < token.expires_at - timedelta(minutes=5):  # 5 min buffer
                return token.access_token
        
        # Token expired or doesn't exist, need to login
        return await self._login_agent(tenant_id)
    
    async def _login_agent(self, tenant_id: str) -> Optional[str]:
        """
        Login agent and store token.
        
        Args:
            tenant_id: Tenant ID to login for
            
        Returns:
            Access token or None if login fails
        """
        if tenant_id not in self.credentials:
            logger.error(f"No credentials configured for tenant {tenant_id}")
            return None
        
        creds = self.credentials[tenant_id]
        
        try:
            async with httpx.AsyncClient() as client:
                login_data = {
                    "email": creds.email,
                    "password": creds.password
                }
                
                response = await client.post(
                    f"{config.TMS_API_BASE_URL}/v1/auth/login",
                    json=login_data,
                    headers={"Content-Type": "application/json"},
                    timeout=30.0
                )
                
                if response.status_code != 200:
                    logger.error(f"Agent login failed for tenant {tenant_id}: {response.status_code}")
                    return None
                
                auth_data = response.json()
                
                # Store token
                expires_in = auth_data.get("expires_in", 3600)  # Default 1 hour
                token = AgentToken(
                    access_token=auth_data["access_token"],
                    refresh_token=auth_data["refresh_token"],
                    expires_at=datetime.now() + timedelta(seconds=expires_in),
                    tenant_id=tenant_id,
                    agent_id=auth_data["user"]["id"]
                )
                
                self.tokens[tenant_id] = token
                logger.info(f"Agent authenticated successfully for tenant {tenant_id}")
                
                return token.access_token
                
        except Exception as e:
            logger.error(f"Agent authentication error for tenant {tenant_id}: {e}")
            return None
    
    async def refresh_token(self, tenant_id: str) -> Optional[str]:
        """
        Refresh authentication token.
        
        Args:
            tenant_id: Tenant ID to refresh token for
            
        Returns:
            New access token or None if refresh fails
        """
        if tenant_id not in self.tokens:
            return await self._login_agent(tenant_id)
        
        token = self.tokens[tenant_id]
        
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{config.TMS_API_BASE_URL}/v1/auth/refresh",
                    headers={
                        "Authorization": f"Bearer {token.refresh_token}",
                        "Content-Type": "application/json"
                    },
                    timeout=30.0
                )
                
                if response.status_code != 200:
                    logger.warning(f"Token refresh failed for tenant {tenant_id}, re-authenticating")
                    return await self._login_agent(tenant_id)
                
                auth_data = response.json()
                
                # Update token
                expires_in = auth_data.get("expires_in", 3600)
                token.access_token = auth_data["access_token"]
                token.expires_at = datetime.now() + timedelta(seconds=expires_in)
                
                if "refresh_token" in auth_data:
                    token.refresh_token = auth_data["refresh_token"]
                
                logger.info(f"Token refreshed for tenant {tenant_id}")
                return token.access_token
                
        except Exception as e:
            logger.error(f"Token refresh error for tenant {tenant_id}: {e}")
            return await self._login_agent(tenant_id)
    
    def get_agent_info(self, tenant_id: str) -> Optional[Tuple[str, str]]:
        """
        Get agent ID and tenant ID for a tenant.
        
        Args:
            tenant_id: Tenant ID
            
        Returns:
            Tuple of (agent_id, tenant_id) or None
        """
        if tenant_id in self.tokens:
            token = self.tokens[tenant_id]
            return token.agent_id, token.tenant_id
        return None
    
    def add_agent_credentials(self, tenant_id: str, email: str, password: str):
        """
        Add agent credentials for a tenant (for dynamic configuration).
        
        Args:
            tenant_id: Tenant ID
            email: Agent email
            password: Agent password
        """
        self.credentials[tenant_id] = AgentCredentials(
            email=email,
            password=password,
            tenant_id=tenant_id
        )
        logger.info(f"Added agent credentials for tenant {tenant_id}")
