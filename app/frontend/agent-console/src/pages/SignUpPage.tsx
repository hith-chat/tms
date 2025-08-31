import React, { useState } from 'react'
import { Eye, EyeOff, Loader2, Shield, Lock, Mail, User, CheckCircle, ArrowLeft } from 'lucide-react'
import { Link } from 'react-router-dom'
import { apiClient } from '@/lib/api'

// Simplified components matching our enterprise design (same as LoginPage)
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
      : variant === 'default'
      ? 'border-border'
      : 'border-green-500/50 text-green-700 dark:text-green-400 [&>svg]:text-green-600'
  }`}>
    {children}
  </div>
)

// OTP Input Component
const OTPInput = ({ value, onChange, disabled }: { value: string; onChange: (value: string) => void; disabled: boolean }) => {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value.replace(/\D/g, '').slice(0, 6)
    onChange(newValue)
  }

  return (
    <div className="space-y-2">
      <label className="text-sm font-medium leading-none">Verification Code</label>
      <Input
        type="text"
        value={value}
        onChange={handleChange}
        placeholder="Enter 6-digit code"
        maxLength={6}
        disabled={disabled}
        className="text-center text-lg tracking-widest font-mono"
        autoComplete="one-time-code"
      />
      <p className="text-xs text-muted-foreground">
        Enter the 6-digit code sent to your email
      </p>
    </div>
  )
}

type SignUpStep = 'signup' | 'verify' | 'success'

// Personal/Consumer email domains that should be blocked for corporate signup
const blockedEmailDomains = new Set([
  // Google
  'gmail.com', 'googlemail.com',
  
  // Microsoft
  'hotmail.com', 'outlook.com', 'live.com', 'msn.com',
  
  // Yahoo
  'yahoo.com', 'yahoo.co.uk', 'yahoo.ca', 'yahoo.co.in', 'yahoo.com.au', 
  'yahoo.fr', 'yahoo.de', 'yahoo.it', 'yahoo.es', 'ymail.com', 'rocketmail.com',
  
  // Apple
  'icloud.com', 'me.com', 'mac.com',
  
  // AOL
  'aol.com', 'aim.com',
  
  // Other common personal email providers
  'protonmail.com', 'proton.me', 'tutanota.com', 'fastmail.com',
  'mailbox.org', 'posteo.de', 'hushmail.com', 'mailfence.com',
  
  // Disposable/Temporary email providers
  'guerrillamail.com', '10minutemail.com', 'tempmail.org', 'mailinator.com',
  'yopmail.com', 'dispostable.com', 'throwaway.email', 'emailondeck.com',
  'getnada.com', 'temp-mail.org', 'fakeinbox.com', 'sharklasers.com',
  'guerrillamailblock.com', 'pokemail.net', 'spam4.me', 'maildrop.cc',
  'mohmal.com', 'nada.email', 'tempail.com', 'disposablemail.com',
  
  // More disposable patterns
  '0-mail.com', '1secmail.com', '2prong.com', '3d-painting.com',
  '4warding.com', '7tags.com', '9ox.net', 'aaathats3as.com',
  'abyssmail.com', 'afrobacon.com', 'ajaxapp.net', 'amilegit.com',
  'amiri.net', 'amiriindustries.com', 'anonmails.de', 'anonymbox.com',
])

// Additional patterns to check
const suspiciousPatterns = [
  'temp', 'throw', 'fake', 'disposable', 'trash', 'delete', 
  'remove', 'destroy', 'kill', 'burn', '10min', '20min', 
  'minute', 'hour', 'day', 'week', 'short', 'quick', 'fast',
  'instant', 'now', 'asap', 'test', 'demo', 'sample'
]

// Validate if email is from a corporate domain
const isValidCorporateEmail = (email: string): { isValid: boolean; error?: string } => {
  if (!email) return { isValid: false, error: 'Email is required' }
  
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(email)) {
    return { isValid: false, error: 'Please enter a valid email address' }
  }
  
  const domain = email.split('@')[1]?.toLowerCase()
  if (!domain) {
    return { isValid: false, error: 'Invalid email format' }
  }
  
  if (blockedEmailDomains.has(domain)) {
    return { 
      isValid: false, 
      error: 'Personal email addresses (e.g., Gmail, Yahoo, Hotmail) are not allowed. Please use your company email address.' 
    }
  }
  
  // Check for suspicious patterns in domain
  const hasSuspiciousPattern = suspiciousPatterns.some(pattern => 
    domain.includes(pattern.toLowerCase())
  )
  
  if (hasSuspiciousPattern) {
    return { 
      isValid: false, 
      error: 'Temporary or disposable email addresses are not allowed. Please use your company email address.' 
    }
  }
  
  // Additional validation: domain should have at least one dot and be longer than 4 chars
  if (domain.length < 4 || !domain.includes('.')) {
    return {
      isValid: false,
      error: 'Please enter a valid company email address.'
    }
  }
  
  return { isValid: true }
}

export function SignUpPage() {
  const [step, setStep] = useState<SignUpStep>('signup')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [otp, setOtp] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [resendCooldown, setResendCooldown] = useState(0)
  const [emailError, setEmailError] = useState('')

  // Validate email when it changes
  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newEmail = e.target.value
    setEmail(newEmail)
    
    if (newEmail) {
      const validation = isValidCorporateEmail(newEmail)
      setEmailError(validation.isValid ? '' : validation.error || '')
    } else {
      setEmailError('')
    }
  }

  const handleSignUp = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError('')

    // Validate email before submitting
    const emailValidation = isValidCorporateEmail(email)
    if (!emailValidation.isValid) {
      setError(emailValidation.error || 'Invalid email address')
      setIsLoading(false)
      return
    }

    try {
      const response = await apiClient.signup({
        email,
        password,
        name
      })

      setSuccess(response.message)
      setStep('verify')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create account')
    } finally {
      setIsLoading(false)
    }
  }

  const handleVerifyOTP = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError('')

    try {
      await apiClient.verifySignupOTP({
        email,
        otp
      })

      // Account created successfully and tokens are stored
      // Redirect to dashboard
      // navigate('/dashboard')
      window.location.assign("/dashboard")
    } catch (err: any) {
      setError(err.response?.data?.error || 'Invalid verification code')
    } finally {
      setIsLoading(false)
    }
  }

  const handleResendOTP = async () => {
    if (resendCooldown > 0) return

    setIsLoading(true)
    setError('')

    try {
      const response = await apiClient.resendSignupOTP({
        email
      })

      setSuccess(response.message)
      
      // Start cooldown
      setResendCooldown(60)
      const interval = setInterval(() => {
        setResendCooldown((prev) => {
          if (prev <= 1) {
            clearInterval(interval)
            return 0
          }
          return prev - 1
        })
      }, 1000)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to resend code')
    } finally {
      setIsLoading(false)
    }
  }

  const renderSignUpForm = () => (
    <>
      {/* Header */}
      <div className="text-center space-y-2">
        <div className="mx-auto h-16 w-16 rounded-2xl bg-primary flex items-center justify-center mb-6 shadow-lg">
          <Shield className="h-8 w-8 text-primary-foreground" />
        </div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">
          Create Account
        </h1>
        <p className="text-muted-foreground">
          Join TMS to start managing tickets efficiently
        </p>
      </div>

      {/* SignUp Card */}
      <Card className="p-8">
        <form onSubmit={handleSignUp} className="space-y-6">
          {/* Error Alert */}
          {error && (
            <Alert variant="destructive">
              <div className="flex items-center space-x-2">
                <Lock className="h-4 w-4" />
                <span className="text-sm font-medium">{error}</span>
              </div>
            </Alert>
          )}

          {/* Name Field */}
          <div className="space-y-2">
            <label htmlFor="name" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
              Full Name
            </label>
            <div className="relative">
              <User className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
              <Input
                id="name"
                type="text"
                value={name}
                onChange={(e: any) => setName(e.target.value)}
                placeholder="Enter your full name"
                className="pl-10"
                disabled={isLoading}
                required
                autoComplete="name"
              />
            </div>
          </div>

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
                onChange={handleEmailChange}
                placeholder="Enter your company email address"
                className={`pl-10 ${emailError ? 'border-destructive focus-visible:ring-destructive' : ''}`}
                disabled={isLoading}
                required
                autoComplete="email"
              />
            </div>
            {emailError && (
              <p className="text-sm text-destructive">{emailError}</p>
            )}
            <p className="text-xs text-muted-foreground">
              Please use your company email address. Personal emails (Gmail, Yahoo, etc.) are not allowed.
            </p>
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
                placeholder="Create a strong password"
                className="pl-10 pr-10"
                disabled={isLoading}
                required
                autoComplete="new-password"
                minLength={8}
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
            <p className="text-xs text-muted-foreground">
              Password must be at least 8 characters long
            </p>
          </div>

          {/* Create Account Button */}
          <Button
            type="submit"
            size="lg"
            className="w-full"
            disabled={isLoading || !email || !password || !name || password.length < 8 || emailError !== ''}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Creating account...
              </>
            ) : (
              <>
                <Shield className="mr-2 h-4 w-4" />
                Create Account
              </>
            )}
          </Button>
        </form>

        {/* Login Link */}
        <div className="mt-6 text-center">
          <p className="text-sm text-muted-foreground">
            Already have an account?{' '}
            <Link to="/login" className="text-primary hover:text-primary/80 font-medium">
              Sign in here
            </Link>
          </p>
        </div>
      </Card>
    </>
  )

  const renderVerifyForm = () => (
    <>
      {/* Header */}
      <div className="text-center space-y-2">
        <div className="mx-auto h-16 w-16 rounded-2xl bg-primary flex items-center justify-center mb-6 shadow-lg">
          <Mail className="h-8 w-8 text-primary-foreground" />
        </div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">
          Verify Email
        </h1>
        <p className="text-muted-foreground">
          We sent a verification code to <strong>{email}</strong>
        </p>
      </div>

      {/* Verify Card */}
      <Card className="p-8">
        <form onSubmit={handleVerifyOTP} className="space-y-6">
          {/* Success Alert */}
          {success && (
            <Alert variant="success">
              <div className="flex items-center space-x-2">
                <CheckCircle className="h-4 w-4" />
                <span className="text-sm font-medium">{success}</span>
              </div>
            </Alert>
          )}

          {/* Error Alert */}
          {error && (
            <Alert variant="destructive">
              <div className="flex items-center space-x-2">
                <Lock className="h-4 w-4" />
                <span className="text-sm font-medium">{error}</span>
              </div>
            </Alert>
          )}

          {/* OTP Input */}
          <OTPInput
            value={otp}
            onChange={setOtp}
            disabled={isLoading}
          />

          {/* Verify Button */}
          <Button
            type="submit"
            size="lg"
            className="w-full"
            disabled={isLoading || !otp || otp.length !== 6}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Verifying...
              </>
            ) : (
              <>
                <CheckCircle className="mr-2 h-4 w-4" />
                Verify Code
              </>
            )}
          </Button>

          {/* Resend Section */}
          <div className="text-center space-y-2">
            <p className="text-sm text-muted-foreground">
              Didn't receive the code?
            </p>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={handleResendOTP}
              disabled={resendCooldown > 0 || isLoading}
            >
              {resendCooldown > 0 
                ? `Resend code (${resendCooldown}s)` 
                : 'Resend verification code'
              }
            </Button>
          </div>
        </form>

        {/* Back Link */}
        <div className="mt-6 text-center">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setStep('signup')}
            disabled={isLoading}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to signup
          </Button>
        </div>
      </Card>
    </>
  )

  const renderSuccessForm = () => (
    <>
      {/* Header */}
      <div className="text-center space-y-2">
        <div className="mx-auto h-16 w-16 rounded-2xl bg-green-600 flex items-center justify-center mb-6 shadow-lg">
          <CheckCircle className="h-8 w-8 text-white" />
        </div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">
          Account Created!
        </h1>
        <p className="text-muted-foreground">
          Your TMS account has been successfully created and verified
        </p>
      </div>

      {/* Success Card */}
      <Card className="p-8">
        <div className="space-y-6 text-center">
          {/* Success Alert */}
          <Alert variant="success">
            <div className="flex items-center space-x-2">
              <CheckCircle className="h-4 w-4" />
              <span className="text-sm font-medium">{success}</span>
            </div>
          </Alert>

          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Welcome to TMS! You can now sign in with your credentials to start managing tickets.
            </p>

            <Button
              asChild
              size="lg"
              className="w-full"
            >
              <Link to="/login">
                <Shield className="mr-2 h-4 w-4" />
                Continue to Sign In
              </Link>
            </Button>
          </div>
        </div>
      </Card>
    </>
  )

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <div className="w-full max-w-md space-y-8">
        {step === 'signup' && renderSignUpForm()}
        {step === 'verify' && renderVerifyForm()}
        {step === 'success' && renderSuccessForm()}

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
