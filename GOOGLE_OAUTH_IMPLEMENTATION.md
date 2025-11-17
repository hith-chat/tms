# Google OAuth Login Implementation

## Overview
This document describes the implementation of Google OAuth login for the TMS (Ticket Management System) application. The implementation allows users to sign in using their Google accounts, with automatic account creation for new users.

## Implementation Summary

### Backend Changes

#### 1. Configuration (`internal/config/config.go`)
Added Google OAuth configuration structure:
```go
type OAuthConfig struct {
    Google GoogleOAuthConfig `mapstructure:"google"`
}

type GoogleOAuthConfig struct {
    ClientID     string `mapstructure:"client_id"`
    ClientSecret string `mapstructure:"client_secret"`
    RedirectURL  string `mapstructure:"redirect_url"`
}
```

Environment variables:
- `GOOGLE_OAUTH_CLIENT_ID` - Google OAuth Client ID
- `GOOGLE_OAUTH_CLIENT_SECRET` - Google OAuth Client Secret
- `GOOGLE_OAUTH_REDIRECT_URL` - OAuth redirect URL (default: `http://localhost:3000/auth/google/callback`)

#### 2. Google OAuth Service (`internal/service/google_oauth.go`)
Created a new service to handle Google OAuth operations:
- **GenerateStateToken()** - Generates CSRF protection token
- **ValidateStateToken()** - Validates state token
- **GetAuthURL()** - Returns Google OAuth authorization URL
- **ExchangeCode()** - Exchanges authorization code for tokens
- **GetUserInfo()** - Fetches user information from Google
- **HandleGoogleLogin()** - Complete OAuth login flow

Features:
- CSRF protection using state tokens
- Email verification check
- Automatic account creation for new users
- Corporate email validation (if enabled)
- JWT token generation

#### 3. Auth Service Updates (`internal/service/auth.go`)
Added new method:
- **SignUpWithOAuth()** - Creates user account via OAuth without OTP verification
  - Creates tenant and default project
  - Assigns tenant admin role
  - Generates JWT tokens
  - Sends welcome email

#### 4. Auth Handler Updates (`internal/handlers/auth.go`)
Added two new endpoints:
- **GoogleOAuthLogin()** - `GET /v1/auth/google/login`
  - Initiates OAuth flow
  - Returns authorization URL and state token
  
- **GoogleOAuthCallback()** - `GET /v1/auth/google/callback`
  - Handles OAuth callback from Google
  - Processes authorization code
  - Returns JWT tokens and user info

#### 5. Route Registration (`cmd/api/main.go`)
- Initialized GoogleOAuthService with configuration
- Registered OAuth routes in auth routes group
- Added conditional initialization (only if credentials are configured)

### Frontend Changes

#### 1. API Client Updates (`src/lib/api.ts`)
Added two new methods:
```typescript
async getGoogleOAuthURL(): Promise<{ auth_url: string; state: string }>
async handleGoogleCallback(code: string, state: string): Promise<LoginResponse>
```

#### 2. Login Page Updates (`src/pages/LoginPage.tsx`)
- Added "Sign in with Google" button with Google logo
- Implemented `handleGoogleLogin()` function
- Added visual divider between email/password and OAuth login
- Proper error handling for OAuth initiation

#### 3. OAuth Callback Page (`src/pages/OAuthCallbackPage.tsx`)
New page to handle OAuth redirect:
- Extracts code and state from URL parameters
- Calls backend to exchange code for tokens
- Stores authentication tokens
- Redirects to dashboard on success
- Shows error messages and redirects to login on failure
- Loading state with spinner

#### 4. Routing Updates (`src/App.tsx`)
- Added route for `/auth/google/callback`
- Imported OAuthCallbackPage component

## OAuth Flow

### 1. User Initiates Login
1. User clicks "Sign in with Google" button on login page
2. Frontend calls `GET /v1/auth/google/login`
3. Backend generates state token and returns Google OAuth URL
4. Frontend redirects user to Google OAuth consent screen

### 2. Google Authorization
1. User authenticates with Google
2. User grants permissions
3. Google redirects to callback URL with authorization code and state

### 3. Callback Processing
1. Frontend receives callback at `/auth/google/callback`
2. Extracts `code` and `state` from URL parameters
3. Calls `GET /v1/auth/google/callback?code=...&state=...`
4. Backend validates state token
5. Backend exchanges code for Google tokens
6. Backend fetches user info from Google

