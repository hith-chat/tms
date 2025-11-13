import { Wand2, Settings } from 'lucide-react'

export type BuilderMode = 'manual' | 'ai'

interface BuilderModeToggleProps {
  mode: BuilderMode
  onChange: (mode: BuilderMode) => void
}

export function BuilderModeToggle({ mode, onChange }: BuilderModeToggleProps) {
  return (
    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between w-full min-w-0 mb-6 p-5 rounded-lg border border-border bg-card shadow-sm gap-4">
      <div>
        <h3 className="text-base font-semibold text-foreground">
          Widget Builder Mode
        </h3>
        <p className="text-sm text-muted-foreground mt-1">
          Choose how you want to create your widget
        </p>
      </div>

      <div className="flex gap-2 bg-muted/50 p-1 rounded-lg">
        <button
          type="button"
          onClick={() => onChange('ai')}
          className={`px-4 py-2 rounded-md text-sm font-medium transition-all flex items-center gap-2 ${
            mode === 'ai'
              ? 'bg-primary text-primary-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <Wand2 className="h-4 w-4" />
          AI Builder
        </button>

        <button
          type="button"
          onClick={() => onChange('manual')}
          className={`px-4 py-2 rounded-md text-sm font-medium transition-all flex items-center gap-2 ${
            mode === 'manual'
              ? 'bg-primary text-primary-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <Settings className="h-4 w-4" />
          Manual
        </button>
      </div>
    </div>
  )
}