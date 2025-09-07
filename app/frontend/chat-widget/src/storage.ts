// Session storage and persistence utilities
import type { SessionData, ChatMessage } from './types'

const STORAGE_KEYS = {
  SESSION: 'tms_chat_session',
  MESSAGES: 'tms_chat_messages',
  VISITOR_INFO: 'tms_visitor_info',
  WIDGET_STATE: 'tms_widget_state'
} as const

export interface StoredVisitorInfo {
  name: string
  email?: string
  fingerprint: string
  last_visit: string
}

export interface WidgetStateData {
  is_minimized: boolean
  unread_count: number
  last_interaction: string
}

export class SessionStorage {
  private storagePrefix: string

  constructor(widgetId: string) {
    // we only need a storage prefix derived from widgetId; the raw widgetId
    // is not read elsewhere so we avoid keeping an unused field.
    this.storagePrefix = `tms_${widgetId}_`
  }

  private getKey(key: string): string {
    return `${this.storagePrefix}${key}`
  }

  private isStorageAvailable(): boolean {
    try {
      const test = '__tms_storage_test__'
      localStorage.setItem(test, 'test')
      localStorage.removeItem(test)
      return true
    } catch {
      return false
    }
  }

  // Session management
  saveSession(session: SessionData): void {
    if (!this.isStorageAvailable()) return

    try {
      const sessionData = {
        ...session,
        last_activity: new Date().toISOString()
      }
      localStorage.setItem(this.getKey(STORAGE_KEYS.SESSION), JSON.stringify(sessionData))
    } catch (error) {
      console.warn('Failed to save session:', error)
    }
  }

  getSession(): SessionData | null {
    if (!this.isStorageAvailable()) return null

    try {
      const stored = localStorage.getItem(this.getKey(STORAGE_KEYS.SESSION))
      if (!stored) return null
      const session = JSON.parse(stored) as SessionData
      return session
    } catch (error) {
      console.warn('Failed to get session:', error)
      return null
    }
  }

  clearSession(): void {
    if (!this.isStorageAvailable()) return

    try {
      localStorage.removeItem(this.getKey(STORAGE_KEYS.SESSION))
      localStorage.removeItem(this.getKey(STORAGE_KEYS.MESSAGES))
    } catch (error) {
      console.warn('Failed to clear session:', error)
    }
  }

  updateSessionActivity(): void {
    const session = this.getSession()
    if (session) {
      session.last_activity = new Date().toISOString()
      this.saveSession(session)
    }
  }

  // Messages management
  saveMessages(messages: ChatMessage[]): void {
    if (!this.isStorageAvailable()) return

    try {
      // Only store last 50 messages to prevent storage bloat
      const messagesToStore = messages.slice(-50)
      localStorage.setItem(this.getKey(STORAGE_KEYS.MESSAGES), JSON.stringify(messagesToStore))
    } catch (error) {
      console.warn('Failed to save messages:', error)
    }
  }

  getMessages(): ChatMessage[] {
    if (!this.isStorageAvailable()) return []

    try {
      const stored = localStorage.getItem(this.getKey(STORAGE_KEYS.MESSAGES))
      return stored ? JSON.parse(stored) : []
    } catch (error) {
      console.warn('Failed to get messages:', error)
      return []
    }
  }

  addMessage(message: ChatMessage): void {
    const messages = this.getMessages()
    messages.push(message)
    this.saveMessages(messages)
  }

  // Visitor info management
  saveVisitorInfo(info: StoredVisitorInfo): void {
    if (!this.isStorageAvailable()) return

    try {
      localStorage.setItem(this.getKey(STORAGE_KEYS.VISITOR_INFO), JSON.stringify(info))
    } catch (error) {
      console.warn('Failed to save visitor info:', error)
    }
  }

