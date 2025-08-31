import { FC, useEffect, useState } from 'react'
import { Ticket, AlertCircle, Clock, CheckCircle, Inbox, AlertTriangle } from 'lucide-react'
import { Link } from 'react-router-dom'
import { apiClient } from '../lib/api'
import { useAuth } from '../hooks/useAuth'
import { useHowlingAlarms } from '../hooks/useHowlingAlarms'

interface DashboardStats {
  openTickets: number
  inProgressTickets: number
  resolvedTickets: number
  urgentTickets: number
  avgResponseTime: string
  activeAgents: number
}

export const DashboardPage: FC = () => {
  const { user } = useAuth()
  const { alarms, stats: alarmStats } = useHowlingAlarms()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [recentTickets, setRecentTickets] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const loadDashboardData = async () => {
      try {
        setLoading(true)
        setError(null)
        
        // Fetch tickets to calculate stats
        console.log('Fetching tickets...')
        const tickets = await apiClient.getTickets()
        console.log('Tickets received:', tickets)
        
        // Ensure tickets is an array
        if (!Array.isArray(tickets)) {
          throw new Error('Expected tickets to be an array, got: ' + typeof tickets)
        }
        
        const openTickets = tickets.filter(t => t.status === 'open').length
        const inProgressTickets = tickets.filter(t => t.status === 'pending').length
        const resolvedTickets = tickets.filter(t => t.status === 'resolved').length
        const urgentTickets = tickets.filter(t => t.priority === 'urgent').length
        
        setStats({
          openTickets,
          inProgressTickets,
          resolvedTickets,
          urgentTickets,
          avgResponseTime: '2.4h', // TODO: Calculate from actual data
          activeAgents: 8, // TODO: Fetch from agents endpoint
        })
        
        // Get recent tickets (last 10)
        const recent = tickets
          .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
          .slice(0, 10)
        setRecentTickets(recent)
        
      } catch (err: any) {
        console.error('Dashboard error:', err)
        setError(err.message || 'Failed to load dashboard data')
      } finally {
        setLoading(false)
      }
    }

    if (user) {
      loadDashboardData()
    }
  }, [user])

  if (loading) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {[1, 2, 3, 4].map(i => (
              <div key={i} className="h-32 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 max-w-7xl mx-auto">
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          Error loading dashboard: {error}
        </div>
      </div>
    )
  }

  const statCards = [
    {
      title: 'Open Tickets',
      value: stats?.openTickets.toString() || '0',
      change: '+12%',
      changeType: 'increase' as const,
      icon: Ticket,
      color: 'bg-blue-500',
    },
    {
      title: 'In Progress',
      value: stats?.inProgressTickets.toString() || '0',
      change: '+5%',
      changeType: 'increase' as const,
      icon: Clock,
      color: 'bg-yellow-500',
    },
    {
      title: 'Resolved Today',
      value: stats?.resolvedTickets.toString() || '0',
      change: '+8%',
      changeType: 'increase' as const,
      icon: CheckCircle,
      color: 'bg-green-500',
    },
    {
      title: 'Urgent Tickets',
      value: stats?.urgentTickets.toString() || '0',
      change: '0',
      changeType: 'neutral' as const,
      icon: AlertCircle,
      color: 'bg-red-500',
    },
    // Phase 4: Alarm Statistics
    {
      title: 'Active Alarms',
      value: alarms.length.toString(),
      change: alarmStats ? `${alarmStats.total_active - alarms.length} resolved` : '0',
      changeType: alarms.length > 0 ? 'decrease' as const : 'neutral' as const,
      icon: AlertTriangle,
      color: alarms.length > 0 ? 'bg-red-600' : 'bg-gray-500',
    },
  ]

  return (
    <div className="h-full flex flex-col bg-gradient-to-br from-background via-background to-slate-50/20 dark:to-slate-950/20">
      {/* Enhanced Header with gradient and glass effect */}
      <div className="border-b border-border/50 bg-background/80 backdrop-blur-xl supports-[backdrop-filter]:bg-background/60 shadow-sm">
        <div className="px-6 py-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-4">
              <div className="relative">
                <div className="absolute -inset-1 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg blur opacity-25"></div>
                <div className="relative p-3 bg-gradient-to-br from-blue-50 to-purple-50 dark:from-blue-950 dark:to-purple-950 rounded-lg border border-blue-200/50 dark:border-blue-800/50">
                  <Inbox className="w-6 h-6 text-blue-600 dark:text-blue-400" />
                </div>
              </div>
              <div>
                <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                  Dashboard
                </h1>
                <div className="flex items-center gap-3 mt-1">
                  <p className="text-sm text-muted-foreground">
                    Welcome back, {user?.name}! Here's what's happening with your support team.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-7xl mx-auto">

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {statCards.map((stat, index) => (
          <div
            key={index}
            className="bg-card p-6 rounded-lg border border-border shadow-sm"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  {stat.title}
                </p>
                <p className="text-2xl font-bold text-foreground mt-2">
                  {stat.value}
                </p>
                <p className={`text-sm mt-1 ${
                  stat.changeType === 'increase' ? 'text-green-600' :
                  stat.changeType === 'neutral' ? 'text-muted-foreground' :
                  'text-red-600'
                }`}>
                  {stat.change} from last week
                </p>
              </div>
              <div className="h-12 w-12 bg-primary/10 rounded-lg flex items-center justify-center">
                <stat.icon className="h-6 w-6 text-primary" />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Phase 4: Active Alarms Alert */}
      {alarms.length > 0 && (
        <div className="mb-6">
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <AlertTriangle className="w-5 h-5 text-red-500" />
                <div>
                  <h3 className="text-sm font-medium text-red-800">
                    {alarms.length} Active Howling Alarm{alarms.length !== 1 ? 's' : ''}
                  </h3>
                  <p className="text-sm text-red-600">
                    Critical issues require immediate attention
                  </p>
                </div>
              </div>
              <Link
                to="/alarms"
                className="px-4 py-2 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 transition-colors"
              >
                Manage Alarms
              </Link>
            </div>
          </div>
        </div>
      )}

      {/* Quick Actions */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-card p-6 rounded-lg border border-border">
          <h2 className="text-xl font-semibold text-foreground mb-4">
            Recent Tickets
          </h2>
          <div className="space-y-4">
            {recentTickets.length > 0 ? (
              recentTickets.slice(0, 5).map((ticket) => (
                <div key={ticket.id} className="flex items-center justify-between p-3 bg-background rounded border">
                  <div className="flex-1">
                    <p className="font-medium text-foreground truncate">{ticket.title}</p>
                    <p className="text-sm text-muted-foreground">
                      {ticket.customer?.name || 'Unknown Customer'} • {new Date(ticket.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex items-center space-x-2">
                    <span className={`px-2 py-1 text-xs rounded-full ${
                      ticket.priority === 'urgent' ? 'bg-red-100 text-red-800' :
                      ticket.priority === 'high' ? 'bg-orange-100 text-orange-800' :
                      ticket.priority === 'medium' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-green-100 text-green-800'
                    }`}>
                      {ticket.priority}
                    </span>
                    <span className={`px-2 py-1 text-xs rounded-full ${
                      ticket.status === 'open' ? 'bg-blue-100 text-blue-800' :
                      ticket.status === 'in_progress' ? 'bg-yellow-100 text-yellow-800' :
                      ticket.status === 'resolved' ? 'bg-green-100 text-green-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {ticket.status.replace('_', ' ')}
                    </span>
                  </div>
                </div>
              ))
            ) : (
              <p className="text-muted-foreground">No recent tickets found.</p>
            )}
          </div>
        </div>

        <div className="bg-card p-6 rounded-lg border border-border">
          <h2 className="text-xl font-semibold text-foreground mb-4">
            Team Performance
          </h2>
          <div className="space-y-4">
            <p className="text-muted-foreground">
              Team metrics and performance charts will appear here...
            </p>
          </div>
        </div>
      </div>
        </div>
      </div>
    </div>
  )
}