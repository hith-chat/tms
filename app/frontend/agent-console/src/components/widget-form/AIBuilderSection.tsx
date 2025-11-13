import { useState, useEffect, useRef } from 'react'
import {
  Wand2,
  Globe,
  Loader2,
  CheckCircle2,
  XCircle,
  Circle,
  Sparkles,
  Database,
  Link2,
  AlertTriangle
} from 'lucide-react'
import { apiClient } from '../../lib/api'

interface AIBuilderSectionProps {
  onThemeGenerated: (theme: any) => void
  loading: boolean
  error: string | null
  onLoadingChange?: (loading: boolean) => void
  onError?: (error: string | null) => void
  initialUrl?: string
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
  { key: 'faq', label: 'Knowledge Q&A' },
]

export function AIBuilderSection({
  onThemeGenerated,
  onLoadingChange,
  onError,
  initialUrl
}: AIBuilderSectionProps) {
  const [urlInput, setUrlInput] = useState(initialUrl || '')
  const [urlError, setUrlError] = useState<string | null>(null)
  const [events, setEvents] = useState<BuilderEvent[]>([])
  const [status, setStatus] = useState<'idle' | 'running' | 'completed' | 'error'>('idle')
  const [widgetId, setWidgetId] = useState<string | null>(null)
  const [iframeError, setIframeError] = useState(false)
  const [iframeLoaded, setIframeLoaded] = useState(false)

  const eventSourceRef = useRef<EventSource | null>(null)
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const timelineRef = useRef<HTMLDivElement | null>(null)

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

    events.forEach((event) => {
      if (!event.stage) return
      if (!map.has(event.stage)) return

      if (event.type === 'error' || event.type?.endsWith('_error')) {
        map.set(event.stage, 'error')
        return
      }

      if (
        event.type === 'completed' ||
        event.type?.endsWith('_completed') ||
        event.type === 'faq_ready' ||
        event.type === 'widget_ready' ||
        event.type === 'knowledge_links_chosen'
      ) {
        map.set(event.stage, 'complete')
        return
      }

      map.set(event.stage, 'active')
    })

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
    setIframeError(false)
    setIframeLoaded(false)
    onLoadingChange?.(true)
    onError?.(null)

    try {
      const streamUrl = apiClient.getAIBuilderStreamUrl()
      const token = localStorage.getItem('auth_token')
      const depth = 3

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

    // Handle widget_ready event
    if (event.type === 'widget_ready' && event.data?.widget_id) {
      setWidgetId(event.data.widget_id)
      onThemeGenerated(event.data)
    }

    // Handle embed_code_ready event
    if (event.type === 'embed_code_ready' && event.data?.widget_id) {
      setWidgetId(event.data.widget_id)
    }

    // Handle completion
    if (event.type === 'completed') {
      setStatus('completed')
      onLoadingChange?.(false)
    }

    // Handle errors
    if (event.type === 'error' || event.type?.endsWith('_error')) {
      setStatus('error')
      const errorMsg = event.message || event.detail || 'An error occurred'
      setUrlError(errorMsg)
      onError?.(errorMsg)
      onLoadingChange?.(false)
    }
  }

  const handleIframeLoad = () => {
    setIframeLoaded(true)
    setIframeError(false)
  }

  const handleIframeError = () => {
    setIframeError(true)
    setIframeLoaded(false)
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

  return (
    <div className="flex flex-col w-full min-w-0 space-y-6">
      {/* URL Input Card */}
      <div className="rounded-lg border border-border bg-card shadow-sm">
        <div className="flex items-center gap-3 p-6 pb-4">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-primary">
            <Wand2 className="h-5 w-5 text-white" />
          </div>
          <div className="flex flex-col space-y-1">
            <h3 className="text-lg font-semibold gradient-text">AI Widget Builder</h3>
            <p className="text-sm text-muted-foreground">
              Enter your website URL to automatically generate a Chat Widget + Knowledge Base + FAQs
            </p>
          </div>
        </div>

        <div className="px-6 pb-6">
          <div className="space-y-4">
            {/* URL Input */}
            <div className="space-y-2">
              <label htmlFor="website-url" className="text-sm font-medium">
                Website URL <span className="text-destructive">*</span>
              </label>
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                    <Globe className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <input
                    id="website-url"
                    type="url"
                    value={urlInput}
                    onChange={(e) => {
                      setUrlInput(e.target.value)
                      setUrlError(null)
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
                      <Sparkles className="h-4 w-4" />
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

      {/* Progress Steps */}
      {status !== 'idle' && (
        <div className="rounded-lg border border-border bg-card shadow-sm p-6">
          <div className="flex items-center gap-3 mb-6">
            <Database className="h-5 w-5 text-primary" />
            <h3 className="text-lg font-semibold">Build Progress</h3>
          </div>

          <div className="space-y-4">
            {STAGE_ORDER.map((stage) => (
              <div key={stage.key} className="flex items-center gap-3">
                {getStatusIcon(stage.key)}
                <span className={`text-sm font-medium ${getStatusColor(stage.key)}`}>
                  {stage.label}
                </span>
              </div>
            ))}
          </div>

          {/* Event Timeline */}
          {events.length > 0 && (
            <div className="mt-6 pt-6 border-t border-border">
              <h4 className="text-sm font-medium mb-3">Activity Log</h4>
              <div
                ref={timelineRef}
                className="space-y-2 max-h-48 overflow-y-auto text-xs text-muted-foreground custom-scrollbar"
              >
                {events.map((event, idx) => (
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

      {/* Preview iframe */}
      {widgetId && urlInput && (
        <div className="rounded-lg border border-border bg-card shadow-sm p-6">
          <div className="flex items-center gap-3 mb-4">
            <Link2 className="h-5 w-5 text-primary" />
            <h3 className="text-lg font-semibold">Live Preview</h3>
          </div>

          <div className="relative rounded-lg overflow-hidden border border-border bg-background">
            {!iframeLoaded && !iframeError && (
              <div className="absolute inset-0 flex items-center justify-center bg-background">
                <Loader2 className="h-6 w-6 animate-spin text-primary" />
              </div>
            )}

            {iframeError ? (
              <div className="h-[600px] flex flex-col items-center justify-center bg-gradient-to-br from-background to-muted/20 p-8">
                <AlertTriangle className="h-12 w-12 text-amber-500 mb-4" />
                <h3 className="text-lg font-semibold mb-2">Preview Not Available</h3>
                <p className="text-sm text-muted-foreground text-center max-w-md mb-6">
                  The website cannot be displayed in an iframe due to security restrictions (X-Frame-Options).
                  Your chat widget will still work when embedded on the actual website.
                </p>
                <div className="w-full max-w-2xl rounded-lg border border-border bg-card p-6">
                  <p className="text-sm text-center text-muted-foreground">
                    Widget is ready! Copy the embed code from the next step to add it to your website.
                  </p>
                </div>
              </div>
            ) : (
              <iframe
                ref={iframeRef}
                src={urlInput.startsWith('http') ? urlInput : `https://${urlInput}`}
                className="w-full h-[600px]"
                onLoad={handleIframeLoad}
                onError={handleIframeError}
                sandbox="allow-same-origin allow-scripts allow-popups allow-forms"
                title="Website Preview"
              />
            )}
          </div>

          <p className="text-xs text-muted-foreground mt-3 flex items-center gap-2">
            <Sparkles className="h-3 w-3" />
            Your chat widget is being loaded on the preview. It may take a few seconds to appear.
          </p>
        </div>
      )}
    </div>
  )
}
