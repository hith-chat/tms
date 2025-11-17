import React, { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Loader2, AlertCircle } from 'lucide-react'
import { apiClient } from '../lib/api'

const Alert = ({ children, variant = 'default' }: any) => (
  <div className={`relative w-full rounded-lg border p-4 ${
    variant === 'destructive' 
      ? 'border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive' 
      : 'border-border'
  }`}>
    {children}
  </div>
)

export function OAuthCallbackPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)
  const [isProcessing, setIsProcessing] = useState(true)

  useEffect(() => {
    const handleCallback = async () => {
      const code = searchParams.get('code')
      const state = searchParams.get('state')
      const errorParam = searchParams.get('error')

      // Check if user denied access
      if (errorParam) {
        setError('Google login was cancelled or denied')
        setIsProcessing(false)
        setTimeout(() => navigate('/login'), 3000)
        return
      }

      // Validate required parameters
      if (!code || !state) {
        setError('Invalid OAuth callback - missing parameters')
        setIsProcessing(false)
        setTimeout(() => navigate('/login'), 3000)
        return
      }

      try {
        // Exchange code for tokens
        const response = await apiClient.handleGoogleCallback(code, state)
        
        // Store user data
        localStorage.setItem('user_data', JSON.stringify(response.user))
        
        // Redirect to dashboard
        window.location.assign('/tickets')
      } catch (err: any) {
        console.error('OAuth callback error:', err)
        setError(err.response?.data?.error || 'Failed to complete Google login')
        setIsProcessing(false)
        setTimeout(() => navigate('/login'), 3000)
      }
    }

    handleCallback()
  }, [searchParams, navigate])

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <div className="w-full max-w-md space-y-8">
        {/* Header */}
        <div className="text-center space-y-2">
          <div className="mx-auto h-16 w-16 rounded-2xl bg-primary flex items-center justify-center mb-6 shadow-lg">
            <img src="/images/logo.svg" alt="Hith Logo" className="h-10 w-10" />
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-foreground">
            {isProcessing ? 'Completing Sign In' : 'Sign In Failed'}
          </h1>
          <p className="text-muted-foreground">
            {isProcessing 
              ? 'Please wait while we complete your Google sign in...'
              : 'Redirecting you back to login...'
            }
          </p>
        </div>

        {/* Content Card */}
        <div className="rounded-lg border bg-card text-card-foreground shadow-sm p-8">
          {isProcessing ? (
            <div className="flex flex-col items-center justify-center space-y-4">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
              <p className="text-sm text-muted-foreground text-center">
                Authenticating with Google...
              </p>
            </div>
          ) : error ? (
            <Alert variant="destructive">
              <div className="flex items-center space-x-2">
                <AlertCircle className="h-4 w-4" />
                <span className="text-sm font-medium">{error}</span>
              </div>
            </Alert>
          ) : null}
        </div>

        {/* Footer */}
        <div className="text-center">
          <p className="text-xs text-muted-foreground">
            Having trouble? <a href="/login" className="text-primary hover:text-primary/80 font-medium">Return to login</a>
          </p>
        </div>
      </div>
    </div>
  )
}
