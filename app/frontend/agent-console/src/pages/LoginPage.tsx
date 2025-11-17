import React, { useState } from 'react'
import { Eye, EyeOff, Loader2, Shield, Lock, Mail } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { apiClient } from '../lib/api'

// Simplified components matching our enterprise design
const Button = ({ children, variant = 'default', size = 'default', className = '', disabled = false, ...props }: any) => (
  <button 
    className={`inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none ring-offset-background ${
      variant === 'outline' ? 'border border-input hover:bg-accent hover:text-accent-foreground' :
      variant === 'ghost' ? 'hover:bg-accent hover:text-accent-foreground' :
      'bg-primary text-primary-foreground hover:bg-primary/90'
    } ${
      size === 'sm' ? 'h-9 px-3 rounded-md' : 
      size === 'lg' ? 'h-11 px-8 rounded-md' :
      'h-10 py-2 px-4'
    } ${className}`}
    disabled={disabled}
    {...props}
  >
    {children}
  </button>
)

const Input = ({ className = '', error, ...props }: any) => (
  <input
    className={`flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 ${
      error ? 'border-destructive focus-visible:ring-destructive' : ''
    } ${className}`}
    {...props}
  />
)

const Card = ({ children, className = '' }: any) => (
  <div className={`rounded-lg border bg-card text-card-foreground shadow-sm ${className}`}>
    {children}
  </div>
)

const Alert = ({ children, variant = 'default' }: any) => (
  <div className={`relative w-full rounded-lg border p-4 ${
    variant === 'destructive' 
      ? 'border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive' 
      : 'border-border'
  }`}>
    {children}
  </div>
)

export function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const { login } = useAuth()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError('')

    try {
      await login({ email, password })
    } catch (err) {
      setError('Invalid credentials. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const handleDemoLogin = () => {
    setEmail('admin@acme.com')
    setPassword('password')
  }

  const handleGoogleLogin = async () => {
    setIsLoading(true)
    setError('')

    try {
      const { auth_url } = await apiClient.getGoogleOAuthURL()
      // Redirect to Google OAuth
      window.location.href = auth_url
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to initiate Google login')
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <div className="w-full max-w-md space-y-8">
        {/* Header */}
        <div className="text-center space-y-2">
          <div className="mx-auto h-16 w-16 rounded-2xl bg-primary flex items-center justify-center mb-6 shadow-lg">
            <img src="/images/logo.svg" alt="Hith Logo" className="h-10 w-10" />
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-foreground">
            Welcome Back
          </h1>
          <p className="text-muted-foreground">
            Sign in to your Hith account to continue
          </p>
        </div>

        {/* Login Card */}
        <Card className="p-8">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Error Alert */}
            {error && (
              <Alert variant="destructive">
                <div className="flex items-center space-x-2">
                  <Lock className="h-4 w-4" />
                  <span className="text-sm font-medium">{error}</span>
                </div>
              </Alert>
            )}

            {/* Email Field */}
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                Email Address
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                <Input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e: any) => setEmail(e.target.value)}
                  placeholder="Enter your email address"
                  className="pl-10"
                  disabled={isLoading}
                  required
                  autoComplete="email"
                />
              </div>
            </div>

            {/* Password Field */}
            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e: any) => setPassword(e.target.value)}
                  placeholder="Enter your password"
                  className="pl-10 pr-10"
                  disabled={isLoading}
                  required
                  autoComplete="current-password"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  disabled={isLoading}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </div>

            {/* Remember me and Forgot password */}
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <input
                  id="remember"
                  type="checkbox"
                  className="h-4 w-4 rounded border-input text-primary focus:ring-2 focus:ring-ring focus:ring-offset-2"
                />
                <label htmlFor="remember" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                  Remember me
                </label>
              </div>
              <button
                type="button"
                className="text-sm text-primary hover:text-primary/80 font-medium"
              >
                Forgot password?
              </button>
            </div>

            {/* Sign In Button */}
            <Button
              type="submit"
              size="lg"
              className="w-full"
              disabled={isLoading || !email || !password}
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Signing in...
                </>
              ) : (
                <>
                  <Shield className="mr-2 h-4 w-4" />
                  Sign In
                </>
              )}
            </Button>

            {/* Divider */}
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-border" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-card px-2 text-muted-foreground">Or continue with</span>
              </div>
            </div>

            {/* Google Sign In Button */}
            <Button
              type="button"
              variant="outline"
              size="lg"
              className="w-full"
              onClick={handleGoogleLogin}
              disabled={isLoading}
            >
              <svg className="mr-2 h-4 w-4" viewBox="0 0 24 24">
                <path
                  fill="currentColor"
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                />
                <path
                  fill="currentColor"
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                />
                <path
                  fill="currentColor"
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                />
                <path
                  fill="currentColor"
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                />
              </svg>
              Sign in with Google
            </Button>
          </form>

          {/* Demo Credentials */}
          <div className="mt-8">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-border" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-card px-2 text-muted-foreground">Development Mode</span>
              </div>
            </div>

            <div className="mt-6 p-4 rounded-lg bg-muted/50 border border-border">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-foreground">Demo Account</p>
                  <p className="text-xs text-muted-foreground">Quick access for testing</p>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleDemoLogin}
                  disabled={isLoading}
                >
                  Use Demo
                </Button>
              </div>
              <div className="mt-3 p-3 rounded bg-background border border-border">
                <div className="grid grid-cols-2 gap-2 text-xs font-mono">
                  <div>
                    <span className="text-muted-foreground">Email:</span>
                    <div className="text-foreground">admin@acme.com</div>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Password:</span>
                    <div className="text-foreground">password</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Card>

        {/* SignUp Link */}
        <div className="text-center">
          <p className="text-sm text-muted-foreground">
            Don't have an account?{' '}
            <Link to="/signup" className="text-primary hover:text-primary/80 font-medium">
              Create one here
            </Link>
          </p>
        </div>

        {/* Footer */}
        <div className="text-center space-y-2">
          <p className="text-xs text-muted-foreground">
            Secured by enterprise-grade authentication
          </p>
          <div className="flex items-center justify-center space-x-4 text-xs text-muted-foreground">
            <button className="hover:text-foreground transition-colors">Privacy Policy</button>
            <span>•</span>
            <button className="hover:text-foreground transition-colors">Terms of Service</button>
            <span>•</span>
            <button className="hover:text-foreground transition-colors">Contact Support</button>
          </div>
        </div>
      </div>
    </div>
  )
}
