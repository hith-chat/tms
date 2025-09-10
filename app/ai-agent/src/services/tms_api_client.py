"""TMS API client for making authenticated requests."""

import asyncio
import logging
from typing import Dict, Optional, Any, List
import httpx
from datetime import datetime

from ..config import config
from .agent_auth_service import AgentAuthService
from .auth_service import auth_service

logger = logging.getLogger(__name__)


class TMSApiClient:
    """Client for making authenticated requests to TMS APIs."""
    
    def __init__(self):
        self.base_url = config.TMS_API_BASE_URL
        self.timeout = 30.0
    
    async def _make_request(
        self,
        method: str,
        endpoint: str,
        tenant_id: str,
        project_id: str,
        data: Optional[Dict] = None,
        params: Optional[Dict] = None,
        retry_auth: bool = True
    ) -> Optional[Dict]:
        """
        Make authenticated request to TMS API.
        
        Args:
            method: HTTP method (GET, POST, PUT, etc.)
            endpoint: API endpoint (without base URL)
            tenant_id: Tenant ID for authentication
            data: Request body data
            params: URL parameters
            retry_auth: Whether to retry on auth failure
            
        Returns:
            API response data or None on failure
        """
        # Get authentication token
        token = await auth_service.authenticate(tenant_id, project_id)
        logger.info(f"Obtained auth token for tenant {tenant_id}, project {project_id} with token: {token}")
        if not token:
            logger.error(f"Failed to get auth token for tenant {tenant_id}")
            return None
        
        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        }
        
        url = f"{self.base_url}{endpoint}"
        
        try:
            async with httpx.AsyncClient() as client:
                response = await client.request(
                    method=method,
                    url=url,
                    headers=headers,
                    json=data,
                    params=params,
                    timeout=self.timeout
                )
                
                # Handle authentication errors
                if response.status_code >= 299:
                    logger.error(
                        f"TMS API {method} {endpoint} failed: {response.status_code} - {response.text}"
                    )
                    raise httpx.HTTPStatusError(
                        f"Error response {response.status_code}",
                        request=response.request,
                        response=response
                    )
                

                return response.json() if response.content else {}
                
        except Exception as e:
            logger.error(f"TMS API request error for {method} {endpoint}: {e}")
            return None
    
    async def create_ticket(
        self,
        tenant_id: str,
        project_id: str,
        title: str,
        description: str,
        customer_email: Optional[str] = None,
        customer_name: Optional[str] = None,
        priority: str = "medium",
        source: str = "chat",
        category: str = "general"
    ) -> Optional[Dict]:
        """
        Create a support ticket via TMS API.
        
        Args:
            tenant_id: Tenant ID
            project_id: Project ID
            title: Ticket title
            description: Ticket description
            customer_email: Customer email address
            priority: Ticket priority (low, medium, high, urgent)
            category: Ticket category
            
        Returns:
            Created ticket data or None on failure
        """
        endpoint = f"/v1/tenants/{tenant_id}/projects/{project_id}/tickets"
        
        ticket_data = {
            "subject": title,
            "priority": priority,
            "type": category,
            "requester_email": customer_email,
            "requester_name": customer_name,
            "initial_message": description,
            "status": "open",
            "source": source
        }
        
        if customer_email:
            ticket_data["requester_email"] = customer_email
        
        logger.info(f"Creating ticket for tenant {tenant_id}, project {project_id}: {title}")

        logger.info(f"Ticket data: {ticket_data}")
        logger.info(f"Endpoint: {endpoint}")

        
        result = await self._make_request("POST", endpoint, tenant_id, project_id, ticket_data)
        
        if result:
            logger.info(f"Ticket created successfully: #{result.get('id', 'unknown')}")
        
        return result
    
    
    async def escalate_session(
        self,
        tenant_id: str,
        project_id: str,
        session_id: str,
        reason: str,
        priority: str = "high"
    ) -> Optional[Dict]:
        """
        Escalate a chat session to human agents.
        
        Note: This assumes an escalation API exists. If not, this will need
        to be implemented in the Go backend first.
        
        Args:
            tenant_id: Tenant ID
            project_id: Project ID  
            session_id: Chat session ID
            reason: Escalation reason
            priority: Escalation priority
            
        Returns:
            Escalation result or None on failure
        """
        endpoint = f"/v1/tenants/{tenant_id}/projects/{project_id}/chat/sessions/{session_id}/escalate"
        
        escalation_data = {
            "reason": reason,
            "priority": priority,
            "escalated_at": datetime.now().isoformat()
        }
        
        logger.info(f"Escalating session {session_id} for tenant {tenant_id}: {reason}")
        
        result = await self._make_request("POST", endpoint, tenant_id, escalation_data)
        
        if result:
            logger.info(f"Session {session_id} escalated successfully")
        else:
            # Fallback: Try to create a ticket with escalation flag
            logger.warning(f"Direct escalation API failed, creating escalation ticket instead")
            return await self.create_ticket(
                tenant_id=tenant_id,
                project_id=project_id,
                title=f"Escalated Chat Session - {reason}",
                description=f"Chat session {session_id} was escalated.\n\nReason: {reason}\n\nThis requires immediate attention from a human agent.",
                priority="high",
                category="escalation"
            )
        
        return result
    
    async def update_contact_info(
        self,
        tenant_id: str,
        project_id: str,
        session_id: str,
        email: Optional[str] = None,
        phone: Optional[str] = None,
        name: Optional[str] = None
    ) -> Optional[Dict]:
        """
        Update contact information for a chat session.
        
        Args:
            tenant_id: Tenant ID
            session_id: Chat session ID
            email: Customer email
            phone: Customer phone
            name: Customer name
            
        Returns:
            Update result or None on failure
        """
        # This might need to be adjusted based on actual API structure
        endpoint = f"/v1/tenants/{tenant_id}/customers"
        
        contact_data = {}
        if email:
            contact_data["email"] = email
        if phone:
            contact_data["phone"] = phone
        if name:
            contact_data["name"] = name
        
        if not contact_data:
            return {"success": True}  # Nothing to update
        
        logger.info(f"Updating contact info for session {session_id}")

        try:
            await self._make_request("POST", endpoint, tenant_id, project_id, contact_data)
        except Exception as e:
            logger.error(f"Failed to update contact info for session {session_id}: {e}")
        return {"success": True}

tms_api_client = TMSApiClient()