import { Save, X, Loader2 } from 'lucide-react'

interface FormActionsProps {
  submitting: boolean
  widgetId?: string
  onCancel: () => void
}

export function FormActions({ submitting, widgetId, onCancel }: FormActionsProps) {
  return (
    <div className="flex items-center justify-end gap-4">

      {/* Right side - action buttons */}
      <div className="flex items-center gap-3 justify-end">
        <button
          type="button"
          onClick={onCancel}
          disabled={submitting}
          className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2 disabled:cursor-not-allowed disabled:opacity-50"
        >
          <X className="h-4 w-4 mr-2" aria-hidden="true" />
          Cancel
        </button>
        
        <button
          type="submit"
          disabled={submitting}
          className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {submitting ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" aria-hidden="true" />
              <span>{widgetId ? 'Updating...' : 'Creating...'}</span>
            </>
          ) : (
            <>
              <Save className="h-4 w-4 mr-2" aria-hidden="true" />
              <span>{widgetId ? 'Update Widget' : 'Create Widget'}</span>
            </>
          )}
        </button>
      </div>
    </div>
  )
}
