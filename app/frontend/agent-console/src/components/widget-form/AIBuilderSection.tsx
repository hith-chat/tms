import { useState } from 'react'
import { Wand2, Globe, AlertCircle, CheckCircle, Loader2 } from 'lucide-react'
import type { CreateChatWidgetRequest } from '../../hooks/useChatWidgetForm'
import { apiClient } from '../../lib/api'

interface AIBuilderSectionProps {
  onThemeGenerated: (theme: Partial<CreateChatWidgetRequest>) => void
  loading: boolean
  error: string | null
  onLoadingChange?: (loading: boolean) => void
  onError?: (error: string | null) => void
}

export function AIBuilderSection({
  onThemeGenerated,
  loading,
  error,
  onLoadingChange,
  onError
}: AIBuilderSectionProps) {
  const [url, setUrl] = useState('')
  const [urlError, setUrlError] = useState<string | null>(null)

  const validateUrl = (input: string): boolean => {
    try {
      const urlObj = new URL(input.startsWith('http') ? input : `https://${input}`)
      return urlObj.protocol === 'http:' || urlObj.protocol === 'https:'
    } catch {
      return false
    }
  }

  const handleUrlChange = (value: string) => {
    setUrl(value)
    setUrlError(null)
  }

  const handleGenerate = async () => {
    if (!url.trim()) {
      setUrlError('Please enter a website URL')
      return
    }

    const normalizedUrl = url.startsWith('http') ? url : `https://${url}`

    if (!validateUrl(normalizedUrl)) {
      setUrlError('Please enter a valid website URL')
      return
    }

    onLoadingChange?.(true)
    onError?.(null)
    setUrlError(null)

    try {
      const themeData = await apiClient.scrapeWebsiteTheme(normalizedUrl)
      onThemeGenerated(themeData)
    } catch (err: any) {
      const errorMessage = err.message || 'Failed to analyze website. Please try again.'
      setUrlError(errorMessage)
      onError?.(errorMessage)
    } finally {
      onLoadingChange?.(false)
    }
  }

  return (
    <div className="flex flex-col w-full min-w-0">
      <div className="rounded-lg border border-border bg-card text-card-foreground shadow-sm">
        {/* Header */}
        <div className="flex items-center gap-3 p-6 pb-4">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10">
            <Wand2 className="h-4 w-4 text-primary" aria-hidden="true" />
          </div>
          <div className="flex flex-col space-y-1">
            <h3 className="text-base font-semibold leading-none tracking-tight">
              AI Theme Generator
            </h3>
            <p className="text-sm text-muted-foreground">
              Enter your website URL and let AI automatically generate a matching theme
            </p>
          </div>
        </div>

        {/* Form content */}
        <div className="px-6 pb-6">
          <div className="space-y-4">
            {/* URL Input */}
            <div className="space-y-2">
              <label
                htmlFor="website-url"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              >
                Website URL <span className="text-destructive">*</span>
              </label>
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                    <Globe className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <input
                    id="website-url"
                    type="url"
                    value={url}
                    onChange={(e) => handleUrlChange(e.target.value)}
                    placeholder="example.com or https://example.com"
                    className={`flex h-10 w-full rounded-md border bg-background pl-10 pr-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 ${
                      urlError ? 'border-destructive' : 'border-input'
                    }`}
                    disabled={loading}
                  />
                </div>
                <button
                  type="button"
                  onClick={handleGenerate}
                  disabled={loading || !url.trim()}
                  className="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2"
                >
                  {loading ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      Analyzing...
                    </>
                  ) : (
                    <>
                      <Wand2 className="h-4 w-4" />
                      Generate Theme
                    </>
                  )}
                </button>
              </div>
              {urlError && (
                <div className="flex items-center gap-2 text-sm text-destructive">
                  <AlertCircle className="h-4 w-4" />
                  {urlError}
                </div>
              )}
            </div>

            {/* Error Message */}
            {error && (
              <div className="flex items-center gap-2 p-3 rounded-md bg-destructive/10 border border-destructive/20 text-destructive">
                <AlertCircle className="h-4 w-4" />
                <span className="text-sm">{error}</span>
              </div>
            )}

            {/* Info Box */}
            <div className="rounded-md border border-border bg-muted/50 p-4">
              <div className="flex items-start gap-3">
                <CheckCircle className="h-5 w-5 text-green-600 mt-0.5 flex-shrink-0" />
                <div className="space-y-2 text-sm">
                  <p className="font-medium text-foreground">
                    How AI Theme Generation Works:
                  </p>
                  <ul className="space-y-1 text-muted-foreground">
                    <li>• Analyzes your website's color palette and design</li>
                    <li>• Extracts brand colors, fonts, and visual style</li>
                    <li>• Generates a matching chat widget theme</li>
                    <li>• Creates personalized welcome messages</li>
                  </ul>
                  <p className="text-xs text-muted-foreground mt-3">
                    You can always customize the generated theme in Manual Mode.
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