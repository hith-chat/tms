import { useState, useEffect, useRef } from 'react'
import { 
  Upload, 
  FileText, 
  Globe, 
  Trash2, 
  Loader, 
  CheckCircle, 
  XCircle, 
  AlertCircle,
  Plus,
  Search,
  Save,
  X
} from 'lucide-react'
import { apiClient, KnowledgeDocument, KnowledgeScrapingJob, ScrapedLinkPreview } from '../lib/api'

interface KnowledgeManagementProps {
  projectId: string | null
}

interface IndexingLogEntry {
  type: string
  message: string
  url?: string
  timestamp: string
}

interface IndexingProgressState {
  status: 'idle' | 'running' | 'completed' | 'error'
  total: number
  completed: number
  pending: number
  totalTokens: number
  logs: IndexingLogEntry[]
}

interface ScrapingLogEntry {
  type: string
  message: string
  url?: string
  current_depth?: number
  max_depth?: number
  links_found?: number
  timestamp: string
}

interface ScrapingProgressState {
  status: 'idle' | 'running' | 'completed' | 'error'
  jobId: string | null
  linksFound: number
  currentDepth: number
  maxDepth: number
  logs: ScrapingLogEntry[]
}

const INITIAL_INDEXING_STATE: IndexingProgressState = {
  status: 'idle',
  total: 0,
  completed: 0,
  pending: 0,
  totalTokens: 0,
  logs: []
}

const INITIAL_SCRAPING_STATE: ScrapingProgressState = {
  status: 'idle',
  jobId: null,
  linksFound: 0,
  currentDepth: 0,
  maxDepth: 0,
  logs: []
}

