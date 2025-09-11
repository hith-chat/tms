import React, { useState, useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { 
  Search, 
  Settings, 
  User, 
  Mail, 
  Moon, 
  Sun, 
  Monitor,
  ChevronLeft,
  ChevronRight,
  Home,
  Ticket,
  LogOut,
  MessageCircle,
} from 'lucide-react'
import { useTheme } from '../components/ThemeProvider'
import { useAuth } from '../hooks/useAuth'
import { useAgentWebSocket } from '../hooks/useAgentWebSocket'
import { useHowlingAlarms } from '../hooks/useHowlingAlarms'
import { NotificationBell } from './NotificationBell'
import { FloatingAlarmWidget } from './FloatingAlarmWidget'
import { CommandPalette } from './CommandPalette'
import { ProjectSelector } from './ProjectSelector'
import { apiClient } from '../lib/api'
import { ConnectionStatus } from './chat/ConnectionStatus'

interface AppShellProps {
  children: React.ReactNode
}

export function AppShell({ children }: AppShellProps) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false)
  const [currentProjectId, setCurrentProjectId] = useState<string | undefined>(
    localStorage.getItem('project_id') || undefined
  )
  const { theme, setTheme } = useTheme()
  const { user, logout } = useAuth()
  const location = useLocation()
  
  // WebSocket connection for real-time chat updates
  const { isConnected: wsConnected, isConnecting: wsConnecting, error: wsError } = useAgentWebSocket()
  
  // Phase 4: Howling Alarms integration
  const { 
    alarms, 
    acknowledgeAlarm, 
    soundEnabled, 
    setSoundEnabled
  } = useHowlingAlarms()
  
  // Determine if we're on a chat session page and extract session ID
  const isChatSessionPage = location.pathname.startsWith('/chat/sessions/')
  const selectedSession = isChatSessionPage ? { id: location.pathname.split('/').pop() } : null

  const handleProjectChange = (projectId: string) => {
    setCurrentProjectId(projectId)
    apiClient.setProjectId(projectId)
    // Optionally refresh current page data when project changes
    window.location.reload()
  }

  // Command palette keyboard shortcut
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setCommandPaletteOpen(true)
      }
    }
    
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  const navigation = [
    // { name: 'Dashboard', icon: Home, href: '/dashboard' },
    // { name: 'Inbox', icon: Mail, href: '/inbox' },
    { name: 'Tickets', icon: Ticket, href: '/tickets' },
    { name: 'Chat', icon: MessageCircle, href: '/chat/sessions' },
    // { name: 'Analytics', icon: BarChart3, href: '/analytics' },
    // { name: 'Integrations', icon: Zap, href: '/integrations' },
    { name: 'Settings', icon: Settings, href: '/settings' },
  ]

  const toggleTheme = () => {
    if (theme === 'light') setTheme('dark')
    else if (theme === 'dark') setTheme('system')
    else setTheme('light')
  }

  const getThemeIcon = () => {
    if (theme === 'light') return <Sun className="h-4 w-4" />
    if (theme === 'dark') return <Moon className="h-4 w-4" />
    return <Monitor className="h-4 w-4" />
  }

  const handleLogout = () => {
    logout()
  }

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <div 
        className={`${sidebarCollapsed ? 'w-16' : 'w-64'} transition-all duration-300 ease-in-out flex flex-col bg-card border-r border-border/60 shadow-sm`}
      >
  {/* Logo */}
        <div className="flex h-16 items-center justify-start border-b border-border/60 bg-card px-3">
          {sidebarCollapsed ? (
            <div className="h-8 w-8 rounded-lg bg-primary flex items-center justify-center shadow-sm">
              <img src="/sounds/images/taral-svg.svg" alt="T" className="h-5 w-5" />
            </div>
          ) : (
            <div className="flex items-center space-x-3">
              <div className="h-8 w-8 rounded-lg bg-primary flex items-center justify-center shadow-sm">
                <img src="/sounds/images/taral-svg.svg" alt="T" className="h-5 w-5" />
              </div>
              <span className="font-semibold text-foreground text-lg">Taral</span>
            </div>
          )}
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-2 p-3">
          {navigation.map((item) => {
            const isActive = location.pathname === item.href || location.pathname.startsWith(item.href + '/')
            return (
              <Link
                key={item.name}
                to={item.href}
                className={`
                  group flex items-center rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-200 ease-in-out
                  ${isActive
                    ? 'bg-primary text-primary-foreground shadow-sm'
                    : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground hover:shadow-sm'
                  }
                  ${sidebarCollapsed ? 'justify-center' : ''}
                `}
              >
                <item.icon
                  className={`h-5 w-5 flex-shrink-0 ${
                    sidebarCollapsed ? '' : 'mr-3'
                  } ${isActive ? 'text-primary-foreground' : ''}`}
                  aria-hidden="true"
                />
                {!sidebarCollapsed && (
                  <span className="truncate">{item.name}</span>
                )}
              </Link>
            )
          })}
        </nav>

        {/* Collapse toggle */}
        <div className="border-t border-border/60 p-3 bg-muted/20">
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="w-full flex items-center justify-center rounded-lg px-3 py-2.5 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-all duration-200 ease-in-out hover:shadow-sm"
            title={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {sidebarCollapsed ? (
              <ChevronRight className="h-5 w-5" />
            ) : (
              <>
                <ChevronLeft className="h-5 w-5 mr-3" />
                <span className="truncate">Collapse</span>
              </>
            )}
          </button>
        </div>
      </div>

      {/* Main content */}
      <div className="flex flex-1 flex-col min-w-0">
        {/* Top bar */}
        <header className="flex h-16 items-center justify-between border-b border-border/60 bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/60 px-6 shadow-sm">
          {/* Search */}
          <div className="flex flex-1 items-center space-x-6">
            <div className="relative w-80">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search tickets... (Cmd+K)"
                className="flex h-10 w-full rounded-lg border border-input bg-background px-3 py-2 pl-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-shadow"
                onFocus={() => setCommandPaletteOpen(true)}
                readOnly
              />
            </div>
            
            {/* Project Selector */}
            <ProjectSelector
              currentProjectId={currentProjectId}
              onProjectChange={handleProjectChange}
            />
          </div>

          {/* Actions */}
          <div className="flex items-center space-x-2">
            {/* Connection Status */}
            <ConnectionStatus 
              isConnected={wsConnected}
              isConnecting={wsConnecting}
              error={wsError}
              selectedSession={selectedSession}
              variant="md"
            />
            
            {/* Theme toggle */}
            <button
              onClick={toggleTheme}
              className="rounded-lg p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-all duration-200 hover:shadow-sm"
              title={`Theme: ${theme}`}
            >
              {getThemeIcon()}
            </button>

            {/* Notifications */}
            <NotificationBell />

            {/* User menu */}
            <div className="flex items-center space-x-3 pl-4 border-l border-border/60">
              <div className="text-right text-sm">
                <div className="font-medium text-foreground leading-tight">{user?.name || 'Loading...'}</div>
                <div className="text-muted-foreground text-xs">{user?.email || ''}</div>
              </div>
              <div className="relative flex items-center">
                <button className="rounded-lg p-1.5 text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-all duration-200 hover:shadow-sm">
                  <User className="h-5 w-5" />
                </button>
                {/* Simple logout for now - could be expanded to dropdown menu */}
                <button
                  onClick={handleLogout}
                  className="ml-2 rounded-lg p-1.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive transition-all duration-200 hover:shadow-sm"
                  title="Logout"
                >
                  <LogOut className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </header>

  {/* Page content */}
  <main className="flex-1 overflow-hidden bg-background">
          {children}
        </main>
      </div>
      
      {/* Command Palette */}
      <CommandPalette 
        isOpen={commandPaletteOpen} 
        onClose={() => setCommandPaletteOpen(false)} 
      />
      
      {/* Phase 4: Floating Howling Alarm Widget */}
      <FloatingAlarmWidget
        alarms={alarms}
        onAcknowledge={acknowledgeAlarm}
        soundEnabled={soundEnabled}
        onToggleSound={async () => await setSoundEnabled(!soundEnabled)}
      />
    </div>
  )
}
