export interface HandoffRequestData {
  session_id: string
  customer_name?: string
  customer_email?: string
  handoff_reason: string
  urgency_level: 'low' | 'normal' | 'high' | 'critical'
  requested_at: string
  timeout_at?: string
  session_metadata: {
    messages_count: number
    session_duration: number
    last_ai_response: string
  }
}

export interface HandoffNotification {
  id: string
  session_id: string
  customer_name?: string
  customer_email?: string
  handoff_reason: string
  urgency_level: 'low' | 'normal' | 'high' | 'critical'
  requested_at: string
  timeout_at?: string
  session_metadata: {
    messages_count: number
    session_duration: number
    last_ai_response: string
  }
  timestamp: Date
  isRead: boolean
  isActive: boolean
}
