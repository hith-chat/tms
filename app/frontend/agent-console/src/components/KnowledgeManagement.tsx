import { useState, useEffect } from 'react'
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
  Search 
} from 'lucide-react'
import { apiClient, KnowledgeDocument, KnowledgeScrapingJob } from '../lib/api'

interface KnowledgeManagementProps {
  projectId: string | null
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

  // Search state
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<any>(null)

  useEffect(() => {
    loadData()
  }, [projectId])

  // Polling effect for updating running/pending jobs
  useEffect(() => {
    const hasActiveJobs = scrapingJobs.some(job => 
      job.status === 'running' || job.status === 'pending'
    )
    
    if (!hasActiveJobs) return

    const pollInterval = setInterval(() => {
      loadData()
    }, 3000) // Poll every 3 seconds

    return () => clearInterval(pollInterval)
  }, [scrapingJobs])

  const loadData = async () => {
    if (!projectId) {
      console.warn('No project ID provided to KnowledgeManagement component')
      return
    }
    
    setLoading(true)
    setError(null)
    try {
      console.log('Loading knowledge data for project:', projectId)
      const [documentsData, jobsData] = await Promise.all([
        apiClient.getDocuments(projectId),
        apiClient.getScrapingJobs(projectId)
      ])
      console.log('Successfully loaded:', { documents: documentsData?.length, jobs: jobsData?.length })
      setDocuments(documentsData || [])
      setScrapingJobs(jobsData || [])
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

  const searchKnowledge = async () => {
    if (!projectId) {
      setError('No project selected')
      return
    }
    if (!searchQuery.trim()) return

    setLoading(true)
    try {
      const results = await apiClient.searchKnowledge(projectId, searchQuery)
      setSearchResults(results)
    } catch (err) {
      setError('Search failed')
      console.error('Search error:', err)
    } finally {
      setLoading(false)
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ready':
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'processing':
      case 'running':
      case 'pending':
        return <Loader className="h-4 w-4 text-blue-500 animate-spin" />
      case 'error':
        return <XCircle className="h-4 w-4 text-red-500" />
      default:
        return <AlertCircle className="h-4 w-4 text-gray-500" />
    }
  }

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // Early return if no projectId is provided
  if (!projectId) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Please select a project to manage knowledge base.</p>
      </div>
    )
  }

  if (loading && !documents.length && !scrapingJobs.length) {
    return (
      <div className="flex items-center justify-center h-64">
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
                  Drop PDF files here or click to browse
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
              <span className="text-sm text-muted-foreground">Uploading and processing documents...</span>
            </div>
          )}
        </div>
      </div>

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
              Add Website
            </button>
          </div>

          {showScrapingForm && (
            <div className="border rounded-lg p-4 bg-muted/50">
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">Website URL</label>
                  <input
                    type="url"
                    value={scrapingUrl}
                    onChange={(e) => setScrapingUrl(e.target.value)}
                    placeholder="https://example.com"
                    className="w-full px-3 py-2 border border-input rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-ring"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-2">Max Depth</label>
                  <select
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
                    onClick={createScrapingJob}
                    disabled={scrapingInProgress}
                    className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {scrapingInProgress ? (
                      <Loader className="h-4 w-4 animate-spin mr-2" />
                    ) : (
                      <Globe className="h-4 w-4 mr-2" />
                    )}
                    Start Scraping
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
                        className="text-muted-foreground hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" />
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
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(job.status)}
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        job.status === 'completed' 
                          ? 'bg-green-100 text-green-800' 
                          : job.status === 'running'
                          ? 'bg-blue-100 text-blue-800'
                          : job.status === 'pending'
                          ? 'bg-yellow-100 text-yellow-800'
                          : 'bg-red-100 text-red-800'
                      }`}>
                        {job.status}
                      </span>
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
              placeholder="Search your knowledge base..."
              className="flex-1 px-3 py-2 border border-input rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-ring"
              onKeyPress={(e) => e.key === 'Enter' && searchKnowledge()}
            />
            <button
              onClick={searchKnowledge}
              className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90"
            >
              <Search className="h-4 w-4" />
            </button>
          </div>

          {searchResults && (
            <div className="border rounded-lg p-4 bg-muted/30">
              <p className="text-sm text-muted-foreground mb-2">
                Found {searchResults.total_results} results
              </p>
              {/* Add search results display here */}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
