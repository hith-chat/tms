import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { format, formatDistanceToNow } from 'date-fns'
import { 
  ArrowLeft, 
  Mail, 
  User, 
  Paperclip, 
  Reply, 
  Forward,
  Eye,
  EyeOff,
  Clock,
  ExternalLink,
  AlertTriangle,
} from 'lucide-react'
import { 
  Button, 
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  PageHeader,
  PageHeaderContent,
  PageHeaderTitle,
  PageHeaderBreadcrumb,
  PageHeaderActions,
  DataCard,
  DataCardHeader,
  DataCardContent,
  DetailSection,
  DetailSectionHeader,
  DetailSectionTitle,
  DetailSectionContent,
  DetailItem,
  StatusIndicator
} from '@tms/shared'
import { apiClient, EmailInbox, ConvertToTicketRequest } from '../lib/api'

export function EmailDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [email, setEmail] = useState<EmailInbox | null>(null)
  const [threadEmails, setThreadEmails] = useState<EmailInbox[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [toggling, setToggling] = useState(false)
  const [converting, setConverting] = useState(false)
  const [showConvertDialog, setShowConvertDialog] = useState(false)
  const [showReplyDialog, setShowReplyDialog] = useState(false)
  const [replyForm, setReplyForm] = useState({
    body: '',
    subject: '',
    cc_addresses: [] as string[],
    is_private: false
  })
  const [replying, setReplying] = useState(false)
  const [convertForm, setConvertForm] = useState<ConvertToTicketRequest>({
    type: 'question',
    priority: 'normal'
  })

  useEffect(() => {
    if (id) {
      loadEmailData(id)
    }
  }, [id])

  const loadEmailData = async (emailId: string) => {
    try {
      setLoading(true)
      setError(null)
      
      // First load the specific email to get its thread_id
      const foundEmail = await apiClient.getEmailFromId(emailId)
      setEmail(foundEmail)
      
      // If the email has a thread_id, load all emails in the thread
      if (foundEmail.thread_id) {
        const threadResponse = await apiClient.getEmailInbox({ 
          thread_id: foundEmail.thread_id,
          limit: 100 // Get all emails in thread
        })
        // Sort by sent_at/received_at to show chronological order
        const sortedEmails = threadResponse.emails.sort((a, b) => {
          const aTime = new Date(a.sent_at || a.received_at).getTime()
          const bTime = new Date(b.sent_at || b.received_at).getTime()
          return aTime - bTime
        })
        setThreadEmails(sortedEmails)
      } else {
        // If no thread_id, just show the single email
        setThreadEmails([foundEmail])
      }
    } catch (err) {
      console.error('Failed to load email:', err)
      setError('Failed to load email. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const handleToggleRead = async () => {
    if (!email) return
    
    try {
      setToggling(true)
      if (!email.is_read) {
        await apiClient.markEmailAsRead(email.id)
        setEmail(prev => prev ? { ...prev, is_read: true } : null)
      }
      // Note: There's no markAsUnread method in the API, so we can only mark as read
    } catch (err) {
      console.error('Failed to toggle read status:', err)
    } finally {
      setToggling(false)
    }
  }

  const handleConvertToTicket = async () => {
    if (!email) return
    
    try {
      setConverting(true)
      await apiClient.convertEmailToTicket(email.id, convertForm)
      
      // Refresh email data to get updated ticket_id
      await loadEmailData(email.id)
      setShowConvertDialog(false)
      
      // Show success message
      alert('Email successfully converted to ticket!')
    } catch (err) {
      console.error('Failed to convert email:', err)
      alert('Failed to convert email to ticket')
    } finally {
      setConverting(false)
    }
  }

  const handleReply = async () => {
    if (!email) return
    
    try {
      setReplying(true)
      await apiClient.replyToEmail(email.id, {
        body: replyForm.body,
        subject: replyForm.subject || undefined,
        cc_addresses: replyForm.cc_addresses.length > 0 ? replyForm.cc_addresses : undefined,
        is_private: replyForm.is_private
      })
      
      setShowReplyDialog(false)
      setReplyForm({
        body: '',
        subject: '',
        cc_addresses: [],
        is_private: false
      })
      
      // Refresh the thread to show the new reply
      await loadEmailData(email.id)
      
      // Show success message
      alert('Reply sent successfully!')
    } catch (err) {
      console.error('Failed to send reply:', err)
      alert('Failed to send reply')
    } finally {
      setReplying(false)
    }
  }

  const openReplyDialog = () => {
    if (!email) return
    
    // Set default reply subject
    const defaultSubject = email.subject.toLowerCase().startsWith('re:') 
      ? email.subject 
      : `Re: ${email.subject}`
    
    setReplyForm({
      body: '',
      subject: defaultSubject,
      cc_addresses: [],
      is_private: false
    })
    setShowReplyDialog(true)
  }

  const formatEmailDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffInHours = Math.abs(now.getTime() - date.getTime()) / (1000 * 60 * 60)
    
    if (diffInHours < 24) {
      return formatDistanceToNow(date, { addSuffix: true })
    } else if (diffInHours < 168) { // 7 days
      return format(date, 'EEE p')
    } else {
      return format(date, 'MMM d, yyyy')
    }
  }

  const getFromDisplay = (email: EmailInbox) => {
    return email.from_address
  }

  if (loading) {
    return (
      <div className="container mx-auto px-6 py-8 max-w-7xl">
        <PageHeader>
          <PageHeaderContent>
            <div className="flex items-center gap-4">
              <div className="h-10 w-24 bg-muted rounded animate-pulse"></div>
              <div className="h-8 w-64 bg-muted rounded animate-pulse"></div>
            </div>
            <div className="h-10 w-32 bg-muted rounded animate-pulse"></div>
          </PageHeaderContent>
        </PageHeader>
        
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mt-8">
          <div className="lg:col-span-2 space-y-6">
            <DataCard>
              <DataCardHeader>
                <div className="h-6 w-48 bg-muted rounded animate-pulse"></div>
              </DataCardHeader>
              <DataCardContent>
                <div className="space-y-4">
                  <div className="h-4 w-full bg-muted rounded animate-pulse"></div>
                  <div className="h-4 w-3/4 bg-muted rounded animate-pulse"></div>
                  <div className="h-64 w-full bg-muted rounded animate-pulse"></div>
                </div>
              </DataCardContent>
            </DataCard>
          </div>
          
          <div className="space-y-6">
            <DataCard>
              <DataCardHeader>
                <div className="h-5 w-24 bg-muted rounded animate-pulse"></div>
              </DataCardHeader>
              <DataCardContent>
                <div className="space-y-3">
                  <div className="h-4 w-full bg-muted rounded animate-pulse"></div>
                  <div className="h-4 w-2/3 bg-muted rounded animate-pulse"></div>
                  <div className="h-4 w-3/4 bg-muted rounded animate-pulse"></div>
                </div>
              </DataCardContent>
            </DataCard>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container mx-auto px-6 py-8 max-w-7xl">
        <PageHeader>
          <PageHeaderContent>
            <div>
              <PageHeaderTitle>Email Not Found</PageHeaderTitle>
            </div>
            <PageHeaderActions>
              <Button onClick={() => navigate('/inbox')} variant="outline">
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Inbox
              </Button>
            </PageHeaderActions>
          </PageHeaderContent>
        </PageHeader>
        
        <div className="mt-8">
          <DataCard>
            <DataCardContent>
              <div className="flex items-center justify-center py-12">
                <div className="text-center space-y-4">
                  <AlertTriangle className="h-12 w-12 text-destructive mx-auto" />
                  <div>
                    <h3 className="text-lg font-semibold">Unable to Load Email</h3>
                    <p className="text-muted-foreground mt-1">{error}</p>
                  </div>
                </div>
              </div>
            </DataCardContent>
          </DataCard>
        </div>
      </div>
    )
  }

  if (!email) {
    return (
      <div className="container mx-auto px-6 py-8 max-w-7xl">
        <PageHeader>
          <PageHeaderContent>
            <div>
              <PageHeaderTitle>Email Not Found</PageHeaderTitle>
            </div>
            <PageHeaderActions>
              <Button onClick={() => navigate('/inbox')} variant="outline">
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Inbox
              </Button>
            </PageHeaderActions>
          </PageHeaderContent>
        </PageHeader>
        
        <div className="mt-8">
          <DataCard>
            <DataCardContent>
              <div className="flex items-center justify-center py-12">
                <div className="text-center space-y-4">
                  <Mail className="h-12 w-12 text-muted-foreground mx-auto" />
                  <div>
                    <h3 className="text-lg font-semibold">Email not found</h3>
                    <p className="text-muted-foreground mt-1">
                      The email you're looking for doesn't exist or you don't have permission to view it.
                    </p>
                  </div>
                </div>
              </div>
            </DataCardContent>
          </DataCard>
        </div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-6 py-8 max-w-7xl h-full flex flex-col">
      {/* Page Header */}
      <PageHeader>
        <PageHeaderBreadcrumb>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate('/inbox')}
            className="text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Inbox
          </Button>
        </PageHeaderBreadcrumb>
        
        <PageHeaderContent>
          <div className="space-y-2">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <PageHeaderTitle>{email.subject || '(No Subject)'}</PageHeaderTitle>
              </div>
              <div className="flex items-center gap-2">
                {email.is_converted_to_ticket && (
                  <StatusIndicator status="success" size="sm" showDot={false}>
                    <ExternalLink className="w-3 h-3 mr-1" />
                    Converted
                  </StatusIndicator>
                )}
              </div>
            </div>
          </div>
          
          <PageHeaderActions>
            <Button
              variant="outline"
              size="sm"
              onClick={handleToggleRead}
              disabled={toggling || email.is_read}
            >
              {email.is_read ? (
                <>
                  <Eye className="w-4 h-4 mr-2" />
                  Read
                </>
              ) : (
                <>
                  <EyeOff className="w-4 h-4 mr-2" />
                  Mark Read
                </>
              )}
            </Button>
            
            {!email.is_converted_to_ticket && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowConvertDialog(true)}
              >
                <ExternalLink className="w-4 h-4 mr-2" />
                Convert to Ticket
              </Button>
            )}
            
            <Button variant="default" size="sm" onClick={openReplyDialog}>
              <Reply className="w-4 h-4 mr-2" />
              Reply
            </Button>
            
            <Button variant="outline" size="sm">
              <Forward className="w-4 h-4 mr-2" />
              Forward
            </Button>
          </PageHeaderActions>
        </PageHeaderContent>
      </PageHeader>

      {/* Main Content */}
      <div className="flex-1 min-h-0 mt-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 h-full">
        {/* Email Conversation */}
        <div className="lg:col-span-2 overflow-y-auto min-h-0 space-y-6 pr-1">
          <div className="space-y-4">
            {threadEmails.map((threadEmail) => (
              <DataCard key={threadEmail.id} className={threadEmail.id === email?.id ? 'ring-2 ring-primary/20' : ''}>
                <DataCardHeader>
                  <div className="flex items-start justify-between w-full">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                        <User className="w-5 h-5 text-primary" />
                      </div>
                      <div>
                        <p className="font-medium text-card-foreground">{getFromDisplay(threadEmail)}</p>
                        <p className="text-sm text-muted-foreground">
                          to {threadEmail.mailbox_address}
                        </p>
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Clock className="w-4 h-4" />
                      <span>{formatEmailDate(threadEmail.sent_at || threadEmail.received_at)}</span>
                      {threadEmail.is_reply && (
                        <StatusIndicator status="default" size="sm" showDot={false}>
                          <Reply className="w-3 h-3 ml-1" />
                        </StatusIndicator>
                      )}
                    </div>
                  </div>
                </DataCardHeader>
                
                <DataCardContent>
                  <div className="prose prose-sm max-w-none dark:prose-invert">
                    {threadEmail.body_html ? (
                      <div 
                        dangerouslySetInnerHTML={{ __html: threadEmail.body_html }}
                        className="email-content"
                      />
                    ) : (
                      <pre className="whitespace-pre-wrap font-sans text-sm leading-relaxed text-card-foreground">
                        {threadEmail.body_text || threadEmail.snippet}
                      </pre>
                    )}
                  </div>
                  
                  {/* Attachments */}
                  {threadEmail.has_attachments && (
                    <div className="mt-8 pt-6 border-t border-border/40">
                      <div className="flex items-center gap-2 text-sm font-medium mb-4">
                        <Paperclip className="w-4 h-4" />
                        Attachments ({threadEmail.attachment_count})
                      </div>
                      <div className="rounded-lg bg-muted/50 p-4">
                        <p className="text-sm text-muted-foreground">
                          This email has {threadEmail.attachment_count} attachment(s).
                        </p>
                      </div>
                    </div>
                  )}
                </DataCardContent>
              </DataCard>
            ))}
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6 min-h-0">
          {/* Email Details */}
          <DetailSection>
            <DetailSectionHeader>
              <DetailSectionTitle>Details</DetailSectionTitle>
            </DetailSectionHeader>
            <DetailSectionContent>
              <DetailItem
                label="From"
                value={getFromDisplay(email)}
              />
              <DetailItem
                label="To"
                value={email.mailbox_address}
              />
              <DetailItem
                label="Received"
                value={format(new Date(email.received_at), 'MMM d, yyyy h:mm a')}
              />
              {email.sent_at && (
                <DetailItem
                  label="Sent"
                  value={format(new Date(email.sent_at), 'MMM d, yyyy h:mm a')}
                />
              )}
              
            </DetailSectionContent>
          </DetailSection>

          {/* Linked Ticket */}
          {email.ticket_id && (
            <DetailSection>
              <DetailSectionHeader>
                <DetailSectionTitle>Linked Ticket</DetailSectionTitle>
              </DetailSectionHeader>
              <DetailSectionContent>
                <p className="text-sm text-muted-foreground mb-4">
                  This email has been converted to a ticket.
                </p>
                <Button 
                  variant="default" 
                  size="sm" 
                  className="w-full"
                  onClick={() => navigate(`/tickets/${email.ticket_id}`)}
                >
                  <ExternalLink className="w-4 h-4 mr-2" />
                  View Ticket
                </Button>
              </DetailSectionContent>
            </DetailSection>
          )}
        </div>
        </div>
      </div>

      {/* Convert to Ticket Dialog */}
      <Dialog open={showConvertDialog} onOpenChange={setShowConvertDialog}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Convert Email to Ticket</DialogTitle>
          </DialogHeader>
          
          <div className="space-y-6">
            <div className="space-y-2">
              <label className="text-sm font-medium">Ticket Type</label>
              <Select value={convertForm.type} onValueChange={(value) => setConvertForm(prev => ({ ...prev, type: value }))}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose ticket type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="question">Question</SelectItem>
                  <SelectItem value="incident">Incident</SelectItem>
                  <SelectItem value="problem">Problem</SelectItem>
                  <SelectItem value="task">Task</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="space-y-2">
              <label className="text-sm font-medium">Priority</label>
              <Select value={convertForm.priority} onValueChange={(value) => setConvertForm(prev => ({ ...prev, priority: value }))}>
                <SelectTrigger>
                  <SelectValue placeholder="Choose priority" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="normal">Normal</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="urgent">Urgent</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="flex justify-end gap-3 pt-4 border-t border-border/40">
              <Button
                variant="outline"
                onClick={() => setShowConvertDialog(false)}
                disabled={converting}
              >
                Cancel
              </Button>
              <Button
                onClick={handleConvertToTicket}
                disabled={converting}
              >
                {converting ? 'Converting...' : 'Convert to Ticket'}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Reply Dialog */}
      <Dialog open={showReplyDialog} onOpenChange={setShowReplyDialog}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Reply to Email</DialogTitle>
            {email && (
              <p className="text-sm text-muted-foreground">
                Replying to: <span className="font-medium">{getFromDisplay(email)}</span>
              </p>
            )}
          </DialogHeader>
          
          <div className="space-y-6">
            <div className="space-y-2">
              <label className="text-sm font-medium">Subject</label>
              <input
                type="text"
                value={replyForm.subject}
                onChange={(e) => setReplyForm(prev => ({ ...prev, subject: e.target.value }))}
                className="w-full px-3 py-2 border border-input rounded-md bg-background text-sm"
                placeholder="Reply subject"
              />
            </div>
            
            <div className="space-y-2">
              <label className="text-sm font-medium">CC Addresses (optional)</label>
              <input
                type="text"
                value={replyForm.cc_addresses.join(', ')}
                onChange={(e) => setReplyForm(prev => ({ 
                  ...prev, 
                  cc_addresses: e.target.value.split(',').map(addr => addr.trim()).filter(Boolean)
                }))}
                className="w-full px-3 py-2 border border-input rounded-md bg-background text-sm"
                placeholder="email1@example.com, email2@example.com"
              />
            </div>
            
            <div className="space-y-2">
              <label className="text-sm font-medium">Reply Message</label>
              <textarea
                value={replyForm.body}
                onChange={(e) => setReplyForm(prev => ({ ...prev, body: e.target.value }))}
                className="w-full px-3 py-2 border border-input rounded-md bg-background text-sm min-h-[200px] resize-y"
                placeholder="Type your reply message here..."
                required
              />
            </div>
            
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="is_private"
                checked={replyForm.is_private}
                onChange={(e) => setReplyForm(prev => ({ ...prev, is_private: e.target.checked }))}
                className="rounded border-input"
              />
              <label htmlFor="is_private" className="text-sm">
                Private reply (internal note, won't send external email)
              </label>
            </div>
            
            <div className="flex justify-end gap-3 pt-4 border-t border-border/40">
              <Button
                variant="outline"
                onClick={() => setShowReplyDialog(false)}
                disabled={replying}
              >
                Cancel
              </Button>
              <Button
                onClick={handleReply}
                disabled={replying || !replyForm.body.trim()}
              >
                {replying ? 'Sending...' : 'Send Reply'}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <style>{`
        .email-content {
          max-width: 100%;
          overflow-wrap: break-word;
        }
        .email-content img {
          max-width: 100%;
          height: auto;
        }
        .email-content table {
          width: 100%;
          border-collapse: collapse;
        }
        .email-content blockquote {
          border-left: 3px solid hsl(var(--border));
          padding-left: 1rem;
          margin: 1rem 0;
          color: hsl(var(--muted-foreground));
        }
      `}</style>
    </div>
  )
}
