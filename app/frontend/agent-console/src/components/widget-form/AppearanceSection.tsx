import { Palette } from 'lucide-react'
import type { CreateChatWidgetRequest } from '../../hooks/useChatWidgetForm'
import { widgetShapes, bubbleStyles } from '../../utils/widgetHelpers'

interface AppearanceSectionProps {
  formData: CreateChatWidgetRequest
  onUpdate: (updates: Partial<CreateChatWidgetRequest>) => void
}

export function AppearanceSection({ 
  formData, 
  onUpdate 
}: AppearanceSectionProps) {
  return (
    <div className="flex flex-col w-full min-w-0">
      {/* Card container with enterprise styling */}
      <div className="rounded-lg border border-border bg-card text-card-foreground shadow-sm">
        {/* Header */}
        <div className="flex items-center gap-3 p-6 pb-4">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10">
            <Palette className="h-4 w-4 text-primary" aria-hidden="true" />
          </div>
          <div className="flex flex-col space-y-1">
            <h3 className="text-base font-semibold leading-none tracking-tight">
              Appearance & Styling
            </h3>
            <p className="text-sm text-muted-foreground">
              Customize the visual appearance and behavior of your chat widget
            </p>
          </div>
        </div>

        {/* Form content */}
        <div className="px-6 pb-6">
          <div className="space-y-6">
            {/* Widget Shape, Bubble Style, Widget Size, and Position */}
            <div className="grid gap-4 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-4">
              <div className="space-y-2">
                <label
                  htmlFor="widget-shape"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Widget Shape <span className="text-destructive">*</span>
                </label>
                <select
                  id="widget-shape"
                  value={formData.widget_shape}
                  onChange={(e) => onUpdate({ widget_shape: e.target.value as any })}
                  className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  aria-describedby="widget-shape-description"
                >
                  {widgetShapes.map((shape) => (
                    <option key={shape.value} value={shape.value}>
                      {shape.preview} {shape.label}
                    </option>
                  ))}
                </select>
                <p id="widget-shape-description" className="text-xs text-muted-foreground">
                  Overall shape of widget button
                </p>
              </div>

              <div className="space-y-2">
                <label
                  htmlFor="bubble-style"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Bubble Style <span className="text-destructive">*</span>
                </label>
                <select
                  id="bubble-style"
                  value={formData.chat_bubble_style}
                  onChange={(e) => onUpdate({ chat_bubble_style: e.target.value as any })}
                  className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  aria-describedby="bubble-style-description"
                >
                  {bubbleStyles.map((style) => (
                    <option key={style.value} value={style.value}>
                      {style.label}
                    </option>
                  ))}
                </select>
                <p id="bubble-style-description" className="text-xs text-muted-foreground">
                  Style for chat message bubbles
                </p>
              </div>

              <div className="space-y-2">
                <label
                  htmlFor="widget-size"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Widget Size <span className="text-destructive">*</span>
                </label>
                <select
                  id="widget-size"
                  value={formData.widget_size}
                  onChange={(e) => onUpdate({ widget_size: e.target.value as any })}
                  className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  aria-describedby="widget-size-description"
                >
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
                <p id="widget-size-description" className="text-xs text-muted-foreground">
                  Size of the chat widget
                </p>
              </div>

              <div className="space-y-2">
                <label
                  htmlFor="position"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Position <span className="text-destructive">*</span>
                </label>
                <select
                  id="position"
                  value={formData.position}
                  onChange={(e) => onUpdate({ position: e.target.value as any })}
                  className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  aria-describedby="position-description"
                >
                  <option value="bottom-right">Bottom Right</option>
                  <option value="bottom-left">Bottom Left</option>
                </select>
                <p id="position-description" className="text-xs text-muted-foreground">
                  Location on the page
                </p>
              </div>
            </div>

            {/* Colors */}
            <div className="grid gap-4 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
              <div className="flex items-center gap-3 p-3 rounded-md border border-input bg-background/50">
                <input
                  id="primary-color"
                  type="color"
                  value={formData.primary_color}
                  onChange={(e) => onUpdate({ primary_color: e.target.value })}
                  className="h-10 w-10 rounded-md border-0 cursor-pointer"
                  aria-describedby="primary-color-description"
                />
                <div className="flex-1 min-w-0">
                  <label
                    htmlFor="primary-color"
                    className="text-sm font-medium leading-none cursor-pointer block"
                  >
                    Primary Color <span className="text-destructive">*</span>
                  </label>
                  <p id="primary-color-description" className="text-xs text-muted-foreground mt-1">
                    Main theme color
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3 p-3 rounded-md border border-input bg-background/50">
                <input
                  id="secondary-color"
                  type="color"
                  value={formData.secondary_color || '#6b7280'}
                  onChange={(e) => onUpdate({ secondary_color: e.target.value })}
                  className="h-10 w-10 rounded-md border-0 cursor-pointer"
                  aria-describedby="secondary-color-description"
                />
                <div className="flex-1 min-w-0">
                  <label
                    htmlFor="secondary-color"
                    className="text-sm font-medium leading-none cursor-pointer block"
                  >
                    Secondary Color
                  </label>
                  <p id="secondary-color-description" className="text-xs text-muted-foreground mt-1">
                    Agent messages
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3 p-3 rounded-md border border-input bg-background/50">
                <input
                  id="background-color"
                  type="color"
                  value={formData.background_color || '#ffffff'}
                  onChange={(e) => onUpdate({ background_color: e.target.value })}
                  className="h-10 w-10 rounded-md border-0 cursor-pointer"
                  aria-describedby="background-color-description"
                />
                <div className="flex-1 min-w-0">
                  <label
                    htmlFor="background-color"
                    className="text-sm font-medium leading-none cursor-pointer block"
                  >
                    Background Color
                  </label>
                  <p id="background-color-description" className="text-xs text-muted-foreground mt-1">
                    Chat window background
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
