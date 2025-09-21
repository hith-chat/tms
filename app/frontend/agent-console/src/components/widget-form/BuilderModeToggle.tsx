import { Wand2, Settings } from 'lucide-react'

export type BuilderMode = 'manual' | 'ai'

interface BuilderModeToggleProps {
  mode: BuilderMode
  onChange: (mode: BuilderMode) => void
}

export function BuilderModeToggle({ mode, onChange }: BuilderModeToggleProps) {
  return (
    <div className="flex flex-col w-full min-w-0 mb-6">
      <div className="rounded-lg border border-border bg-card text-card-foreground shadow-sm">
        <div className="p-6">
          <div className="flex flex-col space-y-4">
            <div>
              <h3 className="text-lg font-semibold leading-none tracking-tight">
                Widget Builder Mode
              </h3>
              <p className="text-sm text-muted-foreground mt-1">
                Choose how you want to configure your chat widget
              </p>
            </div>

            <div className="flex gap-4">
              <button
                type="button"
                onClick={() => onChange('ai')}
                className={`flex-1 p-4 rounded-lg border-2 transition-all ${
                  mode === 'ai'
                    ? 'border-primary bg-primary/5 text-primary'
                    : 'border-border bg-background hover:bg-muted/50 text-muted-foreground hover:text-foreground'
                }`}
              >
                <div className="flex items-center gap-3">
                  <div className={`flex h-10 w-10 items-center justify-center rounded-lg ${
                    mode === 'ai' ? 'bg-primary/10' : 'bg-muted'
                  }`}>
                    <Wand2 className={`h-5 w-5 ${mode === 'ai' ? 'text-primary' : 'text-muted-foreground'}`} />
                  </div>
                  <div className="text-left">
                    <div className="font-semibold">AI Builder</div>
                    <div className="text-sm opacity-80">
                      Auto-generate theme from your website
                    </div>
                  </div>
                </div>
              </button>

              <button
                type="button"
                onClick={() => onChange('manual')}
                className={`flex-1 p-4 rounded-lg border-2 transition-all ${
                  mode === 'manual'
                    ? 'border-primary bg-primary/5 text-primary'
                    : 'border-border bg-background hover:bg-muted/50 text-muted-foreground hover:text-foreground'
                }`}
              >
                <div className="flex items-center gap-3">
                  <div className={`flex h-10 w-10 items-center justify-center rounded-lg ${
                    mode === 'manual' ? 'bg-primary/10' : 'bg-muted'
                  }`}>
                    <Settings className={`h-5 w-5 ${mode === 'manual' ? 'text-primary' : 'text-muted-foreground'}`} />
                  </div>
                  <div className="text-left">
                    <div className="font-semibold">Manual Mode</div>
                    <div className="text-sm opacity-80">
                      Customize every detail yourself
                    </div>
                  </div>
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}