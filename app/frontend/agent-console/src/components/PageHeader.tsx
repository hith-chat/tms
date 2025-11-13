import { LucideIcon } from 'lucide-react'
import { ReactNode } from 'react'

interface PageHeaderProps {
  icon: LucideIcon
  title: string
  subtitle: string
  gradientFrom: string
  gradientTo: string
  actions?: ReactNode
  children?: ReactNode
}

export function PageHeader({
  icon: Icon,
  title,
  subtitle,
  gradientFrom,
  gradientTo,
  actions,
  children
}: PageHeaderProps) {
  return (
    <div className="border-b border-border/50 bg-background/80 backdrop-blur-xl supports-[backdrop-filter]:bg-background/60 shadow-sm">
      <div className="px-6 py-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="relative">
              <div
                className="absolute -inset-1 rounded-lg blur opacity-25"
                style={{
                  background: `linear-gradient(to right, ${gradientFrom}, ${gradientTo})`
                }}
              />
              <div
                className="relative p-3 rounded-lg border"
                style={{
                  background: `linear-gradient(to bottom right, ${gradientFrom}10, ${gradientTo}10)`,
                  borderColor: `${gradientFrom}30`
                }}
              >
                <Icon
                  className="w-6 h-6"
                  style={{
                    color: gradientFrom
                  }}
                />
              </div>
            </div>

            <div>
              <h1
                className="text-2xl font-bold bg-clip-text text-transparent"
                style={{
                  backgroundImage: `linear-gradient(to right, ${gradientFrom}, ${gradientTo})`
                }}
              >
                {title}
              </h1>
              <div className="flex items-center gap-3 mt-1">
                <p className="text-sm text-muted-foreground">{subtitle}</p>
              </div>
            </div>
          </div>

          {/* Actions like tabs, buttons, etc */}
          {actions && <div className="flex items-center gap-4">{actions}</div>}
        </div>

        {/* Additional content like error/success messages */}
        {children}
      </div>
    </div>
  )
}
