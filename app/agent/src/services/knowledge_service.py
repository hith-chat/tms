"""Knowledge base search service using async SQLAlchemy."""

import logging
from typing import List, Dict, Any, Optional
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, text
from openai import AsyncOpenAI

from ..config import config
from ..models.knowledge import KnowledgePage, KnowledgeScrapedPage

logger = logging.getLogger(__name__)


class KnowledgeService:
    """Service for knowledge base operations."""
    
    def __init__(self):
        self.openai_client = AsyncOpenAI(api_key=config.AI_API_KEY)
    
    async def search_knowledge_base(
        self,
        session: AsyncSession,
        query: str,
        tenant_id: Optional[str] = None,
        project_id: Optional[str] = None,
        limit: int = None
    ) -> List[Dict[str, Any]]:
        """
        Search knowledge base using vector similarity.
        
        Args:
            session: Database session
            query: Search query
            tenant_id: Optional tenant filter
            project_id: Optional project filter
            limit: Number of results to return
            
        Returns:
            List of relevant knowledge base articles with similarity scores
        """
        try:
            # Generate embedding for query
            response = await self.openai_client.embeddings.create(
                model="text-embedding-3-small",
                input=query
            )
            query_embedding = response.data[0].embedding
            
            # Use configured limit if not provided
            if limit is None:
                limit = config.KB_MAX_RESULTS
            
            # Search in deduplicated pages table first
            page_query = select(
                KnowledgePage.url,
                KnowledgePage.title,
                KnowledgePage.content,
                KnowledgePage.metadata,
                text("1 - (embedding <=> :embedding::vector) as similarity")
            ).where(
                text("embedding IS NOT NULL")
            ).order_by(
                text("embedding <=> :embedding::vector")
            ).limit(limit)
            
            # Add tenant/project filters if provided
            if tenant_id:
                page_query = page_query.where(KnowledgePage.tenant_id == tenant_id)
            if project_id:
                page_query = page_query.where(KnowledgePage.project_id == project_id)
            
            result = await session.execute(page_query, {"embedding": query_embedding})
            pages = result.fetchall()
            
            # If no results in pages table, fall back to scraped pages
            if not pages:
                scraped_query = select(
                    KnowledgeScrapedPage.url,
                    KnowledgeScrapedPage.title, 
                    KnowledgeScrapedPage.content,
                    KnowledgeScrapedPage.metadata,
                    text("1 - (embedding <=> :embedding::vector) as similarity")
                ).where(
                    text("embedding IS NOT NULL")
                ).order_by(
                    text("embedding <=> :embedding::vector")
                ).limit(limit)
                
                result = await session.execute(scraped_query, {"embedding": query_embedding})
                pages = result.fetchall()
            
            # Filter by similarity threshold and format results
            results = []
            for page in pages:
                if page.similarity >= config.KB_SIMILARITY_THRESHOLD:
                    results.append({
                        "url": page.url,
                        "title": page.title,
                        "content": page.content,
                        "metadata": page.metadata or {},
                        "similarity": float(page.similarity)
                    })
            
            return results
            
        except Exception as e:
            logger.error(f"Knowledge base search error: {e}")
            return []
    
    async def format_kb_response(self, articles: List[Dict[str, Any]]) -> Optional[str]:
        """
        Format knowledge base articles into a response.
        
        Args:
            articles: List of knowledge base articles
            
        Returns:
            Formatted response string or None if no relevant articles
        """
        if not articles:
            return None
        
        # Check if the top result meets the threshold
        if articles[0].get('similarity', 0) < config.KB_SIMILARITY_THRESHOLD:
            return None
        
        # Format response from KB articles
        response = "Based on our knowledge base:\n\n"
        for article in articles[:2]:  # Use top 2 results
            metadata = article.get('metadata', {})
            title = metadata.get('title', article.get('title', 'Information'))
            response += f"**{title}**\n"
            response += f"{article['content']}\n\n"
        
        return response
