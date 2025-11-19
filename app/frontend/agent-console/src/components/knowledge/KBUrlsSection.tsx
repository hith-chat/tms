import { useState, useEffect } from 'react'
import { ExternalLink, Trash2, Eye, RefreshCw, Loader } from 'lucide-react'
import { apiClient, ProjectKnowledgePage } from '../../lib/api'

interface KBUrlsSectionProps {
  projectId: string
}

export function KBUrlsSection({ projectId }: KBUrlsSectionProps) {
  const [pages, setPages] = useState<ProjectKnowledgePage[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedPage, setSelectedPage] = useState<ProjectKnowledgePage | null>(null)
  const [showContentModal, setShowContentModal] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const loadPages = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await apiClient.getWidgetKnowledgePages(projectId)
      setPages(data)
    } catch (err) {
      console.error('Failed to load knowledge pages:', err)
      setError('Failed to load knowledge pages')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadPages()
  }, [projectId])

  const handleDelete = async (mappingId: string) => {
    if (!confirm('Are you sure you want to remove this page from the knowledge base?')) {
      return
    }

    setDeletingId(mappingId)
    try {
      await apiClient.deleteWidgetKnowledgePageMapping(mappingId)
      setPages(pages.filter(p => p.id !== mappingId))
    } catch (err) {
      console.error('Failed to delete page mapping:', err)
      alert('Failed to delete page mapping')
    } finally {
      setDeletingId(null)
    }
  }

  const handleViewContent = (page: ProjectKnowledgePage) => {
    setSelectedPage(page)
    setShowContentModal(true)
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  if (loading) {
    return (
      <div className="border rounded-lg p-6 bg-card">
        <div className="flex items-center justify-center py-8">
          <Loader className="h-6 w-6 animate-spin text-muted-foreground" />
          <span className="ml-2 text-muted-foreground">Loading knowledge pages...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="border rounded-lg p-6 bg-card">
        <div className="text-center py-8">
          <p className="text-destructive">{error}</p>
          <button
            onClick={loadPages}
            className="mt-4 inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <>
      <div className="border rounded-lg p-6 bg-card">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-medium text-foreground">Knowledge Base URLs</h3>
              <p className="text-sm text-muted-foreground mt-1">
                All pages indexed in your widget's knowledge base ({pages.length} total)
              </p>
            </div>
            <button
              onClick={loadPages}
              disabled={loading}
              className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors"
            >
              <RefreshCw className={`h-4 w-4 mr-1 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </button>
          </div>

          {pages.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>No pages have been indexed yet.</p>
              <p className="text-sm mt-2">Start by creating a scraping job in the KB-Jobs tab.</p>
            </div>
          ) : (
            <div className="border rounded-lg overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-muted/50 border-b">
                    <tr>
                      <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">URL</th>
                      <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Title</th>
                      <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Tokens</th>
                      <th className="text-left px-4 py-3 text-sm font-medium text-muted-foreground">Scraped</th>
                      <th className="text-right px-4 py-3 text-sm font-medium text-muted-foreground">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {pages.map((page) => (
                      <tr key={page.id} className="hover:bg-muted/30 transition-colors">
                        <td className="px-4 py-3">
                          <a
                            href={page.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-sm text-primary hover:underline inline-flex items-center max-w-md truncate"
                          >
                            <span className="truncate">{page.url}</span>
                            <ExternalLink className="h-3 w-3 ml-1 flex-shrink-0" />
                          </a>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-sm text-foreground truncate max-w-xs block">
                            {page.title || '-'}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-sm text-muted-foreground">{page.token_count.toLocaleString()}</span>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-sm text-muted-foreground">{formatDate(page.scraped_at)}</span>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center justify-end space-x-2">
                            <button
                              onClick={() => handleViewContent(page)}
                              className="inline-flex items-center px-2 py-1 text-xs font-medium rounded border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors"
                              title="View details"
                            >
                              <Eye className="h-3 w-3 mr-1" />
                              View
                            </button>
                            <button
                              onClick={() => handleDelete(page.id)}
                              disabled={deletingId === page.id}
                              className="inline-flex items-center px-2 py-1 text-xs font-medium rounded border border-destructive/50 text-destructive hover:bg-destructive hover:text-destructive-foreground disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                              title="Remove from knowledge base"
                            >
                              {deletingId === page.id ? (
                                <Loader className="h-3 w-3 animate-spin" />
                              ) : (
                                <>
                                  <Trash2 className="h-3 w-3 mr-1" />
                                  Remove
                                </>
                              )}
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Content Modal */}
      {showContentModal && selectedPage && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50" onClick={() => setShowContentModal(false)}>
          <div className="bg-card border rounded-lg shadow-lg max-w-2xl w-full max-h-[80vh] overflow-hidden" onClick={(e) => e.stopPropagation()}>
            <div className="p-6 border-b">
              <h3 className="text-lg font-semibold text-foreground">Page Details</h3>
            </div>
            <div className="p-6 space-y-4 overflow-y-auto max-h-[60vh]">
              <div>
                <label className="text-sm font-medium text-muted-foreground">URL</label>
                <a
                  href={selectedPage.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block text-sm text-primary hover:underline mt-1"
                >
                  {selectedPage.url}
                </a>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">Title</label>
                <p className="text-sm text-foreground mt-1">{selectedPage.title || 'No title'}</p>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Token Count</label>
                  <p className="text-sm text-foreground mt-1">{selectedPage.token_count.toLocaleString()}</p>
                </div>
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Scraped At</label>
                  <p className="text-sm text-foreground mt-1">{formatDate(selectedPage.scraped_at)}</p>
                </div>
              </div>
            </div>
            <div className="p-6 border-t flex justify-end">
              <button
                onClick={() => setShowContentModal(false)}
                className="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
