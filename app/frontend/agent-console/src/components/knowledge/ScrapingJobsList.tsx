import { Globe, CheckCircle, Loader, XCircle, AlertCircle } from 'lucide-react'
import { KnowledgeScrapingJob } from '../../lib/api'

interface ScrapingJobsListProps {
  scrapingJobs: KnowledgeScrapingJob[]
  onOpenScrapingJobLinks: (job: KnowledgeScrapingJob) => void
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'ready':
    case 'completed':
      return <CheckCircle className="h-4 w-4 text-green-500" />
    case 'processing':
    case 'running':
    case 'indexing':
      return <Loader className="h-4 w-4 text-blue-500 animate-spin" />
    case 'pending':
    case 'awaiting_selection':
      return <AlertCircle className="h-4 w-4 text-amber-500" />
    case 'failed':
    case 'error':
      return <XCircle className="h-4 w-4 text-red-500" />
    case 'cancelled':
      return <AlertCircle className="h-4 w-4 text-gray-400" />
    default:
      return <AlertCircle className="h-4 w-4 text-gray-500" />
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

export function ScrapingJobsList({ scrapingJobs, onOpenScrapingJobLinks }: ScrapingJobsListProps) {
  return (
    <div className="border rounded-lg p-6 bg-card">
      <div className="space-y-4">
        <div>
          <h3 className="font-medium text-foreground">Web Scraping Jobs</h3>
          <p className="text-sm text-muted-foreground mt-1">Monitor your website scraping progress</p>
        </div>

        <div className="border rounded-lg overflow-hidden">
          <div className="bg-muted/50 px-4 py-3 border-b">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-foreground">Scraping Jobs</span>
              <span className="text-sm text-muted-foreground">{scrapingJobs.length} jobs</span>
            </div>
          </div>
          <div className="divide-y">
            {scrapingJobs.map((job) => (
              <div key={job.id} className="flex items-center justify-between p-4">
                <div className="flex items-center space-x-3">
                  <Globe className="h-5 w-5 text-muted-foreground" />
                  <div>
                    <p className="text-sm font-medium text-foreground">{job.url}</p>
                    <p className="text-xs text-muted-foreground">
                      Depth: {job.max_depth} • Pages: {job.pages_scraped}/{job.total_pages || '?'} •{' '}
                      {new Date(job.created_at).toLocaleDateString()}
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
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusBadgeClass(
                        job.status
                      )}`}
                    >
                      {job.status.replace(/_/g, ' ')}
                    </span>
                  </div>
                  {(job.status === 'awaiting_selection' ||
                    job.status === 'indexing' ||
                    job.status === 'failed') && (
                    <button
                      onClick={() => onOpenScrapingJobLinks(job)}
                      className="inline-flex items-center px-3 py-1.5 text-xs font-medium rounded-md border border-input bg-background text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
                    >
                      {job.status === 'awaiting_selection'
                        ? 'Review Links'
                        : job.status === 'indexing'
                        ? 'View Progress'
                        : 'Review & Retry'}
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
