import { useState, useEffect, useRef } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { User, Globe, Loader, CheckCircle, XCircle } from 'lucide-react'
import { apiClient, KnowledgeDocument, KnowledgeScrapingJob, KnowledgeFAQItem, ScrapedLinkPreview } from '../lib/api'
import { AboutMeTab } from './knowledge/AboutMeTab'
import { KnowledgeBaseTab } from './knowledge/KnowledgeBaseTab'
import { KBUrlsSection } from './knowledge/KBUrlsSection'
import {
  IndexingProgressState,
  ScrapingProgressState,
  INITIAL_INDEXING_STATE,
  INITIAL_SCRAPING_STATE
} from './knowledge/types'

interface KnowledgeManagementProps {
  projectId: string | null
}

type KnowledgeTab = 'about-me' | 'knowledge-base' | 'kb-urls'

export function KnowledgeManagement({ projectId }: KnowledgeManagementProps) {
  const navigate = useNavigate()
  const location = useLocation()

  // Derive active tab from URL
  const getActiveTabFromPath = (): KnowledgeTab => {
    const path = location.pathname
    if (path.includes('/kb-jobs')) {
      return 'knowledge-base'
    }
    if (path.includes('/kb-urls')) {
      return 'kb-urls'
    }
    return 'about-me' // Default to about-me
  }

  const activeTab = getActiveTabFromPath()

  // Data state
  const [documents, setDocuments] = useState<KnowledgeDocument[]>([])
  const [scrapingJobs, setScrapingJobs] = useState<KnowledgeScrapingJob[]>([])
  const [faqItems, setFaqItems] = useState<KnowledgeFAQItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  // About Me state
  const [aboutMeContent, setAboutMeContent] = useState('')

  // Scraping state
  const [scrapingInProgress, setScrapingInProgress] = useState(false)
  const [activeScrapingJob, setActiveScrapingJob] = useState<KnowledgeScrapingJob | null>(null)
  const [stagedLinks, setStagedLinks] = useState<ScrapedLinkPreview[]>([])
  const [stagedLinksLoading, setStagedLinksLoading] = useState(false)
  const [stagedLinksError, setStagedLinksError] = useState<string | null>(null)
  const [maxSelectableLinks, setMaxSelectableLinks] = useState<number>(10)
  const [selectedLinkUrls, setSelectedLinkUrls] = useState<Set<string>>(new Set<string>())
  const [indexingProgress, setIndexingProgress] = useState<IndexingProgressState>(INITIAL_INDEXING_STATE)
  const indexingEventSourceRef = useRef<EventSource | null>(null)
  const [scrapingProgress, setScrapingProgress] = useState<ScrapingProgressState>(INITIAL_SCRAPING_STATE)
  const scrapingStreamAbortRef = useRef<AbortController | null>(null)
  const [_streamingJobId, setStreamingJobId] = useState<string | null>(null)
  const streamingJobRef = useRef<KnowledgeScrapingJob | null>(null)

  // Sidebar tabs
  const tabs = [
    // { id: 'about-me' as KnowledgeTab, name: 'About Me', icon: User, path: '/knowledge/about-me' },
    { id: 'knowledge-base' as KnowledgeTab, name: 'Scraping Jobs', icon: Globe, path: '/knowledge/kb-jobs' },
    { id: 'kb-urls' as KnowledgeTab, name: 'KB URLs', icon: Globe, path: '/knowledge/kb-urls' }
  ]

  // Navigate to tab
  const handleTabChange = (tab: typeof tabs[0]) => {
    navigate(tab.path)
  }

  // Load data on mount and project change
  useEffect(() => {
    loadData()
  }, [projectId])

  // Polling for active jobs
  useEffect(() => {
    const hasActiveJobs = scrapingJobs.some(
      (job) => job.status === 'running' || job.status === 'pending' || job.status === 'indexing'
    )

    if (!hasActiveJobs) return

    const pollInterval = setInterval(() => {
      loadData()
    }, 3000)

    return () => clearInterval(pollInterval)
  }, [scrapingJobs])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (indexingEventSourceRef.current) {
        indexingEventSourceRef.current.close()
        indexingEventSourceRef.current = null
      }
      if (scrapingStreamAbortRef.current) {
        scrapingStreamAbortRef.current.abort()
        scrapingStreamAbortRef.current = null
      }
    }
  }, [])

  // Close active scraping selection if project changes
  useEffect(() => {
    if (activeScrapingJob && projectId && activeScrapingJob.project_id !== projectId) {
      closeScrapingSelection()
    }
  }, [projectId, activeScrapingJob])

  const loadData = async () => {
    if (!projectId) {
      console.warn('No project ID provided to KnowledgeManagement component')
      return
    }

    setLoading(true)
    setError(null)
    try {
      const [documentsData, jobsData, aboutMeData, faqData] = await Promise.all([
        apiClient.getDocuments(projectId),
        apiClient.getScrapingJobs(projectId),
        apiClient.getAboutMeSettings().catch(() => ({ content: '' })),
        apiClient.getKnowledgeFAQ(projectId).catch(() => [])
      ])
      setDocuments(documentsData || [])
      setScrapingJobs(jobsData || [])
      setFaqItems(faqData || [])
      setAboutMeContent(aboutMeData?.content || '')
    } catch (err: any) {
      console.error('Error loading knowledge data:', err)
      setError(`Failed to load knowledge base data: ${err?.response?.data?.error || err?.message || 'Unknown error'}`)
    } finally {
      setLoading(false)
    }
  }

  const deleteDocument = async (documentId: string) => {
    if (!projectId) {
      setError('No project selected')
      return
    }
    if (!confirm('Are you sure you want to delete this document?')) return

    try {
      await apiClient.deleteDocument(projectId, documentId)
      setSuccessMessage('Document deleted successfully')
      setTimeout(() => setSuccessMessage(null), 3000)
      await loadData()
    } catch (err) {
      setError('Failed to delete document')
      setTimeout(() => setError(null), 5000)
      console.error('Delete error:', err)
    }
  }

  const createScrapingJob = async (url: string, depth: number) => {
    if (!projectId) {
      setError('No project selected')
      return
    }

    setScrapingInProgress(true)
    setError(null)

    try {
      await apiClient.createScrapingJob(projectId, {
        url,
        max_depth: depth
      })
      setSuccessMessage('Scraping job created successfully')
      setTimeout(() => setSuccessMessage(null), 3000)
      await loadData()
    } catch (err) {
      setError('Failed to create scraping job')
      setTimeout(() => setError(null), 5000)
      console.error('Scraping error:', err)
    } finally {
      setScrapingInProgress(false)
    }
  }

  const createScrapingJobWithStream = async (url: string, depth: number) => {
    if (!projectId) {
      setError('No project selected')
      return
    }

    setScrapingInProgress(true)
    setError(null)
    resetScrapingSession()

    const controller = new AbortController()
    scrapingStreamAbortRef.current = controller

    try {
      const streamUrl = apiClient.getScrapingJobStreamUrl()
      const token = localStorage.getItem('auth_token')
      const response = await fetch(streamUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
          Accept: 'text/event-stream'
        },
        body: JSON.stringify({ url, max_depth: depth }),
        signal: controller.signal
      })

      if (!response.ok || !response.body) {
        throw new Error('Failed to open scraping stream')
      }

      setScrapingProgress({
        status: 'running',
        jobId: null,
        linksFound: 0,
        currentDepth: 0,
        maxDepth: depth,
        logs: [
          {
            type: 'info',
            message: `Starting to scrape ${url}`,
            timestamp: new Date().toISOString()
          }
        ]
      })

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

          if (eventName === 'job_created') {
            const job = payload as KnowledgeScrapingJob
            streamingJobRef.current = job
            setStreamingJobId(job.id)
            setScrapingJobs((current) => {
              const existingIndex = current.findIndex((existing) => existing.id === job.id)
              if (existingIndex !== -1) {
                const copy = [...current]
                copy[existingIndex] = job
                return copy
              }
              return [job, ...current]
            })
            return
          }

          if (eventName) {
            handleScrapingEvent(eventName, payload)
          }
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
      if (err?.name === 'AbortError') {
        console.info('Scraping stream aborted by user')
        setStreamingJobId(null)
        streamingJobRef.current = null
      } else {
        console.error('Scraping error:', err)
        setError(err?.response?.data?.error || err?.message || 'Failed to create scraping job')
        setTimeout(() => setError(null), 5000)
        setScrapingProgress((prev) => ({ ...prev, status: 'error' }))
      }
    } finally {
      scrapingStreamAbortRef.current = null
      setScrapingInProgress(false)
    }
  }

  const handleScrapingEvent = (eventType: string, payload: any) => {
    const timestamp = payload?.timestamp || new Date().toISOString()
    const message = describeScrapingEvent(eventType, payload)

    setScrapingProgress((prev) => {
      const linksFound = typeof payload?.links_found === 'number' ? payload.links_found : prev.linksFound
      const currentDepth = typeof payload?.current_depth === 'number' ? payload.current_depth : prev.currentDepth
      const maxDepth = typeof payload?.max_depth === 'number' ? payload.max_depth : prev.maxDepth
      let status: ScrapingProgressState['status'] = prev.status

      if (eventType === 'started' || eventType === 'link_found' || eventType === 'visiting') {
        status = 'running'
      } else if (eventType === 'completed') {
        status = 'completed'
      } else if (eventType === 'error') {
        status = 'error'
      }

      return {
        status,
        jobId: payload?.job_id ?? prev.jobId,
        linksFound,
        currentDepth,
        maxDepth,
        logs: [
          ...prev.logs,
          {
            type: eventType,
            message,
            url: payload?.url,
            current_depth: payload?.current_depth,
            max_depth: payload?.max_depth,
            links_found: payload?.links_found,
            timestamp
          }
        ]
      }
    })

    if (eventType === 'completed') {
      cancelScrapingStream()
      setSuccessMessage('Website scraping completed successfully')
      setTimeout(() => setSuccessMessage(null), 3000)
      loadData()
      setScrapingInProgress(false)
      setStreamingJobId(null)
      streamingJobRef.current = null
    } else if (eventType === 'error') {
      cancelScrapingStream()
      setError(payload?.message || 'Scraping failed')
      setTimeout(() => setError(null), 5000)
      setScrapingInProgress(false)
      setStreamingJobId(null)
      streamingJobRef.current = null
    }
  }

  const describeScrapingEvent = (eventType: string, payload: any) => {
    switch (eventType) {
      case 'job_created':
        return `Scraping job created for ${payload?.url || 'website'}`
      case 'started':
        return `Started scraping ${payload?.url || 'website'}`
      case 'visiting':
        return `Visiting: ${payload?.url || 'page'} (depth ${payload?.current_depth || 0})`
      case 'link_found':
        return `Discovered link: ${payload?.url || payload?.message || 'New link'}`
      case 'completed':
        return `Scraping completed! Found ${payload?.links_found || 0} links`
      case 'error':
        return payload?.message || 'Scraping encountered an error'
      case 'warning':
        return payload?.message || 'Warning during scraping'
      default:
        return eventType
    }
  }

  const resetScrapingSession = () => {
    cancelScrapingStream()
    setStreamingJobId(null)
    streamingJobRef.current = null
    setScrapingProgress(INITIAL_SCRAPING_STATE)
  }

  // Simplified scraping that directly indexes URLs into knowledge base
  const scrapeAndIndexURLs = async (url: string, _depth: number) => {
    if (!projectId) {
      setError('No project selected')
      return
    }

    setScrapingInProgress(true)
    setError(null)

    try {
      const result = await apiClient.scrapeURLs([url], false)

      if (result.pages_added > 0) {
        setSuccessMessage(`Successfully scraped and indexed ${result.pages_added} page(s)`)
      } else if (result.pages_skipped > 0) {
        setSuccessMessage(`Page was already indexed recently (${result.pages_skipped} skipped)`)
      } else {
        setError('No pages could be scraped from this URL')
      }

      setTimeout(() => setSuccessMessage(null), 5000)
      await loadData()
    } catch (err) {
      console.error('Scraping error:', err)
      setError(err instanceof Error ? err.message : 'Failed to scrape URL')
      setTimeout(() => setError(null), 5000)
    } finally {
      setScrapingInProgress(false)
    }
  }

  const cancelScrapingStream = () => {
    if (scrapingStreamAbortRef.current) {
      scrapingStreamAbortRef.current.abort()
      scrapingStreamAbortRef.current = null
    }
  }

  const openScrapingJobLinks = async (job: KnowledgeScrapingJob) => {
    setActiveScrapingJob(job)
    setStagedLinks([])
    setStagedLinksError(null)
    setStagedLinksLoading(true)
    resetIndexingSession()

    try {
      const { links, maxSelectableLinks: limit } = await apiClient.getScrapingJobLinks(job.id)
      const safeLimit = limit || 10
      setMaxSelectableLinks(safeLimit)

      const initialSelection = new Set<string>()
      const preselected = links.filter((link) => link.selected)
      const selectionPool = (preselected.length > 0 ? preselected : links).slice(0, safeLimit)
      selectionPool.forEach((link) => initialSelection.add(link.url))
      setSelectedLinkUrls(initialSelection)

      const updatedLinks = links.map((link) => ({
        ...link,
        selected: initialSelection.has(link.url)
      }))
      setStagedLinks(updatedLinks)

      if ((preselected.length > 0 ? preselected.length : links.length) > safeLimit) {
        setStagedLinksError(
          `Only ${safeLimit} link${safeLimit === 1 ? '' : 's'} can be selected. The first ${safeLimit} are pre-selected.`
        )
      } else {
        setStagedLinksError(null)
      }
    } catch (err: any) {
      console.error('Failed to load staged links', err)
      const message = err?.response?.data?.error || err?.message || 'Failed to load scraped links'
      setStagedLinksError(message)
      setMaxSelectableLinks(10)
      setSelectedLinkUrls(new Set<string>())
    } finally {
      setStagedLinksLoading(false)
    }
  }

  const closeScrapingSelection = () => {
    resetIndexingSession()
    setActiveScrapingJob(null)
    setStagedLinks([])
    setSelectedLinkUrls(new Set<string>())
    setStagedLinksError(null)
    setMaxSelectableLinks(10)
  }

  const toggleLinkSelection = (url: string) => {
    setSelectedLinkUrls((prev) => {
      const next = new Set(prev)
      if (next.has(url)) {
        next.delete(url)
        if (next.size <= maxSelectableLinks) {
          setStagedLinksError(null)
        }
      } else {
        if (next.size >= maxSelectableLinks) {
          setStagedLinksError(`You can select up to ${maxSelectableLinks} link${maxSelectableLinks === 1 ? '' : 's'}.`)
          return prev
        }
        next.add(url)
        setStagedLinksError(null)
      }
      return next
    })
  }

  const selectAllLinks = () => {
    const urls = stagedLinks.map((link) => link.url)
    const limited = urls.slice(0, maxSelectableLinks)
    if (urls.length > maxSelectableLinks) {
      setStagedLinksError(
        `Only ${maxSelectableLinks} link${maxSelectableLinks === 1 ? '' : 's'} can be selected. Selected the first ${maxSelectableLinks}.`
      )
    } else {
      setStagedLinksError(null)
    }
    setSelectedLinkUrls(new Set(limited))
  }

  const clearLinkSelection = () => {
    setSelectedLinkUrls(new Set<string>())
    setStagedLinksError(null)
  }

  const startIndexing = async () => {
    if (!activeScrapingJob) {
      setStagedLinksError('No scraping job selected')
      return
    }

    const urls = Array.from(selectedLinkUrls)
    if (urls.length === 0) {
      setStagedLinksError('Select at least one link to index')
      return
    }
    if (urls.length > maxSelectableLinks) {
      setStagedLinksError(
        `You can only index up to ${maxSelectableLinks} link${maxSelectableLinks === 1 ? '' : 's'} at a time`
      )
      return
    }

    setStagedLinksError(null)

    try {
      const selectionResponse = await apiClient.selectScrapingJobLinks(activeScrapingJob.id, urls)
      setMaxSelectableLinks(selectionResponse.maxSelectableLinks)
      setStagedLinks((current) => current.map((link) => ({ ...link, selected: urls.includes(link.url) })))
      setScrapingJobs((current) =>
        current.map((job) =>
          job.id === activeScrapingJob.id ? { ...job, status: 'indexing', selected_links: urls } : job
        )
      )
      setActiveScrapingJob((current) => (current ? { ...current, status: 'indexing', selected_links: urls } : current))

      const streamUrl = apiClient.getScrapingJobIndexStreamUrl(activeScrapingJob.id)

      if (indexingEventSourceRef.current) {
        indexingEventSourceRef.current.close()
      }

      const source = new EventSource(streamUrl)
      indexingEventSourceRef.current = source

      setIndexingProgress({
        status: 'running',
        total: urls.length,
        completed: 0,
        pending: urls.length,
        totalTokens: 0,
        logs: [
          {
            type: 'info',
            message: `Submitted ${urls.length} link(s) for indexing`,
            timestamp: new Date().toISOString()
          }
        ]
      })

      const eventsToHandle = ['started', 'warning', 'skipped', 'embedding_started', 'embedding_completed', 'completed', 'error']
      eventsToHandle.forEach((eventName) => {
        source.addEventListener(eventName, (event: MessageEvent) => {
          try {
            const payload = JSON.parse(event.data)
            handleIndexingEvent(eventName, payload)
          } catch (_err) {
            handleIndexingEvent('warning', { message: `Failed to parse ${eventName} event payload` })
          }
        })
      })

      source.onerror = () => {
        source.close()
        indexingEventSourceRef.current = null
        handleIndexingEvent('error', { message: 'Connection lost while streaming indexing progress' })
      }
    } catch (err: any) {
      console.error('Failed to start indexing', err)
      const message = err?.response?.data?.error || err?.message || 'Failed to start indexing'
      setStagedLinksError(message)
    }
  }

  const handleIndexingEvent = (eventType: string, payload: any) => {
    const timestamp = payload?.timestamp || new Date().toISOString()
    const message = describeProgressEvent(eventType, payload)

    setIndexingProgress((prev) => {
      const total = payload?.total ?? prev.total
      let completed = payload?.completed ?? prev.completed
      let pending = payload?.pending ?? Math.max(total - completed, 0)
      const totalTokens = payload?.total_tokens ?? prev.totalTokens
      let status: IndexingProgressState['status'] = prev.status

      if (eventType === 'started') {
        completed = payload?.completed ?? 0
        pending = payload?.pending ?? Math.max(total - completed, 0)
        status = 'running'
      } else if (eventType === 'completed') {
        completed = payload?.completed ?? total
        pending = 0
        status = 'completed'
      } else if (eventType === 'error') {
        status = 'error'
      }

      return {
        status,
        total,
        completed,
        pending,
        totalTokens,
        logs: [...prev.logs, { type: eventType, message, url: payload?.url, timestamp }]
      }
    })

    if (eventType === 'completed') {
      closeIndexingEventSource()
      setSuccessMessage('Website indexing completed successfully')
      setTimeout(() => setSuccessMessage(null), 3000)
      loadData()
    } else if (eventType === 'error') {
      closeIndexingEventSource()
      setError(payload?.message || 'Indexing failed')
      setTimeout(() => setError(null), 5000)
    }
  }

  const describeProgressEvent = (eventType: string, payload: any) => {
    switch (eventType) {
      case 'started':
        return `Indexing started for ${payload?.total ?? 'selected'} link(s)`
      case 'warning':
        return payload?.message || 'Warning received during indexing'
      case 'skipped':
        return `Skipped ${payload?.url ?? 'link'} (duplicate content)`
      case 'embedding_started':
        return `Generating embeddings for ${payload?.total ?? 0} page(s)`
      case 'embedding_completed':
        return `Embeddings completed for ${payload?.completed ?? 0} page(s)`
      case 'completed':
        return 'Indexing completed successfully'
      case 'error':
        return payload?.message || 'Indexing encountered an error'
      default:
        return eventType
    }
  }

  const closeIndexingEventSource = () => {
    if (indexingEventSourceRef.current) {
      indexingEventSourceRef.current.close()
      indexingEventSourceRef.current = null
    }
  }

  const resetIndexingSession = () => {
    closeIndexingEventSource()
    setIndexingProgress(INITIAL_INDEXING_STATE)
  }

  // Render tab content
  const renderTabContent = () => {
    switch (activeTab) {
      case 'about-me':
        return (
          <AboutMeTab
            projectId={projectId}
            initialContent={aboutMeContent}
            onContentChange={setAboutMeContent}
            onSaveSuccess={(msg) => {
              setSuccessMessage(msg)
              setTimeout(() => setSuccessMessage(null), 3000)
            }}
            onSaveError={(err) => {
              setError(err)
              setTimeout(() => setError(null), 5000)
            }}
          />
        )
      case 'knowledge-base':
        return (
          <KnowledgeBaseTab
            projectId={projectId}
            documents={documents}
            scrapingJobs={scrapingJobs}
            faqItems={faqItems}
            scrapingInProgress={scrapingInProgress}
            scrapingProgress={scrapingProgress}
            activeScrapingJob={activeScrapingJob}
            stagedLinks={stagedLinks}
            stagedLinksLoading={stagedLinksLoading}
            stagedLinksError={stagedLinksError}
            maxSelectableLinks={maxSelectableLinks}
            selectedLinkUrls={selectedLinkUrls}
            indexingProgress={indexingProgress}
            onStartScraping={scrapeAndIndexURLs}
            onStartLegacyScraping={createScrapingJob}
            onDeleteDocument={deleteDocument}
            onOpenScrapingJobLinks={openScrapingJobLinks}
            onCloseScrapingSelection={closeScrapingSelection}
            onToggleLinkSelection={toggleLinkSelection}
            onSelectAllLinks={selectAllLinks}
            onClearLinkSelection={clearLinkSelection}
            onStartIndexing={startIndexing}
          />
        )
      case 'kb-urls':
        return projectId ? <KBUrlsSection projectId={projectId} /> : null
      default:
        return null
    }
  }

  if (!projectId) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Please select a project to manage knowledge base.</p>
      </div>
    )
  }

  if (loading && documents.length === 0 && scrapingJobs.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader className="h-8 w-8 animate-spin text-primary" />
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Error/Success Messages */}
      {(error || successMessage) && (
        <div className="p-4 space-y-3">
          {error && (
            <div className="border rounded-lg p-4 bg-destructive/5 border-destructive/20">
              <div className="flex items-start">
                <XCircle className="h-5 w-5 text-destructive mr-3 mt-0.5 flex-shrink-0" />
                <p className="text-sm text-destructive">{error}</p>
              </div>
            </div>
          )}

          {successMessage && (
            <div className="border rounded-lg p-4 bg-green-50 border-green-200">
              <div className="flex items-start">
                <CheckCircle className="h-5 w-5 text-green-600 mr-3 mt-0.5 flex-shrink-0" />
                <p className="text-sm text-green-800">{successMessage}</p>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Content with Sidebar */}
      <div className="flex-1 flex min-h-0">
        {/* Sidebar Navigation */}
        <div className="w-64 flex-shrink-0 border-r border-border bg-card/50">
          <nav className="p-4 space-y-1">
            {tabs.map((tab) => {
              const Icon = tab.icon
              return (
                <button
                  key={tab.id}
                  onClick={() => handleTabChange(tab)}
                  className={`w-full flex items-center space-x-3 px-3 py-2 text-left rounded-md transition-colors focus-visible-ring ${
                    activeTab === tab.id
                      ? 'bg-primary/10 text-primary border-l-2 border-primary'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
                  }`}
                  aria-current={activeTab === tab.id ? 'page' : undefined}
                >
                  <Icon className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
                  <span className="text-sm font-medium">{tab.name}</span>
                </button>
              )
            })}
          </nav>
        </div>

        {/* Main Content */}
        <div className="flex-1 overflow-hidden">
          <div className="h-full overflow-y-auto custom-scrollbar">
            <div className="p-6">{renderTabContent()}</div>
          </div>
        </div>
      </div>
    </div>
  )
}
