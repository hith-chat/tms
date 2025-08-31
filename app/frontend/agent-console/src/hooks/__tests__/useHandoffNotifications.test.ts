import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import React from 'react'
import { useHandoffNotifications } from '../useHandoffNotifications'
import { apiClient } from '../../lib/api'

// Mock the API client methods
vi.mock('../../lib/api', () => ({
  apiClient: {
    acceptHandoff: vi.fn(),
    declineHandoff: vi.fn(),
    getHandoffStatus: vi.fn(),
  }
}))

// Mock react-router-dom navigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

// Mock useAgentWebSocket
const mockHandoffHandler = vi.fn()
vi.mock('../useAgentWebSocket', () => ({
  useAgentWebSocket: vi.fn((options) => {
    if (options?.onHandoffRequest) {
      mockHandoffHandler.mockImplementation(options.onHandoffRequest)
    }
    return {
      isConnected: true,
      connect: vi.fn(),
      disconnect: vi.fn(),
    }
  })
}))

// Mock browser APIs
const mockAudio = {
  play: vi.fn().mockResolvedValue(undefined),
  pause: vi.fn(),
  volume: 0.5,
}

// Mock HTMLAudioElement constructor
global.Audio = vi.fn(() => mockAudio) as any

// Mock Notification API
Object.defineProperty(global, 'Notification', {
  value: class MockNotification {
    static permission = 'default' as NotificationPermission
    static requestPermission = vi.fn().mockResolvedValue('granted')
    constructor(_title: string, _options?: NotificationOptions) {
      // Mock notification instance
    }
  },
  writable: true,
})

// Test wrapper component
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  
  return ({ children }: { children: React.ReactNode }) => (
    React.createElement(QueryClientProvider, { client: queryClient },
      React.createElement(BrowserRouter, null, children)
    )
  )
}

