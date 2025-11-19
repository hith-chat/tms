import { Loader } from 'lucide-react'
import { KnowledgeDocument, KnowledgeScrapingJob, KnowledgeFAQItem, ScrapedLinkPreview } from '../../lib/api'
import { WebScrapingSection } from './WebScrapingSection'
import { ScrapingProgressSection } from './ScrapingProgressSection'
import { PrepareIndexingSection } from './PrepareIndexingSection'
import { DocumentsList } from './DocumentsList'
import { ScrapingJobsList } from './ScrapingJobsList'
import { FAQList } from './FAQList'
import { IndexingProgressState, ScrapingProgressState } from './types'

interface KnowledgeBaseTabProps {
  projectId: string | null
  documents: KnowledgeDocument[]
  scrapingJobs: KnowledgeScrapingJob[]
  faqItems: KnowledgeFAQItem[]
  scrapingInProgress: boolean
  scrapingProgress: ScrapingProgressState
  activeScrapingJob: KnowledgeScrapingJob | null
  stagedLinks: ScrapedLinkPreview[]
  stagedLinksLoading: boolean
  stagedLinksError: string | null
  maxSelectableLinks: number
  selectedLinkUrls: Set<string>
  indexingProgress: IndexingProgressState
  onStartScraping: (url: string, depth: number) => void
  onStartLegacyScraping: (url: string, depth: number) => void
  onDeleteDocument: (documentId: string) => void
  onOpenScrapingJobLinks: (job: KnowledgeScrapingJob) => void
  onCloseScrapingSelection: () => void
  onToggleLinkSelection: (url: string) => void
  onSelectAllLinks: () => void
  onClearLinkSelection: () => void
  onStartIndexing: () => void
}

export function KnowledgeBaseTab(props: KnowledgeBaseTabProps) {
  const selectedLinkCount = props.selectedLinkUrls.size
  const totalDiscoveredTokens = props.stagedLinks.reduce((sum, link) => sum + (link.token_count || 0), 0)
  const indexingActive = props.indexingProgress.status === 'running'

  return (
    <div className="space-y-6">
      {/* Prepare Indexing Section */}
      {props.activeScrapingJob && (
        <PrepareIndexingSection
          activeScrapingJob={props.activeScrapingJob}
          stagedLinks={props.stagedLinks}
          stagedLinksLoading={props.stagedLinksLoading}
          stagedLinksError={props.stagedLinksError}
          maxSelectableLinks={props.maxSelectableLinks}
          selectedLinkUrls={props.selectedLinkUrls}
          selectedLinkCount={selectedLinkCount}
          totalDiscoveredTokens={totalDiscoveredTokens}
          indexingProgress={props.indexingProgress}
          indexingActive={indexingActive}
          onClose={props.onCloseScrapingSelection}
          onToggleLinkSelection={props.onToggleLinkSelection}
          onSelectAllLinks={props.onSelectAllLinks}
          onClearLinkSelection={props.onClearLinkSelection}
          onStartIndexing={props.onStartIndexing}
        />
      )}

      {/* Scraping Progress Section */}
      {props.scrapingProgress.status !== 'idle' && (
        <ScrapingProgressSection scrapingProgress={props.scrapingProgress} />
      )}

      {/* Simple Scraping Loading Indicator */}
      {props.scrapingInProgress && props.scrapingProgress.status === 'idle' && (
        <div className="border rounded-lg p-6 bg-card">
          <div className="flex items-center space-x-3">
            <Loader className="h-5 w-5 animate-spin text-primary" />
            <div>
              <h3 className="font-medium text-foreground">Scraping in Progress</h3>
              <p className="text-sm text-muted-foreground">
                Fetching and indexing content from URL...
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Web Scraping Section */}
      <WebScrapingSection
        onStartScraping={props.onStartScraping}
        onStartLegacyScraping={props.onStartLegacyScraping}
        scrapingInProgress={props.scrapingInProgress}
      />

      {/* Documents List */}
      {props.documents.length > 0 && (
        <DocumentsList
          documents={props.documents}
          onDeleteDocument={props.onDeleteDocument}
        />
      )}

      {/* Scraping Jobs List */}
      {props.scrapingJobs.length > 0 && (
        <ScrapingJobsList
          scrapingJobs={props.scrapingJobs}
          onOpenScrapingJobLinks={props.onOpenScrapingJobLinks}
        />
      )}

      {/* FAQ Items */}
      {props.faqItems.length > 0 && (
        <FAQList faqItems={props.faqItems} />
      )}
    </div>
  )
}
