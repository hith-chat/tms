import { MessageCircle, X, Send, Paperclip, AlertTriangle, Maximize2 } from 'lucide-react'
import { useState, useEffect, useRef } from 'react'
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

export function WidgetSimulation({ formData }: WidgetSimulationProps) {
  const {
    isWidgetOpen,
    isTyping,
    simulationMessages,
    toggleWidget,
    startTypingDemo
  } = useWidgetSimulation(formData.welcome_message, formData.agent_name)

  const [iframeError, setIframeError] = useState(false)
  const [showMockup, setShowMockup] = useState(false)
  const [iframeLoaded, setIframeLoaded] = useState(false)
  const [isFullScreen, setIsFullScreen] = useState(false)
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const timeoutRef = useRef<NodeJS.Timeout>()

  // Normalize domain_url to add protocol if missing
  const getIframeUrl = () => {
    const domain = formData.domain_url || 'hith.chat'
    if (domain.startsWith('http://') || domain.startsWith('https://')) {
      return domain
    }
    return `https://${domain}`
  }

  const handleIframeError = () => {
    console.warn('Iframe failed to load or blocked by X-Frame-Options')
    setIframeError(true)
    setShowMockup(true)
    setIframeLoaded(false)
  }

  const handleIframeLoad = () => {
    // Clear the timeout as iframe loaded successfully
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    // Try to access iframe content to verify it's not blocked
    try {
      // If we can't access contentWindow or it's restricted, it might be blocked
      // const iframeWindow = iframeRef.current?.contentWindow
      // if (!iframeWindow || iframeWindow.location.href === 'about:blank') {
      //   // Likely blocked by X-Frame-Options
      //   handleIframeError()
      //   return
      // }

      // Successfully loaded
      setIframeError(false)
      setIframeLoaded(true)
      setShowMockup(false)
    } catch (e) {
      console.warn('Iframe load check failed:', e)
      // Cross-origin or blocked - this is actually normal for cross-origin iframes
      // So we'll consider it loaded unless we hit the timeout
      setIframeError(false)
      setIframeLoaded(true)
      setShowMockup(false)
    }
  }

  // Monitor iframe loading with timeout
  useEffect(() => {
    // Reset states when domain changes
    setIframeLoaded(false)
    setShowMockup(false)
    setIframeError(false)

    // Set a timeout to check if iframe loads within 3 seconds
    timeoutRef.current = setTimeout(() => {
      if (!iframeLoaded) {
        // Iframe didn't load in time, show mockup
        setIframeError(true)
        setShowMockup(true)
      }
    }, 3000)

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [formData.domain_url])

  return (
    <div className="sticky top-6 space-y-4">
      <div className="rounded border border-border bg-card p-4">
        <div className="flex items-center gap-2 mb-4">
          <MessageCircle className="h-4 w-4 text-primary" />
          <h3 className="text-sm font-medium text-foreground">Live Preview</h3>
        </div>

        {/* Website Container */}
        <div className="rounded-lg h-[690px] relative overflow-hidden border border-border flex flex-col">
          {/* Mockup Browser UI - Only show in mockup mode */}
          {showMockup && (
            <div className="bg-white dark:bg-slate-800 rounded-t border-b border-slate-200 dark:border-slate-700 p-3 flex-shrink-0">
              <div className="flex items-center gap-2">
                <div className="flex gap-1">
                  <div className="w-3 h-3 rounded-full bg-red-400"></div>
                  <div className="w-3 h-3 rounded-full bg-yellow-400"></div>
                  <div className="w-3 h-3 rounded-full bg-green-400"></div>
                </div>
                <div className="flex-1 bg-slate-100 dark:bg-slate-700 rounded px-3 py-1 text-xs text-slate-600 dark:text-slate-400">
                  {formData.domain_url || 'hith.chat'}
                </div>
              </div>
            </div>
          )}

          {/* Website Content Area */}
          <div className={`flex-1 relative overflow-hidden ${showMockup ? 'bg-white dark:bg-slate-800 rounded-b' : ''}`}>
            {/* Show iframe error message if blocked */}
            {iframeError && showMockup && (
              <div className="absolute top-3 left-1/2 transform -translate-x-1/2 z-10 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-3 max-w-md">
                <div className="flex items-start gap-2">
                  <AlertTriangle className="h-4 w-4 text-yellow-600 dark:text-yellow-400 flex-shrink-0 mt-0.5" />
                  <div className="text-xs text-yellow-800 dark:text-yellow-200">
                    <p className="font-medium">Preview not available</p>
                    <p className="mt-1">This website blocks iframe embedding.</p>
                  </div>
                </div>
              </div>
            )}

            {/* Try to load real website in iframe */}
            {!showMockup && (
              <iframe
                ref={iframeRef}
                src={getIframeUrl()}
                className="absolute inset-0 w-full h-full rounded-lg"
                onError={handleIframeError}
                onLoad={handleIframeLoad}
                sandbox="allow-same-origin allow-scripts"
                title="Website Preview"
              />
            )}

            {/* Fallback mockup content */}
            {showMockup && (
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
            )}

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
        <div className="mt-4 flex gap-2 justify-end">
          <button
            type="button"
            onClick={startTypingDemo}
            className="px-3 py-2 text-xs rounded-md font-medium transition-all hover:opacity-90 border border-border flex items-center gap-1"
          >
            Demo Typing
          </button>
          <button
            type="button"
            onClick={() => setIsFullScreen(true)}
            className="px-3 py-2 text-xs rounded-md flex items-center gap-1 font-medium transition-all hover:opacity-90"
            style={{
              backgroundColor: formData.primary_color || '#3b82f6',
              color: 'white'
            }}
          >
            <Maximize2 className="h-3 w-3" />
            Full Screen
          </button>
        </div>
      </div>

      {/* Full Screen Modal */}
      {isFullScreen && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <div className="relative w-full h-full max-w-7xl max-h-[90vh] bg-card border border-border rounded-lg shadow-lg overflow-hidden flex flex-col">
              {/* Modal Header */}
              <div className="flex items-center justify-between p-4 border-b border-border bg-card">
                <div className="flex items-center gap-2">
                  <MessageCircle className="h-5 w-5 text-primary" />
                  <h2 className="text-lg font-semibold text-foreground">Full Screen Preview</h2>
                </div>
                <button
                  type="button"
                  onClick={() => setIsFullScreen(false)}
                  className="rounded-md p-2 hover:bg-accent transition-colors"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>

              {/* Modal Content - Full Preview */}
              <div className="flex-1 relative overflow-hidden">
                {/* Mockup Browser UI - Only show in mockup mode */}
                {showMockup && (
                  <div className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 p-3">
                    <div className="flex items-center gap-2">
                      <div className="flex gap-1">
                        <div className="w-3 h-3 rounded-full bg-red-400"></div>
                        <div className="w-3 h-3 rounded-full bg-yellow-400"></div>
                        <div className="w-3 h-3 rounded-full bg-green-400"></div>
                      </div>
                      <div className="flex-1 bg-slate-100 dark:bg-slate-700 rounded px-3 py-1 text-xs text-slate-600 dark:text-slate-400">
                        {formData.domain_url || 'hith.chat'}
                      </div>
                    </div>
                  </div>
                )}

                {/* Website Content Area */}
                <div className={`absolute inset-0 ${showMockup ? 'top-[52px] bg-white dark:bg-slate-800' : ''}`}>
                  {/* Show iframe error message if blocked */}
                  {iframeError && showMockup && (
                    <div className="absolute top-4 left-1/2 transform -translate-x-1/2 z-10 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-3 max-w-md">
                      <div className="flex items-start gap-2">
                        <AlertTriangle className="h-4 w-4 text-yellow-600 dark:text-yellow-400 flex-shrink-0 mt-0.5" />
                        <div className="text-xs text-yellow-800 dark:text-yellow-200">
                          <p className="font-medium">Preview not available</p>
                          <p className="mt-1">This website blocks iframe embedding.</p>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Try to load real website in iframe */}
                  {!showMockup && (
                    <iframe
                      src={getIframeUrl()}
                      className="absolute inset-0 w-full h-full"
                      sandbox="allow-same-origin allow-scripts"
                      title="Website Preview Fullscreen"
                    />
                  )}

                  {/* Fallback mockup content */}
                  {showMockup && (
                    <div className="p-8 space-y-8 h-full overflow-auto">
                      {/* Main headline placeholder */}
                      <div className="h-12 bg-slate-200 dark:bg-slate-700 rounded w-2/3 max-w-2xl"></div>

                      {/* Content paragraphs placeholder */}
                      <div className="space-y-4 max-w-4xl">
                        <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded"></div>
                        <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded w-5/6"></div>
                        <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded w-4/5"></div>
                      </div>

                      {/* Additional content sections */}
                      <div className="space-y-4 max-w-3xl">
                        <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded w-3/4"></div>
                        <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded w-2/3"></div>
                      </div>
                    </div>
                  )}

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

                      {/* Chat Window - Same as regular preview */}
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
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
