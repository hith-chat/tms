import { Globe, Loader, X } from 'lucide-react'
import { KnowledgeScrapingJob, ScrapedLinkPreview } from '../../lib/api'
import { IndexingProgressState } from './types'

interface PrepareIndexingSectionProps {
  activeScrapingJob: KnowledgeScrapingJob
  stagedLinks: ScrapedLinkPreview[]
  stagedLinksLoading: boolean
  stagedLinksError: string | null
  maxSelectableLinks: number
  selectedLinkUrls: Set<string>
  selectedLinkCount: number
  totalDiscoveredTokens: number
  indexingProgress: IndexingProgressState
  indexingActive: boolean
  onClose: () => void
  onToggleLinkSelection: (url: string) => void
  onSelectAllLinks: () => void
  onClearLinkSelection: () => void
  onStartIndexing: () => void
}

export function PrepareIndexingSection(props: PrepareIndexingSectionProps) {
  return (
    <div className="border rounded-lg p-6 bg-card">
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="font-medium text-foreground">Prepare Indexing</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Review staged links discovered from {props.activeScrapingJob.url}
          </p>
          <p className="text-xs text-muted-foreground mt-1">
            {props.stagedLinks.length} link(s) discovered • ≈ {props.totalDiscoveredTokens.toLocaleString()} tokens
          </p>
          <p className="text-xs text-muted-foreground">
            Select up to {props.maxSelectableLinks} link{props.maxSelectableLinks === 1 ? '' : 's'} for indexing.
          </p>
        </div>
        <button
          onClick={props.onClose}
          className="inline-flex items-center px-3 py-1.5 text-xs font-medium rounded-md border border-input bg-background text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
        >
          <X className="h-4 w-4 mr-1" />
          Close
        </button>
      </div>

      {props.stagedLinksError && (
        <div className="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
          {props.stagedLinksError}
        </div>
      )}

      {props.stagedLinksLoading ? (
        <div className="flex items-center justify-center py-10">
          <Loader className="h-5 w-5 mr-2 animate-spin text-primary" />
          <span className="text-sm text-muted-foreground">Loading staged links...</span>
        </div>
      ) : (
        <div className="space-y-4">
          {props.stagedLinks.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No staged links available yet. Please wait for the scraper to finish.
            </p>
          ) : (
            <>
              <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between text-sm">
                <span className="text-foreground">
                  {props.selectedLinkCount} of {props.maxSelectableLinks} link(s) selected •{' '}
                  {props.stagedLinks.length} discovered
                </span>
                <div className="space-x-3">
                  <button
                    onClick={props.onSelectAllLinks}
                    className="text-xs font-medium text-primary hover:underline"
                  >
                    Select All
                  </button>
                  <button
                    onClick={props.onClearLinkSelection}
                    className="text-xs font-medium text-muted-foreground hover:underline"
                  >
                    Clear
                  </button>
                </div>
              </div>

              <div className="max-h-64 overflow-y-auto space-y-2 pr-1">
                {props.stagedLinks.map((link) => {
                  const isSelected = props.selectedLinkUrls.has(link.url)
                  return (
                    <label
                      key={link.url}
                      className={`flex items-start space-x-3 rounded-md border p-3 transition-colors cursor-pointer ${
                        isSelected
                          ? 'border-primary bg-primary/5'
                          : 'border-border hover:border-primary/40'
                      }`}
                    >
                      <input
                        type="checkbox"
                        className="mt-1"
                        checked={isSelected}
                        onChange={() => props.onToggleLinkSelection(link.url)}
                      />
                      <div className="space-y-1 flex-1">
                        <div className="flex flex-col gap-1 md:flex-row md:items-center md:justify-between">
                          <span className="text-sm font-medium text-foreground break-words md:max-w-xl">
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
                        <p className="text-xs text-muted-foreground break-all">{link.url}</p>
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
                {props.indexingProgress.status === 'idle'
                  ? 'Awaiting indexing'
                  : props.indexingProgress.status === 'running'
                  ? 'Indexing in progress'
                  : props.indexingProgress.status === 'completed'
                  ? 'Completed'
                  : 'Error'}
              </span>
              <br />
              Processed: {props.indexingProgress.completed}/{props.indexingProgress.total} • Pending:{' '}
              {props.indexingProgress.pending}
              {props.indexingProgress.totalTokens > 0 && (
                <> • Tokens ≈ {props.indexingProgress.totalTokens.toLocaleString()}</>
              )}
            </div>

            <div className="flex items-center space-x-2">
              <button
                onClick={props.onStartIndexing}
                disabled={props.indexingActive || props.selectedLinkCount === 0}
                className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {props.indexingActive ? (
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

          {props.indexingProgress.logs.length > 0 && (
            <div className="border rounded-md bg-muted/20 p-3 max-h-52 overflow-y-auto space-y-2 text-xs">
              {props.indexingProgress.logs.map((log, idx) => (
                <div key={`${log.timestamp}-${idx}`} className="flex items-start justify-between gap-3">
                  <div>
                    <span className="font-medium text-foreground">{log.message}</span>
                    {log.url && <div className="text-muted-foreground break-all">{log.url}</div>}
                  </div>
                  <span className="text-muted-foreground whitespace-nowrap">
                    {new Date(log.timestamp).toLocaleTimeString()}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
