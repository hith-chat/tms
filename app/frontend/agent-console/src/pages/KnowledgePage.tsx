import { KnowledgeManagement } from '../components/KnowledgeManagement'
import { PageHeader } from '../components/PageHeader'
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
      <PageHeader
        icon={Brain}
        title="AI Knowledge Base"
        subtitle="Manage documents, web sources, and AI agent information"
        gradientFrom="#2563eb"
        gradientTo="#9333ea"
      />

      {/* Content Area */}
      <div className="flex-1 overflow-hidden">
        <KnowledgeManagement projectId={currentProjectId} />
      </div>
    </div>
  )
}