export function KnowledgeManagement({ projectId }: KnowledgeManagementProps) {
  // State for documents
  const [documents, setDocuments] = useState<KnowledgeDocument[]>([])
  const [scrapingJobs, setScrapingJobs] = useState<KnowledgeScrapingJob[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  // Upload state
  const [dragOver, setDragOver] = useState(false)
  const [uploading, setUploading] = useState(false)

  // Scraping state
  const [showScrapingForm, setShowScrapingForm] = useState(false)
  const [scrapingUrl, setScrapingUrl] = useState('')
  const [scrapingDepth, setScrapingDepth] = useState(3)
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
  const [streamingJobId, setStreamingJobId] = useState<string | null>(null)
  const streamingJobRef = useRef<KnowledgeScrapingJob | null>(null)

  // Search state
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<any>(null)
  const [searchLoading, setSearchLoading] = useState(false)

  // About Me state
  const [aboutMeContent, setAboutMeContent] = useState('')
  const [aboutMeSaving, setAboutMeSaving] = useState(false)

  useEffect(() => {
    loadData()
  }, [projectId])

  // Polling effect for updating running/pending jobs
  useEffect(() => {
    const hasActiveJobs = scrapingJobs.some(job => 
      job.status === 'running' || job.status === 'pending' || job.status === 'indexing'
    )
    
    if (!hasActiveJobs) return

    const pollInterval = setInterval(() => {
      loadData()
    }, 3000) // Poll every 3 seconds

    return () => clearInterval(pollInterval)
  }, [scrapingJobs])

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

  const loadData = async () => {
    if (!projectId) {
      console.warn('No project ID provided to KnowledgeManagement component')
      return
    }
    
    setLoading(true)
    setError(null)
    try {
      console.log('Loading knowledge data for project:', projectId)
      const [documentsData, jobsData, aboutMeData] = await Promise.all([
        apiClient.getDocuments(projectId),
        apiClient.getScrapingJobs(projectId),
        apiClient.getAboutMeSettings().catch(() => ({ content: '' })) // Default to empty if not found
      ])
      console.log('Successfully loaded:', { documents: documentsData?.length, jobs: jobsData?.length, aboutMe: aboutMeData?.content?.length })
      setDocuments(documentsData || [])
      setScrapingJobs(jobsData || [])
      setAboutMeContent(aboutMeData?.content || '')
    } catch (err: any) {
      console.error('Error loading knowledge data:', {
        error: err,
        message: err?.message,
        response: err?.response?.data,
        status: err?.response?.status
      })
      setError(`Failed to load knowledge base data: ${err?.response?.data?.error || err?.message || 'Unknown error'}`)
    } finally {
      setLoading(false)
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(true)
  }

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
  }

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    
    const files = Array.from(e.dataTransfer.files)
    await uploadFiles(files)
  }

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || [])
    await uploadFiles(files)
  }

  const uploadFiles = async (files: File[]) => {
    if (!projectId) {
      setError('No project selected')
      return
    }

    const pdfFiles = files.filter(file => file.type === 'application/pdf')
    
    if (pdfFiles.length === 0) {
      setError('Please select PDF files only')
      return
    }

    setUploading(true)
    setError(null)

    try {
      for (const file of pdfFiles) {
        await apiClient.uploadDocument(projectId, file)
      }
      setSuccessMessage(`Successfully uploaded ${pdfFiles.length} document(s)`)
      await loadData()
    } catch (err) {
      setError('Failed to upload documents')
      console.error('Upload error:', err)
    } finally {
      setUploading(false)
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
      await loadData()
    } catch (err) {
      setError('Failed to delete document')
      console.error('Delete error:', err)
    }
  }

  const createScrapingJob = async () => {
    if (!projectId) {
      setError('No project selected')
      return
    }
    if (!scrapingUrl.trim()) {
      setError('Please enter a valid URL')
      return
    }

    setScrapingInProgress(true)
    setError(null)

    try {
      await apiClient.createScrapingJob(projectId, {
        url: scrapingUrl,
        max_depth: scrapingDepth
      })
      setSuccessMessage('Scraping job created successfully')
      setScrapingUrl('')
      setShowScrapingForm(false)
      await loadData()
    } catch (err) {
      setError('Failed to create scraping job')
      console.error('Scraping error:', err)
    } finally {
      setScrapingInProgress(false)
    }
  }

  const createScrapingJobWithStream = async () => {
    if (!projectId) {
      setError('No project selected')
      return
    }
    if (!scrapingUrl.trim()) {
      setError('Please enter a valid URL')
      return
    }

    setScrapingInProgress(true)
    setError(null)
    resetScrapingSession()

    const controller = new AbortController()
    scrapingStreamAbortRef.current = controller

    const appendLog = (eventType: string, payload: any) => {
      const timestamp = payload?.timestamp || new Date().toISOString()
      const message = describeScrapingEvent(eventType, payload)
      setScrapingProgress(prev => ({
        status: prev.status === 'idle' ? 'running' : prev.status,
        jobId: prev.jobId,
        linksFound: prev.linksFound,
        currentDepth: prev.currentDepth,
        maxDepth: prev.maxDepth,
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
      }))
    }

    try {
      const streamUrl = apiClient.getScrapingJobStreamUrl()
      const token = localStorage.getItem('auth_token')
      const response = await fetch(streamUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
          Accept: 'text/event-stream'
        },
        body: JSON.stringify({
          url: scrapingUrl,
          max_depth: scrapingDepth
        }),
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
        maxDepth: scrapingDepth,
        logs: [
          {
            type: 'info',
            message: `Starting to scrape ${scrapingUrl}`,
            timestamp: new Date().toISOString()
          }
        ]
      })

      setScrapingUrl('')
      setShowScrapingForm(false)

      const reader = response.body.getReader()
      const decoder = new TextDecoder('utf-8')
      let buffer = ''

      const dispatchEvent = (rawEvent: string) => {
        if (!rawEvent.trim()) {
          return
        }

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
        if (!dataText) {
          return
        }

        try {
          const payload = JSON.parse(dataText)

          if (eventName === 'job_created') {
            const job = payload as KnowledgeScrapingJob
            streamingJobRef.current = job
            setStreamingJobId(job.id)
            setScrapingProgress(prev => ({
              status: 'running',
              jobId: job.id,
              linksFound: prev.linksFound,
              currentDepth: prev.currentDepth,
              maxDepth: job.max_depth ?? prev.maxDepth,
              logs: [
                ...prev.logs,
                {
                  type: 'job_created',
                  message: describeScrapingEvent('job_created', job),
                  timestamp: new Date().toISOString()
                }
              ]
            }))
            setScrapingJobs(current => {
              const existingIndex = current.findIndex(existing => existing.id === job.id)
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
          } else {
            appendLog('message', payload)
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
        if (done) {
          break
        }
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
        setScrapingProgress(prev => ({
          ...prev,
          status: 'error'
        }))
      }
    } finally {
      scrapingStreamAbortRef.current = null
      setScrapingInProgress(false)
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

  const cancelScrapingStream = () => {
    if (scrapingStreamAbortRef.current) {
      scrapingStreamAbortRef.current.abort()
      scrapingStreamAbortRef.current = null
    }
  }

  const resetScrapingSession = () => {
    cancelScrapingStream()
    setStreamingJobId(null)
    streamingJobRef.current = null
    setScrapingProgress(INITIAL_SCRAPING_STATE)
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
      const preselected = links.filter(link => link.selected)
      const selectionPool = (preselected.length > 0 ? preselected : links).slice(0, safeLimit)
      selectionPool.forEach(link => initialSelection.add(link.url))
      setSelectedLinkUrls(initialSelection)

      const updatedLinks = links.map(link => ({
        ...link,
        selected: initialSelection.has(link.url)
      }))
      setStagedLinks(updatedLinks)

      if ((preselected.length > 0 ? preselected.length : links.length) > safeLimit) {
        setStagedLinksError(`Only ${safeLimit} link${safeLimit === 1 ? '' : 's'} can be selected. The first ${safeLimit} are pre-selected.`)
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
    setSelectedLinkUrls(prev => {
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
    const urls = stagedLinks.map(link => link.url)
    const limited = urls.slice(0, maxSelectableLinks)
    if (urls.length > maxSelectableLinks) {
      setStagedLinksError(`Only ${maxSelectableLinks} link${maxSelectableLinks === 1 ? '' : 's'} can be selected. Selected the first ${maxSelectableLinks}.`)
    } else {
      setStagedLinksError(null)
    }
    setSelectedLinkUrls(new Set(limited))
  }

  const clearLinkSelection = () => {
    setSelectedLinkUrls(new Set<string>())
    setStagedLinksError(null)
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

  const handleScrapingEvent = (eventType: string, payload: any) => {
    const eventJobId = payload?.job_id || streamingJobId
    if (streamingJobId && eventJobId && eventJobId !== streamingJobId) {
      return
    }

    if (!streamingJobId && payload?.job_id) {
      setStreamingJobId(payload.job_id)
    }

    const timestamp = payload?.timestamp || new Date().toISOString()
    const message = describeScrapingEvent(eventType, payload)
    const logEntry: ScrapingLogEntry = {
      type: eventType,
      message,
      url: payload?.url,
      current_depth: payload?.current_depth,
      max_depth: payload?.max_depth,
      links_found: payload?.links_found,
      timestamp
    }

    setScrapingProgress(prev => {
      const linksFound = typeof payload?.links_found === 'number' ? payload.links_found : prev.linksFound
      const currentDepth = typeof payload?.current_depth === 'number' ? payload.current_depth : prev.currentDepth
      const maxDepth = typeof payload?.max_depth === 'number' ? payload.max_depth : prev.maxDepth
      let status: ScrapingProgressState['status'] = prev.status === 'idle' ? 'running' : prev.status

      if (eventType === 'started' || eventType === 'link_found' || eventType === 'visiting') {
        status = 'running'
      } else if (eventType === 'completed') {
        status = 'completed'
      } else if (eventType === 'error') {
        status = 'error'
      }

      return {
        status,
        jobId: eventJobId ?? prev.jobId ?? null,
        linksFound,
        currentDepth,
        maxDepth,
        logs: [...prev.logs, logEntry]
      }
    })

    if (eventType === 'completed') {
      cancelScrapingStream()
      setSuccessMessage('Website scraping completed successfully')
      loadData()
      setScrapingInProgress(false)
      setStreamingJobId(null)
      streamingJobRef.current = null
    } else if (eventType === 'error') {
      cancelScrapingStream()
      setError(payload?.message || 'Scraping failed')
      setScrapingInProgress(false)
      setStreamingJobId(null)
      streamingJobRef.current = null
    }
  }

  const handleIndexingEvent = (eventType: string, payload: any) => {
    const timestamp = payload?.timestamp || new Date().toISOString()
    const message = describeProgressEvent(eventType, payload)
    const logEntry: IndexingLogEntry = {
      type: eventType,
      message,
      url: payload?.url,
      timestamp
    }

    setIndexingProgress(prev => {
      const total = payload?.total ?? prev.total
      let completed = payload?.completed ?? prev.completed
      let pending = payload?.pending ?? Math.max(total - completed, 0)
      const totalTokens = payload?.total_tokens ?? prev.totalTokens
      let status: IndexingProgressState['status'] = prev.status === 'idle' ? 'running' : prev.status

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
        logs: [...prev.logs, logEntry]
      }
    })

    if (eventType === 'completed') {
      closeIndexingEventSource()
      setSuccessMessage('Website indexing completed successfully')
      loadData()
    } else if (eventType === 'error') {
      closeIndexingEventSource()
      setError(payload?.message || 'Indexing failed')
    }
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
      setStagedLinksError(`You can only index up to ${maxSelectableLinks} link${maxSelectableLinks === 1 ? '' : 's'} at a time`)
      return
    }

    setStagedLinksError(null)

    try {
      const selectionResponse = await apiClient.selectScrapingJobLinks(activeScrapingJob.id, urls)
      setMaxSelectableLinks(selectionResponse.maxSelectableLinks)
      setStagedLinks(current => current.map(link => ({ ...link, selected: urls.includes(link.url) })))
      setScrapingJobs(current => current.map(job => job.id === activeScrapingJob.id ? { ...job, status: 'indexing', selected_links: urls } : job))
      setActiveScrapingJob(current => current ? { ...current, status: 'indexing', selected_links: urls } : current)

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
      eventsToHandle.forEach(eventName => {
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

  useEffect(() => {
    if (activeScrapingJob && projectId && activeScrapingJob.project_id !== projectId) {
      closeScrapingSelection()
    }
  }, [projectId])

  const saveAboutMe = async () => {
    if (!projectId) {
      setError('No project selected')
      return
    }

    setAboutMeSaving(true)
    setError(null)

    try {
      await apiClient.updateAboutMeSettings({ content: aboutMeContent })
      setSuccessMessage('About me information saved successfully')
      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err: any) {
      setError(`Failed to save about me information: ${err?.response?.data?.error || err?.message || 'Unknown error'}`)
      setTimeout(() => setError(null), 5000)
    } finally {
      setAboutMeSaving(false)
    }
  }

  const searchKnowledge = async () => {
    if (!projectId) {
      setError('No project selected')
      return
    }
    if (!searchQuery.trim()) return

    setSearchLoading(true)
    try {
      const results = await apiClient.searchKnowledge(projectId, searchQuery)
      setSearchResults(results)
    } catch (err) {
      setError('Search failed')
      console.error('Search error:', err)
    } finally {
      setSearchLoading(false)
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ready':
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-500" data-testid="check-circle-icon" />
      case 'processing':
      case 'running':
      case 'indexing':
        return <Loader className="h-4 w-4 text-blue-500 animate-spin" data-testid="loader-icon" />
      case 'pending':
      case 'awaiting_selection':
        return <AlertCircle className="h-4 w-4 text-amber-500" data-testid="alert-circle-icon" />
      case 'failed':
      case 'error':
        return <XCircle className="h-4 w-4 text-red-500" data-testid="x-circle-icon" />
      case 'cancelled':
        return <AlertCircle className="h-4 w-4 text-gray-400" data-testid="alert-circle-icon" />
      default:
        return <AlertCircle className="h-4 w-4 text-gray-500" data-testid="alert-circle-icon" />
    }
  }

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800'
      case 'running':
      case 'indexing':
        return 'bg-blue-100 text-blue-800'
      case 'pending':
        return 'bg-yellow-100 text-yellow-800'
      case 'awaiting_selection':
        return 'bg-amber-100 text-amber-800'
      case 'failed':
      case 'error':
        return 'bg-red-100 text-red-800'
      case 'cancelled':
        return 'bg-gray-100 text-gray-700'
      default:
        return 'bg-gray-100 text-gray-700'
    }
  }

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const selectedLinkCount = selectedLinkUrls.size
  const totalDiscoveredTokens = stagedLinks.reduce((sum, link) => sum + (link.token_count || 0), 0)
  const indexingActive = indexingProgress.status === 'running'

  // Early return if no projectId is provided
  if (!projectId) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Please select a project to manage knowledge base.</p>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64" data-testid="knowledge-loading">
        <Loader className="h-8 w-8 animate-spin text-blue-500" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Error/Success Messages */}
      {error && (
        <div className="border rounded-lg p-4 bg-destructive/5 border-destructive/20">
          <div className="flex items-start">
            <XCircle className="h-5 w-5 text-destructive mr-3 mt-0.5" />
            <div>
              <p className="text-sm text-destructive">{error}</p>
            </div>
          </div>
        </div>
      )}

      {activeScrapingJob && (
        <div className="border rounded-lg p-6 bg-card">
          <div className="flex items-start justify-between">
            <div>
              <h3 className="font-medium">Prepare Indexing</h3>
              <p className="text-sm text-muted-foreground">Review staged links discovered from {activeScrapingJob.url}</p>
              <p className="text-xs text-muted-foreground mt-1">
                {stagedLinks.length} link(s) discovered • ≈ {totalDiscoveredTokens.toLocaleString()} tokens
              </p>
              <p className="text-xs text-muted-foreground">
                Select up to {maxSelectableLinks} link{maxSelectableLinks === 1 ? '' : 's'} for indexing.
              </p>
            </div>
            <button
              onClick={closeScrapingSelection}
              className="inline-flex items-center px-3 py-1.5 text-xs font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground"
            >
              <X className="h-4 w-4 mr-1" />
              Close
            </button>
          </div>

          {stagedLinksError && (
            <div className="mt-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
              {stagedLinksError}
            </div>
          )}

          {stagedLinksLoading ? (
            <div className="flex items-center justify-center py-10">
              <Loader className="h-5 w-5 mr-2 animate-spin text-primary" />
              <span className="text-sm text-muted-foreground">Loading staged links...</span>
            </div>
          ) : (
            <div className="mt-4 space-y-4">
              {stagedLinks.length === 0 ? (
                <p className="text-sm text-muted-foreground">No staged links available yet. Please wait for the scraper to finish.</p>
              ) : (
                <>
                  <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between text-sm">
                    <span>{selectedLinkCount} of {maxSelectableLinks} link(s) selected • {stagedLinks.length} discovered</span>
                    <div className="space-x-3">
                      <button onClick={selectAllLinks} className="text-xs font-medium text-primary hover:underline">
                        Select All
                      </button>
                      <button onClick={clearLinkSelection} className="text-xs font-medium text-muted-foreground hover:underline">
                        Clear
                      </button>
                    </div>
                  </div>

                  <div className="max-h-64 overflow-y-auto space-y-2 pr-1">
                    {stagedLinks.map(link => {
                      const isSelected = selectedLinkUrls.has(link.url)
                      return (
                        <label
                          key={link.url}
                          className={`flex items-start space-x-3 rounded-md border p-3 transition-colors ${isSelected ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/40'}`}
                        >
                          <input
                            type="checkbox"
                            className="mt-1"
                            checked={isSelected}
                            onChange={() => toggleLinkSelection(link.url)}
                          />
                          <div className="space-y-1">
                            <div className="flex flex-col gap-1 md:flex-row md:items-center md:justify-between">
                              <span className="text-sm font-medium break-words md:max-w-xl">
                                {link.title || link.url}
                              </span>
                              <span className="text-xs text-muted-foreground whitespace-nowrap">
                                Depth {link.depth} • {link.token_count.toLocaleString()} tokens
                              </span>
                            </div>
                            {link.content_preview && (
                              <p className="text-xs text-muted-foreground line-clamp-3">
                                {link.content_preview}
                              </p>
                            )}
                            <p className="text-xs text-muted-foreground break-all">
                              {link.url}
                            </p>
                          </div>
                        </label>
                      )
                    })}
                  </div>
                </>
              )}

              <div className="flex flex-col gap-3 border rounded-md bg-muted/20 p-3 text-sm md:flex-row md:items-center md:justify-between">
                <div className="text-xs text-muted-foreground">
                  Status:{' '}
                  <span className="font-medium text-foreground">
                    {indexingProgress.status === 'idle'
                      ? 'Awaiting indexing'
                      : indexingProgress.status === 'running'
                      ? 'Indexing in progress'
                      : indexingProgress.status === 'completed'
                      ? 'Completed'
                      : 'Error'}
                  </span>
                  <br />
                  Processed: {indexingProgress.completed}/{indexingProgress.total} • Pending: {indexingProgress.pending}
                  {indexingProgress.totalTokens > 0 && (
                    <> • Tokens ≈ {indexingProgress.totalTokens.toLocaleString()}</>
                  )}
                </div>

                <div className="flex items-center space-x-2">
                  <button
                    onClick={startIndexing}
                    disabled={indexingActive || selectedLinkCount === 0}
                    className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {indexingActive ? (
                      <>
                        <Loader className="h-4 w-4 mr-2 animate-spin" />
                        Streaming progress...
                      </>
                    ) : (
                      <>
                        <Globe className="h-4 w-4 mr-2" />
                        Start Indexing
                      </>
                    )}
                  </button>
                </div>
              </div>

              {indexingProgress.logs.length > 0 && (
                <div className="border rounded-md bg-muted/20 p-3 max-h-52 overflow-y-auto space-y-2 text-xs">
                  {indexingProgress.logs.map((log, idx) => (
                    <div key={`${log.timestamp}-${idx}`} className="flex items-start justify-between gap-3">
                      <div>
                        <span className="font-medium text-foreground">{log.message}</span>
                        {log.url && <div className="text-muted-foreground break-all">{log.url}</div>}
                      </div>
                      <span className="text-muted-foreground whitespace-nowrap">{new Date(log.timestamp).toLocaleTimeString()}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {successMessage && (
        <div className="border rounded-lg p-4 bg-green-50 border-green-200">
          <div className="flex items-start">
            <CheckCircle className="h-5 w-5 text-green-600 mr-3 mt-0.5" />
            <div>
              <p className="text-sm text-green-800">{successMessage}</p>
            </div>
          </div>
        </div>
      )}

      {/* About Me and Document Upload Sections */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* About Me Section */}
        <div className="border rounded-lg p-6 bg-card">
          <div className="space-y-4">
            <div>
              <h3 className="font-medium">About Me & Information</h3>
              <p className="text-sm text-muted-foreground">Provide information about yourself and other relevant details for the AI agent</p>
            </div>
            
            <div className="space-y-4">
              <textarea
                value={aboutMeContent}
                onChange={(e) => setAboutMeContent(e.target.value)}
                placeholder="Tell me about yourself, your role, preferences, and any other information that would help the AI agent assist you better..."
                className="w-full h-64 px-3 py-2 border border-input rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                disabled={aboutMeSaving}
              />
              
              <div className="flex items-center justify-between">
                <div className="text-xs text-muted-foreground">
                  {aboutMeContent.length} characters
                </div>
                <button
                  onClick={saveAboutMe}
                  disabled={aboutMeSaving}
                  className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {aboutMeSaving ? (
                    <>
                      <Loader className="h-4 w-4 animate-spin mr-2" />
                      Saving...
                    </>
                  ) : (
                    <>
                      <Save className="h-4 w-4 mr-2" />
                      Save Information
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Document Upload Section */}
        <div className="border rounded-lg p-6 bg-card">
          <div className="space-y-4">
            <div>
              <h3 className="font-medium">Document Upload</h3>
              <p className="text-sm text-muted-foreground">Upload PDF files to add content to your knowledge base</p>
            </div>
            
            <div
              className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                dragOver 
                  ? 'border-primary bg-primary/5' 
                  : 'border-border hover:border-primary/50'
              }`}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
            >
              <Upload className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <div>
                <label htmlFor="file-upload" className="cursor-pointer">
                  <span className="text-sm font-medium">
                    Drag and drop PDF files here or click to browse
                  </span>
                  <input
                    id="file-upload"
                    name="file-upload"
                    type="file"
                    className="sr-only"
                    multiple
                    accept=".pdf"
                    onChange={handleFileSelect}
                  />
                </label>
                <p className="mt-2 text-xs text-muted-foreground">PDF files only, up to 10MB each</p>
              </div>
            </div>

            {uploading && (
              <div className="flex items-center justify-center py-4">
                <Loader className="h-5 w-5 animate-spin text-primary mr-2" />
                <span className="text-sm text-muted-foreground">Uploading...</span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Scraping Progress Section */}
      {scrapingProgress.status !== 'idle' && (
        <div className="border rounded-lg p-6 bg-card">
          <div className="space-y-4">
            <div>
              <h3 className="font-medium">Scraping Progress</h3>
              <p className="text-sm text-muted-foreground">Real-time progress of website scraping</p>
            </div>

            <div className="flex flex-col gap-3 border rounded-md bg-muted/20 p-3 text-sm md:flex-row md:items-center md:justify-between">
              <div className="text-xs text-muted-foreground">
                Status:{' '}
                <span className="font-medium text-foreground">
                  {scrapingProgress.status === 'running'
                    ? 'Scraping in progress'
                    : scrapingProgress.status === 'completed'
                    ? 'Scraping completed'
                    : scrapingProgress.status === 'error'
                    ? 'Error occurred'
                    : 'Idle'}
                </span>
                <br />
                Links discovered: {scrapingProgress.linksFound}
                {scrapingProgress.currentDepth > 0 && (
                  <> • Current depth: {scrapingProgress.currentDepth}/{scrapingProgress.maxDepth}</>
                )}
              </div>

              {scrapingProgress.status === 'running' && (
                <div className="flex items-center space-x-2">
                  <Loader className="h-4 w-4 animate-spin text-primary" />
                  <span className="text-xs text-muted-foreground">Scraping...</span>
                </div>
              )}
            </div>

            {scrapingProgress.logs.length > 0 && (
              <div className="border rounded-md bg-muted/20 p-3 max-h-52 overflow-y-auto space-y-2 text-xs">
                {scrapingProgress.logs.map((log, idx) => (
                  <div key={`${log.timestamp}-${idx}`} className="flex items-start justify-between gap-3">
                    <div>
                      <span className="font-medium text-foreground">{log.message}</span>
                      {log.url && <div className="text-muted-foreground break-all">{log.url}</div>}
                      {typeof log.links_found === 'number' && (
                        <div className="text-muted-foreground">Links discovered: {log.links_found}</div>
                      )}
                    </div>
                    <span className="text-muted-foreground whitespace-nowrap">{new Date(log.timestamp).toLocaleTimeString()}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Web Scraping Section */}
      <div className="border rounded-lg p-6 bg-card">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-medium">Web Scraping</h3>
              <p className="text-sm text-muted-foreground">Scrape websites to automatically add their content to your knowledge base</p>
            </div>
            <button
              onClick={() => setShowScrapingForm(!showScrapingForm)}
              className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90"
            >
              <Plus className="h-4 w-4 mr-1" />
              Add Web Source
            </button>
          </div>

          {showScrapingForm && (
            <div className="border rounded-lg p-4 bg-muted/50">
              <div className="space-y-4">
                <div>
                  <label htmlFor="website-url" className="block text-sm font-medium mb-2">Website URL</label>
                  <input
                    id="website-url"
                    type="url"
                    value={scrapingUrl}
                    onChange={(e) => setScrapingUrl(e.target.value)}
                    placeholder="https://example.com"
                    className="w-full px-3 py-2 border border-input rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-ring"
                  />
                </div>
                <div>
                  <label htmlFor="max-depth" className="block text-sm font-medium mb-2">Max Depth</label>
                  <select
                    id="max-depth"
                    value={scrapingDepth}
                    onChange={(e) => setScrapingDepth(Number(e.target.value))}
                    className="w-full px-3 py-2 border border-input rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-ring"
                  >
                    <option value={1}>1 level (homepage only)</option>
                    <option value={2}>2 levels</option>
                    <option value={3}>3 levels</option>
                    <option value={4}>4 levels</option>
                    <option value={5}>5 levels (maximum)</option>
                  </select>
                </div>
                <div className="flex space-x-3">
                  <button
                    onClick={createScrapingJobWithStream}
                    disabled={scrapingInProgress}
                    className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {scrapingInProgress ? (
                      <>
                        <Loader className="h-4 w-4 animate-spin mr-2" />
                        Scraping...
                      </>
                    ) : (
                      <>
                        <Globe className="h-4 w-4 mr-2" />
                        Start Streaming Scrape
                      </>
                    )}
                  </button>
                  <button
                    onClick={createScrapingJob}
                    disabled={scrapingInProgress}
                    className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Legacy Scrape
                  </button>
                  <button
                    onClick={() => setShowScrapingForm(false)}
                    className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Documents List */}
      {documents.length > 0 && (
        <div className="border rounded-lg p-6 bg-card">
          <div className="space-y-4">
            <div>
              <h3 className="font-medium">Uploaded Documents</h3>
              <p className="text-sm text-muted-foreground">Manage your uploaded PDF documents</p>
            </div>
            
            <div className="border rounded-lg overflow-hidden">
              <div className="bg-muted/50 px-4 py-3 border-b">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Documents</span>
                  <span className="text-sm text-muted-foreground">{documents.length} files</span>
                </div>
              </div>
              <div className="divide-y">
                {documents.map((doc) => (
                  <div key={doc.id} className="flex items-center justify-between p-4">
                    <div className="flex items-center space-x-3">
                      <FileText className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <p className="text-sm font-medium">{doc.filename}</p>
                        <p className="text-xs text-muted-foreground">
                          {formatFileSize(doc.file_size)} • {new Date(doc.created_at).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-3">
                      <div className="flex items-center space-x-2">
                        {getStatusIcon(doc.status)}
                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          doc.status === 'ready' 
                            ? 'bg-green-100 text-green-800' 
                            : doc.status === 'processing'
                            ? 'bg-blue-100 text-blue-800'
                            : 'bg-red-100 text-red-800'
                        }`}>
                          {doc.status}
                        </span>
                      </div>
                      <button
                        onClick={() => deleteDocument(doc.id)}
                        className="inline-flex items-center px-2 py-1 text-xs font-medium text-muted-foreground hover:text-destructive"
                        aria-label="Delete document"
                      >
                        <Trash2 className="h-3 w-3 mr-1" />
                        Delete
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Scraping Jobs List */}
      {scrapingJobs.length > 0 && (
        <div className="border rounded-lg p-6 bg-card">
          <div className="space-y-4">
            <div>
              <h3 className="font-medium">Web Scraping Jobs</h3>
              <p className="text-sm text-muted-foreground">Monitor your website scraping progress</p>
            </div>
            
            <div className="border rounded-lg overflow-hidden">
              <div className="bg-muted/50 px-4 py-3 border-b">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Scraping Jobs</span>
                  <span className="text-sm text-muted-foreground">{scrapingJobs.length} jobs</span>
                </div>
              </div>
              <div className="divide-y">
                {scrapingJobs.map((job) => (
                  <div key={job.id} className="flex items-center justify-between p-4">
                    <div className="flex items-center space-x-3">
                      <Globe className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <p className="text-sm font-medium">{job.url}</p>
                        <p className="text-xs text-muted-foreground">
                          Depth: {job.max_depth} • Pages: {job.pages_scraped}/{job.total_pages || '?'} • {new Date(job.created_at).toLocaleDateString()}
                        </p>
                        {job.selected_links && job.selected_links.length > 0 && (
                          <p className="text-xs text-muted-foreground">
                            Selected links: {job.selected_links.length}
                          </p>
                        )}
                        {job.status === 'failed' && job.error_message && (
                          <p className="text-xs text-red-500 mt-1">{job.error_message}</p>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center space-x-3">
                      <div className="flex items-center space-x-2">
                        {getStatusIcon(job.status)}
                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusBadgeClass(job.status)}`}>
                          {job.status.replace(/_/g, ' ')}
                        </span>
                      </div>
                      {(job.status === 'awaiting_selection' || job.status === 'indexing' || job.status === 'failed') && (
                        <button
                          onClick={() => openScrapingJobLinks(job)}
                          className="inline-flex items-center px-3 py-1.5 text-xs font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground"
                        >
                          {job.status === 'awaiting_selection' ? 'Review Links' : job.status === 'indexing' ? 'View Progress' : 'Review & Retry'}
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Knowledge Search */}
      <div className="border rounded-lg p-6 bg-card">
        <div className="space-y-4">
          <div>
            <h3 className="font-medium">Search Knowledge Base</h3>
            <p className="text-sm text-muted-foreground">Search through your uploaded documents and scraped content</p>
          </div>
          
          <div className="flex space-x-2">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search knowledge base..."
              className="flex-1 px-3 py-2 border border-input rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-ring"
              onKeyPress={(e) => e.key === 'Enter' && searchKnowledge()}
            />
            <button
              onClick={searchKnowledge}
              disabled={searchLoading}
              className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {searchLoading ? (
                <>
                  <Loader className="h-4 w-4 animate-spin mr-1" />
                  Searching...
                </>
              ) : (
                <>
                  <Search className="h-4 w-4 mr-1" />
                  Search
                </>
              )}
            </button>
          </div>

          {searchResults && (
            <div className="border rounded-lg p-4 bg-muted/30">
              <h4 className="font-medium mb-3">Search Results</h4>
              <p className="text-sm text-muted-foreground mb-4">
                Found {searchResults.total_results} results
              </p>
              {searchResults.chunks && searchResults.chunks.map((chunk: any, index: number) => (
                <div key={index} className="border-b border-border pb-3 mb-3 last:border-b-0">
                  <p className="text-sm mb-2">{chunk.content}</p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span>From: {chunk.document_title || 'Unknown document'}</span>
                    <span>Similarity: {Math.round(chunk.similarity * 100)}%</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
