import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { ThemeProvider } from './components/ThemeProvider'
import { NotificationProvider } from './contexts/NotificationContext'
import { AppShell } from './components/AppShell'
import { useAuth } from './hooks/useAuth'
import { LoginPage } from './pages/LoginPage'
import { SignUpPage } from './pages/SignUpPage'
import { GoogleCallbackPage } from './pages/GoogleCallbackPage'
import { InboxPage } from './pages/InboxPage'
import { EmailDetailPage } from './pages/EmailDetailPage'
import { AddInboxPage } from './pages/AddInboxPage'
import { EmailConnectorsPage } from './pages/EmailConnectorsPage'
import { EmailMailboxesPage } from './pages/EmailMailboxesPage'
import { CreateEmailMailboxPage } from './pages/CreateEmailMailboxPage'
import { TicketsPage } from './pages/TicketsPage'
import { DashboardPage } from './pages/DashboardPage'
import { TicketDetailPage } from './pages/TicketDetailPage'
import { AnalyticsPage } from './pages/AnalyticsPage'
import { IntegrationsPage } from './pages/IntegrationsPage'
import { SettingsPage } from './pages/SettingsPage'
import { KnowledgePage } from './pages/KnowledgePage'
import { NotificationsPage } from './pages/NotificationsPage'
import { ChatPage } from './pages/ChatPage'
import { CustomersPage } from './pages/CustomersPage'
import './index.css'

// Configure React Query client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60 * 1000, // 1 minute
      retry: 2,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 1,
    },
  },
})

function AppContent() {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-background">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <Routes>
      <Route path="/login" element={
        isAuthenticated ? <Navigate to="/inbox" replace /> : <LoginPage />
      } />
      <Route path="/signup" element={
        isAuthenticated ? <Navigate to="/inbox" replace /> : <SignUpPage />
      } />
      <Route path="/auth/google/callback" element={<GoogleCallbackPage />} />
      <Route path="/" element={
        isAuthenticated ? <Navigate to="/inbox" replace /> : <Navigate to="/login" replace />
      } />
      {isAuthenticated ? (
        <Route path="/*" element={
          <NotificationProvider>
            <AppShell>
              <Routes>
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/inbox" element={<InboxPage />} />
                <Route path="/inbox/emails/:id" element={<EmailDetailPage />} />
                <Route path="/inbox/add" element={<AddInboxPage />} />
                <Route path="/inbox/connectors" element={<EmailConnectorsPage />} />
                <Route path="/inbox/mailboxes" element={<EmailMailboxesPage />} />
                <Route path="/inbox/mailboxes/create" element={<CreateEmailMailboxPage />} />
                <Route path="/tickets" element={<TicketsPage />} />
                <Route path="/tickets/:id" element={<TicketDetailPage />} />
                <Route path="/customers" element={<CustomersPage />} />
                <Route path="/knowledge/*" element={<KnowledgePage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/analytics" element={<AnalyticsPage />} />
                <Route path="/integrations" element={<IntegrationsPage />} />
                <Route path="/notifications" element={<NotificationsPage />} />
                <Route path="/chat/*" element={<ChatPage />} />
                <Route path="*" element={<Navigate to="/dashboard" replace />} />
              </Routes>
            </AppShell>
          </NotificationProvider>
        } />
      ) : (
        <Route path="*" element={<Navigate to="/login" replace />} />
      )}
    </Routes>
  )
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <Router>
          <AppContent />
        </Router>
      </ThemeProvider>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  )
}

export default App
