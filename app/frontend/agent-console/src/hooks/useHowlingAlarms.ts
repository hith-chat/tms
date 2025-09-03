import { useState, useEffect, useCallback, useRef } from 'react'
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
  setSoundEnabled: (enabled: boolean) => Promise<void>
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
  const [audioInitialized, setAudioInitialized] = useState(false)
  const [previousAlarmCount, setPreviousAlarmCount] = useState(0)

  const audioRef = useRef<HTMLAudioElement | null>(null)
  const escalationSoundRef = useRef<HTMLAudioElement | null>(null)

  // Initialize audio elements
  useEffect(() => {
    if (!audioRef.current) {
      audioRef.current = new Audio()
      audioRef.current.preload = 'auto'
      audioRef.current.volume = 0.1
      audioRef.current.src = '/sounds/alarm-soft.mp3'
    }
    
    if (!escalationSoundRef.current) {
      escalationSoundRef.current = new Audio()
      escalationSoundRef.current.preload = 'auto'
      escalationSoundRef.current.volume = 0.3
      escalationSoundRef.current.src = '/sounds/alarm-urgent.mp3'
    }
  }, [])

  // Initialize audio on first user interaction
  const initializeAudio = useCallback(async () => {
    if (audioInitialized) return
    
    try {
      if (audioRef.current) {
        await audioRef.current.load()
      }
      if (escalationSoundRef.current) {
        await escalationSoundRef.current.load()
      }
      setAudioInitialized(true)
      console.log('Alarm audio initialized successfully')
    } catch (error) {
      console.error('Failed to initialize alarm audio:', error)
    }
  }, [audioInitialized])

  // Fallback function to create beep using Web Audio API
  const createBeep = useCallback((frequency: number, duration: number, volume: number = 0.1) => {
    try {
      const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
      const oscillator = audioContext.createOscillator()
      const gainNode = audioContext.createGain()
      
      oscillator.connect(gainNode)
      gainNode.connect(audioContext.destination)
      
      oscillator.frequency.setValueAtTime(frequency, audioContext.currentTime)
      oscillator.type = 'square'
      
      // Create beeping pattern with envelope
      gainNode.gain.setValueAtTime(0, audioContext.currentTime)
      gainNode.gain.linearRampToValueAtTime(volume, audioContext.currentTime + 0.01)
      gainNode.gain.linearRampToValueAtTime(volume, audioContext.currentTime + 0.1)
      gainNode.gain.linearRampToValueAtTime(0, audioContext.currentTime + 0.2)
      gainNode.gain.linearRampToValueAtTime(volume, audioContext.currentTime + 0.3)
      gainNode.gain.linearRampToValueAtTime(volume, audioContext.currentTime + 0.4)
      gainNode.gain.linearRampToValueAtTime(0, audioContext.currentTime + 0.5)
      gainNode.gain.linearRampToValueAtTime(0, audioContext.currentTime + duration)
      
      oscillator.start(audioContext.currentTime)
      oscillator.stop(audioContext.currentTime + duration)
      
      console.log(`Playing fallback beep at ${frequency}Hz`)
    } catch (error) {
      console.error('Failed to create fallback beep:', error)
    }
  }, [])

  // Play alarm sound when new alarms are detected
  useEffect(() => {
    const playAlarmSound = async () => {
      // Don't play sound if disabled or no alarms
      if (!soundEnabled || alarms.length === 0) {
        setPreviousAlarmCount(alarms.length)
        return
      }
      
      // Auto-initialize audio if not done yet
      if (!audioInitialized) {
        await initializeAudio()
      }
      
      // Check if we have new alarms or if this is the first load with alarms
      const shouldPlaySound = alarms.length > previousAlarmCount || (previousAlarmCount === 0 && alarms.length > 0)
      
      if (!shouldPlaySound) {
        setPreviousAlarmCount(alarms.length)
        return
      }
      
      console.log(`Playing alarm sound for ${alarms.length} alarms (was ${previousAlarmCount})`)
      
      try {
        const criticalAlarms = alarms.filter(alarm => 
          alarm.current_level === 'critical' || alarm.current_level === 'urgent'
        )
        
        // Always use fallback beep for better reliability
        if (criticalAlarms.length > 0) {
          console.log('Playing urgent alarm sound')
          createBeep(1200, 1.5, 0.2) // Urgent beep: 1200Hz, 1.5s
        } else if (alarms.length > 0) {
          console.log('Playing normal alarm sound')
          createBeep(800, 1.0, 0.1)  // Soft beep: 800Hz, 1s
        }
      } catch (error) {
        console.error('Error playing alarm sound:', error)
        // Fallback to generated beep
        const isCritical = alarms.some(alarm => 
          alarm.current_level === 'critical' || alarm.current_level === 'urgent'
        )
        if (isCritical) {
          createBeep(1200, 1.5, 0.2)
        } else {
          createBeep(800, 1.0, 0.1)
        }
      }
      
      // Update the previous count
      setPreviousAlarmCount(alarms.length)
    }
    
    playAlarmSound()
  }, [alarms, soundEnabled, audioInitialized, previousAlarmCount, initializeAudio, createBeep])

  // Auto-initialize audio when sound is enabled
  useEffect(() => {
    if (soundEnabled && !audioInitialized) {
      // Add a click listener to initialize audio on first user interaction
      const initializeOnInteraction = async () => {
        await initializeAudio()
        document.removeEventListener('click', initializeOnInteraction)
        document.removeEventListener('keydown', initializeOnInteraction)
      }
      
      document.addEventListener('click', initializeOnInteraction)
      document.addEventListener('keydown', initializeOnInteraction)
      
      return () => {
        document.removeEventListener('click', initializeOnInteraction)
        document.removeEventListener('keydown', initializeOnInteraction)
      }
    }
  }, [soundEnabled, audioInitialized, initializeAudio])

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
          const newAlarms = [...prev, newAlarm]
          console.log('New alarm added via WebSocket, total alarms:', newAlarms.length)
          return newAlarms
        })
        break
      }
        
      case 'alarm_escalated': {
        // Update existing alarm with escalation info
        setAlarms(prev => {
          const updated = prev.map(alarm => 
            alarm.id === alarmData.alarm_id 
              ? { ...alarm, current_level: alarmData.current_level, escalation_count: alarmData.escalation_count }
              : alarm
          )
          console.log('Alarm escalated via WebSocket:', alarmData.alarm_id)
          return updated
        })
        break
      }
        
      case 'alarm_acknowledged': {
        // Remove acknowledged alarm from the list
        setAlarms(prev => {
          const filtered = prev.filter(alarm => alarm.id !== alarmData.alarm_id)
          console.log('Alarm acknowledged via WebSocket, remaining alarms:', filtered.length)
          return filtered
        })
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
    setSoundEnabled: useCallback(async (enabled: boolean) => {
      setSoundEnabled(enabled)
      if (enabled && !audioInitialized) {
        await initializeAudio()
      }
    }, [audioInitialized, initializeAudio])
  }
}
