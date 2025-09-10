import React, { useState, useEffect, useMemo, useCallback } from 'react'
import { 
  Plus, 
  Search, 
  MoreHorizontal, 
  Clock, 
  User, 
  AlertCircle, 
  Download,
  RefreshCw,
  Inbox,
  CheckCircle,
  XCircle,
  Settings,
  Archive,
  Star,
  Calendar,
  Tag
} from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import {
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Badge,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  useToast,
  Toaster,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@tms/shared'
import { apiClient, Ticket, CreateTicketRequest } from '../lib/api'

// Enterprise color schemes with CSS variables
const statusConfig = {
  new: { color: 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800', icon: AlertCircle },
  open: { color: 'bg-green-50 text-green-700 border-green-200 dark:bg-green-950 dark:text-green-300 dark:border-green-800', icon: CheckCircle },
  pending: { color: 'bg-yellow-50 text-yellow-700 border-yellow-200 dark:bg-yellow-950 dark:text-yellow-300 dark:border-yellow-800', icon: Clock },
  resolved: { color: 'bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-950 dark:text-purple-300 dark:border-purple-800', icon: CheckCircle },
  closed: { color: 'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-950 dark:text-gray-300 dark:border-gray-800', icon: XCircle },
}

const priorityConfig = {
  low: { color: 'bg-gray-50 text-gray-600 border-gray-200 dark:bg-gray-950 dark:text-gray-400 dark:border-gray-800', level: 1 },
  normal: { color: 'bg-blue-50 text-blue-600 border-blue-200 dark:bg-blue-950 dark:text-blue-400 dark:border-blue-800', level: 2 },
  high: { color: 'bg-orange-50 text-orange-600 border-orange-200 dark:bg-orange-950 dark:text-orange-400 dark:border-orange-800', level: 3 },
  urgent: { color: 'bg-red-50 text-red-600 border-red-200 dark:bg-red-950 dark:text-red-400 dark:border-red-800', level: 4 },
}

interface TicketFilters {
  search: string
  status: string
  priority: string
  assignee: string
  dateRange: string
}

export const TicketsPage: React.FC = () => {
  const navigate = useNavigate()
  const { toast } = useToast()
  const [tickets, setTickets] = useState<Ticket[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filters, setFilters] = useState<TicketFilters>({
    search: '',
    status: 'all',
    priority: 'all',
    assignee: 'all',
    dateRange: 'all'
  })
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [creating, setCreating] = useState(false)
  const [refreshing, setRefreshing] = useState(false)

  useEffect(() => {
    loadTickets()
  }, [])

  const loadTickets = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const ticketList = await apiClient.getTickets()
      setTickets(ticketList)
    } catch (err) {
      console.error('Failed to load tickets:', err)
      setError('Failed to load tickets. Please try again.')
      toast({
        title: "Error",
        description: "Failed to load tickets. Please try again.",
        variant: "destructive"
      })
    } finally {
      setLoading(false)
    }
  }, [toast])

  const handleRefresh = useCallback(async () => {
    try {
      setRefreshing(true)
      await loadTickets()
      toast({
        title: "Refreshed",
        description: "Ticket list has been updated.",
        variant: "default"
      })
    } finally {
      setRefreshing(false)
    }
  }, [loadTickets, toast])

  // Memoized filtered and sorted tickets for performance
  const filteredTickets = useMemo(() => {
    return tickets.filter(ticket => {
      const matchesSearch = !filters.search || 
        ticket.subject.toLowerCase().includes(filters.search.toLowerCase()) ||
        ticket.customer?.name?.toLowerCase().includes(filters.search.toLowerCase()) ||
        ticket.number.toString().includes(filters.search)
      
      const matchesStatus = filters.status === 'all' || ticket.status === filters.status
      const matchesPriority = filters.priority === 'all' || ticket.priority === filters.priority
      // Add more filter logic here for assignee, date range, etc.
      
      return matchesSearch && matchesStatus && matchesPriority
    }).sort((a, b) => {
      // Sort by priority (urgent first) then by creation date (newest first)
      const priorityDiff = priorityConfig[b.priority].level - priorityConfig[a.priority].level
      if (priorityDiff !== 0) return priorityDiff
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    })
  }, [tickets, filters])

  const handleCreateTicket = useCallback(() => {
    setShowCreateDialog(true)
  }, [])

  const handleCreateTicketSubmit = useCallback(async (ticketData: CreateTicketRequest) => {
    try {
      setCreating(true)
      const newTicket = await apiClient.createTicket(ticketData)
      setTickets(prev => [newTicket, ...prev])
      setShowCreateDialog(false)
      
      toast({
        title: "Ticket created",
        description: `Ticket #${newTicket.number} has been created successfully.`,
        variant: "default"
      })
    } catch (error) {
      console.error('Failed to create ticket:', error)
      toast({
        title: "Error",
        description: "Failed to create ticket. Please try again.",
        variant: "destructive"
      })
    } finally {
      setCreating(false)
    }
  }, [toast])

  const activeFiltersCount = useMemo(() => {
    return Object.entries(filters).filter(([key, value]) => 
      key !== 'search' && value !== 'all' && value !== ''
    ).length
  }, [filters])

  if (error) {
    return (
      <div className="flex-1 bg-background">
        <div className="p-6">
          <Card className="p-8 text-center">
            <AlertCircle className="h-12 w-12 text-destructive mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-foreground mb-2">Unable to load tickets</h2>
            <p className="text-muted-foreground mb-6">{error}</p>
            <Button onClick={loadTickets}>
              <RefreshCw className="h-4 w-4 mr-2" />
              Try Again
            </Button>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <>
      <div className="flex-1 bg-background">
        {/* Enhanced Header */}
        <div className="border-b bg-card">
          <div className="px-6 py-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex items-center gap-3">
                  <div className="p-1.5 bg-primary/10 rounded-md">
                    <Inbox className="h-5 w-5 text-primary" />
                  </div>
                  <div>
                    <h1 className="text-xl font-semibold text-foreground">Tickets</h1>
                    <p className="text-xs text-muted-foreground">
                      {filteredTickets.length} of {tickets.length} tickets
                    </p>
                  </div>
                </div>
              </div>
              
              <div className="flex items-center gap-2">
                <Button 
                  variant="outline" 
                  size="sm" 
                  onClick={handleRefresh}
                  disabled={refreshing}
                  aria-label="Refresh tickets"
                >
                  <RefreshCw className={`h-4 w-4 ${refreshing ? 'animate-spin' : ''}`} />
                </Button>
                
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm" aria-label="More actions">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem>
                      <Download className="h-4 w-4 mr-2" />
                      Export tickets
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <Settings className="h-4 w-4 mr-2" />
                      Manage views
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem>
                      <Archive className="h-4 w-4 mr-2" />
                      Bulk actions
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                
                <Button onClick={handleCreateTicket} size="sm" className="gap-1.5">
                  <Plus className="h-4 w-4" />
                  New Ticket
                </Button>
              </div>
            </div>
          </div>
        </div>

        {/* Enhanced Filters Bar */}
        <div className="border-b bg-card/50">
          <div className="px-6 py-4">
            <div className="flex items-center gap-3">
              <div className="relative flex-1 max-w-sm">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="Search tickets..."
                  value={filters.search}
                  onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
                  className="pl-10 h-9"
                  aria-label="Search tickets"
                />
              </div>
              
              <div className="flex items-center gap-2">
                <Select value={filters.status} onValueChange={(value) => setFilters(prev => ({ ...prev, status: value }))}>
                  <SelectTrigger className="w-28 h-9">
                    <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Status</SelectItem>
                    <SelectItem value="new">New</SelectItem>
                    <SelectItem value="open">Open</SelectItem>
                    <SelectItem value="pending">Pending</SelectItem>
                    <SelectItem value="resolved">Resolved</SelectItem>
                    <SelectItem value="closed">Closed</SelectItem>
                  </SelectContent>
                </Select>
                
                <Select value={filters.priority} onValueChange={(value) => setFilters(prev => ({ ...prev, priority: value }))}>
                  <SelectTrigger className="w-28 h-9">
                    <SelectValue placeholder="Priority" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Priority</SelectItem>
                    <SelectItem value="urgent">Urgent</SelectItem>
                    <SelectItem value="high">High</SelectItem>
                    <SelectItem value="normal">Normal</SelectItem>
                    <SelectItem value="low">Low</SelectItem>
                  </SelectContent>
                </Select>
                
                {activeFiltersCount > 0 && (
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={() => setFilters({ search: '', status: 'all', priority: 'all', assignee: 'all', dateRange: 'all' })}
                    className="h-9 px-3 text-xs"
                  >
                    Clear filters
                  </Button>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="p-6">
          {loading ? (
            <Card>
              <div className="p-6">
                <TicketListSkeleton />
              </div>
            </Card>
          ) : filteredTickets.length === 0 ? (
            <EmptyState 
              hasFilters={Object.values(filters).some(f => f !== 'all' && f !== '')}
              onCreateTicket={handleCreateTicket}
              onClearFilters={() => setFilters({ search: '', status: 'all', priority: 'all', assignee: 'all', dateRange: 'all' })}
            />
          ) : (
            <Card>
              <div className="divide-y divide-border">
                {filteredTickets.map((ticket) => (
                  <TicketListItem 
                    key={ticket.id}
                    ticket={ticket}
                    onClick={() => navigate(`/tickets/${ticket.id}`)}
                  />
                ))}
              </div>
            </Card>
          )}
        </div>
      </div>

      {/* Create Ticket Dialog */}
      <CreateTicketDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        onSubmit={handleCreateTicketSubmit}
        isLoading={creating}
      />
      
      {/* Toast Notifications */}
      <Toaster />
    </>
  )
}

// Enhanced Ticket List Item Component
interface TicketListItemProps {
  ticket: Ticket
  onClick: () => void
}

const TicketListItem: React.FC<TicketListItemProps> = ({ ticket, onClick }) => {
  const StatusIcon = statusConfig[ticket.status].icon
  
  return (
    <div 
      className="flex items-center gap-3 p-4 hover:bg-muted/50 cursor-pointer transition-colors group"
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          onClick()
        }
      }}
      tabIndex={0}
      role="button"
      aria-label={`View ticket #${ticket.number}: ${ticket.subject}`}
    >
      <div className="flex-shrink-0">
        <div className="w-8 h-8 rounded-full bg-muted flex items-center justify-center">
          <StatusIcon className="h-4 w-4 text-muted-foreground" />
        </div>
      </div>
      
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <h3 className="font-medium text-sm text-foreground truncate">
            #{ticket.number} {ticket.subject}
          </h3>
          <Badge className={`text-xs px-1.5 py-0.5 ${statusConfig[ticket.status].color}`} variant="outline">
            {ticket.status}
          </Badge>
          <Badge className={`text-xs px-1.5 py-0.5 ${priorityConfig[ticket.priority].color}`} variant="outline">
            {ticket.priority}
          </Badge>
        </div>
        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          <span className="flex items-center gap-1">
            <User className="h-3 w-3" />
            {ticket.customer?.name || 'Unknown Customer'}
          </span>
          {ticket.assigned_agent?.name && (
            <span>Assigned to {ticket.assigned_agent.name}</span>
          )}
          <span className="flex items-center gap-1">
            <Calendar className="h-3 w-3" />
            {new Date(ticket.created_at).toLocaleDateString()}
          </span>
        </div>
      </div>
      
      <div className="flex-shrink-0">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button 
              variant="ghost" 
              size="sm"
              className="h-8 w-8 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
              onClick={(e) => e.stopPropagation()}
              aria-label={`Actions for ticket #${ticket.number}`}
            >
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem>
              <Star className="h-4 w-4 mr-2" />
              Add to favorites
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Tag className="h-4 w-4 mr-2" />
              Add tags
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <Archive className="h-4 w-4 mr-2" />
              Archive
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}

