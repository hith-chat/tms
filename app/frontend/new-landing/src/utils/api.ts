export interface AIBuilderEvent {
  type: string
  stage: string
  message: string
  detail?: string
  data?: Record<string, any>
  timestamp: string
}

export interface WidgetData {
  project_id: string
  build_id: string
  widget_id: string
  expires_at: string
}

export class APIClient {
  private baseURL: string

  constructor() {
    // Use environment variable or default to window origin
    this.baseURL = import.meta.env.VITE_API_URL || window.location.origin
  }

  /**
   * Creates an EventSource connection to stream widget build events
   */
  streamWidgetBuild(
    url: string,
    depth: number = 3,
    onEvent: (event: AIBuilderEvent) => void,
    onError: (error: Error) => void,
    onComplete: (data: WidgetData | null) => void
  ): EventSource {
    const encodedUrl = encodeURIComponent(url)
    const sseUrl = `${this.baseURL}/api/public/ai-widget-builder?url=${encodedUrl}&depth=${depth}`

    const eventSource = new EventSource(sseUrl)

    eventSource.onmessage = (event) => {
      try {
        const data: AIBuilderEvent = JSON.parse(event.data)
        onEvent(data)

        // Check if build is complete
        if (data.type === 'completed') {
          onComplete(data.data as WidgetData)
          eventSource.close()
        } else if (data.type === 'error') {
          onError(new Error(data.message || 'Build failed'))
          eventSource.close()
        }
      } catch (err) {
        console.error('Error parsing SSE data:', err)
        onError(err as Error)
      }
    }

    eventSource.onerror = (err) => {
      console.error('EventSource error:', err)
      onError(new Error('Connection to server lost'))
      eventSource.close()
    }

    return eventSource
  }

  /**
   * Checks if a URL can be loaded in an iframe
   */
  async canLoadInIframe(url: string): Promise<boolean> {
    try {
      const response = await fetch(url, {
        method: 'HEAD',
        mode: 'no-cors',
      })

      // If we can make a request, check X-Frame-Options header
      const xFrameOptions = response.headers.get('X-Frame-Options')
      if (xFrameOptions && (xFrameOptions.toLowerCase() === 'deny' || xFrameOptions.toLowerCase() === 'sameorigin')) {
        return false
      }

      return true
    } catch {
      // If no-cors fails, try to load it anyway
      return true
    }
  }

  /**
   * Gets chat widget by ID
   */
  async getWidget(widgetId: string) {
    const response = await fetch(`${this.baseURL}/api/public/chat/widgets/${widgetId}`)
    if (!response.ok) {
      throw new Error('Failed to fetch widget')
    }
    return response.json()
  }
}

export const apiClient = new APIClient()
