import { useState, useEffect, useRef } from 'react'
import { AlertTriangle, Volume2, VolumeX, Check, Clock, TrendingUp, X } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
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
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  const audioRef = useRef<HTMLAudioElement>(null)
  const escalationSoundRef = useRef<HTMLAudioElement>(null)

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

  // Play alarm sounds based on escalation level
  useEffect(() => {
    if (soundEnabled && alarms.length > 0) {
      const criticalAlarms = alarms.filter(alarm => 
        alarm.current_level === 'critical' || alarm.current_level === 'urgent'
      )
      
      if (criticalAlarms.length > 0) {
        escalationSoundRef.current?.play()
      } else if (alarms.length > 0) {
        audioRef.current?.play()
      }
    }
  }, [alarms, soundEnabled])

  const loadAlarms = async () => {
    try {
      const response = await fetch('/api/v1/howling-alarms/active', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        }
      })
      
      if (response.ok) {
        const data = await response.json()
        setAlarms(data.alarms || [])
      }
    } catch (err) {
      console.error('Failed to load alarms:', err)
    }
  }

  const loadStats = async () => {
    try {
      const response = await fetch('/api/v1/howling-alarms/stats', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        }
      })
      
      if (response.ok) {
        const data = await response.json()
        setStats(data.stats)
      }
    } catch (err) {
      console.error('Failed to load alarm stats:', err)
    }
  }

  const acknowledgeAlarm = async (alarm: HowlingAlarm) => {
    if (!acknowledgmentText.trim()) {
      setError('Please provide an acknowledgment response')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const response = await fetch(`/api/v1/howling-alarms/${alarm.id}/acknowledge`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          response: acknowledgmentText
        })
      })

      if (response.ok) {
        setSelectedAlarm(null)
        setAcknowledgmentText('')
        loadAlarms() // Refresh the alarm list
        loadStats() // Refresh stats
      } else {
        const errorData = await response.json()
        setError(errorData.error || 'Failed to acknowledge alarm')
      }
    } catch (err) {
      setError('Network error occurred')
      console.error('Acknowledgment error:', err)
    } finally {
      setLoading(false)
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
    <div className={`bg-white rounded-lg shadow-lg border ${className}`}>
      {/* Audio elements for alarm sounds */}
      <audio ref={audioRef} preload="auto">
        <source src="/sounds/alarm-soft.mp3" type="audio/mpeg" />
      </audio>
      <audio ref={escalationSoundRef} preload="auto">
        <source src="/sounds/alarm-urgent.mp3" type="audio/mpeg" />
      </audio>

      {/* Header */}
      <div className="p-4 border-b border-gray-200 flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <AlertTriangle className="w-5 h-5 text-red-500" />
          <h2 className="text-lg font-semibold text-gray-900">
            Howling Alarms
            {alarms.length > 0 && (
              <span className="ml-2 px-2 py-1 bg-red-100 text-red-800 text-sm rounded-full">
                {alarms.length} Active
              </span>
            )}
          </h2>
        </div>
        
        <div className="flex items-center space-x-2">
          <button
            onClick={() => setSoundEnabled(!soundEnabled)}
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

      {/* Alarm Statistics */}
      {stats && (
        <div className="p-4 bg-gray-50 border-b border-gray-200">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-900">{stats.total_active}</div>
              <div className="text-gray-600">Active Alarms</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {stats.average_duration.toFixed(1)}s
              </div>
              <div className="text-gray-600">Avg Duration</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-orange-600">
                {Object.values(stats.by_level).reduce((a, b) => Math.max(a, b), 0)}
              </div>
              <div className="text-gray-600">Peak Level</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-red-600">
                {stats.escalation_counts['3_escalations'] || 0}
              </div>
              <div className="text-gray-600">High Escalations</div>
            </div>
          </div>
        </div>
      )}

      {/* Active Alarms List */}
      <div className="max-h-96 overflow-y-auto">
        {alarms.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            <AlertTriangle className="w-12 h-12 mx-auto mb-3 text-gray-300" />
            <p className="text-lg font-medium">No Active Alarms</p>
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
              
              {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                  <p className="text-sm text-red-600">{error}</p>
                </div>
              )}
              
              <div className="flex space-x-3">
                <button
                  onClick={() => acknowledgeAlarm(selectedAlarm)}
                  disabled={loading || !acknowledgmentText.trim()}
                  className="flex-1 flex items-center justify-center space-x-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Check className="w-4 h-4" />
                  <span>{loading ? 'Acknowledging...' : 'Acknowledge'}</span>
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
