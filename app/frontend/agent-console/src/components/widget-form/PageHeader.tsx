import { AlertCircle, AlertTriangle } from 'lucide-react'
import { Link } from 'react-router-dom'
import type { DomainValidation } from '../../lib/api'

interface PageHeaderProps {
  widgetId?: string
  error: string | null
  domains: DomainValidation[]
}

export function PageHeader({ widgetId: _widgetId, error, domains }: PageHeaderProps) {
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

      {/* Domain Warning */}
      {domains.length === 0 && (
        <div className="rounded-lg border border-warning/50 bg-warning/10 p-4">
          <div className="flex items-start gap-3">
            <AlertTriangle className="h-5 w-5 text-warning flex-shrink-0 mt-0.5" aria-hidden="true" />
            <div className="flex flex-col gap-2">
              <h4 className="text-sm font-medium text-warning-foreground">
                Domain Verification Required
              </h4>
              <p className="text-sm text-warning-foreground/80">
                You need to verify at least one domain before creating chat widgets. 
                Verified domains ensure your widgets can only be embedded on authorized websites.
              </p>
              <Link 
                to="/settings?tab=domains" 
                className="inline-flex items-center rounded-md bg-warning text-warning-foreground hover:bg-warning/90 px-3 py-1.5 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              >
                Verify Domains in Settings
              </Link>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
