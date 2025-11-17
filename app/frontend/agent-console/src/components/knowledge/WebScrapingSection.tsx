import { useState } from 'react'
import { Globe, Loader, Plus } from 'lucide-react'

interface WebScrapingSectionProps {
  onStartScraping: (url: string, depth: number) => void
  onStartLegacyScraping: (url: string, depth: number) => void
  scrapingInProgress: boolean
}

export function WebScrapingSection({
  onStartScraping,
  onStartLegacyScraping,
  scrapingInProgress
}: WebScrapingSectionProps) {
  const [showForm, setShowForm] = useState(false)
  const [url, setUrl] = useState('')
  const [depth, setDepth] = useState(3)

  const handleStartScraping = () => {
    onStartScraping(url, depth)
    setUrl('')
    setShowForm(false)
  }

  const handleLegacyScraping = () => {
    onStartLegacyScraping(url, depth)
    setUrl('')
    setShowForm(false)
  }

  return (
    <div className="border rounded-lg p-6 bg-card">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-medium text-foreground">Web Scraping</h3>
            <p className="text-sm text-muted-foreground mt-1">
              Scrape websites to automatically add their content to your knowledge base
            </p>
          </div>
          <button
            onClick={() => setShowForm(!showForm)}
            className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            <Plus className="h-4 w-4 mr-1" />
            Add Web Source
          </button>
        </div>

        {showForm && (
          <div className="border rounded-lg p-4 bg-muted/50">
            <div className="space-y-4">
              <div>
                <label htmlFor="website-url" className="block text-sm font-medium mb-2 text-foreground">
                  Website URL
                </label>
                <input
                  id="website-url"
                  type="url"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://example.com"
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                />
              </div>
              <div>
                <label htmlFor="max-depth" className="block text-sm font-medium mb-2 text-foreground">
                  Max Depth
                </label>
                <select
                  id="max-depth"
                  value={depth}
                  onChange={(e) => setDepth(Number(e.target.value))}
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value={1}>1 level (homepage only)</option>
                  <option value={2}>2 levels</option>
                  {/* <option value={3}>3 levels</option> */}
                  {/* <option value={4}>4 levels</option> */}
                  {/* <option value={5}>5 levels (maximum)</option> */}
                </select>
              </div>
              <div className="flex space-x-3">
                <button
                  onClick={handleStartScraping}
                  disabled={scrapingInProgress || !url.trim()}
                  className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
                  onClick={handleLegacyScraping}
                  disabled={scrapingInProgress || !url.trim()}
                  className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md border border-input bg-background text-foreground hover:bg-accent hover:text-accent-foreground disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  Legacy Scrape
                </button>
                <button
                  onClick={() => setShowForm(false)}
                  className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md border border-input bg-background text-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
