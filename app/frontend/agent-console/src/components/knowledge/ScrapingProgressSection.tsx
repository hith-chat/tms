import { Loader } from 'lucide-react'
import { ScrapingProgressState } from './types'

interface ScrapingProgressSectionProps {
  scrapingProgress: ScrapingProgressState
}

export function ScrapingProgressSection({ scrapingProgress }: ScrapingProgressSectionProps) {
  return (
    <div className="border rounded-lg p-6 bg-card">
      <div className="space-y-4">
        <div>
          <h3 className="font-medium text-foreground">Scraping Progress</h3>
          <p className="text-sm text-muted-foreground mt-1">Real-time progress of website scraping</p>
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
                <span className="text-muted-foreground whitespace-nowrap">
                  {new Date(log.timestamp).toLocaleTimeString()}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
