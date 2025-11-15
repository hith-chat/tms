import { useEffect, useMemo, useRef, useState, ChangeEvent } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  Button,
  Input,
  Badge
} from '@tms/shared'
import { apiClient } from '../lib/api'
import {
  Sparkles,
  Globe,
  Loader2,
  CheckCircle2,
  XCircle,
  Circle,
  Link2,
  ClipboardCopy,
  CopyCheck,
} from 'lucide-react'
import { cn } from '../lib/utils'

interface AIBuilderModalProps {
  open: boolean
  onClose: () => void
  defaultUrl?: string
  onCompleted?: () => void
}

interface BuilderEvent {
  type: string
  stage?: string
  message?: string
  detail?: string
  timestamp?: string
  data?: Record<string, any>
}

interface KnowledgeLinkSelectionView {
  url: string
  category?: string
  rationale?: string
}

type BuilderStatus = 'idle' | 'running' | 'completed' | 'error'

type StageStatus = 'pending' | 'active' | 'complete' | 'error'

const STAGE_ORDER: Array<{ key: string; label: string }> = [
  { key: 'widget', label: 'Chat Widget' },
  { key: 'scraping', label: 'Discovery' },
  { key: 'indexing', label: 'Indexing' },
  { key: 'faq', label: 'Knowledge Q&A' },
]

