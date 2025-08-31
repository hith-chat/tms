import { useState, useRef, useEffect } from 'react'
import { Bell, Check, CheckCheck, X } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { useNotificationContext } from '../contexts/NotificationContext'
import type { Notification } from '../types/notifications'

interface NotificationBellProps {
  className?: string
}

export function NotificationBell({ className = '' }: NotificationBellProps) {
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  
  const {
    notifications,
    count,
    loading,
    error,
    markAsRead,
    markAllAsRead,
    loadMore,
    clearError,
    refresh
  } = useNotificationContext()

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Refresh notifications when dropdown opens
  useEffect(() => {
    if (isOpen) {
      refresh()
    }
  }, [isOpen, refresh])

  const handleNotificationClick = async (notification: Notification) => {
    if (!notification.is_read) {
      await markAsRead(notification.id)
    }
    
    // Navigate to action URL if provided
    if (notification.action_url) {
      // You can implement navigation logic here
      // e.g., navigate to notification.action_url
      console.log('Navigate to:', notification.action_url)
    }
  }

  const getNotificationIcon = (type: Notification['type']) => {
    switch (type) {
      case 'message_received':
        return '💬'
      case 'ticket_assigned':
        return '🎫'
      case 'ticket_updated':
        return '📝'
      case 'ticket_escalated':
        return '⬆️'
      case 'ticket_resolved':
        return '✅'
      case 'mention_received':
        return '@'
      case 'sla_warning':
        return '⚠️'
      case 'sla_breach':
        return '🚨'
      case 'system_alert':
        return '🔔'
      case 'maintenance_notice':
        return '🔧'
      case 'feature_announcement':
        return '🎉'
      // Phase 4: New notification types
      case 'agent_assignment':
        return '👤'
      case 'howling_alarm':
        return '🚨'
      case 'agent_auto_assigned':
        return '🤖'
      case 'knowledge_response':
        return '📚'
      case 'greeting_response':
        return '👋'
      default:
        return '📌'
    }
  }

  return (
    <div className={`relative ${className}`} ref={dropdownRef}>
      {/* Bell Icon Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2"
        aria-label={`Notifications${count.unread > 0 ? ` (${count.unread} unread)` : ''}`}
      >
        <Bell className="w-5 h-5" />
        
        {/* Unread Count Badge */}
        {count.unread > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs font-medium min-w-[1.25rem] h-5 rounded-full flex items-center justify-center px-1">
            {count.unread > 99 ? '99+' : count.unread}
          </span>
        )}
      </button>

      {/* Dropdown Panel */}
      {isOpen && (
        <div className="absolute right-0 top-full mt-2 w-96 bg-white rounded-lg shadow-lg border border-gray-200 z-50 max-h-96 flex flex-col">
          {/* Header */}
          <div className="p-4 border-b border-gray-200 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-gray-900">
              Notifications
              {count.unread > 0 && (
                <span className="ml-2 text-sm font-normal text-gray-500">
                  ({count.unread} unread)
                </span>
              )}
            </h3>
            
            <div className="flex items-center gap-2">
              {count.unread > 0 && (
                <button
                  onClick={markAllAsRead}
                  className="text-xs text-primary hover:text-primary/80 font-medium flex items-center gap-1"
                  title="Mark all as read"
                >
                  <CheckCheck className="w-3 h-3" />
                  Mark all read
                </button>
              )}
              
              <button
                onClick={() => setIsOpen(false)}
                className="text-gray-400 hover:text-gray-600 p-1"
                aria-label="Close notifications"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="p-4 bg-red-50 border-b border-red-200">
              <div className="flex items-center justify-between">
                <p className="text-sm text-red-600">{error}</p>
                <button
                  onClick={clearError}
                  className="text-red-400 hover:text-red-600"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}

          {/* Notifications List */}
          <div className="flex-1 overflow-y-auto">
            
            {loading && notifications.length === 0 ? (
              <div className="p-4 text-center text-gray-500">
                <div className="animate-spin w-6 h-6 border-2 border-primary border-t-transparent rounded-full mx-auto mb-2"></div>
                Loading notifications...
              </div>
            ) : notifications.length === 0 ? (
              <div className="p-8 text-center text-gray-500">
                <Bell className="w-12 h-12 mx-auto mb-2 text-gray-300" />
                <p className="text-sm">No notifications yet</p>
              </div>
            ) : (
              <>
                {notifications.map((notification) => (
                  <div
                    key={notification.id}
                    className={`p-4 border-b border-gray-100 hover:bg-gray-50 cursor-pointer transition-colors duration-150 ${
                      !notification.is_read ? 'bg-blue-50 border-l-4 border-l-blue-500' : ''
                    }`}
                    onClick={() => handleNotificationClick(notification)}
                  >
                    <div className="flex items-start gap-3">
                      {/* Icon */}
                      <span className="text-lg flex-shrink-0 mt-0.5">
                        {getNotificationIcon(notification.type)}
                      </span>
                      
                      {/* Content */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-start justify-between gap-2">
                          <h4 className={`text-sm font-medium ${
                            notification.is_read ? 'text-gray-700' : 'text-gray-900'
                          }`}>
                            {notification.title}
                          </h4>
                          
                          {!notification.is_read && (
                            <button
                              onClick={(e) => {
                                e.stopPropagation()
                                markAsRead(notification.id)
                              }}
                              className="text-gray-400 hover:text-gray-600 p-1 flex-shrink-0"
                              title="Mark as read"
                            >
                              <Check className="w-3 h-3" />
                            </button>
                          )}
                        </div>
                        
                        <p className={`text-sm mt-1 ${
                          notification.is_read ? 'text-gray-500' : 'text-gray-600'
                        }`}>
                          {notification.message}
                        </p>
                        
                        <p className="text-xs text-gray-400 mt-2">
                          {formatDistanceToNow(new Date(notification.created_at), { addSuffix: true })}
                        </p>
                      </div>
                    </div>
                  </div>
                ))}
                
                {/* Load More Button */}
                {notifications.length >= 20 && (
                  <div className="p-4 border-t border-gray-200">
                    <button
                      onClick={loadMore}
                      disabled={loading}
                      className="w-full text-sm text-primary hover:text-primary/80 font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {loading ? 'Loading...' : 'Load more'}
                    </button>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
