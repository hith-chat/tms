import { FC } from 'react'
import { HowlingAlarmPanel } from './HowlingAlarmPanel'
import { EnhancedNotificationSettings } from './EnhancedNotificationSettings'

interface AlertsSettingsProps {
  className?: string
}

export const AlertsSettings: FC<AlertsSettingsProps> = ({ className = '' }) => {
  return (
    <div className={`space-y-6 ${className}`}>
      <div>
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">
          Alert Settings
        </h2>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          Monitor active alerts and manage your notification preferences
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Alerts Panel */}
        <div className="lg:col-span-2">
          <HowlingAlarmPanel />
        </div>
        
        {/* Notification Settings Panel */}
        <div className="lg:col-span-1">
          <EnhancedNotificationSettings />
        </div>
      </div>
    </div>
  )
}
