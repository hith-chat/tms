import { EventEmitter } from './events'
import { ChatAPI } from './api'
import { ChatMessage, ChatSession, ChatWidget, ChatWidgetOptions, WSMessage } from './types'
import { SessionStorage, generateVisitorFingerprint, isBusinessHours } from './storage'
import { injectWidgetCSS, playNotificationSound } from './themes'

type Events = {
  'message:received': ChatMessage
  'message:sent': ChatMessage
  'session:started': ChatSession
  'session:ended': ChatSession
  'agent:joined': { agent_name: string }
  'agent:typing': { agent_name: string }
  'error': string
}

export class TMSChatWidget {
  private api: ChatAPI
  private emitter = new EventEmitter<Events>()
  private widget: ChatWidget | null = null
  private session: ChatSession | null = null
  private container: HTMLElement | null = null
  private toggleButton: HTMLElement | null = null
  private websocket: WebSocket | null = null
  private isOpen: boolean = false
  private messages: ChatMessage[] = []
  private isTyping: boolean = false
  private typingTimeout: number | null = null
  private storage: SessionStorage
  private unreadCount: number = 0
  private isBusinessHoursOpen: boolean = true
  private reconnectAttempts: number = 0
  private maxReconnectAttempts: number = 5
  private reconnectDelay: number = 3000
  private isConnected: boolean = false
  private poweredBadge: HTMLElement | null = null
  
  // Session timeout in milliseconds (1 hour)
  private readonly SESSION_TIMEOUT_MS = 7 * 24 * 60 *60 * 1000 // 7 days

  constructor(private options: ChatWidgetOptions) {
    this.api = new ChatAPI(options.apiUrl)
    this.storage = new SessionStorage(options.widgetId)
    this.init()
  }

  private getBubbleStyleIcon(style?: string): string {
    switch (style) {
      case 'modern':
        return `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
        </svg>`
      case 'classic':
        return `<svg width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12c0 1.54.36 2.98.97 4.29L1 23l6.71-1.97C9.02 21.64 10.46 22 12 22c5.52 0 10-4.48 10-10S17.52 2 12 2z"/>
        </svg>`
      case 'minimal':
        return `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
          <polyline points="22,6 12,13 2,6"/>
        </svg>`
      case 'bot':
        return `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 8V4H8"/>
          <rect width="16" height="12" x="4" y="8" rx="2"/>
          <path d="M2 14h2"/>
          <path d="M20 14h2"/>
          <path d="M15 13v2"/>
          <path d="M9 13v2"/>
        </svg>`
      default:
        return `<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
        </svg>`
    }
  }

  private async init() {
    try {
      // Check for existing session first
      await this.restoreSession()
      
      // Get widget configuration
      this.widget = await this.api.getWidgetById(this.options.widgetId)
      
      if (!this.widget) {
        throw new Error('Widget not found for domain')
      }

      // Check business hours
      this.isBusinessHoursOpen = isBusinessHours(this.widget.business_hours)
      
      // Inject CSS styles
      injectWidgetCSS(this.widget)
      
      // Create and inject widget UI
      this.createWidget()
      
      
      // Auto-open if configured and no existing session
      if (this.widget.auto_open_delay > 0 && !this.session) {
        setTimeout(() => this.open(), this.widget.auto_open_delay * 1000)
      }
      
    } catch (error) {
      console.error('Failed to initialize chat widget:', error)
      this.emitter.emit('error', 'Failed to initialize chat widget')
    }
  }

  private isSessionExpired(lastActivity: string): boolean {
    if (!lastActivity) return true
    
    const lastActivityTime = new Date(lastActivity).getTime()
    const currentTime = Date.now()
    const timeDiff = currentTime - lastActivityTime
    
    return timeDiff > this.SESSION_TIMEOUT_MS
  }

  private async restoreSession() {
    const existingSession = this.storage.getSession()
    if (existingSession) {
      // Check if session has expired (older than 1 hour)

      if( await this.api.verifySessionToken(existingSession.widget_id, existingSession.token)) {
        this.storage.clearSession()
        return
      }

      if (this.isSessionExpired(existingSession.last_activity)) {
        this.storage.clearSession()
        return
      }
      
      this.session = {
        id: existingSession.session_id,
        token: existingSession.token,
        widget_id: existingSession.widget_id,
        status: 'active'
      }
      
      // Load cached messages
      this.messages = this.storage.getMessages()
      this.storage.updateSessionActivity()
    }
  }