### 4. Account Creation/Login
**For New Users:**
1. Backend validates email (corporate email check if enabled)
2. Creates new tenant and default project
3. Creates agent account (no password)
4. Assigns tenant admin role
5. Generates JWT tokens
6. Sends welcome email

**For Existing Users:**
1. Backend retrieves existing agent
2. Gets role bindings
3. Generates JWT tokens

### 5. Frontend Completion
1. Frontend stores tokens in localStorage
2. Stores user data
3. Redirects to tickets page

## Security Features

1. **CSRF Protection**: State token validation prevents CSRF attacks
2. **Email Verification**: Only verified Google emails are accepted
3. **Corporate Email Validation**: Optional check for corporate domains
4. **Token Expiration**: State tokens expire after 5 minutes
5. **JWT Authentication**: Standard JWT-based authentication
6. **Rate Limiting**: Auth routes have rate limiting applied

## Configuration Required

### Backend Environment Variables
```bash
# Google OAuth Configuration
GOOGLE_OAUTH_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_OAUTH_CLIENT_SECRET=your-client-secret
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:3000/auth/google/callback

# For production
GOOGLE_OAUTH_REDIRECT_URL=https://yourdomain.com/auth/google/callback
```

### Google Cloud Console Setup
1. Create a new project in Google Cloud Console
2. Enable Google+ API
3. Create OAuth 2.0 credentials
4. Add authorized redirect URIs:
   - `http://localhost:3000/auth/google/callback` (development)
   - `https://yourdomain.com/auth/google/callback` (production)
5. Copy Client ID and Client Secret to environment variables

## Testing Checklist

### Backend Testing
- [ ] OAuth initiation endpoint returns valid Google URL
- [ ] State token is generated and stored
- [ ] Callback endpoint validates state token
- [ ] Authorization code is exchanged for tokens
- [ ] User info is fetched from Google
- [ ] New user account is created successfully
- [ ] Existing user can login
- [ ] JWT tokens are generated correctly
- [ ] Corporate email validation works (if enabled)
- [ ] Error handling for invalid codes/tokens

### Frontend Testing
- [ ] Google Sign In button is visible and styled correctly
- [ ] Clicking button redirects to Google OAuth
- [ ] Callback page handles success correctly
- [ ] Callback page handles errors correctly
- [ ] Tokens are stored in localStorage
- [ ] User is redirected to dashboard after login
- [ ] Error messages are displayed properly
- [ ] Loading states work correctly

### Integration Testing
- [ ] Complete OAuth flow works end-to-end
- [ ] New user signup via Google works
- [ ] Existing user login via Google works
- [ ] User can access protected routes after OAuth login
- [ ] Refresh token works for OAuth users
- [ ] Logout clears OAuth session

## Dependencies

### Backend
- `golang.org/x/oauth2` - OAuth 2.0 client library
- `golang.org/x/oauth2/google` - Google OAuth provider

### Frontend
- No additional dependencies required (uses existing axios and react-router-dom)

## Error Handling

### Backend Errors
- Invalid state token → 400 Bad Request
- Missing code/state → 400 Bad Request
- OAuth exchange failure → 401 Unauthorized
- User info fetch failure → 401 Unauthorized
- Email not verified → 401 Unauthorized
- Account creation failure → 400 Bad Request

### Frontend Errors
- OAuth initiation failure → Error message displayed
- Callback parameter missing → Redirect to login
- Token exchange failure → Error message + redirect to login
- User denied access → Error message + redirect to login

## Future Enhancements

1. **Multiple OAuth Providers**: Add support for GitHub, Microsoft, etc.
2. **Account Linking**: Allow users to link multiple OAuth providers
3. **OAuth Token Storage**: Store OAuth tokens for API access
4. **Profile Picture**: Use Google profile picture for user avatar
5. **Redis State Storage**: Move state token storage to Redis for scalability
6. **OAuth Scopes**: Add additional scopes for calendar, drive integration

## Notes

- OAuth users don't have passwords (PasswordHash is nil)
- OAuth users can still use email/password login if they set a password later
- State tokens are currently stored in memory (use Redis in production)
- Corporate email validation can be enabled/disabled via feature flags
- All OAuth users are created with tenant admin role by default
