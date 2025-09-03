import { useHandoffNotifications } from '../../hooks/useHandoffNotifications'

interface HandoffNotificationBadgeProps {
  className?: string
  showCount?: boolean
}

export function HandoffNotificationBadge({ 
  className = '', 
  showCount = true 
}: HandoffNotificationBadgeProps) {
  const { unreadCount } = useHandoffNotifications()

  if (unreadCount === 0) {
    return null
  }

  return (
    <span className={`bg-red-500 text-white text-xs font-medium rounded-full ${
      showCount ? 'px-2 py-1' : 'w-2 h-2'
    } ${className}`}>
      {showCount ? unreadCount : ''}
    </span>
  )
}
