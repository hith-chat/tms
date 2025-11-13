import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { apiClient } from '../lib/api'
import type { DomainValidation } from '../lib/api'

export interface CreateChatWidgetRequest {
  name: string
  domain_id: string // Deprecated: use domain_url instead
  domain_url: string
  welcome_message?: string
  custom_greeting?: string
  away_message?: string
  primary_color?: string
  secondary_color?: string
  background_color?: string
  position?: 'bottom-right' | 'bottom-left'
  widget_shape?: 'rounded' | 'square' | 'minimal' | 'professional' | 'modern' | 'classic'
  chat_bubble_style?: 'modern' | 'classic' | 'minimal' | 'bot'
  widget_size?: 'small' | 'medium' | 'large'
  animation_style?: 'smooth' | 'bounce' | 'fade' | 'slide'
  agent_name?: string
  agent_avatar_url?: string
  allow_file_uploads?: boolean
  show_agent_avatars?: boolean
  require_email: boolean
  require_name: boolean
  sound_enabled?: boolean
  show_powered_by?: boolean
  use_ai?: boolean
  auto_open_delay?: number
  embed_code?: string
}

const defaultFormData: CreateChatWidgetRequest = {
  name: '',
  domain_id: '', // Deprecated
  domain_url: 'hith.chat',
  welcome_message: 'Hello! How can we help you today?',
  custom_greeting: 'Hi there! 👋 How can we help you today?',
  away_message: 'We\'re currently away. Leave us a message and we\'ll get back to you!',
  primary_color: '#3b82f6',
  secondary_color: '#6b7280',
  background_color: '#ffffff',
  position: 'bottom-right',
  widget_shape: 'rounded',
  chat_bubble_style: 'modern',
  widget_size: 'medium',
  animation_style: 'smooth',
  agent_name: 'Support Agent',
  agent_avatar_url: '',
  allow_file_uploads: false,
  show_agent_avatars: true,
  require_email: false,
  require_name: false,
  sound_enabled: true,
  show_powered_by: true, // Show branding by default
  use_ai: true, // Enable AI assistance by default
  auto_open_delay: 0
}

export function useChatWidgetForm() {
  const { widgetId } = useParams<{ widgetId: string }>()
  const [domains, setDomains] = useState<DomainValidation[]>([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [formData, setFormData] = useState<CreateChatWidgetRequest>(defaultFormData)

  const loadData = async () => {
    try {
      setLoading(true)
      setError(null)
      
      const domainsData = await apiClient.getDomainValidations()
      setDomains(domainsData.filter(d => d.status === 'verified'))
      
      // If editing, load the widget data
      if (widgetId) {
        const widget = await apiClient.getChatWidget(widgetId)
        setFormData({
          name: widget.name,
          domain_id: widget.domain_id || '', // Deprecated
          domain_url: widget.domain_url || 'hith.chat',
          welcome_message: widget.welcome_message || 'Hello! How can we help you today?',
          custom_greeting: widget.custom_greeting || 'Hi there! 👋 How can we help you today?',
          away_message: widget.away_message || 'We\'re currently away. Leave us a message and we\'ll get back to you!',
          primary_color: widget.primary_color || '#3b82f6',
          secondary_color: widget.secondary_color || '#6b7280',
          background_color: widget.background_color || '#ffffff',
          position: widget.position || 'bottom-right',
          widget_shape: widget.widget_shape || 'rounded',
          chat_bubble_style: widget.chat_bubble_style || 'modern',
          widget_size: widget.widget_size || 'medium',
          animation_style: widget.animation_style || 'smooth',
          agent_name: widget.agent_name || 'Support Agent',
          agent_avatar_url: widget.agent_avatar_url || '',
          allow_file_uploads: widget.allow_file_uploads || false,
          show_agent_avatars: widget.show_agent_avatars !== false,
          require_email: widget.require_email || false,
          require_name: widget.require_name || false,
          sound_enabled: widget.sound_enabled !== false,
          show_powered_by: widget.show_powered_by !== false,
          use_ai: widget.use_ai || false,
          auto_open_delay: widget.auto_open_delay || 0,
          embed_code: widget.embed_code || ''
        })
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load data')
    } finally {
      setLoading(false)
    }
  }

  const submitForm = async () => {

    try {
      setSubmitting(true)
      setError(null)
      if(formData.agent_avatar_url === '') {
        formData.agent_avatar_url = undefined
      }
      
      if (widgetId) {
        await apiClient.updateChatWidget(widgetId, formData)
      } else {
        await apiClient.createChatWidget(formData)
      }
      return true
    } catch (err: any) {
      setError(`Failed to ${widgetId ? 'update' : 'create'} widget: ${err.message}`)
      return false
    } finally {
      setSubmitting(false)
    }
  }

  const updateFormData = (updates: Partial<CreateChatWidgetRequest>) => {
    setFormData(prev => ({ ...prev, ...updates }))
  }

  useEffect(() => {
    loadData()
  }, [widgetId])

  return {
    widgetId,
    domains,
    loading,
    submitting,
    error,
    formData,
    updateFormData,
    submitForm,
    setError
  }
}
