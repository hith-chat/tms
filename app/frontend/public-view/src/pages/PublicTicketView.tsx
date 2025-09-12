import { useState } from 'react'
import type React from 'react'
import { useParams, Navigate } from 'react-router-dom'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  Badge,
  Button,
  Textarea,
  cn
} from '@tms/shared'
import { MessageCircle, Clock, User, Building, AlertCircle } from 'lucide-react'
import { ThemeToggle } from '@shared/theme'

// API Response types (different from internal types)
interface PublicMessage {
  id: string
  tenant_id: string
  project_id: string
  ticket_id: string
  author_type: 'customer' | 'agent' | 'system'
  author_id: string
  body: string
  is_private: boolean
  created_at: string
  user_info: {
    id: string
    name: string
    email: string
  }
}

interface PublicTicket {
  id: string
  tenant_id: string
  project_id: string
  number: number
  subject: string
  status: string
  priority: string
  type: string
  source: string
  customer_id: string
  customer_name: string
  created_at: string
  updated_at: string
}

interface TokenValidationResponse {
  valid: boolean
  ticket?: PublicTicket
  messages?: PublicMessage[]
}

export function PublicTicketView() {
  const { ticketId } = useParams<{ ticketId: string }>()
  const [newMessage, setNewMessage] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Validate token and fetch ticket data
  const { data: tokenData, isLoading, error } = useQuery({
    queryKey: ['public-ticket', ticketId],
    queryFn: async (): Promise<TokenValidationResponse> => {
      if (!ticketId) throw new Error('No ticketId provided')
      
      const response = await fetch(`/api/public/tickets/${ticketId}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error('Ticket not found or invalid token')
        }
        throw new Error('Failed to load ticket')
      }

      return response.json()
    },
    enabled: !!ticketId,
    retry: false
  })

  // Auto-refresh messages every 30 seconds
  // Messages query handled by React Query so the UI updates automatically
  const queryClient = useQueryClient()

  const { data: messagesData } = useQuery<PublicMessage[]>({
    queryKey: ['public-messages', ticketId],
    queryFn: async () => {
      if (!ticketId) throw new Error('No token')
      const resp = await fetch(`/api/public/tickets/${ticketId}/messages`)
      if (!resp.ok) throw new Error('Failed to load messages')
  const json = await resp.json()
  // API may return either an array or an object { messages: [...] }
  if (Array.isArray(json)) return json
  if (json && Array.isArray(json.messages)) return json.messages
  return []
    },
    enabled: !!ticketId && !!tokenData?.valid,
    // Auto refetch every 30s and update the cache/UI
    refetchInterval: 30000,
  // Use messages included in token validation as initial data to avoid flash
  initialData: Array.isArray(tokenData?.messages) ? tokenData?.messages : undefined,
  })

  // Mutation for posting a message with optimistic update
  const postMessage = useMutation({
    mutationFn: async (body: string) => {
      if (!ticketId) throw new Error('No token')
      const resp = await fetch(`/api/public/tickets/${ticketId}/messages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ body }),
      })
      if (!resp.ok) throw new Error('Failed to send message')
      return resp.json()
    },
    // Optimistic update: append a temp message immediately
    onMutate: async (body) => {
      await queryClient.cancelQueries({ queryKey: ['public-messages', ticketId] })
      const previous = queryClient.getQueryData<PublicMessage[]>(['public-messages', ticketId])

      const optimistic: PublicMessage = {
        id: `optimistic-${Date.now()}`,
        tenant_id: tokenData?.ticket?.tenant_id || '',
        project_id: tokenData?.ticket?.project_id || '',
        ticket_id: tokenData?.ticket?.id || '',
        author_type: 'customer',
        author_id: tokenData?.ticket?.customer_id || '',
        body,
        is_private: false,
        created_at: new Date().toISOString(),
        user_info: {
          id: tokenData?.ticket?.customer_id || '',
          name: tokenData?.ticket?.customer_name || 'You',
          email: '',
        },
      }

      queryClient.setQueryData<PublicMessage[]>(['public-messages', ticketId], (old) => {
        if (!old) return [optimistic]
        return [...old, optimistic]
      })

      return { previous }
    },
    onError: (_err, _vars, context: any) => {
      if (context?.previous) {
        queryClient.setQueryData(['public-messages', ticketId], context.previous)
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['public-messages', ticketId] })
      queryClient.invalidateQueries({ queryKey: ['public-ticket', ticketId] })
    },
  })

  const handleSubmitMessage = () => {
    if (!newMessage.trim() || !ticketId || isSubmitting) return
    setIsSubmitting(true)
    postMessage.mutate(newMessage.trim(), {
      onSuccess: () => setNewMessage(''),
      onSettled: () => setIsSubmitting(false),
    })
  }

  // Redirect if no token
  if (!ticketId) {
    return <Navigate to="/" replace />
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  // Error or invalid token
  if (error || !tokenData?.valid) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full">
          <CardHeader className="text-center">
            <div className="mx-auto w-12 h-12 bg-destructive/10 rounded-full flex items-center justify-center mb-4">
              <AlertCircle className="w-6 h-6 text-destructive" />
            </div>
            <CardTitle>Access Denied</CardTitle>
          </CardHeader>
          <CardContent className="text-center">
            <p className="text-muted-foreground mb-4">
              {error?.message || 'The ticket link is invalid or has expired.'}
            </p>
            <p className="text-sm text-muted-foreground">
              Please check your email for the correct link or contact support if you need assistance.
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  const ticket = tokenData?.ticket
  const liveMessages = messagesData ?? tokenData?.messages ?? []

  if (!ticket) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground mb-2">Ticket Not Found</h1>
          <p className="text-muted-foreground">This ticket could not be found.</p>
        </div>
      </div>
    )
  }

  const getPriorityColor = (priority: string) => {
    switch (priority.toLowerCase()) {
      case 'low': return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      case 'medium': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'high': return 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300'
      case 'urgent': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
      default: return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'open': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'in_progress': return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
      case 'waiting_on_customer': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'resolved': return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300'
      case 'closed': return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300'
      default: return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300'
    }
  }

  

  return (
    <div className="h-screen flex flex-col bg-background">
      {/* Fixed Header */}
      <header className="shrink-0 border-b bg-card/95 backdrop-blur">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div>
                <h1 className="text-xl font-bold text-foreground flex items-center gap-2">
                  <Building className="w-5 h-5" />
                  Support Ticket #{ticket.number}
                </h1>
                <p className="text-sm text-muted-foreground">{ticket.subject}</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Badge className={cn('text-xs', getPriorityColor(ticket.priority))}>
                {ticket.priority}
              </Badge>
              <Badge className={cn('text-xs', getStatusColor(ticket.status))}>
                {ticket.status.replace('_', ' ')}
              </Badge>
              <ThemeToggle />
            </div>
          </div>
        </div>
      </header>

      {/* Main Content - Flex Layout */}
      <div className="flex-1 flex flex-col min-h-0 max-w-4xl mx-auto w-full">
        {/* Ticket Details - Collapsible */}
        <div className="shrink-0 p-4 border-b bg-muted/20">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">
            <div className="flex items-center gap-2 text-muted-foreground">
              <User className="w-4 h-4" />
              <span>Customer: {ticket.customer_name || 'Customer'}</span>
            </div>
            <div className="flex items-center gap-2 text-muted-foreground">
              <Clock className="w-4 h-4" />
              <span>Created: {new Date(ticket.created_at).toLocaleDateString()} {new Date(ticket.created_at).toLocaleTimeString()}</span>
            </div>
            <div className="flex items-center gap-2 text-muted-foreground">
              <Clock className="w-4 h-4" />
              <span>Updated: {new Date(ticket.updated_at).toLocaleDateString()} {new Date(ticket.updated_at).toLocaleTimeString()}</span>
            </div>
          </div>
        </div>

        {/* Messages Header */}
        <div className="shrink-0 p-4 border-b bg-card/50">
            <h3 className="font-medium text-sm flex items-center gap-2">
            <MessageCircle className="w-4 h-4" />
            Conversation ({liveMessages.length} messages)
          </h3>
        </div>
        
        {/* Messages - Scrollable Container */}
        <div className="flex-1 overflow-auto">
          <div className="p-4 space-y-4">
            {liveMessages.length === 0 ? (
              <div className="text-center py-12">
                <MessageCircle className="w-12 h-12 text-muted-foreground/30 mx-auto mb-4" />
                <p className="text-muted-foreground">No messages yet. Start the conversation below.</p>
              </div>
              ) : (
              liveMessages.map((message, index) => (
                <div
                  key={message.id || index}
                  className={cn(
                    'flex gap-3 p-4 rounded-lg border',
                    message.author_type === 'agent'
                      ? 'bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 border-l-4 border-l-blue-500'
                      : 'bg-muted/30 border-border'
                  )}
                >
                  <div className="shrink-0">
                    {message.author_type === 'agent' ? (
                      <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center">
                        <User className="w-4 h-4 text-white" />
                      </div>
                    ) : (
                      <div className="w-8 h-8 rounded-full bg-muted flex items-center justify-center">
                        <User className="w-4 h-4 text-muted-foreground" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <span className="font-medium text-sm">
                        {message.user_info.name}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {new Date(message.created_at).toLocaleDateString()} at {new Date(message.created_at).toLocaleTimeString()}
                      </span>
                      {message.author_type === 'agent' && (
                        <Badge variant="secondary" className="text-xs">Support</Badge>
                      )}
                    </div>
                    <div className="text-sm whitespace-pre-wrap break-words leading-relaxed">{message.body}</div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Reply Form - Fixed at Bottom */}
        {ticket.status !== 'closed' && (
          <div className="shrink-0 border-t bg-card">
            <div className="p-4 space-y-4">
              <div className="flex items-center gap-2">
                <MessageCircle className="w-4 h-4" />
                <h4 className="font-medium text-sm">Add Reply</h4>
              </div>
              <Textarea
                placeholder="Type your message here..."
                value={newMessage}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setNewMessage(e.target.value)}
                className="min-h-[100px] resize-none"
                disabled={isSubmitting}
              />
              <div className="flex justify-end">
                <Button
                  onClick={handleSubmitMessage}
                  disabled={!newMessage.trim() || isSubmitting}
                  className="min-w-[100px]"
                >
                  {isSubmitting ? (
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2"></div>
                  ) : (
                    <MessageCircle className="w-4 h-4 mr-2" />
                  )}
                  {isSubmitting ? 'Sending...' : 'Send Reply'}
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
