import { AlertCircle } from 'lucide-react'
import type { DomainValidation } from '../../lib/api'

interface PageHeaderProps {
  widgetId?: string
  error: string | null
  domains: DomainValidation[]
}

export function PageHeader({ widgetId: _widgetId, error, domains: _domains }: PageHeaderProps) {
  return (
    <div className="flex flex-col gap-4">
      {/* Error Alert */}
      {error && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="h-5 w-5 text-destructive flex-shrink-0 mt-0.5" aria-hidden="true" />
            <div className="flex flex-col gap-1">
              <h4 className="text-sm font-medium text-destructive">
                Configuration Error
              </h4>
              <p className="text-sm text-destructive/80">{error}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
