import { useState, useEffect, useRef } from 'react'
import { AlertTriangle, Volume2, VolumeX, Check, Clock, TrendingUp, X } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { apiClient } from '../lib/api'
import type { HowlingAlarm, AlarmStats } from '../types/notifications'

interface HowlingAlarmPanelProps {
  className?: string
}

export function HowlingAlarmPanel({ className = '' }: HowlingAlarmPanelProps) {
  const [alarms, setAlarms] = useState<HowlingAlarm[]>([])
  const [stats, setStats] = useState<AlarmStats | null>(null)
  const [soundEnabled, setSoundEnabled] = useState(true)
  const [selectedAlarm, setSelectedAlarm] = useState<HowlingAlarm | null>(null)
  const [acknowledgmentText, setAcknowledgmentText] = useState('')
  const [_loading, _setLoading] = useState(false)
  const [_error, _setError] = useState<string | null>(null)
  const [audioInitialized, setAudioInitialized] = useState(false)
  const [previousAlarmCount, setPreviousAlarmCount] = useState(0)
  
  const audioRef = useRef<HTMLAudioElement>(null)
  const escalationSoundRef = useRef<HTMLAudioElement>(null)

  // Initialize audio on first user interaction
  const initializeAudio = async () => {
    if (audioInitialized) return
    
    try {
      // Try to load and prepare audio elements
      if (audioRef.current) {
        audioRef.current.volume = 0.1 // Set a reasonable volume
        await audioRef.current.load()
      }
      if (escalationSoundRef.current) {
        escalationSoundRef.current.volume = 0.3 // Urgent sounds louder
        await escalationSoundRef.current.load()
      }
      setAudioInitialized(true)
      console.log('Audio initialized successfully')
    } catch (error) {
      console.error('Failed to initialize audio:', error)
    }
  }

  // Fallback function to create beep using Web Audio API
  const createBeep = (frequency: number, duration: number, volume: number = 0.1) => {
    try {
      const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
      const oscillator = audioContext.createOscillator()
      const gainNode = audioContext.createGain()
      
      oscillator.connect(gainNode)
      gainNode.connect(audioContext.destination)
      
      oscillator.frequency.setValueAtTime(frequency, audioContext.currentTime)
      oscillator.type = 'square' // Square wave for more alarm-like sound
      
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
  }

  // Load alarms and stats
  useEffect(() => {
    loadAlarms()
    loadStats()
    
    // Set up polling for real-time updates
    const interval = setInterval(() => {
      loadAlarms()
      loadStats()
    }, 5000) // Poll every 5 seconds
    
    return () => clearInterval(interval)
  }, [])

  // Auto-initialize audio when sound is enabled
  useEffect(() => {
    if (soundEnabled && !audioInitialized) {
      initializeAudio()
    }
  }, [soundEnabled, audioInitialized])

  // Play alarm sounds based on escalation level and detect new alarms
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
        
        let audioElement: HTMLAudioElement | null = null
        let isCritical = false
        
        if (criticalAlarms.length > 0) {
          audioElement = escalationSoundRef.current
          isCritical = true
        } else if (alarms.length > 0) {
          audioElement = audioRef.current
          isCritical = false
        }
        
        if (audioElement) {
          // Reset audio to beginning
          audioElement.currentTime = 0
          
          // Try to play the audio
          const playPromise = audioElement.play()
          if (playPromise !== undefined) {
            playPromise
              .then(() => {
                console.log('Alarm sound playing successfully')
              })
              .catch((error) => {
                console.warn('Audio file failed to play, using fallback beep:', error)
                // Use fallback beep if audio file fails
                if (isCritical) {
                  createBeep(1200, 1.5, 0.2) // Urgent beep: 1200Hz, 1.5s
                } else {
                  createBeep(800, 1.0, 0.1)  // Soft beep: 800Hz, 1s
                }
              })
          }
        } else {
          // If no audio element, use fallback beep directly
          console.log('No audio element found, using fallback beep')
          if (isCritical) {
            createBeep(1200, 1.5, 0.2)
          } else {
            createBeep(800, 1.0, 0.1)
          }
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
  }, [alarms, soundEnabled, audioInitialized, previousAlarmCount])

  const loadAlarms = async () => {
    try {
      const projectId = localStorage.getItem('project_id')
      if (!projectId) {
        console.error('No project selected')
        return
      }

      const alarms = await apiClient.getActiveAlarms(projectId)
      console.log('Loaded alarms:', alarms?.length || 0, 'alarms', alarms)
      setAlarms(alarms || [])
    } catch (err) {
      console.error('Failed to load alarms:', err)
    }
  }

  const loadStats = async () => {
    try {
      const projectId = localStorage.getItem('project_id')
      if (!projectId) {
        console.error('No project selected')
        return
      }

      const stats = await apiClient.getAlarmStats(projectId)
      setStats(stats)
    } catch (err) {
      console.error('Failed to load alarm stats:', err)
    }
  }

  const acknowledgeAlarm = async (alarmId: string) => {
    try {
      const projectId = localStorage.getItem('project_id')
      if (!projectId) {
        console.error('No project selected')
        return
      }

      await apiClient.acknowledgeAlarm(projectId, alarmId, acknowledgmentText)
      await loadAlarms()
      await loadStats()
      setSelectedAlarm(null)
      setAcknowledgmentText('')
    } catch (err) {
      console.error('Failed to acknowledge alarm:', err)
    }
  }

  const getAlarmLevelColor = (level: HowlingAlarm['current_level']) => {
    switch (level) {
      case 'soft': return 'text-blue-600 bg-blue-50 border-blue-200'
      case 'medium': return 'text-yellow-600 bg-yellow-50 border-yellow-200'
      case 'loud': return 'text-orange-600 bg-orange-50 border-orange-200'
      case 'urgent': return 'text-red-600 bg-red-50 border-red-200'
      case 'critical': return 'text-red-800 bg-red-100 border-red-300 animate-pulse'
      default: return 'text-gray-600 bg-gray-50 border-gray-200'
    }
  }

  const getPriorityColor = (priority: HowlingAlarm['priority']) => {
    switch (priority) {
      case 'normal': return 'text-green-600'
      case 'high': return 'text-yellow-600'
      case 'urgent': return 'text-orange-600'
      case 'critical': return 'text-red-600'
      default: return 'text-gray-600'
    }
  }

  return (
    <div className={`bg-card border border-border rounded-lg shadow-sm ${className}`}>
      {/* Audio elements for alarm sounds */}
      <audio 
        ref={audioRef} 
        preload="auto"
        onError={(e) => console.error('Error loading alarm sound:', e)}
        onLoadedData={() => console.log('Alarm sound loaded successfully')}
      >
        <source src="/sounds/alarm-soft.mp3" type="audio/mpeg" />
        <source src="data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmAaBT+y2+/CdSIELI/X8tOCPAYZaq3w45ZNDwlPoOHyvmAaBT+y2+/CdSIELI/X8tOCPAYZaq3w45ZNDwlPoOHyvmAaBT+y2+/CdSIELI/X8tOCPAYZaq3w45ZNDwlPoOHyvmAaBT+y2+/CdSIELI/X8tOCPAYZaq3w45ZNDwlPjrDbqFsbDT6K1fHCjDAFHm7A7+OZQT+a2/PEhbQBLY/Q8tyPOgUYZrPn4KMOBQlGn+DiYhQELI/X8tOCPAYZaq3w45ZNDwlPhZv3dSIELI/X8tOCPAYZaq3w45ZNDwlPhZvs6qCTBjme3PbEgkAFGnLD7dqJOgUYZrPn4KMOBQlGn+DzAJJgkl3S8tKMPwcYZbDc6pbWET66yO7FgkIFGnLC7duJOAUYZbDc6pbWET66yO7FgkIFGnLC7duJOAUYZbDc6pbWET66yO7FgkIFGnLC7duJOAUYZbDc6pbWET66yO7FgkIFGnLC7duJOAQYA" type="audio/wav" />
      </audio>
      <audio 
        ref={escalationSoundRef} 
        preload="auto"
        onError={(e) => console.error('Error loading escalation sound:', e)}
        onLoadedData={() => console.log('Escalation sound loaded successfully')}
      >
        <source src="/sounds/alarm-urgent.mp3" type="audio/mpeg" />
        <source src="data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmAaBT+y2+/CdSIELI/X8tOCPAYZaq3w45ZNDwlPoOHyvmAaBT+y2+/CdSIELI/X8tOCPAYZaq3w45ZNDwlPoOHyvmAaBT+y2+/CdSIELI/X8tOCPAYZaq3w45ZNDwlPoOHyvmAaBT+y2+/CdSIELI/X8tOCPAYZaq3w45ZNDwlPjrDbqFsbDT6K1fHCjDAFHm7A7+OZQT+a2/PEhbQBLY/Q8tyPOgUYZrPn4KMOBQlGn+DiYhQELI/X8tOCPAYZaq3w45ZNDwlPhZv3dSIELI/X8tOCPAYZaq3w45ZNDwlPhZvs6qCTBjme3PbEgkAFGnLD7dqJOgUYZrPn4KMOBQlGn+DzAJJgkl3S8tKMPwcYZbDc6pbWET66yO7FgkIFGnLC7duJOAUYZbDc6pbWET66yO7FgkIFGnLC7duJOAUYZbDc6pbWET66yO7FgkIFGnLC7duJOAUYZbDc6pbWET66yO7FgkIFGnLC7duJOAQYA" type="audio/wav" />
      </audio>

      {/* Header */}
      <div className="p-6 border-b border-border flex items-center justify-between bg-card/50">
        <div className="flex items-center space-x-2">
          <AlertTriangle className="w-5 h-5 text-red-500" />
          <h2 className="text-lg font-semibold text-foreground">
            Active Alerts
            {alarms.length > 0 && (
              <span className="ml-2 px-2 py-1 bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-200 text-sm rounded-full">
                {alarms.length} Active
              </span>
            )}
          </h2>
        </div>        <div className="flex items-center space-x-2">
          {!audioInitialized && soundEnabled && (
            <div className="text-xs text-amber-600 bg-amber-50 px-2 py-1 rounded-md">
              Click sound button to enable audio
            </div>
          )}
          <button
            onClick={async () => {
              if (!audioInitialized) {
                await initializeAudio()
              }
              setSoundEnabled(!soundEnabled)
            }}
            className={`p-2 rounded-lg transition-colors ${
              soundEnabled 
                ? 'text-blue-600 bg-blue-50 hover:bg-blue-100' 
                : 'text-gray-400 bg-gray-50 hover:bg-gray-100'
            }`}
            title={soundEnabled ? 'Disable alarm sounds' : 'Enable alarm sounds'}
          >
            {soundEnabled ? <Volume2 className="w-4 h-4" /> : <VolumeX className="w-4 h-4" />}
          </button>
        </div>
      </div>

      {/* Alert Statistics */}
      {stats && (
        <div className="p-6 bg-muted/50 border-b border-border">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div className="text-center">
              <div className="text-2xl font-bold text-foreground">{stats?.total_active || 0}</div>
              <div className="text-muted-foreground">Active Alerts</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {(stats?.average_duration || 0).toFixed(1)}s
              </div>
              <div className="text-muted-foreground">Avg Duration</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-orange-600">
                {stats?.by_level ? Object.values(stats.by_level).reduce((a, b) => Math.max(a, b), 0) : 0}
              </div>
              <div className="text-muted-foreground">Peak Level</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-red-600">
                {stats?.escalation_counts?.['3_escalations'] || 0}
              </div>
              <div className="text-muted-foreground">High Escalations</div>
            </div>
          </div>
        </div>
      )}

      {/* Active Alerts List */}
      <div className="max-h-[32rem] overflow-y-auto">
        {alarms.length === 0 ? (
          <div className="p-8 text-center text-muted-foreground">
            <AlertTriangle className="w-12 h-12 mx-auto mb-3 text-muted-foreground/50" />
            <p className="text-lg font-medium">No Active Alerts</p>
            <p className="text-sm">All systems are operating normally</p>
          </div>
        ) : (
          <div className="space-y-2 p-4">
            {alarms.map((alarm) => (
              <div
                key={alarm.id}
                className={`p-4 rounded-lg border-2 transition-all hover:shadow-md cursor-pointer ${getAlarmLevelColor(alarm.current_level)}`}
                onClick={() => setSelectedAlarm(alarm)}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-2">
                      <span className="text-lg font-semibold">{alarm.title}</span>
                      <span className={`px-2 py-1 text-xs font-medium rounded-full ${getPriorityColor(alarm.priority)} bg-current bg-opacity-10`}>
                        {alarm.priority.toUpperCase()}
                      </span>
                      <span className="px-2 py-1 text-xs font-medium rounded-full bg-current bg-opacity-20">
                        Level: {alarm.current_level.toUpperCase()}
                      </span>
                    </div>
                    
                    <p className="text-sm mb-2">{alarm.message}</p>
                    
                    <div className="flex items-center space-x-4 text-xs text-gray-600">
                      <div className="flex items-center space-x-1">
                        <Clock className="w-3 h-3" />
                        <span>{formatDistanceToNow(new Date(alarm.created_at), { addSuffix: true })}</span>
                      </div>
                      
                      {alarm.escalation_count > 0 && (
                        <div className="flex items-center space-x-1">
                          <TrendingUp className="w-3 h-3" />
                          <span>{alarm.escalation_count} escalations</span>
                        </div>
                      )}
                    </div>
                  </div>
                  
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      setSelectedAlarm(alarm)
                    }}
                    className="ml-4 px-3 py-1 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700 transition-colors"
                  >
                    Acknowledge
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Acknowledgment Modal */}
      {selectedAlarm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
            <div className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900">
                  Acknowledge Alarm
                </h3>
                <button
                  onClick={() => setSelectedAlarm(null)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>
              
              <div className="mb-4 p-3 bg-gray-50 rounded-lg">
                <h4 className="font-medium text-gray-900">{selectedAlarm.title}</h4>
                <p className="text-sm text-gray-600 mt-1">{selectedAlarm.message}</p>
              </div>
              
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Acknowledgment Response
                </label>
                <textarea
                  value={acknowledgmentText}
                  onChange={(e) => setAcknowledgmentText(e.target.value)}
                  placeholder="Describe what action you're taking to address this alarm..."
                  className="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  rows={3}
                />
              </div>
              
              {_error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                  <p className="text-sm text-red-600">{_error}</p>
                </div>
              )}
              
              <div className="flex space-x-3">
                <button
                  onClick={() => acknowledgeAlarm(selectedAlarm.id)}
                  disabled={_loading || !acknowledgmentText.trim()}
                  className="flex-1 flex items-center justify-center space-x-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Check className="w-4 h-4" />
                  <span>{_loading ? 'Acknowledging...' : 'Acknowledge'}</span>
                </button>
                
                <button
                  onClick={() => setSelectedAlarm(null)}
                  className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