  getVisitorInfo(): StoredVisitorInfo | null {
    if (!this.isStorageAvailable()) return null

    try {
      const stored = localStorage.getItem(this.getKey(STORAGE_KEYS.VISITOR_INFO))
      return stored ? JSON.parse(stored) : null
    } catch (error) {
      console.warn('Failed to get visitor info:', error)
      return null
    }
  }

  // Widget state management
  saveWidgetState(state: WidgetStateData): void {
    if (!this.isStorageAvailable()) return

    try {
      localStorage.setItem(this.getKey(STORAGE_KEYS.WIDGET_STATE), JSON.stringify(state))
    } catch (error) {
      console.warn('Failed to save widget state:', error)
    }
  }

  getWidgetState(): WidgetStateData | null {
    if (!this.isStorageAvailable()) return null

    try {
      const stored = localStorage.getItem(this.getKey(STORAGE_KEYS.WIDGET_STATE))
      return stored ? JSON.parse(stored) : null
    } catch (error) {
      console.warn('Failed to get widget state:', error)
      return null
    }
  }

  cleanup(): void {
    if (!this.isStorageAvailable()) return

    try {
      // Remove old storage entries for this widget
      const keysToRemove: string[] = []
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i)
        if (key && key.startsWith(this.storagePrefix)) {
          keysToRemove.push(key)
        }
      }
      
      keysToRemove.forEach(key => localStorage.removeItem(key))
    } catch (error) {
      console.warn('Failed to cleanup storage:', error)
    }
  }

  importData(data: Record<string, any>): void {
    if (data.session) this.saveSession(data.session)
    if (data.messages) this.saveMessages(data.messages)
    if (data.visitorInfo) this.saveVisitorInfo(data.visitorInfo)
    if (data.widgetState) this.saveWidgetState(data.widgetState)
  }
}

