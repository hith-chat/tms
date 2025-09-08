import { ChatWidget, InitiateChatRequest } from './types'
import { SignJWT, jwtVerify } from 'jose'

export class ChatAPI {
  private baseUrl: string

  constructor(apiUrl?: string) {
    // Priority for base URL:
    // 1. explicit constructor arg
    // 2. Vite env var VITE_API_URL
    // 3. production default -> https://tms.bareuptime.co/api
    // 4. development default -> http://localhost:8080/api
    // const viteEnv = (import.meta as any)?.env
    // const envUrl: string | undefined = viteEnv?.VITE_API_URL
    const mode: string | undefined = 'production' // viteEnv?.MODE
    const defaultUrl = mode === 'production' ? 'https://tms.bareuptime.co/api' : 'http://localhost:8080/api'
    this.baseUrl = defaultUrl
  }

  async getWidgetByDomain(domain: string): Promise<ChatWidget> {
    const response = await fetch(`${this.baseUrl}/public/chat/widgets/domain/${domain}`)
    if (!response.ok) {
      throw new Error(`Failed to get widget: ${response.statusText}`)
    }
    return response.json()
  }

  private async createSessionToken(widgetId: string, request: InitiateChatRequest): Promise<{ chatSessionToken: string, sessionId: string }> {
    const chatToken = localStorage.getItem('chat_session_token')
    if (chatToken) {
      try {
        const { chatSessionToken, sessionId } = JSON.parse(chatToken)
        if (await this.verifySessionToken(widgetId, chatSessionToken)) {
          return { chatSessionToken, sessionId }
        }
      } catch {
        // Invalid stored token, create new one
        localStorage.removeItem('chat_session_token')
      }
    }
    
    const timestamp = Date.now()
    const sessionId = (request.visitor_info?.fingerprint || 'anon') + '_' + timestamp
    const payload = {
      session_id: sessionId,
      widget_id: widgetId,
      visitor_name: request.visitor_name,
      visitor_email: request.visitor_email,
      visitor_info: request.visitor_info,
      timestamp: Date.now(),
      iat: Math.floor(Date.now() / 1000),
      exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24 hours expiration
    }
    
    const secret = new TextEncoder().encode(widgetId)
    const chatSessionToken = await new SignJWT(payload)
      .setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt()
      .setExpirationTime('24h')
      .sign(secret)
    
    const sessionInfo = { chatSessionToken, sessionId }
    localStorage.setItem('chat_session_token', JSON.stringify(sessionInfo))
    return sessionInfo
  }

  public async verifySessionToken(widgetId: string, token: string): Promise<boolean> {
    try {
      const secret = new TextEncoder().encode(widgetId)
      await jwtVerify(token, secret)
      return true
    } catch (error) {
      return false
    }
  }

  async initiateChat(widgetId: string, request: InitiateChatRequest): Promise<{ session_token: string, session_id: string }> {
    // Create a JWT session token that combines widgetId with request information
    const session = await this.createSessionToken(widgetId, request)
    console.log("session:", session)
    // Return the session token
    return { session_token: session.chatSessionToken, session_id: session.sessionId }
  }


  async markMessagesAsRead(sessionToken: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/public/chat/sessions/${sessionToken}/read`, {
      method: 'POST',
    })
    
    if (!response.ok) {
      throw new Error(`Failed to mark messages as read: ${response.statusText}`)
    }
  }

  getWebSocketUrl(sessionToken: string, widgetId: string): string {
    const wsUrl = this.baseUrl.replace('http', 'ws')
    return `${wsUrl}/public/chat/ws/widgets/${widgetId}/chat/${sessionToken}`
  }
}
