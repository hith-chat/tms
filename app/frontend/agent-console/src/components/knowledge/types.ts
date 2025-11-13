export interface IndexingLogEntry {
  type: string
  message: string
  url?: string
  timestamp: string
}

export interface IndexingProgressState {
  status: 'idle' | 'running' | 'completed' | 'error'
  total: number
  completed: number
  pending: number
  totalTokens: number
  logs: IndexingLogEntry[]
}

export interface ScrapingLogEntry {
  type: string
  message: string
  url?: string
  current_depth?: number
  max_depth?: number
  links_found?: number
  timestamp: string
}

export interface ScrapingProgressState {
  status: 'idle' | 'running' | 'completed' | 'error'
  jobId: string | null
  linksFound: number
  currentDepth: number
  maxDepth: number
  logs: ScrapingLogEntry[]
}

export const INITIAL_INDEXING_STATE: IndexingProgressState = {
  status: 'idle',
  total: 0,
  completed: 0,
  pending: 0,
  totalTokens: 0,
  logs: []
}

export const INITIAL_SCRAPING_STATE: ScrapingProgressState = {
  status: 'idle',
  jobId: null,
  linksFound: 0,
  currentDepth: 0,
  maxDepth: 0,
  logs: []
}
