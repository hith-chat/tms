import { FC } from 'react'
import { HowlingAlarmPanel } from '../components/HowlingAlarmPanel'
import { EnhancedNotificationSettings } from '../components/EnhancedNotificationSettings'

export const AlarmsPage: FC = () => {
  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">
          Alarm Management
        </h1>
        <p className="text-gray-600">
          Monitor active howling alarms and manage your notification preferences
        </p>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        {/* Main Alarms Panel */}
        <div className="xl:col-span-2">
          <HowlingAlarmPanel />
        </div>
        
        {/* Notification Settings Panel */}
        <div className="xl:col-span-1">
          <EnhancedNotificationSettings />
        </div>
      </div>
    </div>
  )
}
