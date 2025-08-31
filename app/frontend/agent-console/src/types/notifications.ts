// Notification types for the frontend
export interface Notification {
  id: string
  tenant_id: string
  project_id?: string
  agent_id: string
  type: 'ticket_assigned' | 'ticket_updated' | 'ticket_escalated' | 'ticket_resolved' | 
        'message_received' | 'mention_received' | 'sla_warning' | 'sla_breach' |
        'system_alert' | 'maintenance_notice' | 'feature_announcement' |
        'agent_assignment' | 'howling_alarm' | 'agent_auto_assigned' | 'knowledge_response' | 'greeting_response'
  title: string
  message: string
  priority: 'low' | 'normal' | 'high' | 'urgent' | 'critical'
  channels: ('web' | 'email' | 'slack' | 'sms' | 'push' | 'audio' | 'desktop' | 'overlay' | 'popup')[]
  action_url?: string
  metadata?: Record<string, any>
  is_read: boolean
  read_at?: string
  expires_at?: string
  created_at: string
  updated_at: string
}

// Phase 4: Howling Alarm types
export interface HowlingAlarm {
  id: string
  assignment_id: string
  agent_id: string
  tenant_id: string
  project_id: string
  title: string
  message: string
  priority: 'normal' | 'high' | 'urgent' | 'critical'
  current_level: 'soft' | 'medium' | 'loud' | 'urgent' | 'critical'
  escalation_count: number
  created_at: string
  updated_at: string
  acknowledged_at?: string
  acknowledged_by?: string
  acknowledged_response?: string
  last_escalation_at?: string
  escalation_interval: number // seconds
}

export interface NotificationCount {
  total: number
  unread: number
}

export interface AlarmStats {
  total_active: number
  by_level: Record<string, number>
  by_priority: Record<string, number>
  average_duration: number
  escalation_counts: Record<string, number>
}

export interface NotificationSettings {
  sound_enabled: boolean
  browser_notifications: boolean
  email_notifications: boolean
  // Phase 4: Enhanced notification channels
  audio_notifications: boolean
  desktop_notifications: boolean
  overlay_notifications: boolean
  popup_notifications: boolean
  alarm_sound_enabled: boolean
  alarm_escalation_sound: boolean
  notification_types: {
    ticket_assigned: boolean
    ticket_updated: boolean
    ticket_escalated: boolean
    ticket_resolved: boolean
    message_received: boolean
    mention_received: boolean
    sla_warning: boolean
    sla_breach: boolean
    system_alert: boolean
    maintenance_notice: boolean
    feature_announcement: boolean
    // Phase 4: New notification types
    agent_assignment: boolean
    howling_alarm: boolean
    agent_auto_assigned: boolean
    knowledge_response: boolean
    greeting_response: boolean
  }
}