export function AIBuilderModal({ open, onClose, defaultUrl, onCompleted }: AIBuilderModalProps) {
  const [urlInput, setUrlInput] = useState(defaultUrl || '')
  const depth = '3' // Fixed crawl depth
  const [events, setEvents] = useState<BuilderEvent[]>([])
  const [status, setStatus] = useState<BuilderStatus>('idle')
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [embedCode, setEmbedCode] = useState<string | null>(null)
  const [selectedLinks, setSelectedLinks] = useState<KnowledgeLinkSelectionView[]>([])
  const [faqCount, setFaqCount] = useState<number | null>(null)
  const [copied, setCopied] = useState(false)

  const eventSourceRef = useRef<EventSource | null>(null)
  const streamAbortRef = useRef<AbortController | null>(null)
  const timelineRef = useRef<HTMLDivElement | null>(null)

  // Reset state when modal opens fresh
  useEffect(() => {
    if (open) {
      setErrorMessage(null)
      setCopied(false)
      if (status === 'idle') {
        setEvents([])
        setEmbedCode(null)
        setSelectedLinks([])
        setFaqCount(null)
      }
    } else {
      eventSourceRef.current?.close()
      eventSourceRef.current = null
      setStatus('idle')
    }
  }, [open, status])

  // Auto-scroll timeline as new events arrive
  useEffect(() => {
    if (!open || !timelineRef.current) return
    timelineRef.current.scrollTo({
      top: timelineRef.current.scrollHeight,
      behavior: 'smooth',
    })
  }, [events, open])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      eventSourceRef.current?.close()
      eventSourceRef.current = null
      if (streamAbortRef.current) {
        streamAbortRef.current.abort()
        streamAbortRef.current = null
      }
    }
  }, [])

  const stageStatus = useMemo(() => {
    const map = new Map<string, StageStatus>()
    STAGE_ORDER.forEach(({ key }) => map.set(key, 'pending'))

    events.forEach((event) => {
      if (!event.stage) return
      if (!map.has(event.stage)) return

      if (event.type === 'error' || event.type?.endsWith('_error')) {
        map.set(event.stage, 'error')
        return
      }

      if (event.type === 'completed' || event.type?.endsWith('_completed') || event.type === 'faq_ready' || event.type === 'widget_ready' || event.type === 'knowledge_links_chosen') {
        map.set(event.stage, 'complete')
        return
      }

      map.set(event.stage, 'active')
    })

    return map
  }, [events])

  const handleStart = async () => {
    if (status === 'running') {
      return
    }

    setCopied(false)
    setErrorMessage(null)
    setEmbedCode(null)
    setSelectedLinks([])
    setFaqCount(null)

    const trimmed = urlInput.trim()
    if (!trimmed) {
      setErrorMessage('Please provide a website URL to analyze.')
      return
    }

    let parsed: URL
    try {
      parsed = new URL(trimmed)
    } catch (_err) {
      setErrorMessage('URL looks invalid. Include http(s):// prefix (e.g., https://example.com).')
      return
    }

    if (!['http:', 'https:'].includes(parsed.protocol)) {
      setErrorMessage('Only HTTPS or HTTP URLs are supported.')
      return
    }

    const depthNumber = Number(depth)

    const streamUrl = apiClient.getAIBuilderStreamUrl(trimmed, Number.isFinite(depthNumber) ? depthNumber : 3)

    setStatus('running')
    setEvents([])

    // Use fetch-based streaming so we can send Authorization header (EventSource doesn't support headers)
    const controller = new AbortController()
    streamAbortRef.current = controller

    const token = localStorage.getItem('auth_token')

    try {
      const response = await fetch(streamUrl, {
        method: 'GET',
        headers: {
          Accept: 'text/event-stream',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        signal: controller.signal,
      })

      if (!response.ok || !response.body) {
        throw new Error(`Failed to connect to AI builder stream: ${response.status} ${response.statusText}`)
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder('utf-8')
      let buffer = ''

      const flushEvent = (raw: string) => {
        // parse SSE event block, extract lines beginning with "data:"
        const lines = raw.split(/\r?\n/)
        const dataLines: string[] = []
        for (const line of lines) {
          if (line.startsWith('data:')) {
            dataLines.push(line.slice(5))
          }
        }
        if (dataLines.length === 0) return
        const dataStr = dataLines.join('\n').trim()
        if (!dataStr) return
        try {
          const payload: BuilderEvent = JSON.parse(dataStr)
          setEvents((prev) => [...prev, payload])

          if (payload.data?.selected_links) {
            setSelectedLinks(payload.data.selected_links as KnowledgeLinkSelectionView[])
          }
          if (payload.data?.faq_count !== undefined) {
            setFaqCount(payload.data.faq_count as number)
          }
          if (payload.data?.embed_code) {
            setEmbedCode(payload.data.embed_code as string)
          }

          if (payload.type === 'completed') {
            setStatus('completed')
            controller.abort()
            streamAbortRef.current = null
            onCompleted?.()
          }

          if (payload.type === 'error') {
            setStatus('error')
            setErrorMessage(payload.detail || payload.message || 'AI builder encountered an error.')
            controller.abort()
            streamAbortRef.current = null
          }
        } catch (err) {
          console.error('Failed to parse AI builder event', err)
        }
      }

      const processBuffer = (text: string) => {
        // SSE events are separated by a blank line
        let boundary = text.indexOf('\n\n')
        while (boundary === -1) {
          boundary = text.indexOf('\r\n\r\n')
          if (boundary !== -1) break
          break
        }

        // If we have at least one event block, process them
        if (boundary !== -1) {
          const chunk = text.slice(0, boundary)
          const rest = text.slice(boundary + (text[boundary] === '\r' ? 4 : 2))
          flushEvent(chunk)
          return rest
        }
        return text
      }

      while (true) {
        const { value, done } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        // try to process as many complete events as possible
        let prev
        do {
          prev = buffer
          buffer = processBuffer(buffer)
        } while (buffer !== prev && buffer.length > 0)
      }

      // process any remaining buffered data
      if (buffer.length > 0) {
        flushEvent(buffer)
      }
    } catch (err: any) {
      if (err?.name === 'AbortError') {
        // expected when we cancel the stream
      } else {
        console.error('AI builder stream error', err)
        setStatus('error')
        setErrorMessage(err?.message || 'Connection to AI builder stream ended unexpectedly.')
      }
    } finally {
      if (streamAbortRef.current) {
        streamAbortRef.current = null
      }
    }
  }

  const handleClose = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    setStatus('idle')
    onClose()
  }

  const handleCopyEmbed = async () => {
    if (!embedCode) return
    try {
      await navigator.clipboard.writeText(embedCode)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch (err) {
      console.error('Failed to copy embed code', err)
    }
  }

  const stageBadgeIcon = (state: StageStatus) => {
    switch (state) {
      case 'complete':
        return <CheckCircle2 className="h-4 w-4 text-emerald-500" />
      case 'error':
        return <XCircle className="h-4 w-4 text-destructive" />
      case 'active':
        return <Loader2 className="h-4 w-4 animate-spin text-primary" />
      default:
        return <Circle className="h-4 w-4 text-muted-foreground" />
    }
  }

  const stageBadgeClass = (state: StageStatus) => {
    switch (state) {
      case 'complete':
        return 'bg-emerald-500/10 text-emerald-600 border-emerald-500/40'
      case 'error':
        return 'bg-destructive/10 text-destructive border-destructive/40'
      case 'active':
        return 'bg-primary/10 text-primary border-primary/40'
      default:
        return 'bg-muted text-muted-foreground border-transparent'
    }
  }

  const renderEventTime = (timestamp?: string) => {
    if (!timestamp) return ''
    const date = new Date(timestamp)
    if (Number.isNaN(date.getTime())) return ''
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  }

  return (
  <Dialog open={open} onOpenChange={(value: boolean) => (value ? null : handleClose())}>
      <DialogContent className="max-w-3xl overflow-hidden">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-xl">
            <Sparkles className="h-5 w-5 text-primary" />
            AI Builder Copilot
          </DialogTitle>
          <DialogDescription>
            Give the builder your website URL and it will craft the chat widget.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          <div className="flex flex-col gap-3 rounded-lg border border-border/60 bg-muted/30 p-4">
            <div className="flex flex-col md:flex-row md:items-end md:gap-3">
              <div className="flex-1">
                <label className="text-xs font-medium text-muted-foreground">Homepage URL</label>
                <div className="mt-1 flex items-center gap-2">
                  <Globe className="h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="https://your-company.com"
                    value={urlInput}
                    onChange={(event: ChangeEvent<HTMLInputElement>) => setUrlInput(event.target.value)}
                    disabled={status === 'running'}
                  />
                </div>
              </div>
              <Button
                onClick={handleStart}
                disabled={status === 'running'}
                className="mt-2 w-full md:w-auto"
              >
                {status === 'running' ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Building...
                  </>
                ) : (
                  <>
                    <Sparkles className="mr-2 h-4 w-4" />
                    Start Build
                  </>
                )}
              </Button>
            </div>
            {errorMessage && (
              <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {errorMessage}
              </div>
            )}
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {STAGE_ORDER.map(({ key, label }) => {
              const state = stageStatus.get(key) ?? 'pending'
              return (
                <div
                  key={key}
                  className={cn(
                    'flex items-center gap-3 rounded-lg border px-3 py-2 transition-colors',
                    state === 'active' && 'border-primary/40 bg-primary/5',
                    state === 'complete' && 'border-emerald-500/40 bg-emerald-500/5',
                    state === 'error' && 'border-destructive/40 bg-destructive/5'
                  )}
                >
                  <div className="flex h-9 w-9 items-center justify-center rounded-full bg-background">
                    {stageBadgeIcon(state)}
                  </div>
                  <div className="flex flex-col">
                    <span className="text-xs text-muted-foreground">{label}</span>
                    <Badge variant="outline" className={cn('w-fit px-2 py-0 text-xs capitalize', stageBadgeClass(state))}>
                      {state}
                    </Badge>
                  </div>
                </div>
              )
            })}
          </div>

          <div>
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-sm font-medium text-muted-foreground">Live progress</h3>
              <span className="text-xs text-muted-foreground">{events.length} updates</span>
            </div>
            <div
              ref={timelineRef}
              className="h-56 overflow-y-auto rounded-lg border border-border/60 bg-card/50 px-3 py-4"
            >
              {events.length === 0 ? (
                <p className="text-sm text-muted-foreground">Once started, you can watch the builder work step-by-step here.</p>
              ) : (
                <ul className="space-y-3">
                  {events.map((event, index) => (
                    <li key={`${event.type}-${index}`} className="flex gap-3">
                      <div className="mt-1">
                        <Circle className="h-3 w-3 text-primary" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium text-foreground">
                            {event.message || event.type}
                          </span>
                          {event.stage && (
                            <Badge variant="secondary" className="uppercase tracking-wide text-[10px]">
                              {event.stage}
                            </Badge>
                          )}
                          <span className="text-xs text-muted-foreground">
                            {renderEventTime(event.timestamp)}
                          </span>
                        </div>
                        {event.detail && (
                          <p className="mt-1 text-xs text-muted-foreground">{event.detail}</p>
                        )}
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>

          {selectedLinks.length > 0 && (
            <div className="rounded-lg border border-border/60 bg-card/60 p-4">
              <h3 className="text-sm font-semibold text-foreground">Selected Knowledge Pages</h3>
              <p className="text-xs text-muted-foreground mb-3">Top URLs chosen automatically for indexing.</p>
              <ul className="space-y-2">
                {selectedLinks.map((link) => (
                  <li key={link.url} className="flex items-start gap-2 text-sm">
                    <Link2 className="mt-1 h-4 w-4 text-primary" />
                    <div>
                      <a
                        href={link.url}
                        target="_blank"
                        rel="noreferrer"
                        className="text-primary hover:underline"
                      >
                        {link.url}
                      </a>
                      <div className="text-xs text-muted-foreground">
                        {[link.category, link.rationale].filter(Boolean).join(' · ')}
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {embedCode && (
            <div className="rounded-lg border border-border/60 bg-card/60 p-4">
              <div className="flex items-center justify-between gap-2 mb-3">
                <div>
                  <h3 className="text-sm font-semibold text-foreground">Embed Code</h3>
                  <p className="text-xs text-muted-foreground">
                    Copy this single line and paste it before the closing &lt;/body&gt; tag in your HTML
                  </p>
                </div>
                <Button variant="outline" size="sm" onClick={handleCopyEmbed}>
                  {copied ? (
                    <>
                      <CopyCheck className="mr-1.5 h-4 w-4" /> Copied
                    </>
                  ) : (
                    <>
                      <ClipboardCopy className="mr-1.5 h-4 w-4" /> Copy
                    </>
                  )}
                </Button>
              </div>
              <div className="rounded-md bg-muted/80 p-3 border border-border/40">
                <code className="text-xs text-foreground font-mono break-all">{embedCode}</code>
              </div>
            </div>
          )}

          {faqCount !== null && (
            <div className="rounded-lg border border-border/60 bg-muted/40 px-4 py-3 text-sm text-muted-foreground">
              Generated <span className="font-medium text-foreground">{faqCount}</span> Q&A items for your knowledge base.
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
