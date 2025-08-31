import { useState, useEffect, useCallback, useRef } from 'react'
import { HandoffRequestData, HandoffNotification } from '../types/handoff'
import { useAgentWebSocket } from './useAgentWebSocket'
import { apiClient, HandoffResponse } from '../lib/api'
import { useNavigate } from 'react-router-dom'

interface UseHandoffNotificationsReturn {
  notifications: HandoffNotification[]
  unreadCount: number
  isLoading: boolean
  error: string | null
  markAsRead: (notificationId: string) => void
  dismissNotification: (notificationId: string) => void
  clearAll: () => void
  acceptHandoff: (notificationId: string) => Promise<void>
  declineHandoff: (notificationId: string) => Promise<void>
  clearError: () => void
}

export function useHandoffNotifications(): UseHandoffNotificationsReturn {
  const [notifications, setNotifications] = useState<HandoffNotification[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const audioRef = useRef<HTMLAudioElement | null>(null)
  const navigate = useNavigate()
  
  // Initialize audio element for notification sounds
  useEffect(() => {
    if (typeof Audio !== 'undefined') {
      audioRef.current = new Audio('/notification-sound.mp3')
      audioRef.current.volume = 0.5
    }
  }, [])

  // Handle incoming handoff requests from WebSocket
  const handleHandoffRequest = useCallback((data: HandoffRequestData) => {
    const notification: HandoffNotification = {
      id: `handoff-${data.session_id}-${Date.now()}`,
      session_id: data.session_id,
      customer_name: data.customer_name,
      customer_email: data.customer_email,
      handoff_reason: data.handoff_reason,
      urgency_level: data.urgency_level,
      requested_at: data.requested_at,
      timeout_at: data.timeout_at,
      session_metadata: data.session_metadata,
      timestamp: new Date(),
      isRead: false,
      isActive: true
    }

    setNotifications(prev => [notification, ...prev])
    setUnreadCount(prev => prev + 1)

    // Play notification sound
    if (audioRef.current) {
      audioRef.current.play().catch(console.error)
    }

    // Show browser notification if available
    if (typeof Notification !== 'undefined' && Notification.permission === 'granted') {
      new Notification('New AI Handoff Request', {
        body: `Customer ${data.customer_email || data.customer_name || 'Unknown'} needs human assistance: ${data.handoff_reason}`,
        icon: '/favicon.ico',
        tag: notification.id,
        requireInteraction: true
      })
    }
  }, [])

  // Subscribe to WebSocket handoff events using the options pattern
  useAgentWebSocket({
    onHandoffRequest: handleHandoffRequest
  })

  // Request notification permission on mount
  useEffect(() => {
    if (typeof Notification !== 'undefined' && Notification.permission === 'default') {
      Notification.requestPermission()
    }
  }, [])

  const markAsRead = useCallback((notificationId: string) => {
    setNotifications(prev => 
      prev.map(n => 
        n.id === notificationId ? { ...n, isRead: true } : n
      )
    )
    setUnreadCount(prev => Math.max(0, prev - 1))
  }, [])

  const dismissNotification = useCallback((notificationId: string) => {
    setNotifications(prev => prev.filter(n => n.id !== notificationId))
    setUnreadCount(prev => {
      const notification = notifications.find(n => n.id === notificationId)
      return notification && !notification.isRead ? Math.max(0, prev - 1) : prev
    })
  }, [notifications])

  const clearAll = useCallback(() => {
    setNotifications([])
    setUnreadCount(0)
  }, [])

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  const acceptHandoff = useCallback(async (notificationId: string) => {
    const notification = notifications.find(n => n.id === notificationId)
    if (!notification) {
      setError('Notification not found')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const response: HandoffResponse = await apiClient.acceptHandoff(notification.session_id)
      
      if (response.success) {
        // Mark notification as handled
        setNotifications(prev => 
          prev.map(n => 
            n.id === notificationId ? { ...n, isActive: false } : n
          )
        )

        // Navigate to chat session using React Router
        navigate(`/chat/${notification.session_id}`)
      } else {
        throw new Error(response.message || 'Failed to accept handoff')
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to accept handoff'
      setError(errorMessage)
      console.error('Failed to accept handoff:', error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }, [notifications, navigate])

  const declineHandoff = useCallback(async (notificationId: string) => {
    const notification = notifications.find(n => n.id === notificationId)
    if (!notification) {
      setError('Notification not found')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const response: HandoffResponse = await apiClient.declineHandoff(notification.session_id)
      
      if (response.success) {
        // Remove notification
        dismissNotification(notificationId)
      } else {
        throw new Error(response.message || 'Failed to decline handoff')
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to decline handoff'
      setError(errorMessage)
      console.error('Failed to decline handoff:', error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }, [notifications, dismissNotification])

  return {
    notifications,
    unreadCount,
    isLoading,
    error,
    markAsRead,
    dismissNotification,
    clearAll,
    acceptHandoff,
    declineHandoff,
    clearError
  }
}
