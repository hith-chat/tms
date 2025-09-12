import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, MessageCircle, Globe, Settings, Copy, Trash2, Eye, EyeOff, AlertCircle, CheckCircle } from 'lucide-react'
import { apiClient } from '../lib/api'
import type { ChatWidget } from '../types/chat'
import type { DomainValidation } from '../lib/api'

export function ChatWidgetsPage() {
  const navigate = useNavigate()
  const [widgets, setWidgets] = useState<ChatWidget[]>([])
  const [domains, setDomains] = useState<DomainValidation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  useEffect(() => {
    loadData()
  }, [])

  // Redirect effect after widgets are loaded
  useEffect(() => {
    if (!loading && widgets.length > 0) {
      // If there are widgets, redirect to the first widget
      navigate(`/chat/widget/edit/${widgets[0].id}`)
    } else if (!loading && widgets.length === 0) {
      // If no widgets but domains are verified, redirect to create widget page
      navigate('/chat/widget/create')
    }
  }, [loading, widgets, domains, navigate])

  const loadData = async () => {
    try {
      setLoading(true)
      const [widgetsData, domainsData] = await Promise.all([
        apiClient.listChatWidgets(),
        apiClient.getDomainValidations()
      ])
      setWidgets(widgetsData)
      setDomains(domainsData.filter(d => d.status === 'verified'))
    } catch (err: any) {
      setError(err.message || 'Failed to load data')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteWidget = async (widgetId: string) => {
    if (!confirm('Are you sure you want to delete this chat widget?')) return

    try {
      await apiClient.deleteChatWidget(widgetId)
      setWidgets(prev => prev.filter(w => w.id !== widgetId))
      setSuccessMessage('Widget deleted successfully')
      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err: any) {
      setError(`Failed to delete widget: ${err.message}`)
      setTimeout(() => setError(null), 5000)
    }
  }

  const handleToggleActive = async (widget: ChatWidget) => {
    try {
      const updatedWidget = await apiClient.updateChatWidget(widget.id, { is_active: !widget.is_active })
      setWidgets(prev => prev.map(w => w.id === widget.id ? updatedWidget : w))
      setSuccessMessage(`Widget ${updatedWidget.is_active ? 'activated' : 'deactivated'}`)
      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err: any) {
      setError(`Failed to update widget: ${err.message}`)
      setTimeout(() => setError(null), 5000)
    }
  }

  const copyEmbedCode = (embedCode: string) => {
    navigator.clipboard.writeText(embedCode)
    setSuccessMessage('Embed code copied to clipboard')
    setTimeout(() => setSuccessMessage(null), 3000)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col bg-background">
      {/* Compact Header */}
      <div >
        {/* Compact Alerts */}
        {error && (
          <div className="mt-2 flex items-center gap-2 p-2 m-2 rounded-md bg-destructive/10 border border-destructive/20">
            <AlertCircle className="h-4 w-4 text-destructive flex-shrink-0" />
            <p className="text-xs text-destructive truncate">{error}</p>
          </div>
        )}
        
        {successMessage && (
          <div className="mt-2 flex items-center gap-2 p-2 m-2 rounded-md bg-emerald-50 border border-emerald-200 dark:bg-emerald-950/50 dark:border-emerald-800">
            <CheckCircle className="h-4 w-4 text-emerald-600 flex-shrink-0" />
            <p className="text-xs text-emerald-700 dark:text-emerald-300 truncate">{successMessage}</p>
          </div>
        )}

        {/* {domains.length === 0 && (
          <div className="mt-2 flex items-center gap-2 p-2 rounded-md bg-amber-50 border border-amber-200 dark:bg-amber-950/50 dark:border-amber-800">
            <AlertCircle className="h-4 w-4 text-amber-600 flex-shrink-0" />
            <div className="flex-1 min-w-0">
              <p className="text-xs text-amber-700 dark:text-amber-300">
                Domain verification required. <a href="/settings?tab=domains" className="underline hover:no-underline font-medium">Go to Settings</a>
              </p>
            </div>
          </div>
        )} */}
      </div>

      {/* Content Area */}
      <div className="flex-1 overflow-auto">
        {widgets.length > 0 || domains.length > 0 ? (
          <div className="p-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
              {/* Create Widget Card */}
              {domains.length > 0 && (
                <div className="group bg-card rounded-lg border-2 border-dashed border-border hover:border-primary/50 p-3 hover:shadow-md transition-all duration-200 h-fit">
                  {/* Header Section */}
                  <div className="flex items-center justify-center mb-3">
                    <div className="w-8 h-8 rounded-md bg-primary/10 flex items-center justify-center">
                      <Plus className="h-4 w-4 text-primary" />
                    </div>
                  </div>

                  {/* Content Section */}
                  <div className="text-center mb-3">
                    <h3 className="text-sm font-medium text-foreground mb-1">Create New Widget</h3>
                    <p className="text-xs text-muted-foreground">
                      Add a chat widget to engage with your visitors
                    </p>
                  </div>

                  {/* Essential Info Placeholder */}
                  <div className="space-y-2 mb-3">
                    <div className="flex items-center justify-between text-xs">
                      <span className="text-muted-foreground">&nbsp;</span>
                      <span className="text-foreground font-medium">&nbsp;</span>
                    </div>
                    
                  </div>

                  {/* Action Section */}
                  <div className="pt-2 border-t border-border">
                    <button
                      onClick={() => navigate('/chat/widget/create')}
                      className="w-full inline-flex items-center justify-center gap-1.5 px-2 py-1.5 rounded-md text-xs font-medium bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                    >
                      <Plus className="h-3 w-3" />
                      Create Widget
                    </button>
                  </div>
                </div>
              )}
              
              {widgets.map((widget) => (
                <WidgetCard
                  key={widget.id}
                  widget={widget}
                  onEdit={() => navigate(`/chat/widget/edit/${widget.id}`)}
                  onDelete={() => handleDeleteWidget(widget.id)}
                  onToggleActive={() => handleToggleActive(widget)}
                  onCopyEmbed={() => widget.embed_code && copyEmbedCode(widget.embed_code)}
                />
              ))}
            </div>
          </div>
        ) : domains.length > 0 ? (
          <div className="flex-1 flex items-center justify-center px-4">
            <div className="text-center max-w-sm">
              <div className="mx-auto w-12 h-12 rounded-full bg-muted flex items-center justify-center mb-3">
                <MessageCircle className="h-6 w-6 text-muted-foreground" />
              </div>
              <h3 className="text-sm font-medium text-foreground mb-1">No chat widgets</h3>
              <p className="text-xs text-muted-foreground mb-4">
                Create your first widget to start engaging visitors
              </p>
              <button
                onClick={() => navigate('/chat/widget/create')}
                className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                <Plus className="h-4 w-4" />
                Create Widget
              </button>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  )
}

interface WidgetCardProps {
  widget: ChatWidget
  onEdit: () => void
  onDelete: () => void
  onToggleActive: () => void
  onCopyEmbed: () => void
}

function WidgetCard({ widget, onEdit, onDelete, onToggleActive, onCopyEmbed }: WidgetCardProps) {
  return (
    <div className="group bg-card rounded-lg border border-border p-3 hover:shadow-md transition-all duration-200 h-fit">
      {/* Compact Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <div 
            className="w-8 h-8 rounded-md flex items-center justify-center flex-shrink-0"
            style={{ backgroundColor: widget.primary_color }}
          >
            <MessageCircle className="h-4 w-4 text-white" />
          </div>
          <div className="min-w-0 flex-1">
            <h3 className="font-medium text-sm text-foreground truncate">{widget.name}</h3>
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <Globe className="h-3 w-3 flex-shrink-0" />
              <span className="truncate">{widget.domain_name}</span>
            </div>
          </div>
        </div>
        
        {/* Status Indicator */}
        <div className="flex items-center gap-1">
          <div className="flex items-center gap-0.5">
            <div className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
          widget.is_active 
            ? 'bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-950/50 dark:text-emerald-300 dark:border-emerald-800'
            : 'bg-muted text-muted-foreground border border-border'
        }`}>
          <div className={`w-1 h-1 rounded-full ${widget.is_active ? 'bg-emerald-500' : 'bg-muted-foreground'}`} />
          {widget.is_active ? 'Active' : 'Inactive'}
        </div>
            <button
              onClick={onToggleActive}
              className="p-1 hover:bg-muted rounded transition-colors"
              title={widget.is_active ? 'Deactivate' : 'Activate'}
            >
              {widget.is_active ? <Eye className="h-5 w-5" /> : <EyeOff className="h-5 w-5" />}
            </button>            
            <button
              onClick={onDelete}
              className="p-1 hover:bg-destructive/10 text-destructive/70 hover:text-destructive rounded transition-colors"
              title="Delete"
            >
              <Trash2 className="h-5 w-5" />
            </button>
          </div>
        </div>
      </div>

      {/* Essential Info */}
      <div className="space-y-2 mb-3">
        <div className="flex items-center justify-between text-xs">
          <span className="text-muted-foreground">Position</span>
          <span className="text-foreground font-medium capitalize">{widget.position.replace('-', ' ')}</span>
        </div>
        <div className="flex items-center justify-between text-xs">
          <span className="text-muted-foreground">Auto-open</span>
          <span className="text-foreground font-medium">
            {widget.auto_open_delay > 0 ? `${widget.auto_open_delay}s` : 'Off'}
          </span>
        </div>
        <div className="flex items-center justify-between text-xs">
          <span className="text-muted-foreground">AI Enabled</span>
          <span className="text-foreground font-medium">
            {widget.use_ai ? 'On' : 'Off'}
          </span>
        </div>
      </div>

      {/* Compact Actions */}
      <div className="flex items-center gap-1.5 pt-2 border-t border-border">
        <button
          onClick={onCopyEmbed}
          className="flex-1 inline-flex items-center justify-center gap-1.5 px-2 py-1.5 rounded-md text-xs font-medium bg-muted hover:bg-muted/80 text-foreground transition-colors"
        >
          <Copy className="h-3 w-3" />
          Copy
        </button>
        <button
          onClick={onEdit}
          className="inline-flex items-center gap-1.5 px-2 py-1.5 rounded-md text-xs font-medium bg-primary/10 hover:bg-primary/20 text-primary transition-colors"
        >
          <Settings className="h-3 w-3" />
          Edit
        </button>
      </div>
    </div>
  )
}
