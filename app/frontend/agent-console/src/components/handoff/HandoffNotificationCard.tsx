import React from 'react'
import { HandoffNotification } from '../../types/handoff'

interface HandoffNotificationCardProps {
  notification: HandoffNotification
  onAccept: (notificationId: string) => Promise<void>
  onDecline: (notificationId: string) => Promise<void>
  onMarkRead: (notificationId: string) => void
  onDismiss: (notificationId: string) => void
}

export function HandoffNotificationCard({
  notification,
  onAccept,
  onDecline,
  onMarkRead,
  onDismiss
}: HandoffNotificationCardProps) {
  const [isAccepting, setIsAccepting] = React.useState(false)
  const [isDeclining, setIsDeclining] = React.useState(false)

  const handleAccept = async () => {
    setIsAccepting(true)
    try {
      await onAccept(notification.id)
    } finally {
      setIsAccepting(false)
    }
  }

  const handleDecline = async () => {
    setIsDeclining(true)
    try {
      await onDecline(notification.id)
    } finally {
      setIsDeclining(false)
    }
  }

  const getUrgencyColor = (urgency: string) => {
    switch (urgency) {
      case 'critical': return 'bg-red-100 border-red-300 text-red-800'
      case 'high': return 'bg-orange-100 border-orange-300 text-orange-800'
      case 'normal': return 'bg-blue-100 border-blue-300 text-blue-800'
      case 'low': return 'bg-gray-100 border-gray-300 text-gray-800'
      default: return 'bg-blue-100 border-blue-300 text-blue-800'
    }
  }

  const getUrgencyIcon = (urgency: string) => {
    switch (urgency) {
      case 'critical': return '🚨'
      case 'high': return '⚠️'
      case 'normal': return 'ℹ️'
      case 'low': return '📌'
      default: return 'ℹ️'
    }
  }

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const hrs = Math.floor(mins / 60)
    
    if (hrs > 0) {
      return `${hrs}h ${mins % 60}m`
    }
    return `${mins}m`
  }

  return (
    <div className={`border rounded-lg p-4 shadow-sm transition-all duration-200 ${
      !notification.isRead ? 'bg-blue-50 border-blue-200' : 'bg-white border-gray-200'
    } ${!notification.isActive ? 'opacity-50' : ''}`}>
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="text-lg">{getUrgencyIcon(notification.urgency_level)}</span>
          <h3 className="font-semibold text-gray-900">AI Handoff Request</h3>
          <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getUrgencyColor(notification.urgency_level)}`}>
            {notification.urgency_level.toUpperCase()}
          </span>
        </div>
        
        <div className="flex items-center gap-2">
          {!notification.isRead && (
            <button
              onClick={() => onMarkRead(notification.id)}
              className="text-blue-600 hover:text-blue-800 text-sm"
              title="Mark as read"
            >
              Mark Read
            </button>
          )}
          <button
            onClick={() => onDismiss(notification.id)}
            className="text-gray-400 hover:text-gray-600"
            title="Dismiss"
          >
            ✕
          </button>
        </div>
      </div>

      {/* Customer Info */}
      <div className="mb-3 p-3 bg-gray-50 rounded border">
        <div className="grid grid-cols-2 gap-2 text-sm">
          <div>
            <span className="font-medium text-gray-600">Customer:</span>
            <p className="text-gray-900">{notification.customer_name || 'Unknown'}</p>
          </div>
          <div>
            <span className="font-medium text-gray-600">Email:</span>
            <p className="text-gray-900">{notification.customer_email || 'N/A'}</p>
          </div>
        </div>
      </div>

      {/* Handoff Reason */}
      <div className="mb-3">
        <span className="font-medium text-gray-600 text-sm">Reason:</span>
        <p className="text-gray-900 mt-1">{notification.handoff_reason}</p>
      </div>

      {/* Session Metadata */}
      <div className="mb-4 text-sm text-gray-600">
        <div className="flex gap-4">
          <span>Messages: {notification.session_metadata.messages_count}</span>
          <span>Duration: {formatDuration(notification.session_metadata.session_duration)}</span>
        </div>
        {notification.session_metadata.last_ai_response && (
          <div className="mt-2">
            <span className="font-medium">Last AI Response:</span>
            <p className="text-gray-700 text-xs mt-1 p-2 bg-gray-100 rounded">
              {notification.session_metadata.last_ai_response.length > 100
                ? `${notification.session_metadata.last_ai_response.substring(0, 100)}...`
                : notification.session_metadata.last_ai_response
              }
            </p>
          </div>
        )}
      </div>

      {/* Timestamp */}
      <div className="text-xs text-gray-500 mb-4">
        Requested: {new Date(notification.requested_at).toLocaleString()}
        {notification.timeout_at && (
          <span className="ml-2">
            • Expires: {new Date(notification.timeout_at).toLocaleString()}
          </span>
        )}
      </div>

      {/* Actions */}
      {notification.isActive && (
        <div className="flex gap-2">
          <button
            onClick={handleAccept}
            disabled={isAccepting || isDeclining}
            className="flex-1 bg-green-600 hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2 rounded font-medium transition-colors"
          >
            {isAccepting ? 'Accepting...' : 'Accept & Take Over'}
          </button>
          <button
            onClick={handleDecline}
            disabled={isAccepting || isDeclining}
            className="flex-1 bg-gray-600 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2 rounded font-medium transition-colors"
          >
            {isDeclining ? 'Declining...' : 'Decline'}
          </button>
        </div>
      )}
      
      {!notification.isActive && (
        <div className="text-center py-2 text-gray-500 text-sm">
          This handoff request has been handled
        </div>
      )}
    </div>
  )
}