// Global utilities
export function generateVisitorFingerprint(): Promise<string> {
  return new Promise((resolve) => {
    // Mixpanel-style fingerprinting approach
    const getDeviceFingerprint = () => {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      let canvasHash = ''
      
      if (ctx) {
        canvas.width = 200
        canvas.height = 50
        ctx.textBaseline = 'top'
        ctx.font = '14px Arial'
        ctx.fillStyle = '#f60'
        ctx.fillRect(125, 1, 62, 20)
        ctx.fillStyle = '#069'
        ctx.fillText('🌍 Hello, world! 123', 2, 15)
        ctx.fillStyle = '#f00'
        ctx.fillText('Canvas fingerprint', 2, 30)
        canvasHash = ctx.getImageData(0, 0, canvas.width, canvas.height).data.slice(0, 100).join('')
      }
      
      // Collect comprehensive device characteristics
      const characteristics = [
        navigator.userAgent,
        navigator.language,
        JSON.stringify(navigator.languages || []),
        screen.width,
        screen.height,
        screen.availWidth,
        screen.availHeight,
        screen.colorDepth,
        screen.pixelDepth,
        new Date().getTimezoneOffset(),
        Intl.DateTimeFormat().resolvedOptions().timeZone || '',
        navigator.platform,
        navigator.cookieEnabled,
        navigator.doNotTrack || '',
        navigator.maxTouchPoints || 0,
        navigator.hardwareConcurrency || 0,
        window.devicePixelRatio || 1,
        // Audio context fingerprint
        (() => {
          try {
            const audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)()
            const oscillator = audioCtx.createOscillator()
            const analyser = audioCtx.createAnalyser()
            const gain = audioCtx.createGain()
            const scriptProcessor = audioCtx.createScriptProcessor(4096, 1, 1)
            
            gain.gain.value = 0
            oscillator.frequency.value = 1000
            oscillator.type = 'triangle'
            
            oscillator.connect(analyser)
            analyser.connect(scriptProcessor)
            scriptProcessor.connect(gain)
            gain.connect(audioCtx.destination)
            
            oscillator.start(0)
            
            const freqData = new Uint8Array(analyser.frequencyBinCount)
            analyser.getByteFrequencyData(freqData)
            
            oscillator.stop()
            audioCtx.close()
            
            return freqData.slice(0, 30).join('')
          } catch {
            return 'no-audio'
          }
        })(),
        canvasHash
      ].join('###')
      
      return characteristics
    }
    
    // Generate multiple hash values like Mixpanel does
    const deviceData = getDeviceFingerprint()
    
    // Create long hash using multiple algorithms (like Mixpanel)
    const createMixpanelStyleHash = (input: string): string => {
      // Multiple hash functions for maximum entropy
      const hashFunctions = [
        // DJB2 hash
        (str: string, seed: number) => {
          let hash = seed
          for (let i = 0; i < str.length; i++) {
            hash = ((hash << 5) + hash) + str.charCodeAt(i)
            hash = hash & 0xFFFFFFFF
          }
          return Math.abs(hash).toString(36)
        },
        
        // FNV-1a hash
        (str: string, seed: number) => {
          let hash = seed
          for (let i = 0; i < str.length; i++) {
            hash ^= str.charCodeAt(i)
            hash = (hash * 16777619) & 0xFFFFFFFF
          }
          return Math.abs(hash).toString(36)
        },
        
        // SDBM hash
        (str: string, seed: number) => {
          let hash = seed
          for (let i = 0; i < str.length; i++) {
            hash = str.charCodeAt(i) + (hash << 6) + (hash << 16) - hash
            hash = hash & 0xFFFFFFFF
          }
          return Math.abs(hash).toString(36)
        },
        
        // Jenkins hash
        (str: string, seed: number) => {
          let hash = seed
          for (let i = 0; i < str.length; i++) {
            hash = (hash + str.charCodeAt(i)) & 0xFFFFFFFF
            hash = (hash + (hash << 10)) & 0xFFFFFFFF
            hash = (hash ^ (hash >>> 6)) & 0xFFFFFFFF
          }
          hash = (hash + (hash << 3)) & 0xFFFFFFFF
          hash = (hash ^ (hash >>> 11)) & 0xFFFFFFFF
          hash = (hash + (hash << 15)) & 0xFFFFFFFF
          return Math.abs(hash).toString(36)
        }
      ]
      
      // Generate multiple segments like Mixpanel
      const seeds = [0x9E3779B9, 0x85EBCA6B, 0xC2B2AE35, 0x27D4EB2F, 0x165667B1, 0xD3A2646C, 0xFD7046C5, 0xB55A4F09]
      let result = ''
      
      // Create 8 segments of 8 characters each
      seeds.forEach((seed, index) => {
        const hashFunc = hashFunctions[index % hashFunctions.length]
        const segment = hashFunc(input + index.toString(), seed).padStart(8, '0').slice(-8)
        result += segment
      })
      
      return result
    }
    
    // Generate the long hash
    const longHash = createMixpanelStyleHash(deviceData)
    
    // Add entropy from current session (but stable within session)
    const sessionSeed = Math.floor(Date.now() / (1000 * 60 * 60)) // Changes hourly
    const entropyHash = createMixpanelStyleHash(deviceData + sessionSeed.toString()).slice(0, 16)
    
    // Combine for final 80-character fingerprint (like Mixpanel's longer IDs)
    const finalFingerprint = longHash + entropyHash
    
    // Return 64 characters for optimal uniqueness (Mixpanel-style length)
    resolve(finalFingerprint.slice(0, 64))
  })
}

export function isBusinessHours(businessHours: any): boolean {
  if (!businessHours?.enabled) return true

  try {
    const now = new Date()
    const dayNames = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat']
    const day = dayNames[now.getDay()]
    const time = now.toTimeString().slice(0, 5) // HH:MM
    
    const schedule = businessHours.schedule?.[day]
    if (!schedule?.enabled) return false
    
    return time >= schedule.open && time <= schedule.close
  } catch {
    return true // Default to available if parsing fails
  }
}
