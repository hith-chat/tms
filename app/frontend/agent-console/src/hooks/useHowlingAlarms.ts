import { useState, useEffect, useCallback } from 'react'
import { apiClient } from '../lib/api'
import { useAgentWebSocket } from './useAgentWebSocket'
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
      const projectId = localStorage.getItem('project_id')
      if (!projectId) {
        throw new Error('No project selected')
      }

      const alarmsData = await apiClient.getActiveAlarms(projectId)
      setAlarms(alarmsData || [])
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load alarms'
      setError(errorMessage)
      console.error('Load alarms error:', err)
    }
  }, [])

  // Load alarm statistics
  const loadStats = useCallback(async () => {
    try {
      const projectId = localStorage.getItem('project_id')
      if (!projectId) {
        return // Stats are not critical, just skip if no project
      }

      const statsData = await apiClient.getAlarmStats(projectId)
      setStats(statsData)
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
      const projectId = localStorage.getItem('project_id')
      if (!projectId) {
        throw new Error('No project selected')
      }
      
      await apiClient.acknowledgeAlarm(projectId, alarmId, response)

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

  // Handle real-time alarm updates via WebSocket
  const handleAlarmMessage = useCallback((alarmData: any) => {
    console.log('Received alarm WebSocket message:', alarmData)
    
    switch (alarmData.type) {
      case 'alarm_triggered': {
        // Add new alarm to the list
        const newAlarm: HowlingAlarm = {
          id: alarmData.alarm_id,
          assignment_id: alarmData.assignment_id || '', // May need to be provided by backend
          agent_id: alarmData.agent_id || '', // May need to be provided by backend
          tenant_id: alarmData.tenant_id,
          project_id: alarmData.project_id,
          title: alarmData.title,
          message: alarmData.message,
          priority: alarmData.priority,
          current_level: alarmData.current_level,
          escalation_count: alarmData.escalation_count,
          created_at: alarmData.start_time || new Date().toISOString(),
          updated_at: alarmData.timestamp || new Date().toISOString(),
          escalation_interval: alarmData.config?.escalation_interval || 300, // 5 minutes default
          acknowledged_at: undefined,
          acknowledged_by: undefined,
          acknowledged_response: undefined,
          last_escalation_at: undefined
        }
        setAlarms(prev => {
          // Check if alarm already exists to avoid duplicates
          const exists = prev.some(alarm => alarm.id === newAlarm.id)
          if (exists) return prev
          return [...prev, newAlarm]
        })
        break
      }
        
      case 'alarm_escalated': {
        // Update existing alarm with escalation info
        setAlarms(prev => prev.map(alarm => 
          alarm.id === alarmData.alarm_id 
            ? { ...alarm, current_level: alarmData.current_level, escalation_count: alarmData.escalation_count }
            : alarm
        ))
        break
      }
        
      case 'alarm_acknowledged': {
        // Remove acknowledged alarm from the list
        setAlarms(prev => prev.filter(alarm => alarm.id !== alarmData.alarm_id))
        break
      }
    }
    
    // Refresh stats after any alarm change
    loadStats()
  }, [loadStats])

  // Set up WebSocket connection for real-time updates
  useAgentWebSocket({
    onAlarm: handleAlarmMessage,
    onError: (wsError: string) => {
      console.error('WebSocket error in alarm hook:', wsError)
      // Don't set error state for WebSocket issues as we have fallback polling
    }
  })

  // Initial load and fallback polling (reduced frequency since we have WebSocket)
  useEffect(() => {
    // Initial load
    refreshAlarms()

    // Set up fallback polling interval (every 30 seconds as backup to WebSocket)
    const interval = setInterval(() => {
      loadAlarms()
      loadStats()
    }, 30000)

    return () => clearInterval(interval)
  }, [refreshAlarms, loadAlarms, loadStats])

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
