import { User, Globe } from 'lucide-react'
import type { CreateChatWidgetRequest } from '../../hooks/useChatWidgetForm'

interface AgentPersonalizationSectionProps {
  formData: CreateChatWidgetRequest
  onUpdate: (updates: Partial<CreateChatWidgetRequest>) => void
}

export function AgentPersonalizationSection({
  formData,
  onUpdate
}: AgentPersonalizationSectionProps) {
  return (
    <div className="rounded border border-border bg-card p-4">

      <div className="space-y-3">
        {/* Domain URL Field */}
        <div className="space-y-1">
          <label htmlFor="domain-url" className="text-sm font-medium text-foreground flex items-center gap-1">
            <Globe className="h-3.5 w-3.5" />
            Domain <span className="text-destructive">*</span>
          </label>
          <input
            id="domain-url"
            type="text"
            value={formData.domain_url || ''}
            onChange={(e) => onUpdate({ domain_url: e.target.value })}
            placeholder="example.com"
            className="h-9 w-full rounded border border-input bg-background px-3 py-2 text-sm"
            required
          />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1">
            <label htmlFor="agent-name" className="text-sm font-medium text-foreground">
              Agent Name
            </label>
            <input
              id="agent-name"
              type="text"
              value={formData.agent_name || ''}
              onChange={(e) => onUpdate({ agent_name: e.target.value })}
              placeholder="Sarah Johnson"
              className="h-9 w-full rounded border border-input bg-background px-3 py-2 text-sm"
            />
          </div>

          <div className="space-y-1">
            <label htmlFor="agent-avatar" className="text-sm font-medium text-foreground">
              Avatar URL (optional)
            </label>
            <input
              id="agent-avatar"
              type="url"
              value={formData.agent_avatar_url || ''}
              onChange={(e) => onUpdate({ agent_avatar_url: e.target.value })}
              placeholder="https://example.com/avatar.jpg"
              className="h-9 w-full rounded border border-input bg-background px-3 py-2 text-sm"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 gap-3">
          <div className="space-y-1">
            <label htmlFor="welcome-message" className="text-sm font-medium text-foreground">
              Welcome Message
            </label>
            <textarea
              id="welcome-message"
              value={formData.welcome_message || ''}
              onChange={(e) => onUpdate({ welcome_message: e.target.value })}
              rows={2}
              placeholder="Hi there! 👋 How can we help you today?"
              className="w-full rounded border border-input bg-background px-3 py-2 text-sm resize-none"
            />
          </div>
        </div>
      </div>
    </div>
  )
}
