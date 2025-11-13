import { Wand2, Settings } from 'lucide-react'

export type BuilderMode = 'manual' | 'ai'

interface BuilderModeToggleProps {
  mode: BuilderMode
  onChange: (mode: BuilderMode) => void
}

export function BuilderModeToggle({ mode, onChange }: BuilderModeToggleProps) {
  return (
    <div className="flex items-center justify-between w-full min-w-0 mb-6 p-4 rounded-lg border border-border bg-card">
      <div>
        <h3 className="text-lg font-medium text-foreground">
          Widget Builder Mode
        </h3>
      </div>

      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => onChange('ai')}
          className={`px-3 py-1.5 rounded-md text-sm font-medium transition-all flex items-center gap-1.5 ${
            mode === 'ai'
              ? 'bg-primary text-primary-foreground'
              : 'bg-muted text-muted-foreground hover:bg-muted/80'
          }`}
        >
          <Wand2 className="h-3.5 w-3.5" />
          Auto-generate using AI
        </button>

        <button
          type="button"
          onClick={() => onChange('manual')}
          className={`px-3 py-1.5 rounded-md text-sm font-medium transition-all flex items-center gap-1.5 ${
            mode === 'manual'
              ? 'bg-primary text-primary-foreground'
              : 'bg-muted text-muted-foreground hover:bg-muted/80'
          }`}
        >
          <Settings className="h-3.5 w-3.5" />
          Manual
        </button>
      </div>
    </div>
  )
}