import { useState } from 'react'
import { Loader, Save } from 'lucide-react'
import { apiClient } from '../../lib/api'

interface AboutMeTabProps {
  projectId: string | null
  initialContent: string
  onContentChange: (content: string) => void
  onSaveSuccess: (message: string) => void
  onSaveError: (error: string) => void
}

export function AboutMeTab({
  projectId,
  initialContent,
  onContentChange,
  onSaveSuccess,
  onSaveError
}: AboutMeTabProps) {
  const [aboutMeContent, setAboutMeContent] = useState(initialContent)
  const [saving, setSaving] = useState(false)

  const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value
    setAboutMeContent(newContent)
    onContentChange(newContent)
  }

  const handleSave = async () => {
    if (!projectId) {
      onSaveError('No project selected')
      return
    }

    setSaving(true)

    try {
      await apiClient.updateAboutMeSettings({ content: aboutMeContent })
      onSaveSuccess('About me information saved successfully')
    } catch (err: any) {
      onSaveError(`Failed to save about me information: ${err?.response?.data?.error || err?.message || 'Unknown error'}`)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-medium text-foreground">About Me & Information</h3>
        <p className="text-sm text-muted-foreground mt-1">
          Provide information about yourself and other relevant details for the AI agent
        </p>
      </div>

      <div className="border rounded-lg bg-card p-6">
        <textarea
          value={aboutMeContent}
          onChange={handleContentChange}
          placeholder="Tell me about yourself, your role, preferences, and any other information that would help the AI agent assist you better..."
          className="w-full h-64 px-3 py-2 rounded-md bg-background border border-input focus:outline-none focus:ring-2 focus:ring-ring resize-none text-foreground"
          disabled={saving}
        />

        <div className="flex items-center justify-between mt-4">
          <div className="text-xs text-muted-foreground">
            {aboutMeContent.length} characters
          </div>
          <button
            onClick={handleSave}
            disabled={saving}
            className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {saving ? (
              <>
                <Loader className="h-4 w-4 animate-spin mr-2" />
                Saving...
              </>
            ) : (
              <>
                <Save className="h-4 w-4 mr-2" />
                Save Information
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}
