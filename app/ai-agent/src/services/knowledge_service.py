"""Knowledge base search service that calls the Hith backend API.

This service provides a small helper that calls the tenant-scoped knowledge
search endpoint and returns parsed results suitable for the agent service.
"""

import logging
from typing import List, Dict, Any, Optional
import httpx

from ..config import config
from ..schemas.knowledge import (
    KnowledgeSearchRequest,
    KnowledgeSearchResponse,
)
from .auth_service import auth_service

logger = logging.getLogger(__name__)


class KnowledgeService:
    """Service for knowledge base operations that calls backend API."""

    def __init__(self):
        # OpenAI client kept for future use (embeddings etc.)
        self.base_url = config.TMS_API_BASE_URL

    async def search_knowledge_base(
        self,
        query: str,
        tenant_id: Optional[str] = None,
        project_id: Optional[str] = None,
        limit: int = None,
    ) -> KnowledgeSearchResponse:
        """Search the Hith backend knowledge endpoint and return results.

        Args:
            session: Database session (unused here but kept for compatibility)
            query: Search query string
            tenant_id: Tenant UUID string
            project_id: Project UUID string
            limit: Maximum number of results to return

        Returns:
            List of result dicts (each matches KnowledgeSearchResult shape)
        """

        if not tenant_id or not project_id:
            logger.debug("Knowledge search called without tenant or project id")
            return []

        # Build request model
        req = KnowledgeSearchRequest(
            query=query,
            max_results=limit or config.KB_MAX_RESULTS,
            similarity_score=getattr(config, "KB_SIMILARITY_THRESHOLD", 0.7),
            include_documents=True,
            include_pages=True,
        )

        url = f"{self.base_url}/v1/tenants/{tenant_id}/projects/{project_id}/knowledge/search"

        headers = {"Content-Type": "application/json", "Accept": "application/json"}

        auth_token = auth_service.authenticate(tenant_id, project_id)
        if auth_token:
            headers["Authorization"] = f"Bearer {auth_token}"

        try:
            async with httpx.AsyncClient(timeout=30.0) as client:
                resp = await client.post(url, json=req.dict(), headers=headers)

                if resp.status_code >= 400:
                    logger.error(f"Knowledge API error {resp.status_code}: {resp.text}")
                    return []

                data = resp.json() if resp.content else {}

                try:
                    parsed = KnowledgeSearchResponse.model_validate(data)
                    return parsed
                except Exception as e:
                    logger.error(f"Failed to parse knowledge response: {e}")
                    # best effort fallback
                    if isinstance(data, dict):
                        return data.get("results", [])
                    return []

        except Exception as e:
            logger.error(f"Knowledge search request failed: {e}")
            return []