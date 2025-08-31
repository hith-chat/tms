import { useState, useEffect, useCallback } from 'react'
import type { HowlingAlarm, AlarmStats } from '../types/notifications'

export interface UseHowlingAlarmsResult {
  alarms: HowlingAlarm[]
  stats: AlarmStats | null
  loading: boolean
  error: string | null
  refreshAlarms: () => Promise<void>
  acknowledgeAlarm: (alarmId: string, response: string) => Promise<boolean>
  soundEnabled: boolean
  setSoundEnabled: (enabled: boolean) => void
}

export function useHowlingAlarms(): UseHowlingAlarmsResult {
  const [alarms, setAlarms] = useState<HowlingAlarm[]>([])
  const [stats, setStats] = useState<AlarmStats | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [soundEnabled, setSoundEnabled] = useState(() => {
    const saved = localStorage.getItem('alarm-sound-enabled')
    return saved ? JSON.parse(saved) : true
  })

  // Save sound preference to localStorage
  useEffect(() => {
    localStorage.setItem('alarm-sound-enabled', JSON.stringify(soundEnabled))
  }, [soundEnabled])

  // Load active alarms
  const loadAlarms = useCallback(async () => {
    try {
      setError(null)
      const response = await fetch('/api/v1/howling-alarms/active', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        throw new Error(`Failed to load alarms: ${response.statusText}`)
      }

      const data = await response.json()
      setAlarms(data.alarms || [])
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load alarms'
      setError(errorMessage)
      console.error('Load alarms error:', err)
    }
  }, [])

  // Load alarm statistics
  const loadStats = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/howling-alarms/stats', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        throw new Error(`Failed to load stats: ${response.statusText}`)
      }

      const data = await response.json()
      setStats(data.stats)
    } catch (err) {
      console.error('Load stats error:', err)
      // Don't set error for stats since it's not critical
    }
  }, [])

  // Refresh both alarms and stats
  const refreshAlarms = useCallback(async () => {
    setLoading(true)
    setError(null)
    
    try {
      await Promise.all([loadAlarms(), loadStats()])
    } finally {
      setLoading(false)
    }
  }, [loadAlarms, loadStats])

  // Acknowledge an alarm
  const acknowledgeAlarm = useCallback(async (alarmId: string, response: string): Promise<boolean> => {
    try {
      setError(null)
      
      const apiResponse = await fetch(`/api/v1/howling-alarms/${alarmId}/acknowledge`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ response })
      })

      if (!apiResponse.ok) {
        const errorData = await apiResponse.json()
        throw new Error(errorData.error || `Failed to acknowledge alarm: ${apiResponse.statusText}`)
      }

      // Remove the acknowledged alarm from the local state
      setAlarms(prev => prev.filter(alarm => alarm.id !== alarmId))
      
      // Refresh stats
      await loadStats()
      
      return true
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to acknowledge alarm'
      setError(errorMessage)
      console.error('Acknowledge alarm error:', err)
      return false
    }
  }, [loadStats])

  // Set up real-time polling
  useEffect(() => {
    // Initial load
    refreshAlarms()

    // Set up polling interval (every 5 seconds)
    const interval = setInterval(() => {
      loadAlarms()
      loadStats()
    }, 5000)

    return () => clearInterval(interval)
  }, [refreshAlarms, loadAlarms, loadStats])

  // WebSocket integration (if available)
  useEffect(() => {
    // Try to connect to WebSocket for real-time updates
    const token = localStorage.getItem('token')
    if (!token) return

    try {
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${wsProtocol}//${window.location.host}/api/v1/ws/alarms?token=${token}`
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        console.log('Connected to alarm WebSocket')
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          
          switch (data.type) {
            case 'alarm_triggered':
              setAlarms(prev => [...prev, data.alarm])
              break
              
            case 'alarm_escalated':
              setAlarms(prev => prev.map(alarm => 
                alarm.id === data.alarm.id ? data.alarm : alarm
              ))
              break
              
            case 'alarm_acknowledged':
              setAlarms(prev => prev.filter(alarm => alarm.id !== data.alarm_id))
              break
              
            case 'stats_updated':
              setStats(data.stats)
              break
          }
        } catch (err) {
          console.error('WebSocket message parse error:', err)
        }
      }

      ws.onclose = () => {
        console.log('Disconnected from alarm WebSocket')
      }

      ws.onerror = (error) => {
        console.error('Alarm WebSocket error:', error)
      }

      return () => {
        ws.close()
      }
    } catch (err) {
      console.error('WebSocket connection error:', err)
    }
  }, [])

  return {
    alarms,
    stats,
    loading,
    error,
    refreshAlarms,
    acknowledgeAlarm,
    soundEnabled,
    setSoundEnabled
  }
}
