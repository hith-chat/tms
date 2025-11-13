import { useState, useEffect } from 'react'
import { useLocation, useNavigate, Routes, Route, useParams } from 'react-router-dom'
import { MessageCircle, Settings } from 'lucide-react'
import { ChatSessionsPage } from './ChatSessionsPage'
import { ChatWidgetsPage } from './ChatWidgetsPage'
import { CreateChatWidgetPage } from './CreateChatWidgetPage'
import { apiClient } from '../lib/api'
import { PageHeader } from '../components/PageHeader'

type ChatTab = 'sessions' | 'widgets'

interface GuidedSetupState {
  hasWidgets: boolean
  hasSessions: boolean
  loading: boolean
}

// Wrapper component to extract sessionId from URL params
function ChatSessionPageWrapper() {
  const { sessionId } = useParams<{ sessionId: string }>()
  return <ChatSessionsPage initialSessionId={sessionId} />
}

export function ChatPage() {
  const location = useLocation()
  const navigate = useNavigate()
  
  // Determine active tab from URL
  const getTabFromPath = (pathname: string): ChatTab => {
    if (pathname.includes('/chat/widgets') || pathname.includes('/chat/widget/create')) return 'widgets'
    return 'sessions'
  }

  // Get page title and description based on current route
  const getPageInfo = (pathname: string) => {
    if (pathname.includes('/chat/widget/create')) {
      return {
        title: 'Create Chat Widget',
        description: 'Configure a new chat widget for your domain'
      }
    }
    if (pathname.includes('/chat/widget/edit/')) {
      return {
        title: 'Edit Chat Widget',
        description: 'Update chat widget configuration and settings'
      }
    }
    if (pathname.includes('/chat/widgets')) {
      return {
        title: 'Chat Widgets',
        description: 'Manage chat widgets for your domains'
      }
    }
    if (pathname.includes('/chat/sessions/')) {
      return {
        title: 'Chat Session',
        description: 'View and manage individual chat conversation'
      }
    }
    // Default to Chat Sessions
    return {
      title: 'Chat Sessions',
      description: 'Manage live chat conversations with customers'
    }
  }
  
  const [activeTab, setActiveTab] = useState<ChatTab>(() => getTabFromPath(location.pathname))
  const [guidedSetup, setGuidedSetup] = useState<GuidedSetupState>({
    hasWidgets: false,
    hasSessions: false,
    loading: true
  })

  // Check setup status
  useEffect(() => {
    checkSetupStatus()
  }, [])

  // Update tab when URL changes
  useEffect(() => {
    const newTab = getTabFromPath(location.pathname)
    setActiveTab(newTab)
  }, [location.pathname])

  const checkSetupStatus = async () => {
    try {
      setGuidedSetup(prev => ({ ...prev, loading: true }))
      
      const [widgets, sessions] = await Promise.all([
        apiClient.listChatWidgets().catch(() => []),
        apiClient.listChatSessions().catch(() => [])
      ])
      
      setGuidedSetup({
        hasWidgets: widgets.length > 0,
        hasSessions: sessions.length > 0,
        loading: false
      })
    } catch (_error) {
      setGuidedSetup(prev => ({ ...prev, loading: false }))
    }
  }

  // Handle tab change and update URL
  const handleTabChange = (tab: ChatTab) => {
    setActiveTab(tab)
    const newPath = tab === 'widgets' ? '/chat/widgets' : '/chat/sessions'
    navigate(newPath, { replace: true })
  }

  // Get current page info
  const pageInfo = getPageInfo(location.pathname)
  const tabs = [
    {
      id: 'widgets' as const,
      name: 'Chat Widgets',
      icon: Settings,
      description: 'Configure chat widgets for your domains',
      disabled: false
    },
    {
      id: 'sessions' as const,
      name: 'Chat Sessions',
      icon: MessageCircle,
      description: 'Manage live chat conversations',
      disabled: !guidedSetup.hasWidgets
    }
  ]

  return (
    <div className="h-full flex flex-col bg-gradient-to-br from-background via-background to-slate-50/20 dark:to-slate-950/20">
      <PageHeader
        icon={MessageCircle}
        title={pageInfo.title}
        subtitle={pageInfo.description}
        gradientFrom="#2563eb"
        gradientTo="#9333ea"
        actions={
          <div className="flex space-x-1 bg-muted/50 p-1 rounded-lg w-fit">
            {tabs.map((tab) => {
              const Icon = tab.icon
              const isActive = activeTab === tab.id

              return (
                <button
                  key={tab.id}
                  onClick={() => !tab.disabled && handleTabChange(tab.id)}
                  disabled={tab.disabled}
                  className={`
                    flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-all duration-200
                    ${tab.disabled
                      ? 'text-muted-foreground/50 cursor-not-allowed'
                      : isActive
                        ? 'bg-background text-foreground shadow-sm border border-border'
                        : 'text-muted-foreground hover:text-foreground hover:bg-background/50'
                    }
                  `}
                >
                  <Icon className="h-4 w-4" />
                  {tab.name}
                </button>
              )
            })}
          </div>
        }
      />

  {/* Content */}
  <div className="flex-1 min-h-0 overflow-auto">
        {guidedSetup.loading ? (
          <div className="flex items-center justify-center h-64">
            <div className="flex flex-col items-center gap-3">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              <p className="text-sm text-muted-foreground">Setting up chat system...</p>
            </div>
          </div>
        ) : (
          <Routes>
            <Route path="/sessions/:sessionId" element={<ChatSessionPageWrapper />} />
            <Route path="/sessions" element={<ChatSessionsPage />} />
            <Route path="/widgets" element={<ChatWidgetsPage />} />
            <Route path="/widget/create" element={<CreateChatWidgetPage />} />
            <Route path="/widget/edit/:widgetId" element={<CreateChatWidgetPage />} />
            <Route path="/" element={<ChatSessionsPage />} />
          </Routes>
        )}
      </div>
    </div>
  )
}