  private createWidget() {
    if (!this.widget) return

    // Create main container
    this.container = document.createElement('div')
    this.container.id = 'tms-chat-widget'
    this.container.className = 'tms-widget-container'

    // Create beautiful header with agent info
    const headerHTML = `
      <div class="tms-chat-header">
        <div class="tms-agent-info">
          ${this.widget.show_agent_avatars && this.widget.agent_avatar_url ? 
            `<div class="tms-agent-avatar">
              <img src="${this.widget.agent_avatar_url}" alt="${this.widget.agent_name}" />
            </div>` :
            `<div class="tms-agent-avatar">${this.widget.agent_name.charAt(0).toUpperCase()}</div>`
          }
          <div class="tms-agent-details">
            <div class="tms-agent-name">${this.widget.agent_name}</div>
            <div class="tms-agent-status">
              <div class="tms-status-indicator"></div>
              ${this.isBusinessHoursOpen ? 'Online now' : 'Away'}
            </div>
          </div>
        </div>
        <button class="tms-header-clear" aria-label="Clear chat">⟳</button>
      </div>
    `

    const visitorForm = `<div id="tms-visitor-info-form" class="tms-visitor-info-form" style="display: none;">
          <div class="tms-visitor-info-title">Start a conversation</div>
          <form id="tms-visitor-form">
          ${this.widget.require_name ? `
            <div class="tms-visitor-form-field">
              <label class="tms-visitor-form-label" for="tms-visitor-name">Name *</label>
              <input 
                id="tms-visitor-name" 
                type="text" 
                class="tms-visitor-form-input"
                placeholder="Enter your name"
                required
              />
            </div>`:``}
            ${this.widget.require_email ? `
              <div class="tms-visitor-form-field">
                <label class="tms-visitor-form-label" for="tms-visitor-email">Email *</label>
                <input 
                  id="tms-visitor-email" 
                  type="email" 
                  class="tms-visitor-form-input"
                  placeholder="Enter your email"
                  required
                />
              </div>
            `:``}

            ${this.widget.require_name || this.widget.require_email ? `<div class="tms-visitor-form-actions">
              <button type="button" id="tms-visitor-cancel" class="tms-visitor-form-button secondary">Cancel</button>
              <button type="submit" id="tms-visitor-start" class="tms-visitor-form-button primary">Start Chat</button>
            </div>`:``}
          </form>
        </div>`

    // Create modern body with enhanced messages area
    const bodyHTML = `
      <div class="tms-chat-body">
        ${this.widget.require_name || this.widget.require_email ? visitorForm : ''}
        <div class="tms-messages-container" id="tms-chat-messages"></div>
        
        <div class="tms-typing-indicator" id="tms-chat-typing" style="display: none;">
          <div class="tms-typing-dots">
            <div class="tms-typing-dot"></div>
            <div class="tms-typing-dot"></div>
            <div class="tms-typing-dot"></div>
          </div>
          <span id="tms-typing-text"></span>
        </div>
        
        <div class="tms-input-area">
          <div class="tms-input-controls">
            ${this.widget.allow_file_uploads ? `
              <label class="tms-file-upload-btn" for="tms-file-input" title="Attach file">
                📎
                <input id="tms-file-input" type="file" style="display: none;" accept="image/*,.pdf,.doc,.docx,.txt" />
              </label>
            ` : ''}
          </div>
          <div class="tms-input-wrapper">
            <div
              id="tms-chat-input"
              class="tms-message-input tms-editable"
              contenteditable="true"
              role="textbox"
              aria-multiline="true"
              data-placeholder="Type your message and press Enter..."
            ></div>
          </div>
        </div>
      </div>
    `

    this.container.innerHTML = headerHTML + bodyHTML

    // Create beautiful toggle button with modern icon
    this.toggleButton = document.createElement('button')
    this.toggleButton.id = 'tms-chat-toggle'
    this.toggleButton.className = 'tms-toggle-button'
    this.toggleButton.setAttribute('aria-label', 'Open chat')
    this.toggleButton.innerHTML = `
      ${this.getBubbleStyleIcon(this.widget.chat_bubble_style)}
      ${this.unreadCount > 0 ? `
        <div class="tms-notification-badge">${this.unreadCount > 9 ? '9+' : this.unreadCount}</div>
      ` : ''}
    `

    // Append to document
    document.body.appendChild(this.container)
    document.body.appendChild(this.toggleButton)

    if (this.widget.show_powered_by) {
      const badge = document.createElement('a')
      badge.className = 'tms-powered-badge'
      badge.href = 'https://bareuptime.com/tms'
      badge.target = '_blank'
      badge.rel = 'noopener noreferrer'
      badge.textContent = 'Powered by TMS'
      badge.style.display = 'none' // Initially hidden
      document.body.appendChild(badge)
      this.poweredBadge = badge
    }

    // Add event listeners
    this.attachEventListeners()

    // Show initial message if we have cached messages
    if (this.messages.length > 0) {
      this.messages.forEach(msg => this.displayMessage(msg))
    } else {
      // Always show welcome message for new or cleared sessions
      this.showWelcomeMessage()
    }
  }

