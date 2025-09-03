import { useState, useEffect } from 'react'
import { AlertTriangle, X, Volume2, VolumeX, Clock } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import type { HowlingAlarm } from '../types/notifications'

interface FloatingAlarmWidgetProps {
  alarms: HowlingAlarm[]
  onAcknowledge: (alarmId: string, response: string) => Promise<boolean>
  soundEnabled: boolean
  onToggleSound: () => void | Promise<void>
  className?: string
}

export function FloatingAlarmWidget({ 
  alarms, 
  onAcknowledge, 
  soundEnabled, 
  onToggleSound,
  className = '' 
}: FloatingAlarmWidgetProps) {
  const [isMinimized, setIsMinimized] = useState(false)
  const [currentAlarmIndex, setCurrentAlarmIndex] = useState(0)
  const [acknowledgmentText, setAcknowledgmentText] = useState('')
  const [showAckModal, setShowAckModal] = useState(false)
  const [selectedAlarm, setSelectedAlarm] = useState<HowlingAlarm | null>(null)

  // Auto-cycle through alarms if there are multiple
  useEffect(() => {
    if (alarms.length > 1) {
      const interval = setInterval(() => {
        setCurrentAlarmIndex(prev => (prev + 1) % alarms.length)
      }, 3000) // Switch every 3 seconds
      
      return () => clearInterval(interval)
    }
  }, [alarms.length])

  // Reset index when alarms change
  useEffect(() => {
    if (currentAlarmIndex >= alarms.length) {
      setCurrentAlarmIndex(0)
    }
  }, [alarms.length, currentAlarmIndex])

  if (alarms.length === 0) {
    return null
  }

  const currentAlarm = alarms[currentAlarmIndex]
  
  const getAlarmAnimation = (level: HowlingAlarm['current_level']) => {
    switch (level) {
      case 'soft': return 'animate-pulse'
      case 'medium': return 'animate-bounce'
      case 'loud': return 'animate-pulse'
      case 'urgent': return 'animate-bounce'
      case 'critical': return 'animate-pulse animate-bounce'
      default: return ''
    }
  }

  const getAlarmColors = (level: HowlingAlarm['current_level']) => {
    switch (level) {
      case 'soft': return 'bg-blue-500 border-blue-600'
      case 'medium': return 'bg-yellow-500 border-yellow-600'
      case 'loud': return 'bg-orange-500 border-orange-600'
      case 'urgent': return 'bg-red-500 border-red-600'
      case 'critical': return 'bg-red-600 border-red-700'
      default: return 'bg-gray-500 border-gray-600'
    }
  }

  const handleAcknowledge = (alarm: HowlingAlarm) => {
    setSelectedAlarm(alarm)
    setShowAckModal(true)
  }

  const submitAcknowledgment = async () => {
    if (!selectedAlarm || !acknowledgmentText.trim()) return

    const success = await onAcknowledge(selectedAlarm.id, acknowledgmentText)
    if (success) {
      setShowAckModal(false)
      setSelectedAlarm(null)
      setAcknowledgmentText('')
    }
  }

  if (isMinimized) {
    return (
      <div className={`fixed bottom-4 right-4 z-50 ${className}`}>
        <button
          onClick={() => setIsMinimized(false)}
          className={`p-3 rounded-full text-white shadow-lg hover:shadow-xl transition-all ${getAlarmColors(currentAlarm.current_level)} ${getAlarmAnimation(currentAlarm.current_level)}`}
        >
          <AlertTriangle className="w-6 h-6" />
          {alarms.length > 1 && (
            <span className="absolute -top-2 -right-2 bg-white text-red-600 text-xs font-bold rounded-full w-6 h-6 flex items-center justify-center">
              {alarms.length}
            </span>
          )}
        </button>
      </div>
    )
  }

  return (
    <>
      <div className={`fixed bottom-4 right-4 z-50 ${className}`}>
        <div className={`bg-white rounded-lg shadow-xl border-2 max-w-sm ${getAlarmColors(currentAlarm.current_level)} ${getAlarmAnimation(currentAlarm.current_level)}`}>
          {/* Header */}
          <div className={`p-4 text-white rounded-t-lg ${getAlarmColors(currentAlarm.current_level)}`}>
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <AlertTriangle className="w-5 h-5" />
                <span className="font-semibold text-sm">
                  HOWLING ALARM
                  {alarms.length > 1 && ` (${currentAlarmIndex + 1}/${alarms.length})`}
                </span>
              </div>
              
              <div className="flex items-center space-x-2">
                <button
                  onClick={onToggleSound}
                  className="p-1 rounded hover:bg-black hover:bg-opacity-20 transition-colors"
                  title={soundEnabled ? 'Disable alarm sounds' : 'Enable alarm sounds'}
                >
                  {soundEnabled ? <Volume2 className="w-4 h-4" /> : <VolumeX className="w-4 h-4" />}
                </button>
                
                <button
                  onClick={() => setIsMinimized(true)}
                  className="p-1 rounded hover:bg-black hover:bg-opacity-20 transition-colors"
                  title="Minimize alarm widget"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>

          {/* Alarm Content */}
          <div className="p-4 bg-white">
            <div className="mb-3">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold text-gray-900">{currentAlarm.title}</h3>
                <span className={`px-2 py-1 text-xs font-medium rounded-full text-white ${getAlarmColors(currentAlarm.current_level)}`}>
                  {currentAlarm.current_level.toUpperCase()}
                </span>
              </div>
              
              <p className="text-sm text-gray-700 mb-2">{currentAlarm.message}</p>
              
              <div className="flex items-center justify-between text-xs text-gray-500">
                <div className="flex items-center space-x-1">
                  <Clock className="w-3 h-3" />
                  <span>{formatDistanceToNow(new Date(currentAlarm.created_at), { addSuffix: true })}</span>
                </div>
                
                {currentAlarm.escalation_count > 0 && (
                  <span className="px-2 py-1 bg-orange-100 text-orange-600 rounded-full">
                    {currentAlarm.escalation_count} escalations
                  </span>
                )}
              </div>
            </div>

            {/* Quick Acknowledge Button */}
            <button
              onClick={() => handleAcknowledge(currentAlarm)}
              className="w-full bg-green-600 text-white py-2 px-4 rounded-lg hover:bg-green-700 transition-colors text-sm font-medium"
            >
              Acknowledge Alarm
            </button>
          </div>

          {/* Multiple Alarms Indicator */}
          {alarms.length > 1 && (
            <div className="px-4 pb-4">
              <div className="flex space-x-1">
                {alarms.map((_, index) => (
                  <button
                    key={index}
                    onClick={() => setCurrentAlarmIndex(index)}
                    className={`w-2 h-2 rounded-full transition-colors ${
                      index === currentAlarmIndex 
                        ? getAlarmColors(alarms[index].current_level) 
                        : 'bg-gray-300'
                    }`}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Acknowledgment Modal */}
      {showAckModal && selectedAlarm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-60 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
            <div className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900">
                  Acknowledge Alarm
                </h3>
                <button
                  onClick={() => setShowAckModal(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>
              
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                <h4 className="font-medium text-red-900">{selectedAlarm.title}</h4>
                <p className="text-sm text-red-700 mt-1">{selectedAlarm.message}</p>
                <p className="text-xs text-red-600 mt-2">
                  Level: {selectedAlarm.current_level.toUpperCase()} | Priority: {selectedAlarm.priority.toUpperCase()}
                </p>
              </div>
              
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Acknowledgment Response *
                </label>
                <textarea
                  value={acknowledgmentText}
                  onChange={(e) => setAcknowledgmentText(e.target.value)}
                  placeholder="Describe what action you're taking to address this alarm..."
                  className="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  rows={3}
                  required
                />
              </div>
              
              <div className="flex space-x-3">
                <button
                  onClick={submitAcknowledgment}
                  disabled={!acknowledgmentText.trim()}
                  className="flex-1 bg-green-600 text-white py-2 px-4 rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  Acknowledge
                </button>
                
                <button
                  onClick={() => setShowAckModal(false)}
                  className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
