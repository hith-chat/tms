import { KnowledgeManagement } from '../components/KnowledgeManagement'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { Brain } from 'lucide-react'

export function KnowledgePage() {
  const currentProjectId = localStorage.getItem('project_id')
  
  if (!currentProjectId) {
    return (
      <div className="h-full flex items-center justify-center bg-gradient-to-br from-background via-background to-slate-50/20 dark:to-slate-950/20">
        <div className="text-center py-12">
          <div className="mb-4">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-muted flex items-center justify-center">
              <Brain className="h-8 w-8 text-muted-foreground" />
            </div>
          </div>
          <h2 className="text-xl font-semibold text-foreground mb-2">No Project Selected</h2>
          <p className="text-muted-foreground max-w-md">
            Please select a project from the top navigation to manage your AI knowledge base.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col bg-gradient-to-br from-background via-background to-slate-50/20 dark:to-slate-950/20">
      {/* Header */}
      <div className="border-b bg-card">
        <div className="px-6 py-4">
          <div className="flex items-center gap-4">
            <div className="relative">
              <div className="absolute -inset-1 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg blur opacity-25"></div>
              <div className="relative p-3 bg-gradient-to-br from-blue-50 to-purple-50 dark:from-blue-950 dark:to-purple-950 rounded-lg border border-blue-200/50 dark:border-blue-800/50">
                <Brain className="w-6 h-6 text-blue-600 dark:text-blue-400" />
              </div>
            </div>

            <div>
              <h1 className="text-2xl font-semibold text-foreground">AI Knowledge Base</h1>
              <p className="text-sm text-muted-foreground">
                Manage documents, web sources, and AI agent information
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Scrollable Content */}
      <div className="flex-1 overflow-y-auto">
        <div className="p-6">
          <ErrorBoundary>
            <KnowledgeManagement projectId={currentProjectId} />
          </ErrorBoundary>
        </div>
      </div>
    </div>
  )
}