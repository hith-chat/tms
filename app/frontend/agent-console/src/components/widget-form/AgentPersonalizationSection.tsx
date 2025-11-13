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
    <div className="rounded-lg border border-border bg-card shadow-sm">
      {/* Section Header */}
      <div className="border-b border-border bg-muted/50 px-6 py-4">
        <div className="flex items-center gap-2">
          <User className="h-5 w-5 text-primary" />
          <h3 className="text-base font-semibold text-foreground">Agent Personalization</h3>
        </div>
        <p className="text-sm text-muted-foreground mt-1">Configure your agent's identity and greeting</p>
      </div>

      {/* Section Content */}
      <div className="p-6 space-y-5">
        {/* Domain URL Field */}
        <div className="space-y-2">
          <label htmlFor="domain-url" className="text-sm font-medium text-foreground flex items-center gap-2">
            <Globe className="h-4 w-4 text-muted-foreground" />
            Domain <span className="text-destructive">*</span>
          </label>
          <input
            id="domain-url"
            type="text"
            value={formData.domain_url || ''}
            onChange={(e) => onUpdate({ domain_url: e.target.value })}
            placeholder="example.com"
            className="h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            required
          />
          <p className="text-xs text-muted-foreground">Enter the domain where this widget will be embedded</p>
        </div>

        {/* Agent Details Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          <div className="space-y-2">
            <label htmlFor="agent-name" className="text-sm font-medium text-foreground">
              Agent Name
            </label>
            <input
              id="agent-name"
              type="text"
              value={formData.agent_name || ''}
              onChange={(e) => onUpdate({ agent_name: e.target.value })}
              placeholder="Sarah Johnson"
              className="h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            />
            <p className="text-xs text-muted-foreground">Name displayed in the chat interface</p>
          </div>

          <div className="space-y-2">
            <label htmlFor="agent-avatar" className="text-sm font-medium text-foreground">
              Avatar URL <span className="text-muted-foreground text-xs">(optional)</span>
            </label>
            <input
              id="agent-avatar"
              type="url"
              value={formData.agent_avatar_url || ''}
              onChange={(e) => onUpdate({ agent_avatar_url: e.target.value })}
              placeholder="https://example.com/avatar.jpg"
              className="h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            />
            <p className="text-xs text-muted-foreground">Profile picture for your agent</p>
          </div>
        </div>

        {/* Welcome Message */}
        <div className="space-y-2">
          <label htmlFor="welcome-message" className="text-sm font-medium text-foreground">
            Welcome Message
          </label>
          <textarea
            id="welcome-message"
            value={formData.welcome_message || ''}
            onChange={(e) => onUpdate({ welcome_message: e.target.value })}
            rows={3}
            placeholder="Hi there! 👋 How can we help you today?"
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm resize-none ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          />
          <p className="text-xs text-muted-foreground">Initial greeting message shown to visitors</p>
        </div>
      </div>
    </div>
  )
}
