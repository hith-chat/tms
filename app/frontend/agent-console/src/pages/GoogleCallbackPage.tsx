import React, { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Loader2 } from 'lucide-react'
import { apiClient } from '@/lib/api'

export function GoogleCallbackPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const handleCallback = async () => {
      const code = searchParams.get('code')
      const state = searchParams.get('state')
      const errorParam = searchParams.get('error')

      // Check for OAuth errors
      if (errorParam) {
        setError('Google authentication was cancelled or failed')
        setTimeout(() => navigate('/login'), 3000)
        return
      }

      // Check for required parameters
      if (!code || !state) {
        setError('Invalid callback parameters')
        setTimeout(() => navigate('/login'), 3000)
        return
      }

      try {
        // Send code to backend to complete OAuth
        const response = await apiClient.googleOAuthCallback(code, state)

        // Store user data and tokens
        localStorage.setItem('auth_token', response.access_token)
        localStorage.setItem('refresh_token', response.refresh_token)
        localStorage.setItem('user_data', JSON.stringify(response.user))

        // Set tenant ID in API client
        apiClient.setTenantId(response.user.tenant_id)

        // Navigate to dashboard
        window.location.href = '/tickets'
      } catch (err: any) {
        console.error('OAuth callback error:', err)
        setError(err.response?.data?.error || 'Failed to complete Google authentication')
        setTimeout(() => navigate('/login'), 3000)
      }
    }

    handleCallback()
  }, [searchParams, navigate])

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center space-y-4">
          <div className="mx-auto h-16 w-16 rounded-2xl bg-primary flex items-center justify-center mb-6 shadow-lg">
            <img src="/images/logo.svg" alt="Hith Logo" className="h-10 w-10" />
          </div>

          {error ? (
            <>
              <div className="text-destructive">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-12 w-12 mx-auto mb-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <h1 className="text-2xl font-bold text-foreground">Authentication Failed</h1>
              <p className="text-muted-foreground">{error}</p>
              <p className="text-sm text-muted-foreground">Redirecting to login page...</p>
            </>
          ) : (
            <>
              <Loader2 className="h-12 w-12 mx-auto animate-spin text-primary" />
              <h1 className="text-2xl font-bold text-foreground">Completing Sign In</h1>
              <p className="text-muted-foreground">
                Please wait while we complete your Google authentication...
              </p>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
