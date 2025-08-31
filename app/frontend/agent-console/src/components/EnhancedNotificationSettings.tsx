import { useState, useEffect } from 'react'
import { Settings, Volume2, Bell, Monitor, Smartphone, Mail, MessageSquare } from 'lucide-react'
import type { NotificationSettings } from '../types/notifications'

interface EnhancedNotificationSettingsProps {
  className?: string
}

export function EnhancedNotificationSettings({ className = '' }: EnhancedNotificationSettingsProps) {
  const [settings, setSettings] = useState<NotificationSettings>({
    sound_enabled: true,
    browser_notifications: true,
    email_notifications: true,
    // Phase 4: Enhanced channels
    audio_notifications: true,
    desktop_notifications: true,
    overlay_notifications: true,
    popup_notifications: true,
    alarm_sound_enabled: true,
    alarm_escalation_sound: true,
    notification_types: {
      ticket_assigned: true,
      ticket_updated: true,
      ticket_escalated: true,
      ticket_resolved: true,
      message_received: true,
      mention_received: true,
      sla_warning: true,
      sla_breach: true,
      system_alert: true,
      maintenance_notice: true,
      feature_announcement: true,
      // Phase 4: New types
      agent_assignment: true,
      howling_alarm: true,
      agent_auto_assigned: true,
      knowledge_response: false,
      greeting_response: false,
    }
  })
  
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    loadSettings()
  }, [])

  const loadSettings = async () => {
    try {
      const response = await fetch('/api/v1/notifications/settings', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        }
      })
      
      if (response.ok) {
        const data = await response.json()
        setSettings(data.settings)
      }
    } catch (err) {
      console.error('Failed to load notification settings:', err)
    }
  }

  const saveSettings = async () => {
    setSaving(true)
    setSaved(false)

    try {
      const response = await fetch('/api/v1/notifications/settings', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ settings })
      })

      if (response.ok) {
        setSaved(true)
        setTimeout(() => setSaved(false), 3000)
      }
    } catch (err) {
      console.error('Failed to save notification settings:', err)
    } finally {
      setSaving(false)
    }
  }

  const updateChannelSetting = (channel: keyof Omit<NotificationSettings, 'notification_types'>, enabled: boolean) => {
    setSettings(prev => ({
      ...prev,
      [channel]: enabled
    }))
  }

  const updateNotificationType = (type: keyof NotificationSettings['notification_types'], enabled: boolean) => {
    setSettings(prev => ({
      ...prev,
      notification_types: {
        ...prev.notification_types,
        [type]: enabled
      }
    }))
  }

  const requestBrowserPermission = async () => {
    if ('Notification' in window) {
      const permission = await Notification.requestPermission()
      updateChannelSetting('browser_notifications', permission === 'granted')
    }
  }

  const testNotification = () => {
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification('Test Notification', {
        body: 'This is a test of your notification settings',
        icon: '/favicon.ico'
      })
    }
  }

  return (
    <div className={`bg-card border border-border rounded-lg shadow-sm ${className}`}>
      <div className="p-6 border-b border-border bg-card/50">
        <div className="flex items-center space-x-2">
          <Settings className="w-5 h-5 text-muted-foreground" />
          <h2 className="text-lg font-semibold text-foreground">
            Alert Preferences
          </h2>
        </div>
        <p className="text-sm text-muted-foreground mt-1">
          Configure how you receive notifications and alerts
        </p>
      </div>

      <div className="p-6 space-y-6">{/* Delivery Channels */}
        <div>
          <h3 className="text-md font-semibold text-foreground mb-4 flex items-center space-x-2">
            <Bell className="w-4 h-4" />
            <span>Delivery Channels</span>
          </h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Traditional Channels */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium text-muted-foreground">Standard Channels</h4>
              
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.browser_notifications}
                  onChange={(e) => updateChannelSetting('browser_notifications', e.target.checked)}
                  className="w-4 h-4 text-primary"
                />
                <Monitor className="w-4 h-4 text-muted-foreground" />
                <span className="text-sm text-foreground">Browser Notifications</span>
                {!('Notification' in window) || Notification.permission === 'denied' ? (
                  <button
                    onClick={requestBrowserPermission}
                    className="text-xs px-2 py-1 bg-primary/10 text-primary rounded hover:bg-primary/20"
                  >
                    Enable
                  </button>
                ) : null}
              </label>

              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.email_notifications}
                  onChange={(e) => updateChannelSetting('email_notifications', e.target.checked)}
                  className="w-4 h-4 text-primary"
                />
                <Mail className="w-4 h-4 text-muted-foreground" />
                <span className="text-sm text-foreground">Email Notifications</span>
              </label>

              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.sound_enabled}
                  onChange={(e) => updateChannelSetting('sound_enabled', e.target.checked)}
                  className="w-4 h-4 text-primary"
                />
                <Volume2 className="w-4 h-4 text-muted-foreground" />
                <span className="text-sm text-foreground">Sound Notifications</span>
              </label>
            </div>

            {/* Phase 4: Enhanced Channels */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium text-muted-foreground">Enhanced Channels (Phase 4)</h4>
              
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.audio_notifications}
                  onChange={(e) => updateChannelSetting('audio_notifications', e.target.checked)}
                  className="w-4 h-4 text-primary"
                />
                <Volume2 className="w-4 h-4 text-orange-500" />
                <span className="text-sm text-foreground">Audio Alerts</span>
              </label>

              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.desktop_notifications}
                  onChange={(e) => updateChannelSetting('desktop_notifications', e.target.checked)}
                  className="w-4 h-4 text-primary"
                />
                <Monitor className="w-4 h-4 text-orange-500" />
                <span className="text-sm text-foreground">Desktop Notifications</span>
              </label>

              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.overlay_notifications}
                  onChange={(e) => updateChannelSetting('overlay_notifications', e.target.checked)}
                  className="w-4 h-4 text-primary"
                />
                <Smartphone className="w-4 h-4 text-orange-500" />
                <span className="text-sm text-foreground">Overlay Notifications</span>
              </label>

              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.popup_notifications}
                  onChange={(e) => updateChannelSetting('popup_notifications', e.target.checked)}
                  className="w-4 h-4 text-primary"
                />
                <MessageSquare className="w-4 h-4 text-orange-500" />
                <span className="text-sm text-foreground">Popup Notifications</span>
              </label>
            </div>
          </div>
        </div>

        {/* Howling Alarm Settings */}
        <div className="border-t border-border pt-6">
          <h3 className="text-md font-semibold text-foreground mb-4 flex items-center space-x-2">
            <Volume2 className="w-4 h-4 text-red-500" />
            <span>Howling Alarm Settings</span>
          </h3>
          
          <div className="space-y-3">
            <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                checked={settings.alarm_sound_enabled}
                onChange={(e) => updateChannelSetting('alarm_sound_enabled', e.target.checked)}
                className="w-4 h-4 text-red-600"
              />
              <span className="text-sm text-foreground">Enable Alarm Sounds</span>
            </label>

            <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                checked={settings.alarm_escalation_sound}
                onChange={(e) => updateChannelSetting('alarm_escalation_sound', e.target.checked)}
                className="w-4 h-4 text-red-600"
              />
              <span className="text-sm text-foreground">Escalation Sound Effects</span>
            </label>
          </div>
        </div>

        {/* Notification Types */}
        <div className="border-t border-border pt-6">
          <h3 className="text-md font-semibold text-foreground mb-4">Notification Types</h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Standard Types */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium text-muted-foreground">Standard Types</h4>
              
              {Object.entries(settings.notification_types)
                .filter(([key]) => !['agent_assignment', 'howling_alarm', 'agent_auto_assigned', 'knowledge_response', 'greeting_response'].includes(key))
                .map(([key, enabled]) => (
                <label key={key} className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={enabled}
                    onChange={(e) => updateNotificationType(key as keyof NotificationSettings['notification_types'], e.target.checked)}
                    className="w-4 h-4 text-primary"
                  />
                  <span className="text-sm text-foreground capitalize">
                    {key.replace(/_/g, ' ')}
                  </span>
                </label>
              ))}
            </div>

            {/* Phase 4 Types */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium text-muted-foreground">Agentic Types (Phase 4)</h4>
              
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.notification_types.agent_assignment}
                  onChange={(e) => updateNotificationType('agent_assignment', e.target.checked)}
                  className="w-4 h-4 text-orange-600"
                />
                <span className="text-sm text-foreground">Agent Assignments</span>
              </label>

              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.notification_types.howling_alarm}
                  onChange={(e) => updateNotificationType('howling_alarm', e.target.checked)}
                  className="w-4 h-4 text-red-600"
                />
                <span className="text-sm text-foreground">Howling Alarms</span>
              </label>

              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.notification_types.agent_auto_assigned}
                  onChange={(e) => updateNotificationType('agent_auto_assigned', e.target.checked)}
                  className="w-4 h-4 text-orange-600"
                />
                <span className="text-sm text-foreground">Auto Assignments</span>
              </label>

              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.notification_types.knowledge_response}
                  onChange={(e) => updateNotificationType('knowledge_response', e.target.checked)}
                  className="w-4 h-4 text-green-600"
                />
                <span className="text-sm text-foreground">Knowledge Responses</span>
              </label>

              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.notification_types.greeting_response}
                  onChange={(e) => updateNotificationType('greeting_response', e.target.checked)}
                  className="w-4 h-4 text-green-600"
                />
                <span className="text-sm text-foreground">Greeting Responses</span>
              </label>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex space-x-3 pt-6 border-t border-border">
          <button
            onClick={saveSettings}
            disabled={saving}
            className="flex-1 bg-primary text-primary-foreground px-4 py-2 rounded-lg hover:bg-primary/90 disabled:opacity-50"
          >
            {saving ? 'Saving...' : 'Save Settings'}
          </button>
          
          <button
            onClick={testNotification}
            className="px-4 py-2 border border-border text-foreground rounded-lg hover:bg-accent"
          >
            Test Notification
          </button>
          
          {saved && (
            <div className="flex items-center text-green-600 text-sm">
              ✓ Settings saved
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