// Loading Skeleton Component
const TicketListSkeleton: React.FC = () => (
  <div className="space-y-4">
    {[...Array(8)].map((_, i) => (
      <div key={i} className="flex items-center gap-4 p-4">
        <div className="w-10 h-10 bg-muted rounded-full animate-pulse" />
        <div className="flex-1 space-y-2">
          <div className="h-4 bg-muted rounded animate-pulse w-3/4" />
          <div className="h-3 bg-muted rounded animate-pulse w-1/2" />
        </div>
        <div className="w-8 h-8 bg-muted rounded animate-pulse" />
      </div>
    ))}
  </div>
)

// Empty State Component
interface EmptyStateProps {
  hasFilters: boolean
  onCreateTicket: () => void
  onClearFilters?: () => void
}

const EmptyState: React.FC<EmptyStateProps> = ({ hasFilters, onCreateTicket, onClearFilters }) => (
  <Card className="p-12 text-center">
    <div className="mx-auto w-16 h-16 bg-muted rounded-full flex items-center justify-center mb-6">
      <Inbox className="h-8 w-8 text-muted-foreground" />
    </div>
    <h3 className="text-lg font-semibold text-foreground mb-2">
      {hasFilters ? 'No tickets match your filters' : 'No tickets yet'}
    </h3>
    <p className="text-muted-foreground mb-6 max-w-md mx-auto">
      {hasFilters 
        ? 'Try adjusting your search criteria or filters to find what you\'re looking for.'
        : 'Get started by creating your first support ticket to help customers.'
      }
    </p>
    <div className="flex items-center justify-center gap-3">
      {hasFilters && onClearFilters && (
        <Button variant="outline" onClick={onClearFilters}>
          Clear Filters
        </Button>
      )}
      <Button onClick={onCreateTicket} className="gap-2">
        <Plus className="h-4 w-4" />
        {hasFilters ? 'Create Ticket' : 'Create your first ticket'}
      </Button>
    </div>
  </Card>
)