describe('useHandoffNotifications', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockNavigate.mockClear()
    // Reset Notification permission
    ;(global.Notification as any).permission = 'default'
  })

  it('should initialize with empty notifications', () => {
    const { result } = renderHook(useHandoffNotifications, {
      wrapper: createWrapper(),
    })

    expect(result.current.notifications).toEqual([])
    expect(result.current.unreadCount).toBe(0)
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('should accept handoff successfully', async () => {
    const mockHandoffResponse = {
      success: true,
      session_id: 'test-session-id',
      agent_id: 'test-agent-id',
      tenant_id: 'test-tenant-id',
      message: 'Handoff accepted successfully',
      accepted_at: '2024-01-01T00:00:00Z'
    }

    const mockHandoffData = {
      session_id: 'test-session-id',
      customer_name: 'Test Customer',
      customer_email: 'test@example.com',
      handoff_reason: 'Need human help',
      urgency_level: 'high' as const,
      requested_at: '2024-01-01T00:00:00Z',
      session_metadata: {
        messages_count: 5,
        session_duration: 300,
        last_ai_response: 'Let me get a human agent for you'
      }
    }

    vi.mocked(apiClient.acceptHandoff).mockResolvedValue(mockHandoffResponse)

    const { result } = renderHook(useHandoffNotifications, {
      wrapper: createWrapper(),
    })

    // Simulate receiving a handoff notification
    act(() => {
      mockHandoffHandler(mockHandoffData)
    })

    // Now accept the handoff using the notification ID
    const notificationId = result.current.notifications[0]?.id

    await act(async () => {
      await result.current.acceptHandoff(notificationId!)
    })

    expect(apiClient.acceptHandoff).toHaveBeenCalledWith('test-session-id')
    expect(mockNavigate).toHaveBeenCalledWith('/chat/test-session-id')
  })

  it('should handle accept handoff error', async () => {
    const errorMessage = 'Session not found'
    const mockHandoffData = {
      session_id: 'test-session-id',
      customer_name: 'Test Customer',
      customer_email: 'test@example.com',
      handoff_reason: 'Need human help',
      urgency_level: 'high' as const,
      requested_at: '2024-01-01T00:00:00Z',
      session_metadata: {
        messages_count: 5,
        session_duration: 300,
        last_ai_response: 'Let me get a human agent for you'
      }
    }

    vi.mocked(apiClient.acceptHandoff).mockRejectedValue(new Error(errorMessage))

    const { result } = renderHook(useHandoffNotifications, {
      wrapper: createWrapper(),
    })

    // Simulate receiving a handoff notification
    act(() => {
      mockHandoffHandler(mockHandoffData)
    })

    const notificationId = result.current.notifications[0]?.id

    await act(async () => {
      try {
        await result.current.acceptHandoff(notificationId!)
      } catch (_error) {
        // Expected to throw
      }
    })

    await waitFor(() => {
      expect(result.current.error).toBe(errorMessage)
    })
  })

  it('should decline handoff successfully', async () => {
    const mockHandoffResponse = {
      success: true,
      session_id: 'test-session-id',
      agent_id: 'test-agent-id',
      tenant_id: 'test-tenant-id',
      message: 'Handoff declined successfully',
      declined_at: '2024-01-01T00:00:00Z'
    }

    const mockHandoffData = {
      session_id: 'test-session-id',
      customer_name: 'Test Customer',
      customer_email: 'test@example.com',
      handoff_reason: 'Need human help',
      urgency_level: 'high' as const,
      requested_at: '2024-01-01T00:00:00Z',
      session_metadata: {
        messages_count: 5,
        session_duration: 300,
        last_ai_response: 'Let me get a human agent for you'
      }
    }

    vi.mocked(apiClient.declineHandoff).mockResolvedValue(mockHandoffResponse)

    const { result } = renderHook(useHandoffNotifications, {
      wrapper: createWrapper(),
    })

    // Simulate receiving a handoff notification
    act(() => {
      mockHandoffHandler(mockHandoffData)
    })

    const notificationId = result.current.notifications[0]?.id

    await act(async () => {
      await result.current.declineHandoff(notificationId!)
    })

    expect(apiClient.declineHandoff).toHaveBeenCalledWith('test-session-id')
  })

  it('should handle decline handoff error', async () => {
    const errorMessage = 'Network error'
    const mockHandoffData = {
      session_id: 'test-session-id',
      customer_name: 'Test Customer',
      customer_email: 'test@example.com',
      handoff_reason: 'Need human help',
      urgency_level: 'high' as const,
      requested_at: '2024-01-01T00:00:00Z',
      session_metadata: {
        messages_count: 5,
        session_duration: 300,
        last_ai_response: 'Let me get a human agent for you'
      }
    }

    vi.mocked(apiClient.declineHandoff).mockRejectedValue(new Error(errorMessage))

    const { result } = renderHook(useHandoffNotifications, {
      wrapper: createWrapper(),
    })

    // Simulate receiving a handoff notification
    act(() => {
      mockHandoffHandler(mockHandoffData)
    })

    const notificationId = result.current.notifications[0]?.id

    await act(async () => {
      try {
        await result.current.declineHandoff(notificationId!)
      } catch (_error) {
        // Expected to throw
      }
    })

    await waitFor(() => {
      expect(result.current.error).toBe(errorMessage)
    })
  })

  it('should clear error', async () => {
    const mockHandoffData = {
      session_id: 'test-session-id',
      customer_name: 'Test Customer',
      customer_email: 'test@example.com',
      handoff_reason: 'Need human help',
      urgency_level: 'high' as const,
      requested_at: '2024-01-01T00:00:00Z',
      session_metadata: {
        messages_count: 5,
        session_duration: 300,
        last_ai_response: 'Let me get a human agent for you'
      }
    }

    const { result } = renderHook(useHandoffNotifications, {
      wrapper: createWrapper(),
    })

    // First trigger an error
    vi.mocked(apiClient.acceptHandoff).mockRejectedValue(new Error('Test error'))
    
    // Simulate receiving a handoff notification
    act(() => {
      mockHandoffHandler(mockHandoffData)
    })

    const notificationId = result.current.notifications[0]?.id

    await act(async () => {
      try {
        await result.current.acceptHandoff(notificationId!)
      } catch (_error) {
        // Expected to throw
      }
    })

    await waitFor(() => {
      expect(result.current.error).toBe('Test error')
    })

    act(() => {
      result.current.clearError()
    })

    expect(result.current.error).toBeNull()
  })

  it('should have required hook methods', () => {
    const { result } = renderHook(useHandoffNotifications, {
      wrapper: createWrapper(),
    })

    // Check that all required methods are available
    expect(typeof result.current.acceptHandoff).toBe('function')
    expect(typeof result.current.declineHandoff).toBe('function')
    expect(typeof result.current.markAsRead).toBe('function')
    expect(typeof result.current.dismissNotification).toBe('function')
    expect(typeof result.current.clearAll).toBe('function')
    expect(typeof result.current.clearError).toBe('function')
  })
})
