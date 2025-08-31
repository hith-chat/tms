import { useRef, useEffect, useCallback, useState } from 'react'
import { useAuth } from './useAuth'
import { apiClient } from '../lib/api'
import type { ChatMessage, ChatSession } from '../types/chat'

interface WSMessage {
  type: 'chat_message' | 'typing_start' | 'typing_stop' | 'session_update' | 'agent_joined' | 'session_assigned' | 'notification' | 'error' | 'pong' | 'alarm_triggered' | 'alarm_acknowledged' | 'alarm_escalated'
  data: any // Changed from ChatMessage to any to handle different data types
  timestamp: string
  from_type: 'visitor' | 'agent' | 'system'
  session_id?: string // Add session_id at top level for typing messages
  // Chat message fields (when type is 'chat_message')
  delivery_type: 'direct' | 'broadcast' | 'self'
  // Error field (when type is 'error')
  error?: string
}

interface UseAgentWebSocketOptions {
  onMessage?: (message: ChatMessage) => void
  onSessionUpdate?: (session: ChatSession) => void
  onTyping?: (data: { isTyping: boolean; agentName?: string; sessionId: string }) => void
  onNotification?: (notification: any) => void
  onAlarm?: (alarm: any) => void
  onError?: (error: string) => void
}

interface WebSocketState {
  isConnected: boolean
  isConnecting: boolean
  error: string | null
  lastMessage: WSMessage | null
}

// SINGLETON WEBSOCKET MANAGER
class AgentWebSocketManager {
  private static instance: AgentWebSocketManager | null = null
  private ws: WebSocket | null = null
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null
  private pingInterval: ReturnType<typeof setInterval> | null = null
  private reconnectAttempts = 0
  private readonly maxReconnectAttempts = 5
  private readonly baseReconnectDelay = 1000
  private subscribers = new Set<{
    id: string
    onMessage?: (message: ChatMessage) => void
    onSessionUpdate?: (session: ChatSession) => void
    onTyping?: (data: { isTyping: boolean; agentName?: string; sessionId: string }) => void
    onNotification?: (notification: any) => void
    onAlarm?: (alarm: any) => void
    onError?: (error: string) => void
    setState: (state: WebSocketState) => void
  }>()
  private currentState: WebSocketState = {
    isConnected: false,
    isConnecting: false,
    error: null,
    lastMessage: null
  }

  static getInstance(): AgentWebSocketManager {
    if (!AgentWebSocketManager.instance) {
      AgentWebSocketManager.instance = new AgentWebSocketManager()
    }
    return AgentWebSocketManager.instance
  }

  subscribe(subscriber: any) {
    this.subscribers.add(subscriber)
    // Immediately sync current state
    subscriber.setState(this.currentState)
    return () => {
      this.subscribers.delete(subscriber)
      // If no more subscribers, disconnect
      if (this.subscribers.size === 0) {
        this.disconnect()
      }
    }
  }

  private updateState(newState: Partial<WebSocketState>) {
    this.currentState = { ...this.currentState, ...newState }
    this.subscribers.forEach(sub => sub.setState(this.currentState))
  }

  private notifySubscribers(type: string, data: any) {
    this.subscribers.forEach(sub => {
      switch (type) {
        case 'message':
          sub.onMessage?.(data)
          break
        case 'sessionUpdate':
          sub.onSessionUpdate?.(data)
          break
        case 'typing':
          sub.onTyping?.(data)
          break
        case 'notification':
          sub.onNotification?.(data)
          break
        case 'alarm':
          sub.onAlarm?.(data)
          break
        case 'error':
          sub.onError?.(data)
          break
      }
    })
  }

