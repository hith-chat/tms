#!/usr/bin/env python3
"""
Example script showing how to interact with the restructured agent service.

This demonstrates the READ-ONLY nature of the agent - it returns JSON instructions
for actions rather than performing database operations directly.
"""

import asyncio
import json
from datetime import datetime

# Mock request/response for demonstration
async def demo_agent_interaction():
    """Demonstrate agent interaction patterns."""
    
    print("🤖 Hith Agent Service Demo")
    print("=" * 50)
    
    # Example 1: Knowledge base search response
    print("\n1. Knowledge Base Query:")
    print("User: How do I reset my password?")
    print("Agent Response:", json.dumps({
        "content": "Based on our knowledge base:\n\n**Password Reset**\nTo reset your password:\n1. Go to the login page\n2. Click 'Forgot Password'\n3. Enter your email address\n4. Check your email for reset instructions",
        "source": "knowledge_base",
        "articles_found": 1
    }, indent=2))
    
    # Example 2: Ticket creation instructions
    print("\n2. Ticket Creation Request:")
    print("User: I can't access my account and the password reset isn't working")
    print("Agent Response:", json.dumps({
        "content": "I'll help you create a support ticket. Here are the details:\n\n**Title:** Account access issue with password reset\n**Description:** User unable to access account and password reset feature not working\n\nPlease confirm if you'd like me to proceed with creating this ticket.",
        "pending_actions": {
            "ticket": {
                "action": "create_ticket",
                "data": {
                    "title": "Account access issue with password reset",
                    "description": "User unable to access account and password reset feature not working",
                    "priority": "medium",
                    "category": "general", 
                    "session_id": "demo-session-123",
                    "user_email": None
                }
            }
        }
    }, indent=2))
    
    # Example 3: Escalation instructions
    print("\n3. Human Escalation Request:")
    print("User: This is urgent, I need to access my account for a client presentation")
    print("Agent Response:", json.dumps({
        "content": "I understand this requires human attention. I'll escalate your case for priority handling. Reason: Urgent account access needed for client presentation",
        "pending_actions": {
            "escalation": {
                "action": "escalate_to_human",
                "data": {
                    "reason": "Urgent account access needed for client presentation",
                    "session_id": "demo-session-123",
                    "priority": "high",
                    "requires_immediate_attention": True
                }
            }
        }
    }, indent=2))
    
    # Example 4: Contact information collection
    print("\n4. Contact Information Request:")
    print("User: My email is user@example.com")
    print("Agent Response:", json.dumps({
        "content": "Thank you for providing your contact information! I've noted it for follow-up purposes.",
        "pending_actions": {
            "contact_update": {
                "action": "update_contact_info",
                "data": {
                    "session_id": "demo-session-123",
                    "email": "user@example.com",
                    "phone": None,
                    "name": None
                }
            }
        }
    }, indent=2))
    
    print("\n" + "=" * 50)
    print("✅ Demo complete! The agent provides structured responses")
    print("   with actionable JSON that external systems can process.")


if __name__ == "__main__":
    asyncio.run(demo_agent_interaction())
