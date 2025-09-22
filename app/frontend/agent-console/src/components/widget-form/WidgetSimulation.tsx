import { MessageCircle, X, Send, Paperclip } from 'lucide-react'
import type { CreateChatWidgetRequest } from '../../hooks/useChatWidgetForm'
import type { DomainValidation } from '../../lib/api'
import { useWidgetSimulation } from '../../hooks/useWidgetSimulation'
import { getWidgetButtonSize, getWidgetWindowSize, getBubbleStyleIconProps } from '../../utils/widgetHelpers'

interface WidgetSimulationProps {
  formData: CreateChatWidgetRequest
  domains: DomainValidation[]
}

// Helper component to render SVG icons
function IconSvg({ iconProps }: { iconProps: ReturnType<typeof getBubbleStyleIconProps> }) {
  return (
    <svg {...iconProps}>
      {iconProps.children.map((child, index) => {
        if (child.tag === 'path') {
          return <path key={index} d={child.d} />
        } else if (child.tag === 'rect') {
          return <rect key={index} {...child} />
        }
        return null
      })}
    </svg>
  )
}

export function WidgetSimulation({ formData, domains }: WidgetSimulationProps) {
  const {
    isWidgetOpen,
    isTyping,
    simulationMessages,
    toggleWidget,
    startTypingDemo
  } = useWidgetSimulation(formData.welcome_message, formData.agent_name)

  return (
    <div className="sticky top-6 space-y-4">
      <div className="rounded border border-border bg-card p-4">
        <div className="flex items-center gap-2 mb-4">
          <MessageCircle className="h-4 w-4 text-primary" />
          <h3 className="text-sm font-medium text-foreground">Live Preview</h3>
        </div>
        
        {/* Website Mockup Container */}
        <div className="bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 rounded-lg p-6 h-[690px] relative overflow-hidden border flex flex-col">
          {/* Mockup Browser UI */}
          <div className="bg-white dark:bg-slate-800 rounded-t border-b border-slate-200 dark:border-slate-700 p-3 flex-shrink-0">
            <div className="flex items-center gap-2">
              <div className="flex gap-1">
                <div className="w-3 h-3 rounded-full bg-red-400"></div>
                <div className="w-3 h-3 rounded-full bg-yellow-400"></div>
                <div className="w-3 h-3 rounded-full bg-green-400"></div>
              </div>
              <div className="flex-1 bg-slate-100 dark:bg-slate-700 rounded px-3 py-1 text-xs text-slate-600 dark:text-slate-400">
                {domains.find(d => d.id === formData.domain_id)?.domain || 'your-website.com'}
              </div>
            </div>
          </div>
          
          {/* Website Content Area */}
          <div className="bg-white dark:bg-slate-800 rounded-b flex-1 relative overflow-hidden">
            <div className="p-6 space-y-6 h-full">
              {/* Main headline placeholder */}
              <div className="h-8 bg-slate-200 dark:bg-slate-700 rounded w-3/4 max-w-md"></div>
              
              {/* Content paragraphs placeholder */}
              <div className="space-y-3 max-w-2xl">
                <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded"></div>
                <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-5/6"></div>
                <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-4/5"></div>
              </div>
              
              {/* Additional content sections */}
              <div className="space-y-3 max-w-xl">
                <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-2/3"></div>
                <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-3/4"></div>
              </div>
            </div>
            
            {/* Chat Widget Simulation */}
            <div className={`absolute ${formData.position === 'bottom-left' ? 'bottom-6 left-6' : 'bottom-6 right-6'} z-50`}>
              {/* Toggle Button */}
              <div className="relative">
                <button
                  type="button"
                  onClick={toggleWidget}
                  className={`
                    ${getWidgetButtonSize(formData.widget_size || 'medium')}
                    rounded-full shadow-lg transition-all duration-300 hover:scale-105 
                    flex items-center justify-center text-white
                    ${formData.widget_shape === 'square' ? 'rounded-lg' : 
                      formData.widget_shape === 'minimal' ? 'rounded-full border-2 border-white' : 
                      'rounded-full'}
                  `}
                  style={{ 
                    backgroundColor: formData.primary_color,
                    animation: formData.animation_style === 'bounce' ? 'bounce 2s infinite' :
                              formData.animation_style === 'fade' ? 'pulse 2s infinite' :
                              formData.animation_style === 'slide' ? 'slideIn 0.5s ease-out' : 'none'
                  }}
                >
                  <IconSvg iconProps={getBubbleStyleIconProps(formData.chat_bubble_style || 'modern')} />
                </button>
                
                {/* Chat Window */}
                {isWidgetOpen && (
                  <div 
                    className={`
                      absolute bottom-full mb-4 
                      ${formData.position === 'bottom-left' ? 'left-0' : 'right-0'}
                      ${getWidgetWindowSize(formData.widget_size || 'medium')}
                      rounded-lg shadow-2xl border border-slate-200 dark:border-slate-700
                      transform transition-all duration-300 origin-bottom
                      ${isWidgetOpen ? 'scale-100 opacity-100' : 'scale-95 opacity-0'}
                    `}
                    style={{ backgroundColor: formData.background_color || '#ffffff' }}
                  >
                    {/* Chat Header */}
                    <div 
                      className="p-4 rounded-t-lg border-b border-slate-200 dark:border-slate-700 flex items-center gap-3"
                      style={{ backgroundColor: formData.primary_color }}
                    >
                      {formData.show_agent_avatars && (
                        <div className="w-10 h-10 rounded-full bg-white/20 flex items-center justify-center text-white font-medium">
                          {formData.agent_avatar_url ? (
                            <img 
                              src={formData.agent_avatar_url} 
                              alt={formData.agent_name}
                              className="w-full h-full rounded-full object-cover"
                            />
                          ) : (
                            formData.agent_name?.charAt(0)?.toUpperCase() || 'S'
                          )}
                        </div>
                      )}
                      <div className="flex-1">
                        <div className="text-white font-medium text-sm">
                          {formData.agent_name || 'Support Agent'}
                        </div>
                        <div className="text-white/80 text-xs">
                          <div className="flex items-center gap-1">
                            <div 
                              className="w-2 h-2 rounded-full"
                              style={{ backgroundColor: '#10b981' }}
                            ></div>
                            Online now
                          </div>
                        </div>
                      </div>
                      <button 
                        type="button"
                        onClick={toggleWidget}
                        className="text-white/80 hover:text-white p-1"
                      >
                        <X className="h-4 w-4" />
                      </button>
                    </div>
                    
                    {/* Messages Area */}
                    <div className="flex-1 p-4 space-y-3 overflow-y-auto max-h-72">
                      {simulationMessages.map((message) => (
                        <div
                          key={message.id}
                          className={`flex ${message.type === 'visitor' ? 'justify-end' : 'justify-start'}`}
                        >
                          <div
                            className={`
                              max-w-[80%] p-3 rounded-lg text-sm
                              ${message.type === 'visitor'
                                ? 'text-white rounded-br-sm'
                                : message.type === 'system'
                                ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-900 dark:text-blue-100 rounded-bl-sm'
                                : 'bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-slate-100 rounded-bl-sm'
                              }
                            `}
                            style={
                              message.type === 'visitor' 
                                ? { backgroundColor: formData.primary_color }
                                : message.type === 'agent'
                                ? { backgroundColor: formData.secondary_color || '#f8fafc' }
                                : {}
                            }
                          >
                            {message.content}
                          </div>
                        </div>
                      ))}
                      
                      {/* Typing Indicator */}
                      {isTyping && (
                        <div className="flex justify-start">
                          <div 
                            className="p-3 rounded-lg rounded-bl-sm"
                            style={{ backgroundColor: formData.secondary_color || '#f8fafc' }}
                          >
                            <div className="flex space-x-1">
                              <div 
                                className="w-2 h-2 rounded-full animate-bounce"
                                style={{ backgroundColor: formData.primary_color || '#6b7280' }}
                              ></div>
                              <div 
                                className="w-2 h-2 rounded-full animate-bounce"
                                style={{ 
                                  backgroundColor: formData.primary_color || '#6b7280',
                                  animationDelay: '0.1s'
                                }}
                              ></div>
                              <div 
                                className="w-2 h-2 rounded-full animate-bounce"
                                style={{ 
                                  backgroundColor: formData.primary_color || '#6b7280',
                                  animationDelay: '0.2s'
                                }}
                              ></div>
                            </div>
                          </div>
                        </div>
                      )}
                    </div>
                    
                    {/* Input Area */}
                    <div className="p-4 border-t border-slate-200 dark:border-slate-700">
                      <div className="flex gap-2">
                        {formData.allow_file_uploads && (
                          <button 
                            type="button" 
                            className="p-2 rounded hover:bg-opacity-80 transition-colors"
                            style={{ 
                              color: formData.primary_color || '#6b7280',
                              backgroundColor: formData.secondary_color ? `${formData.secondary_color}40` : '#f3f4f6'
                            }}
                          >
                            <Paperclip className="h-4 w-4" />
                          </button>
                        )}
                        <input
                          type="text"
                          placeholder="Type your message..."
                          className="flex-1 px-3 py-2 rounded-lg text-sm focus:outline-none border-2 transition-colors"
                          style={{ 
                            backgroundColor: formData.background_color || '#ffffff',
                            borderColor: formData.secondary_color || '#e5e7eb'
                          }}
                          disabled
                        />
                        <button 
                          type="button"
                          className="p-2 text-white rounded-lg hover:opacity-90 transition-opacity"
                          style={{ backgroundColor: formData.primary_color }}
                        >
                          <Send className="h-4 w-4" />
                        </button>
                      </div>
                      
                      {formData.show_powered_by && (
                        <div className="mt-2 text-center">
                          <div className="text-xs text-slate-400">
                            Powered by Hith Chat
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
        
        {/* Quick Actions */}
        <div className="mt-4 flex gap-2">
          <button
            type="button"
            onClick={toggleWidget}
            className="flex-1 px-3 py-2 text-xs rounded hover:opacity-90 transition-opacity font-medium"
            style={{ 
              backgroundColor: formData.secondary_color || '#f3f4f6',
              color: formData.primary_color || '#374151',
              border: `1px solid ${formData.primary_color || '#e5e7eb'}`
            }}
          >
            {isWidgetOpen ? 'Close Widget' : 'Open Widget'}
          </button>
          <button
            type="button"
            onClick={startTypingDemo}
            className="px-3 py-2 text-xs rounded hover:opacity-90 transition-opacity font-medium"
            style={{ 
              backgroundColor: formData.primary_color,
              color: 'white'
            }}
          >
            Demo Typing
          </button>
        </div>
      </div>
    </div>
  )
}