  private showWelcomeMessage() {
    if (!this.widget) return
    
    const welcomeMsg = this.widget.custom_greeting || this.widget.welcome_message
    if (!welcomeMsg) return
    
    const message: ChatMessage = {
      id: 'welcome-' + Date.now(),
      content: welcomeMsg,
      author_type: 'system',
      author_name: this.widget.agent_name,
      created_at: new Date().toISOString(),
      message_type: 'text',
      is_private: false
    }
    
    this.displayMessage(message)
  }

  private displayMessage(message: ChatMessage) {
    const messagesContainer = document.getElementById('tms-chat-messages')
    if (!messagesContainer) return

    if (messagesContainer.style.display === 'none') {
      messagesContainer.style.display = 'block'
    }

    const messageWrapper = document.createElement('div')
    messageWrapper.className = `tms-message-wrapper ${message.author_type}`

    const isAgent = message.author_type === 'agent' || message.author_type === 'ai-agent'

    // Create message bubble with enhanced styling
    const messageBubble = document.createElement('div')
    messageBubble.className = `tms-message-bubble ${message.author_type}`
    messageBubble.innerHTML = this.escapeHtml(message.content)

    // Create timestamp
    const timestamp = document.createElement('div')
    timestamp.className = 'tms-message-time'
    timestamp.textContent = new Date(message.created_at).toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit' 
    })

    messageWrapper.appendChild(messageBubble)
    messageWrapper.appendChild(timestamp)
    messagesContainer.appendChild(messageWrapper)

    // Auto-scroll to bottom with smooth behavior
    requestAnimationFrame(() => {
      messagesContainer.scrollTop = messagesContainer.scrollHeight
    })

    // Increment unread count if widget is closed and message is from agent
    if (!this.isOpen && isAgent) {
      this.unreadCount++
      this.updateNotificationBadge()
    }

    // Play notification sound for new messages
    if (isAgent && this.widget?.sound_enabled) {
      playNotificationSound('message', true)
    }
  }

  private attachEventListeners() {
    if (!this.container || !this.toggleButton) return

    const clearButton = this.container.querySelector('.tms-header-clear')
    const input = this.container.querySelector('#tms-chat-input') as HTMLElement
    const fileInput = this.container.querySelector('#tms-file-input') as HTMLInputElement
    const thumbUpBtn = this.container.querySelector('#tms-thumb-up') as HTMLButtonElement
    const thumbDownBtn = this.container.querySelector('#tms-thumb-down') as HTMLButtonElement
    
    // Visitor form elements
    const visitorForm = this.container.querySelector('#tms-visitor-form') as HTMLFormElement
    const visitorCancelButton = this.container.querySelector('#tms-visitor-cancel')
    const visitorNameInput = this.container.querySelector('#tms-visitor-name') as HTMLInputElement
  // const visitorEmailInput = this.container.querySelector('#tms-visitor-email') as HTMLInputElement

    // Toggle button
    this.toggleButton.addEventListener('click', () => this.toggle())

    // Close button
    clearButton?.addEventListener('click', () => this.clear())

    // Visitor form submission
    visitorForm?.addEventListener('submit', (e) => {
      e.preventDefault()
      this.handleVisitorFormSubmit()
    })

    visitorCancelButton?.addEventListener('click', () => {
      this.hideVisitorForm()
      this.close()
    })

    // Auto-focus first empty field when form is shown
    visitorNameInput?.addEventListener('focus', () => {
      if (!visitorNameInput.value) {
        visitorNameInput.focus()
      }
    })

    // Input field - auto-resize and send on Enter
    input?.addEventListener('keydown', (e: KeyboardEvent) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault()
        this.sendMessage()
      }
    })

    input?.addEventListener('input', () => {
      this.handleTyping()
      this.autoResizeEditable(input)
    })

    // Stop typing when user stops typing (keyup event)
    input?.addEventListener('keyup', () => {
      // If input is empty, stop typing immediately
      if (!input.textContent || !input.textContent.trim()) {
        this.stopTyping()
      }
    })

    // Stop typing when input loses focus
    input?.addEventListener('blur', () => {
      this.stopTyping()
    })

  // File upload
    fileInput?.addEventListener('change', (e) => {
      const files = (e.target as HTMLInputElement).files
      if (files && files.length > 0) {
        this.handleFileUpload(files[0])
      }
    })

  // Reaction buttons with animation
  thumbUpBtn?.addEventListener('click', () => {
    this.animateThumb(thumbUpBtn, '👍')
    this.sendQuickReaction('👍')
  })
  thumbDownBtn?.addEventListener('click', () => {
    this.animateThumb(thumbDownBtn, '👎')
    this.sendQuickReaction('👎')
  })

    // Close on Escape key
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && this.isOpen) {
        this.close()
      }
    })
  }

  private autoResizeEditable(el: HTMLElement) {
    el.style.height = 'auto'
    const maxHeight = 100
    const newHeight = Math.min(el.scrollHeight, maxHeight)
    el.style.height = newHeight + 'px'
    el.style.overflowY = el.scrollHeight > maxHeight ? 'auto' : 'hidden'
  }

  private animateThumb(button: HTMLButtonElement, emoji: string) {
    // Add animation class and temporarily change emoji
    button.classList.add('tms-thumb-animate')
    const originalEmoji = button.textContent
    button.textContent = emoji
    
    // Remove animation after it completes
    setTimeout(() => {
      button.classList.remove('tms-thumb-animate')
      button.textContent = originalEmoji
    }, 600)
  }

  private async handleFileUpload(file: File) {
    if (!this.session || !this.widget?.allow_file_uploads) return

    const maxSize = 10 * 1024 * 1024 // 10MB
    if (file.size > maxSize) {
      this.showError('File size must be less than 10MB')
      return
    }

    try {
      // Here you would implement file upload to your backend
      // For now, just show a placeholder message
      const message: ChatMessage = {
        id: 'file-' + Date.now(),
        content: `📎 Uploaded: ${file.name}`,
        author_type: 'visitor',
        author_name: 'You',
        created_at: new Date().toISOString(),
        message_type: 'file',
        is_private: false
      }
      
      this.addMessage(message)
    } catch (error) {
      this.showError('Failed to upload file')
    }
  }

  private showError(message: string) {
    // Create a temporary error message
    const errorMsg: ChatMessage = {
      id: 'error-' + Date.now(),
      content: `⚠️ ${message}`,
      author_type: 'system',
      author_name: 'System',
      created_at: new Date().toISOString(),
      message_type: 'text',
      is_private: false
    }
    
    this.displayMessage(errorMsg)
    
    if (this.widget?.sound_enabled) {
      playNotificationSound('error', true)
    }
  }

  private showVisitorForm() {
    const visitorForm = document.getElementById('tms-visitor-info-form')
    const messagesContainer = document.getElementById('tms-chat-messages')
    
    if (visitorForm && messagesContainer) {
      visitorForm.style.display = 'block'
      messagesContainer.style.display = 'none'
      
      // Focus the name input
      setTimeout(() => {
        const nameInput = document.getElementById('tms-visitor-name') as HTMLInputElement
        nameInput?.focus()
      }, 100)
    }
  }

  private hideVisitorForm() {
    const visitorForm = document.getElementById('tms-visitor-info-form')
    const messagesContainer = document.getElementById('tms-chat-messages')
    
    if (visitorForm && messagesContainer) {
      visitorForm.style.display = 'none'
      messagesContainer.style.display = 'block'
    }
  }

  private async handleVisitorFormSubmit() {
    console.log("1")
    const nameInput = document.getElementById('tms-visitor-name') as HTMLInputElement
    const emailInput = document.getElementById('tms-visitor-email') as HTMLInputElement
    console.log("2")
    if (!nameInput) return
    console.log("3")
    const name = nameInput.value.trim()
    const email = emailInput?.value.trim()
    console.log("4")
    // Validate required fields
    if (!name) {
      nameInput.focus()
      return
    }
    console.log("5")
    if (this.widget?.require_email && !email) {
      emailInput?.focus()
      return
    }
    console.log("6")
    // Store visitor info for future sessions
    const fingerprint = await generateVisitorFingerprint()
    this.storage.saveVisitorInfo({
      name,
      email,
      fingerprint,
      last_visit: new Date().toISOString()
    })
    console.log("7")
    // Hide the form and show messages
    this.hideVisitorForm()
    
    // Show welcome message now that we have visitor info
    this.showWelcomeMessage()
    console.log("8")
    console.log("9")
    // Start the chat session with visitor info
    await this.startChatSessionWithVisitorInfo({ name, email })
  }

  private async startChatSessionWithVisitorInfo(visitorInfo: { name: string; email?: string }) {
    if (!this.widget) return

    try {
      const fingerprint = await generateVisitorFingerprint()
      
      const sessionData: any = {
        visitor_name: visitorInfo.name,
        visitor_email: visitorInfo.email,
        initial_message: this.widget.welcome_message,
        visitor_info: {
          fingerprint,
          user_agent: navigator.userAgent,
          timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
          language: navigator.language
        }
      }

      const sessionResponse = await this.api.initiateChat(this.widget.id, sessionData)
      
      // Create session object
      this.session = {
        id: sessionResponse.session_id,
        token: sessionResponse.session_token,
        widget_id: this.widget.id,
        status: 'active',
        visitor_name: visitorInfo.name
      }
      
      // Save session to storage
      this.storage.saveSession({
        session_id: this.session.id,
        token: this.session.token,
        widget_id: this.widget.id,
        visitor_name: visitorInfo.name,
        visitor_email: visitorInfo.email,
        created_at: new Date().toISOString(),
        last_activity: new Date().toISOString()
      })

      // Connect WebSocket for real-time communication
      this.connectWebSocket()

      this.emitter.emit('session:started', this.session)

    } catch (error) {
      console.error('Failed to start chat session:', error)
      this.emitter.emit('error', 'Failed to start chat session')
      
      // Show error in chat
      this.showError('Unable to connect. Please try again.')
      
      // Show the visitor form again
      this.showVisitorForm()
    }
  }

  private shouldOpenVisitorForm(): boolean {
    // Check if visitor form is required
    if (this.widget?.require_email && !this.storage.getVisitorInfo()?.email) {
      console.log("1")
      return true
    }
    if (this.widget?.require_name && !this.storage.getVisitorInfo()?.name) {
      console.log("2")
      return true
    }
    console.log("4")
    return false
  }

  private async open() {
    if (!this.widget || !this.container) return
    console.log("visitor form opened")

    // Add opening animation class
    this.container.classList.add('opening')
    this.container.classList.add('open')
    this.isOpen = true

    // Update toggle button accessibility
    if (this.toggleButton) {
      this.toggleButton.setAttribute('aria-label', 'Close chat')
    }
    if (this.poweredBadge) {
      this.poweredBadge.style.display = 'block'
      this.poweredBadge.classList.add('open')
    }

    // Start chat session if not already started
    if (!this.session) {
      console.log("no chat session detected")
      if (!this.shouldOpenVisitorForm()){
        console.log("no visitor form opened")
        this.startChatSessionWithVisitorInfo({
          name: '',
          email: ''
        })
      }
      // Check if we have stored visitor info to auto-start session
      const storedVisitor = this.storage.getVisitorInfo()
      if (storedVisitor && storedVisitor.name) {
        // Auto-start session with stored visitor info
        await this.startChatSessionWithVisitorInfo({
          name: storedVisitor.name,
          email: storedVisitor.email
        })
      } else {
        // Show visitor form for new users
        console.log("no stored visitor info found")
        this.showVisitorForm()
      }
    } else {
      // Update activity for existing session
      console.log("chat session already started")
      this.storage.updateSessionActivity()
      
      // Ensure we have WebSocket connection
      if (!this.isConnected) {
        this.connectWebSocket()
      }
    }

    // Clear unread count and send read receipts for unread agent messages
    if (this.unreadCount > 0) {
      this.unreadCount = 0
      this.updateNotificationBadge()
      
      // Send read receipts for unread agent messages
      this.markUnreadMessagesAsRead()
    }

    // Focus input after opening
    setTimeout(() => {
      const input = document.getElementById('tms-chat-input')
      input?.focus()
    }, 300)
  }

  private markUnreadMessagesAsRead() {
    if (!this.isConnected) return

    // Find recent agent messages that haven't been read
    const recentAgentMessages = this.messages
      .filter(m => m.author_type === 'agent' || m.author_type === 'ai-agent')
      .slice(-5) // Only last 5 messages to avoid spam
    
    recentAgentMessages.forEach(message => {
      this.sendReadReceipt(message.id)
    })
  }

  private close() {
    if (!this.container) return
    
    this.container.classList.add('closing')
    this.container.classList.remove('open')
    this.isOpen = false

    // Update toggle button
    if (this.toggleButton) {
      this.toggleButton.setAttribute('aria-label', 'Open chat')
    }
    if (this.poweredBadge) {
      this.poweredBadge.classList.remove('open')
      this.poweredBadge.style.display = 'none'
    }

    // Clear notification badge
    this.unreadCount = 0
    this.updateNotificationBadge()

    // Save widget state
    this.storage.saveWidgetState({
      is_minimized: false,
      unread_count: 0,
      last_interaction: new Date().toISOString()
    })

    setTimeout(() => {
      this.container?.classList.remove('opening', 'closing')
    }, 300)
  }

  private clear() {
    this.storage.clearSession()
    this.session = null
    this.messages = []
    this.unreadCount = 0

    // Clear messages from UI
    const messagesContainer = document.getElementById('tms-chat-messages')
    if (messagesContainer) {
      messagesContainer.innerHTML = ''
      messagesContainer.style.display = 'none'
    }

    // Show visitor form again
    this.showVisitorForm()
    console.log("cleared session")
    this.close()
    
  }

  private toggle() {
    if (this.isOpen) {
      this.close()
    } else {
      this.open()
    }
  }

  private updateNotificationBadge() {
    if (!this.toggleButton) return

    const badge = this.toggleButton.querySelector('.tms-notification-badge')
    if (this.unreadCount > 0) {
      if (!badge) {
        const badgeEl = document.createElement('div')
        badgeEl.className = 'tms-notification-badge'
        badgeEl.textContent = this.unreadCount > 9 ? '9+' : this.unreadCount.toString()
        this.toggleButton.appendChild(badgeEl)
      } else {
        badge.textContent = this.unreadCount > 9 ? '9+' : this.unreadCount.toString()
      }
    } else if (badge) {
      badge.remove()
    }
  }

  private updateToggleButtonIcon() {
    if (!this.toggleButton || !this.widget) return
    
    const currentBadge = this.toggleButton.querySelector('.tms-notification-badge')
    const badgeHTML = currentBadge ? currentBadge.outerHTML : ''
    
    this.toggleButton.innerHTML = `
      ${this.getBubbleStyleIcon(this.widget.chat_bubble_style)}
      ${badgeHTML}
    `
  }

  private connectWebSocket() {
    if (!this.session) return

    try {
      const wsUrl = this.api.getWebSocketUrl(this.session.token, this.session.widget_id)
      this.websocket = new WebSocket(wsUrl)

      this.websocket.onopen = () => {
        console.log('WebSocket connected')
        this.reconnectAttempts = 0
        this.isConnected = true
        this.updateStatus('Connected')

      }

      this.websocket.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data)
          this.handleWebSocketMessage(message)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      this.websocket.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason)
        this.isConnected = false
        this.updateStatus(this.isBusinessHoursOpen ? 'Connecting...' : 'Away')
        
        // Clear typing state on disconnect
        this.stopTyping()
        
        // Attempt to reconnect with exponential backoff
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          this.reconnectAttempts++
          const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), 30000)
          console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)
          setTimeout(() => this.connectWebSocket(), delay)
        } else {
          this.updateStatus('Connection failed')
        }
      }

      this.websocket.onerror = (error) => {
        console.error('WebSocket error:', error)
        this.isConnected = false
        this.updateStatus('Connection error')
      }

    } catch (error) {
      console.error('Failed to connect WebSocket:', error)
      this.isConnected = false
    }
  }

  private handleWebSocketMessage(message: WSMessage) {
    switch (message.type) {
      case 'chat_message':
        // Only add message if it's not already in our local messages (avoid duplicates)
        const existingMessage = this.messages.find(m => m.id === message.data.id)
        if (!existingMessage) {
          this.addMessage(message.data)
          this.emitter.emit('message:received', message.data)
          
          // Increment unread count if widget is closed and message is from agent
          if (!this.isOpen && message.data.author_type === 'agent') {
            this.unreadCount++
            this.updateNotificationBadge()
            
            if (this.widget?.sound_enabled) {
              playNotificationSound('notification', true)
            }
          }
          
          // Send read receipt if widget is open
          if (this.isOpen && message.data.author_type === 'agent') {
            this.sendReadReceipt(message.data.id)
          }
        }
        break

      case 'agent_joined':
        this.updateStatus(`${message.data.agent_name} joined`)
        this.emitter.emit('agent:joined', message.data)
        
        // Show system message
        const joinMessage: ChatMessage = {
          id: 'agent-joined-' + Date.now(),
          content: `${message.data.agent_name} has joined the conversation`,
          author_type: 'system',
          author_name: 'System',
          created_at: new Date().toISOString(),
          message_type: 'text',
          is_private: false
        }
        this.addMessage(joinMessage)
        break

      case 'typing_start':
        if (message.data.author_type === 'agent') {
          this.showTypingIndicator(message.data.author_name)
          this.emitter.emit('agent:typing', message.data)
        }
        break

      case 'typing_stop':
        this.hideTypingIndicator()
        break

      case 'message_read':
        // Handle read receipts for sent messages
        this.handleMessageRead(message.data.message_id)
        break

      case 'session_update':
        if (message.data.status === 'ended') {
          this.handleSessionEnd()
        }
        break

      case 'error':
        this.emitter.emit('error', message.data.error)
        this.showError(message.data.error)
        break

      default:
        console.warn('Unknown WebSocket message type:', message.type)
    }
  }

  private handleSessionEnd() {
    this.updateStatus('Session ended')
    
    const endMessage: ChatMessage = {
      id: 'session-ended-' + Date.now(),
      content: 'The conversation has ended. Feel free to start a new chat if you need further assistance.',
      author_type: 'system',
      author_name: 'System',
      created_at: new Date().toISOString(),
      message_type: 'text',
      is_private: false
    }
    
    this.addMessage(endMessage)
    
    // Clear session
    this.session = null
    this.storage.clearSession()
    
    if (this.websocket) {
      this.websocket.close()
      this.websocket = null
    }
  }

  private addMessage(message: ChatMessage) {
    this.messages.push(message)
    this.displayMessage(message)
    
    // Save to storage
    this.storage.addMessage(message)
    this.storage.updateSessionActivity()
  }

  private async sendMessage() {
  const input = document.getElementById('tms-chat-input') as HTMLElement
    if (!input || !this.session) return

  const content = (input.textContent || '').trim()
    if (!content) return

    // Clear the input immediately for better UX
  input.innerHTML = ''
    input.style.height = 'auto'
    
    // Stop typing indicator
    this.stopTyping()

    try {
      // Create temporary message for immediate display
      const tempMessage: ChatMessage = {
        id: 'temp-' + Date.now(),
        content,
        author_type: 'visitor',
        author_name: 'You',
        created_at: new Date().toISOString(),
        message_type: 'text',
        is_private: false
      }

      // Display message immediately
      this.addMessage(tempMessage)

      if (this.isConnected && this.websocket) {
        // Send via WebSocket for real-time delivery
        const wsMessage = {
          type: 'chat_message',
          client_session_id: this.session.id,
          data: {
            content,
            message_type: 'text',
            author_type: 'visitor',
            author_name: 'You'
          }
        }
        
        this.websocket.send(JSON.stringify(wsMessage))
        this.emitter.emit('message:sent', tempMessage)
      } else {
        console.error('WebSocket not connected, not able to send the chat message')
      }
      
    } catch (error) {
      console.error('Failed to send message:', error)
      this.emitter.emit('error', 'Failed to send message')
      this.showError('Failed to send message. Please try again.')
      
      // Remove the temporary message on error
      const tempIndex = this.messages.findIndex(m => m.id.startsWith('temp-'))
      if (tempIndex !== -1) {
        this.messages.splice(tempIndex, 1)
        // Refresh the display
        this.refreshMessages()
      }
    }
  }

  private sendQuickReaction(emoji: '👍' | '👎') {
    const input = document.getElementById('tms-chat-input') as HTMLElement
    if (!input) return
    input.textContent = emoji
    this.sendMessage()
  }

  private handleTyping() {
    if (!this.isConnected || !this.websocket || !this.session) return

    // Clear existing timeout
    if (this.typingTimeout) {
      clearTimeout(this.typingTimeout)
      this.typingTimeout = null
    }

    // Send typing start if not already typing
    if (!this.isTyping) {
      this.isTyping = true
      this.sendTypingIndicator(true)
    }

    // Set timeout to stop typing after 2 seconds of inactivity
    this.typingTimeout = window.setTimeout(() => {
      this.stopTyping()
    }, 2000)
  }

  private sendTypingIndicator(isTyping: boolean) {
    if (!this.isConnected || !this.websocket || !this.session) return

    try {
      const message = {
        type: isTyping ? 'typing_start' : 'typing_stop',
        client_session_id: this.session.id,
        data: {
          author_type: 'visitor',
          author_name: 'You'
        }
      }
      
      this.websocket.send(JSON.stringify(message))
    } catch (error) {
      console.error('Failed to send typing indicator:', error)
    }
  }

  private stopTyping() {
    if (this.isTyping) {
      this.isTyping = false
      this.sendTypingIndicator(false)
    }
    
    if (this.typingTimeout) {
      clearTimeout(this.typingTimeout)
      this.typingTimeout = null
    }
  }

  private showTypingIndicator(agentName: string) {
    const typingEl = document.getElementById('tms-chat-typing')
    const typingText = document.getElementById('tms-typing-text')
    if (typingEl && typingText) {
      typingText.textContent = `${agentName} is typing...`
      typingEl.style.display = 'flex'
    }
  }

  private hideTypingIndicator() {
    const typingEl = document.getElementById('tms-chat-typing')
    if (typingEl) {
      typingEl.style.display = 'none'
    }
  }

  private updateStatus(status: string) {
    const statusEl = this.container?.querySelector('.tms-agent-status')
    if (statusEl) {
      // Update the status text while keeping the indicator
      const indicator = statusEl.querySelector('.tms-status-indicator')
      statusEl.innerHTML = ''
      if (indicator) statusEl.appendChild(indicator)
      statusEl.appendChild(document.createTextNode(status))
    }
  }

  private sendReadReceipt(messageId: string) {
    if (!this.isConnected || !this.websocket || !this.session) return

    try {
      const message = {
        type: 'message_read',
        client_session_id: this.session.id,
        data: {
          message_id: messageId,
          read_by: 'visitor'
        }
      }
      
      this.websocket.send(JSON.stringify(message))
    } catch (error) {
      console.error('Failed to send read receipt:', error)
    }
  }

  private handleMessageRead(messageId: string) {
    // Find the message and mark it as read
    const message = this.messages.find(m => m.id === messageId)
    if (message && message.author_type === 'visitor') {
      // Add visual indicator for read receipt if needed
      console.log(`Message ${messageId} was read by agent`)
    }
  }

  private refreshMessages() {
    const messagesContainer = document.getElementById('tms-chat-messages')
    if (!messagesContainer) return

    // Clear container
    messagesContainer.innerHTML = ''
    
    // Re-display all messages
    this.messages.forEach(message => this.displayMessage(message))
  }

  private escapeHtml(text: string): string {
    const map: Record<string, string> = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#039;'
    }
    return text.replace(/[&<>"']/g, (m) => map[m])
  }

  // Public API
  public on<K extends keyof Events>(event: K, handler: (data: Events[K]) => void) {
    this.emitter.on(event, handler)
  }

  public off<K extends keyof Events>(event: K, handler: (data: Events[K]) => void) {
    this.emitter.off(event, handler)
  }

  public destroy() {
    // Clean up typing timeout
    if (this.typingTimeout) {
      clearTimeout(this.typingTimeout)
      this.typingTimeout = null
    }

    // Stop typing indicator
    this.stopTyping()

    if (this.websocket) {
      this.websocket.close()
    }
    if (this.container) {
      this.container.remove()
    }
    if (this.toggleButton) {
      this.toggleButton.remove()
    }
    
    // Remove styles
    const styles = document.getElementById('tms-widget-styles')
    if (styles) {
      styles.remove()
    }
    
    // Clear storage if requested
    if (this.options.enableSessionPersistence === false) {
      this.storage.cleanup()
    }
  }

  // Public API methods for external control
  public openWidget() {
    this.open()
  }

  public closeWidget() {
    this.close()
  }

  public toggleWidget() {
    this.toggle()
  }

  public updateWidgetConfig(updates: Partial<ChatWidget>) {
    if (this.widget) {
      Object.assign(this.widget, updates)
      injectWidgetCSS(this.widget)
      this.updateToggleButtonIcon()
    }
  }
}
