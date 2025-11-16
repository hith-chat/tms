import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Loader2,
  CheckCircle2,
  XCircle,
  Circle,
  Sparkles,
  Database,
  AlertTriangle,
  Wand
} from 'lucide-react'
import { apiClient } from '../../lib/api'

interface AIBuilderSectionProps {
  onThemeGenerated: (theme: any) => void
  loading: boolean
  error: string | null
  onLoadingChange?: (loading: boolean) => void
  onError?: (error: string | null) => void
  initialUrl?: string
  onUrlChange?: (url: string) => void
}

interface BuilderEvent {
  type: string
  stage?: string
  message?: string
  detail?: string
  timestamp?: string
  data?: Record<string, any>
}

type StageStatus = 'pending' | 'active' | 'complete' | 'error'

const STAGE_ORDER: Array<{ key: string; label: string }> = [
  { key: 'widget', label: 'Chat Widget' },
  { key: 'scraping', label: 'Discovery' },
  { key: 'indexing', label: 'Indexing' },
]

export function AIBuilderSection({
  onThemeGenerated,
  onLoadingChange,
  onError,
  initialUrl,
  onUrlChange
}: AIBuilderSectionProps) {
  const navigate = useNavigate()
  const [urlInput, setUrlInput] = useState(initialUrl || '')
  const [urlError, setUrlError] = useState<string | null>(null)
  const [events, setEvents] = useState<BuilderEvent[]>([])
  const [status, setStatus] = useState<'idle' | 'running' | 'completed' | 'error'>('idle')
  const [widgetId, setWidgetId] = useState<string | null>(null)
  const [widgetThemeData, setWidgetThemeData] = useState<any>(null)
  const [completedData, setCompletedData] = useState<any>(null)
  const [isWidgetReady, setIsWidgetReady] = useState(false)
  const [indexingProgress, setIndexingProgress] = useState<{ current: number; total: number } | null>(null)

  const eventSourceRef = useRef<EventSource | null>(null)
  const timelineRef = useRef<HTMLDivElement | null>(null)

  // Map backend stage names to frontend stage keys
  const mapBackendStageToFrontend = (backendStage: string): string => {
    switch (backendStage) {
      case 'initialization':
      case 'theme':
      case 'widget':
        return 'widget'

      case 'url_extraction':
      case 'scraping':
      case 'ai_ranking':
        return 'scraping'

      case 'embedding_storage':
        return 'indexing'

      default:
        return backendStage
    }
  }

  // Auto-scroll timeline as new events arrive
  useEffect(() => {
    if (!timelineRef.current) return
    timelineRef.current.scrollTo({
      top: timelineRef.current.scrollHeight,
      behavior: 'smooth',
    })
  }, [events])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      eventSourceRef.current?.close()
      eventSourceRef.current = null
    }
  }, [])

  // Calculate stage status from events
  const stageStatus = () => {
    const map = new Map<string, StageStatus>()
    STAGE_ORDER.forEach(({ key }) => map.set(key, 'pending'))

    // Track if stage 3 (indexing) has started to ensure clean transition
    let stage3Started = false

    events.forEach((event) => {
      // Check if stage 3 has started
      if (event.type === 'stage_3_started') {
        stage3Started = true
      }

      if (!event.stage) return

      // Map backend stage to frontend stage
      const frontendStage = mapBackendStageToFrontend(event.stage)
      if (!map.has(frontendStage)) return

      // Check for errors (don't override complete status)
      if (event.type === 'error' || event.type?.endsWith('_error')) {
        // Allow partial success: only mark as error if not already complete
        if (map.get(frontendStage) !== 'complete') {
          map.set(frontendStage, 'error')
        }
        return
      }

      // Mark stages complete based on specific completion events
      const completionEvents: Record<string, string[]> = {
        widget: ['widget_ready', 'theme_generation_completed'],
        scraping: ['stage_2_completed', 'url_extraction_completed'],
        indexing: ['stage_3_completed'],
      }

      if (completionEvents[frontendStage]?.includes(event.type)) {
        map.set(frontendStage, 'complete')
        return
      }

      // Mark as active if it's started but not complete
      if (map.get(frontendStage) === 'pending') {
        map.set(frontendStage, 'active')
      }
    })

    // If stage 3 has started, ensure Discovery is marked complete
    if (stage3Started && map.get('scraping') === 'active') {
      map.set('scraping', 'complete')
    }

    return map
  }

  const validateUrl = (input: string): boolean => {
    try {
      const urlObj = new URL(input.startsWith('http') ? input : `https://${input}`)
      return urlObj.protocol === 'http:' || urlObj.protocol === 'https:'
    } catch {
      return false
    }
  }

  const handleStart = async () => {
    if (status === 'running') return

    const trimmed = urlInput.trim()
    if (!trimmed) {
      setUrlError('Please provide a website URL to analyze')
      return
    }

    const normalizedUrl = trimmed.startsWith('http') ? trimmed : `https://${trimmed}`

    if (!validateUrl(normalizedUrl)) {
      setUrlError('Please enter a valid website URL')
      return
    }

    setUrlError(null)
    setStatus('running')
    setEvents([])
    setWidgetId(null)
    onLoadingChange?.(true)
    onError?.(null)

    try {
      const depth = 3
      const streamUrl = apiClient.getAIBuilderStreamUrl(normalizedUrl, depth)
      const token = localStorage.getItem('auth_token')

      const response = await fetch(streamUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
          Accept: 'text/event-stream',
        },
        body: JSON.stringify({ url: normalizedUrl, max_depth: depth }),
      })

      if (!response.ok || !response.body) {
        throw new Error('Failed to start AI builder stream')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder('utf-8')
      let buffer = ''

      const dispatchEvent = (rawEvent: string) => {
        if (!rawEvent.trim()) return

        const lines = rawEvent.split('\n')
        let eventName = 'message'
        const dataLines: string[] = []

        for (const line of lines) {
          if (line.startsWith('event:')) {
            eventName = line.slice(6).trim()
          } else if (line.startsWith('data:')) {
            dataLines.push(line.slice(5))
          }
        }

        const dataText = dataLines.join('\n').trim()
        if (!dataText) return

        try {
          const payload = JSON.parse(dataText)
          handleBuilderEvent({ type: eventName, ...payload })
        } catch (parseError) {
          console.warn('Failed to parse SSE payload', parseError)
        }
      }

      const processBuffer = (text: string): string => {
        let working = text
        let boundary = working.indexOf('\n\n')
        while (boundary !== -1) {
          const rawEvent = working.slice(0, boundary)
          dispatchEvent(rawEvent)
          working = working.slice(boundary + 2)
          boundary = working.indexOf('\n\n')
        }
        return working
      }

      while (true) {
        const { value, done } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        buffer = processBuffer(buffer)
      }

      if (buffer.length > 0) {
        processBuffer(buffer + '\n\n')
      }
    } catch (err: any) {
      console.error('AI Builder error:', err)
      const errorMsg = err?.message || 'Failed to build widget'
      setUrlError(errorMsg)
      onError?.(errorMsg)
      setStatus('error')
    } finally {
      onLoadingChange?.(false)
    }
  }

  const handleBuilderEvent = (event: BuilderEvent) => {
    setEvents((prev) => [...prev, event])
    console.log('Received builder event:', event)

    // Handle widget_ready event - store data and update URL
    if (event.type === 'widget_ready' && event.data?.widget_id) {
      console.log('Processing widget_ready event 1')
      setWidgetId(event.data.widget_id)
      console.log('Processing widget_ready event 2')
      setIsWidgetReady(true)
      console.log('Processing widget_ready event 3')

      console.log('Widget ready event data:', event.data)

      // Extract complete widget data from backend response
      const widget = event.data.widget
      const mappedWidgetData = widget ? {
        widget_id: event.data.widget_id,
        name: widget.name || '',
        domain_url: widget.domain_url || 'hith.chat',
        welcome_message: widget.welcome_message || 'Hello! How can we help you today?',
        custom_greeting: widget.custom_greeting || 'Hi there! 👋 How can we help you today?',
        away_message: widget.away_message || 'We\'re currently away. Leave us a message and we\'ll get back to you!',
        primary_color: widget.primary_color || '#8b5cf6',
        secondary_color: widget.secondary_color || '#e0e7ff',
        background_color: widget.background_color || '#ffffff',
        position: widget.position || 'bottom-right',
        widget_shape: widget.widget_shape || 'rounded',
        chat_bubble_style: widget.chat_bubble_style || 'modern',
        widget_size: widget.widget_size || 'medium',
        animation_style: widget.animation_style || 'smooth',
        agent_name: widget.agent_name || 'Support Agent',
        agent_avatar_url: widget.agent_avatar_url || '',
        allow_file_uploads: widget.allow_file_uploads || false,
        show_agent_avatars: widget.show_agent_avatars !== false,
        require_email: widget.require_email || false,
        require_name: widget.require_name || false,
        sound_enabled: widget.sound_enabled !== false,
        show_powered_by: widget.show_powered_by !== false,
        use_ai: widget.use_ai || false,
        auto_open_delay: widget.auto_open_delay || 0,
        embed_code: widget.embed_code || ''
      } : event.data

      console.log('Processing widget_ready event 4')
      setWidgetThemeData(mappedWidgetData) // Store for later use
      console.log('Processing widget_ready event 5')

      // Immediately update parent component with widget data for live preview
      // onThemeGenerated(mappedWidgetData)
      console.log('Processing widget_ready event 6')

      // Update URL in browser history without navigation (allows task to continue)
      // Always append ?page=ai to maintain AI builder mode
      window.history.replaceState(
        null,
        '',
        `/chat/widget/edit/${event.data.widget_id}?page=ai`
      )
    }

    // Handle embed_code_ready event
    if (event.type === 'embed_code_ready' && event.data?.widget_id) {
      setWidgetId(event.data.widget_id)
    }

    // Handle storage_progress event for indexing progress counter
    if (event.type === 'storage_progress' && event.message) {
      // Extract "page X/Y" from message like "Stored embeddings for page 3/8"
      const match = event.message.match(/page\s+(\d+)\/(\d+)/i)
      if (match) {
        setIndexingProgress({ current: parseInt(match[1]), total: parseInt(match[2]) })
      }
    }

    // Handle completion - store data but DON'T auto-switch
    if (event.type === 'completed') {
      console.log('Processing widget_ready event 10', event)
      setStatus('completed')
      setCompletedData(event.data)
      onLoadingChange?.(false)
      console.log('Processing widget_ready event 11', event)
      // Note: Don't call onThemeGenerated here - let user click "Continue" button
    }
    console.log('Processing widget_ready event 12', event)

    // Handle errors (allow partial success)
    if (event.type === 'error' || event.type?.endsWith('_error')) {
      // Only set overall error status if no stages have completed yet
      const hasCompletedStages = events.some(
        (e) =>
          e.type === 'widget_ready' ||
          e.type === 'stage_3_completed'
      )

      if (!hasCompletedStages) {
        setStatus('error')
      }

      const errorMsg = event.message || event.detail || 'An error occurred'
      setUrlError(errorMsg)
      onError?.(errorMsg)
      onLoadingChange?.(false)
    }
  }

  const getStatusIcon = (stageKey: string) => {
    const statusMap = stageStatus()
    const status = statusMap.get(stageKey) || 'pending'

    switch (status) {
      case 'complete':
        return <CheckCircle2 className="h-5 w-5 text-green-500" />
      case 'active':
        return <Loader2 className="h-5 w-5 text-blue-500 animate-spin" />
      case 'error':
        return <XCircle className="h-5 w-5 text-red-500" />
      default:
        return <Circle className="h-5 w-5 text-gray-300" />
    }
  }

  const getStatusColor = (stageKey: string) => {
    const statusMap = stageStatus()
    const status = statusMap.get(stageKey) || 'pending'

    switch (status) {
      case 'complete':
        return 'text-green-600'
      case 'active':
        return 'text-blue-600'
      case 'error':
        return 'text-red-600'
      default:
        return 'text-gray-400'
    }
  }

  const handleContinue = () => {
    // Merge widget theme data and completed data
    const finalData = {
      ...widgetThemeData,
      ...completedData,
    }
    onThemeGenerated(finalData)
  }

  const getStageSummary = () => {
    const statusMap = stageStatus()
    const completed = Array.from(statusMap.values()).filter((s) => s === 'complete').length
    const errors = Array.from(statusMap.values()).filter((s) => s === 'error').length
    return { total: STAGE_ORDER.length, completed, errors }
  }

  const getStageLabel = (stageKey: string, baseLabel: string) => {
    // Show page counter for indexing stage when it's active
    if (stageKey === 'indexing' && indexingProgress) {
      return `${baseLabel} (${indexingProgress.current}/${indexingProgress.total})`
    }
    return baseLabel
  }

  return (
    <div className="flex flex-col w-full min-w-0 space-y-6">
      {/* URL Input Card */}
      <div className="rounded-lg border border-border bg-card shadow-sm">
        <div className="flex items-center gap-3 p-6 pb-4">
          <div className="flex flex-col space-y-1">
            <p className="text-sm text-muted-foreground">
              Enter URL to generate - Chat Widget + Knowledge Base
            </p>
          </div>
        </div>

        <div className="px-6 pb-6">
          <div className="space-y-4">
            {/* URL Input */}
            <div className="space-y-2">

              <div className="flex gap-2">
                <div className="relative flex-1">
                  <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                    <Wand className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <input
                    id="website-url"
                    type="url"
                    value={urlInput}
                    onChange={(e) => {
                      const newUrl = e.target.value
                      setUrlInput(newUrl)
                      setUrlError(null)
                      // Notify parent component of URL change for live preview
                      onUrlChange?.(newUrl)
                    }}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' && status === 'idle') {
                        handleStart()
                      }
                    }}
                    placeholder="https://example.com"
                    className={`flex h-10 w-full rounded-md border bg-background pl-10 pr-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 ${
                      urlError ? 'border-destructive' : 'border-input'
                    }`}
                    disabled={status === 'running'}
                  />
                </div>
                <button
                  type="button"
                  onClick={handleStart}
                  disabled={status === 'running' || !urlInput.trim()}
                  className="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium bg-gradient-primary text-white hover:opacity-90 h-10 px-6 disabled:pointer-events-none disabled:opacity-50 shadow-glow transition-all"
                >
                  {status === 'running' ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      Building...
                    </>
                  ) : (
                    <>
                      Start Building
                    </>
                  )}
                </button>
              </div>
              {urlError && (
                <div className="flex items-center gap-2 text-sm text-destructive">
                  <AlertTriangle className="h-4 w-4" />
                  {urlError}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Widget Ready Notification - Shows when widget is created while task continues */}
      {isWidgetReady && widgetId && status === 'running' && (
        <div className="rounded-lg border border-green-200 bg-green-50 dark:bg-green-900/10 dark:border-green-800 shadow-sm p-4">
          <div className="flex items-start gap-3">
            <CheckCircle2 className="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
            <div className="flex-1 min-w-0">
              <h4 className="text-sm font-semibold text-green-900 dark:text-green-100">
                Chat Widget Created!
              </h4>
              <p className="text-sm text-green-700 dark:text-green-200 mt-1">
                Your widget is ready and the live preview has been updated. Knowledge base building continues in the background.
              </p>
            </div>
            <button
              type="button"
              onClick={() => navigate(`/chat/widget/edit/${widgetId}?page=ai`)}
              className="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium bg-green-600 text-white hover:bg-green-700 h-9 px-4 transition-colors"
            >
              <Sparkles className="h-4 w-4" />
              Go to Settings
            </button>
          </div>
        </div>
      )}

      {/* Progress Steps */}
      {status !== 'idle' && status !== 'completed' && (
        <div className="rounded-lg border border-border bg-card shadow-sm p-6">
          <div className="flex items-center gap-3 mb-4">
            <Database className="h-5 w-5 text-primary" />
            <h3 className="text-lg font-semibold">Build Progress</h3>
          </div>

          {/* Stage Summary - Same horizontal grid as Build Complete */}
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 mb-6">
            {STAGE_ORDER.map((stage) => {
              const statusMap = stageStatus()
              const stageState = statusMap.get(stage.key) || 'pending'
              return (
                <div
                  key={stage.key}
                  className="flex items-center gap-2 rounded-lg border border-border bg-background p-3"
                >
                  {getStatusIcon(stage.key)}
                  <div className="flex flex-col min-w-0">
                    <span className="text-xs font-medium truncate">{getStageLabel(stage.key, stage.label)}</span>
                    <span className={`text-xs ${getStatusColor(stage.key)} capitalize`}>
                      {stageState}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>

          {/* Event Timeline */}
          {events.length > 0 && (
            <div className="pt-6 border-t border-border">
              <h4 className="text-sm font-medium mb-3">Activity Log</h4>
              <div
                ref={timelineRef}
                className="space-y-2 max-h-48 overflow-y-auto text-xs text-muted-foreground custom-scrollbar"
              >
                {events
                  .filter((event) => !event.type?.includes('faq')) // Filter out FAQ-related events
                  .map((event, idx) => (
                    <div key={`${event.timestamp}-${idx}`} className="flex items-start gap-2">
                      <span className="text-xs text-muted-foreground whitespace-nowrap">
                        {new Date(event.timestamp || Date.now()).toLocaleTimeString()}
                      </span>
                      <span className="flex-1">{event.message || event.type}</span>
                    </div>
                  ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Completion Card */}
      {status === 'completed' && (
        <div className="rounded-lg border border-border bg-card shadow-sm p-6">
          <div className="flex items-center gap-3 mb-4">
            <CheckCircle2 className="h-6 w-6 text-green-500" />
            <div>
              <h3 className="text-lg font-semibold">Build Complete!</h3>
              <p className="text-sm text-muted-foreground">
                {(() => {
                  const { completed, errors, total } = getStageSummary()
                  if (errors > 0) {
                    return `${completed}/${total} stages completed with ${errors} ${errors === 1 ? 'warning' : 'warnings'}`
                  }
                  return `All ${total} stages completed successfully`
                })()}
              </p>
            </div>
          </div>

          {/* Stage Summary */}
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 mb-6">
            {STAGE_ORDER.map((stage) => {
              const statusMap = stageStatus()
              const stageState = statusMap.get(stage.key) || 'pending'
              return (
                <div
                  key={stage.key}
                  className="flex items-center gap-2 rounded-lg border border-border bg-background p-3"
                >
                  {getStatusIcon(stage.key)}
                  <div className="flex flex-col min-w-0">
                    <span className="text-xs font-medium truncate">{getStageLabel(stage.key, stage.label)}</span>
                    <span className={`text-xs ${getStatusColor(stage.key)} capitalize`}>
                      {stageState}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>

          {/* Continue Button */}
          <button
            type="button"
            onClick={handleContinue}
            className="w-full inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium bg-gradient-primary text-white hover:opacity-90 h-10 px-6 transition-all shadow-glow"
          >
            <Sparkles className="h-4 w-4" />
            Continue to Widget Settings
          </button>
        </div>
      )}
    </div>
  )
}