// Enhanced Create Ticket Dialog
interface CreateTicketDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: CreateTicketRequest) => void
  isLoading: boolean
}

const CreateTicketDialog: React.FC<CreateTicketDialogProps> = ({ 
  open, 
  onOpenChange, 
  onSubmit, 
  isLoading 
}) => {
  const [formData, setFormData] = useState<CreateTicketRequest>({
    subject: '',
    initial_message: '',
    priority: 'normal',
    type: 'question',
    requester_name: '',
    requester_email: '',
    source: 'web',
    customer_id: '550e8400-e29b-41d4-a716-446655440050' // Dummy customer ID
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.subject.trim()) return
    onSubmit(formData)
  }

  const resetForm = () => {
    setFormData({
      subject: '',
      initial_message: '',
      priority: 'normal',
      requester_name: '',
      requester_email: '',
      type: 'question',
      source: 'web',
      customer_id: '550e8400-e29b-41d4-a716-446655440050'
    })
  }

  return (
    <Dialog open={open} onOpenChange={(open) => {
      onOpenChange(open)
      if (!open) resetForm()
    }}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Plus className="h-5 w-5" />
            Create New Ticket
          </DialogTitle>
          <DialogDescription>
            Create a new support ticket to track customer issues and requests.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 mt-4">
          <div className="space-y-2">
            <label htmlFor="subject" className="text-sm font-medium text-foreground">
              Subject <span className="text-destructive">*</span>
            </label>
            <Input
              id="subject"
              type="text"
              value={formData.subject}
              onChange={(e) => setFormData(prev => ({ ...prev, subject: e.target.value }))}
              placeholder="Enter ticket subject"
              required
              disabled={isLoading}
              aria-describedby="subject-help"
            />
          </div>

          <div className="space-y-2">
            <label htmlFor="description" className="text-sm font-medium text-foreground">
              Description
            </label>
            <textarea
              id="description"
              value={formData.initial_message}
              onChange={(e) => setFormData(prev => ({ ...prev, initial_message: e.target.value }))}
              placeholder="Enter ticket description"
              disabled={isLoading}
              className="w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring bg-background text-foreground h-24 resize-none"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label htmlFor="name" className="text-sm font-medium text-foreground">
                Name
              </label>
              <Input
                id="name"
                type="text"
                value={formData.requester_name}
                onChange={(e) => setFormData(prev => ({ ...prev, requester_name: e.target.value }))}
                placeholder="Customer Name"
                required
                disabled={isLoading}
                aria-describedby="name"
              />
            </div>
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium text-foreground">
                Email
              </label>
              <Input
                id="email"
                type="text"
                value={formData.requester_email}
                onChange={(e) => setFormData(prev => ({ ...prev, requester_email: e.target.value }))}
                placeholder="Customer Email"
                required
                disabled={isLoading}
                aria-describedby="email"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label htmlFor="priority" className="text-sm font-medium text-foreground">
                Priority
              </label>
              <Select value={formData.priority} onValueChange={(value) => setFormData(prev => ({ ...prev, priority: value as any }))}>
                <SelectTrigger id="priority" disabled={isLoading}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="normal">Normal</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="urgent">Urgent</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label htmlFor="type" className="text-sm font-medium text-foreground">
                Type
              </label>
              <Select value={formData.type} onValueChange={(value) => setFormData(prev => ({ ...prev, type: value as any }))}>
                <SelectTrigger id="type" disabled={isLoading}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="question">Question</SelectItem>
                  <SelectItem value="incident">Incident</SelectItem>
                  <SelectItem value="problem">Problem</SelectItem>
                  <SelectItem value="task">Task</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter className="gap-2 pt-4">
            <Button 
              type="button" 
              variant="outline" 
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button 
              type="submit" 
              disabled={isLoading || !formData.subject.trim()}
              className="gap-2"
            >
              {isLoading ? (
                <>
                  <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Plus className="h-4 w-4" />
                  Create Ticket
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
