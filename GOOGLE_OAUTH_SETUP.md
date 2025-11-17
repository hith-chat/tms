# Google OAuth Setup Guide

## Quick Start

### 1. Google Cloud Console Setup

1. **Create a Google Cloud Project**
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select an existing one

2. **Enable Google+ API**
   - Navigate to "APIs & Services" > "Library"
   - Search for "Google+ API"
   - Click "Enable"

3. **Create OAuth 2.0 Credentials**
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth client ID"
   - Select "Web application"
   - Add authorized redirect URIs:
     ```
     http://localhost:3000/auth/google/callback
     https://yourdomain.com/auth/google/callback
     ```
   - Click "Create"
   - Copy the Client ID and Client Secret

### 2. Backend Configuration

Add these environment variables to your backend configuration:

```bash
# .env or config file
GOOGLE_OAUTH_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_OAUTH_CLIENT_SECRET=your-client-secret
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:3000/auth/google/callback
```

For production:
```bash
GOOGLE_OAUTH_REDIRECT_URL=https://yourdomain.com/auth/google/callback
```

### 3. Frontend Configuration

The frontend automatically uses the backend API endpoints. No additional configuration needed.

### 4. Start the Application

**Backend:**
```bash
cd app/backend
go run cmd/api/main.go
```

**Frontend:**
```bash
cd app/frontend/agent-console
npm install
npm run dev
```

### 5. Test the Integration

1. Navigate to `http://localhost:3000/login`
2. Click "Sign in with Google"
3. Authorize the application
4. You should be redirected back and logged in

## Troubleshooting

### "Google OAuth is not configured" Error
- Ensure `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` are set
- Restart the backend server after setting environment variables

### "Invalid state token" Error
- This is a security feature to prevent CSRF attacks
- State tokens expire after 5 minutes
- Try the login flow again

### "Redirect URI mismatch" Error
- Ensure the redirect URI in Google Cloud Console matches exactly
- Check that `GOOGLE_OAUTH_REDIRECT_URL` matches the frontend callback URL
- Include both `http://` and `https://` versions if needed

### "Email not verified with Google" Error
- The user's Google email must be verified
- Ask the user to verify their email with Google

### Backend Not Starting
- Check that all required dependencies are installed: `go mod tidy`
- Ensure the OAuth configuration is properly loaded

## Security Considerations

1. **Never commit credentials**: Keep Client ID and Secret in environment variables
2. **Use HTTPS in production**: Always use HTTPS for the redirect URL in production
3. **Validate redirect URIs**: Only add trusted redirect URIs in Google Cloud Console
4. **Enable corporate email validation**: Set `REQUIRE_CORPORATE_EMAIL=true` if needed
5. **Monitor OAuth usage**: Check Google Cloud Console for usage and errors

## Development vs Production

### Development
```bash
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:3000/auth/google/callback
```

### Production
```bash
GOOGLE_OAUTH_REDIRECT_URL=https://yourdomain.com/auth/google/callback
```

Make sure to add both URLs to Google Cloud Console authorized redirect URIs.

## Additional Resources

- [Google OAuth 2.0 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [Google Cloud Console](https://console.cloud.google.com/)
- [OAuth 2.0 Best Practices](https://tools.ietf.org/html/rfc6749)