  connect(isAuthenticated: boolean, user: any) {
    if (!isAuthenticated || !user) return

    // Prevent multiple connections
    if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
      console.log('AgentWebSocketManager: Already connected/connecting, skipping')
      return
    }

    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('AgentWebSocketManager: Max reconnect attempts reached')
      return
    }

    console.log('AgentWebSocketManager: Connecting...')
    this.cleanup()
    this.updateState({ isConnecting: true, error: null })

    try {
      const token = localStorage.getItem('auth_token')
      if (!token) {
        this.updateState({ isConnecting: false })
        return
      }

      const wsUrl = `${getAgentWebSocketUrl()}?token=${token}`
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        console.log('AgentWebSocketManager: Connected')
        this.updateState({ isConnected: true, isConnecting: false, error: null })
        this.reconnectAttempts = 0

        this.pingInterval = setInterval(() => {
          if (this.ws?.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type: 'ping' }))
          }
        }, 30000)
      }

      this.ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data)
          console.log('AgentWebSocketManager: Message received', message.type)

          if (message.type !== 'pong') {
            this.updateState({ lastMessage: message })
          }

          switch (message.type) {
            case 'chat_message':
              this.notifySubscribers('message', message.data)
              break
            case 'session_update':
            case 'session_assigned':
              if (message.data) {
                this.notifySubscribers('sessionUpdate', message.data)
              }
              break
            case 'typing_start': {
              const sessionId = message.session_id || message.data?.session_id
              const authorName = message.data?.author_name || message.data?.agentName
              if (sessionId && authorName && authorName !== user?.name) {
                this.notifySubscribers('typing', { isTyping: true, agentName: authorName, sessionId })
              }
              break
            }
            case 'typing_stop': {
              const sessionId = message.session_id || message.data?.session_id
              const authorName = message.data?.author_name || message.data?.agentName
              if (sessionId && authorName && authorName !== user?.name) {
                this.notifySubscribers('typing', { isTyping: false, agentName: authorName, sessionId })
              }
              break
            }
            case 'error': {
              const errorMsg = message.error || 'WebSocket error occurred'
              this.updateState({ error: errorMsg })
              this.notifySubscribers('error', errorMsg)
              break
            }
            case 'notification':
              if (message.data) {
                this.notifySubscribers('notification', message.data)
              }
              break
            case 'alarm_triggered':
            case 'alarm_acknowledged':
            case 'alarm_escalated':
              if (message.data) {
                console.log('AgentWebSocketManager: Alarm message received', message.type, message.data)
                this.notifySubscribers('alarm', { type: message.type, ...message.data })
              }
              break
          }
        } catch (error) {
          console.error('AgentWebSocketManager: Failed to parse message:', error)
        }
      }

      this.ws.onclose = (event) => {
        console.log('AgentWebSocketManager: Closed', event.code)
        this.updateState({ isConnected: false, isConnecting: false })

        if (this.pingInterval) {
          clearInterval(this.pingInterval)
          this.pingInterval = null
        }

        // Auto-reconnect logic
        if (event.code !== 1000 && 
            this.reconnectAttempts < this.maxReconnectAttempts && 
            isAuthenticated && user) {
          
          const delay = this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts)
          this.reconnectAttempts++
          
          this.updateState({ 
            error: `Connection lost. Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`
          })
          
          this.reconnectTimeout = setTimeout(() => {
            this.connect(isAuthenticated, user)
          }, delay)
        } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          const errorMsg = 'Connection failed after multiple attempts. Please refresh the page.'
          this.updateState({ error: errorMsg })
          this.notifySubscribers('error', errorMsg)
        }
      }

      this.ws.onerror = (error) => {
        console.error('AgentWebSocketManager: Error', error)
        this.reconnectAttempts++
        
        const errorMsg = this.reconnectAttempts >= this.maxReconnectAttempts 
          ? 'Connection failed. Please refresh the page.'
          : `Connection error (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
          
        this.updateState({ error: errorMsg, isConnecting: false })
        
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          this.notifySubscribers('error', errorMsg)
        }
      }

    } catch (_error) {
      this.updateState({ error: 'Failed to connect', isConnecting: false })
    }
  }

  private cleanup() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  disconnect() {
    this.cleanup()
    this.updateState({ isConnected: false, isConnecting: false })
  }

  sendMessage(message: any): boolean {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
      return true
    }
    return false
  }

  manualRetry(isAuthenticated: boolean, user: any) {
    this.reconnectAttempts = 0
    this.updateState({ error: null })
    this.cleanup()
    setTimeout(() => this.connect(isAuthenticated, user), 500)
  }

  getState() {
    return this.currentState
  }
}

/**
 * Agent Global WebSocket hook for real-time chat communications
 * Single connection per agent that handles all chat sessions
 */
export function useAgentWebSocket(options: UseAgentWebSocketOptions = {}) {
  const { user, isAuthenticated } = useAuth()
  const [state, setState] = useState<WebSocketState>(() => 
    AgentWebSocketManager.getInstance().getState()
  )
  const subscriberId = useRef(`subscriber-${Math.random().toString(36).substr(2, 9)}`)

  // Subscribe to the singleton manager
  useEffect(() => {
    const manager = AgentWebSocketManager.getInstance()
    
    const subscriber = {
      id: subscriberId.current,
      ...options,
      setState
    }

    const unsubscribe = manager.subscribe(subscriber)
    
    // Auto-connect if authenticated
    if (isAuthenticated && user) {
      manager.connect(isAuthenticated, user)
    }

    return unsubscribe
  }, [isAuthenticated, user, options.onMessage, options.onSessionUpdate, options.onTyping, options.onNotification, options.onAlarm, options.onError])

  const connect = useCallback(() => {
    AgentWebSocketManager.getInstance().connect(isAuthenticated, user)
  }, [isAuthenticated, user])

  const disconnect = useCallback(() => {
    AgentWebSocketManager.getInstance().disconnect()
  }, [])

  const sendMessage = useCallback((message: any) => {
    return AgentWebSocketManager.getInstance().sendMessage(message)
  }, [])

  const sendChatMessage = useCallback(async (sessionId: string, projectId: string, content: string, senderName: string): Promise<boolean> => {
    const messageData = {
      type: 'chat_message',
      agent_session_id: sessionId,
      project_id: projectId,
      data: {
        content: content.trim(),
        message_type: 'text'
      }
    }

    // Try WebSocket first
    const manager = AgentWebSocketManager.getInstance()
    if (manager.sendMessage(messageData)) {
      console.log('Message sent via WebSocket')
      return true
    }

    // Fallback to HTTP API
    try {
      await apiClient.sendChatMessage(sessionId, {
        content: content.trim(),
        message_type: 'text',
        user_name: senderName
      })
      console.log('Message sent via HTTP API fallback')
      return true
    } catch (error) {
      console.error('Failed to send message:', error)
      return false
    }
  }, [])

  const sendTypingIndicator = useCallback((isTyping: boolean, sessionId: string) => {
    if (!sessionId) return false
    
    return sendMessage({
      type: isTyping ? 'typing_start' : 'typing_stop',
      agent_session_id: sessionId,
      data: {
        author_type: 'agent',
        author_name: user?.name || 'Agent'
      }
    })
  }, [sendMessage, user])

  const subscribeToSession = useCallback((sessionId: string) => {
    return sendMessage({
      type: 'session_subscribe',
      agent_session_id: sessionId,
      data: {}
    })
  }, [sendMessage])

  const unsubscribeFromSession = useCallback((sessionId: string) => {
    return sendMessage({
      type: 'session_unsubscribe',
      agent_session_id: sessionId,
      data: {}
    })
  }, [sendMessage])

  const manualRetry = useCallback(() => {
    AgentWebSocketManager.getInstance().manualRetry(isAuthenticated, user)
  }, [isAuthenticated, user])

  return {
    ...state,
    connect,
    disconnect,
    sendMessage,
    sendChatMessage,
    sendTypingIndicator,
    subscribeToSession,
    unsubscribeFromSession,
    manualRetry,
    reconnectAttempts: 0 // Manager handles this internally
  }
}

// Helper function to get agent WebSocket URL
function getAgentWebSocketUrl(): string {
  const tenantId = localStorage.getItem('tenant_id')
  
  if (!tenantId) {
    throw new Error('Tenant ID is required for Agent WebSocket connection')
  }
  
  // Use hardcoded base URL since apiClient.getBaseUrl() doesn't exist
  const baseUrl = 'http://localhost:8080'
  const wsUrl = baseUrl.replace('http', 'ws')
  return `${wsUrl}/v1/tenants/${tenantId}/chat/agent/ws`
}
