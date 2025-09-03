import { useHandoffNotifications } from '../../hooks/useHandoffNotifications'
import { HandoffNotificationCard } from './HandoffNotificationCard'

interface HandoffNotificationsListProps {
  className?: string
}

export function HandoffNotificationsList({ className = '' }: HandoffNotificationsListProps) {
  const {
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
  } = useHandoffNotifications()

  if (notifications.length === 0) {
    return (
      <div className={`p-8 text-center ${className}`}>
        <div className="text-gray-400 text-lg mb-2">📭</div>
        <h3 className="text-lg font-medium text-gray-900 mb-1">No handoff requests</h3>
        <p className="text-gray-500">AI handoff requests will appear here when customers need human assistance.</p>
      </div>
    )
  }

  return (
    <div className={className}>
      {/* Error Banner */}
      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="text-red-600 mr-2">⚠️</div>
              <p className="text-red-800">{error}</p>
            </div>
            <button
              onClick={clearError}
              className="text-red-600 hover:text-red-800 text-sm"
            >
              Dismiss
            </button>
          </div>
        </div>
      )}

      {/* Loading Indicator */}
      {isLoading && (
        <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="flex items-center">
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
            <p className="text-blue-800">Processing handoff request...</p>
          </div>
        </div>
      )}
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <h2 className="text-xl font-semibold text-gray-900">Handoff Requests</h2>
          {unreadCount > 0 && (
            <span className="bg-red-500 text-white text-sm px-2 py-1 rounded-full">
              {unreadCount} new
            </span>
          )}
        </div>
        
        {notifications.length > 0 && (
          <button
            onClick={clearAll}
            className="text-gray-500 hover:text-gray-700 text-sm"
          >
            Clear All
          </button>
        )}
      </div>

      {/* Notifications List */}
      <div className="space-y-4">
        {notifications.map((notification) => (
          <HandoffNotificationCard
            key={notification.id}
            notification={notification}
            onAccept={acceptHandoff}
            onDecline={declineHandoff}
            onMarkRead={markAsRead}
            onDismiss={dismissNotification}
          />
        ))}
      </div>

      {/* Summary */}
      <div className="mt-6 p-4 bg-gray-50 rounded-lg text-sm text-gray-600">
        <div className="flex justify-between">
          <span>Total requests: {notifications.length}</span>
          <span>Active: {notifications.filter(n => n.isActive).length}</span>
          <span>Unread: {unreadCount}</span>
        </div>
      </div>
    </div>
  )
}
