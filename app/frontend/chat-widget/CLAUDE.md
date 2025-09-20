# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the TMS Chat Widget - an embeddable JavaScript widget that provides real-time chat functionality for websites. The widget is built as a UMD library that can be embedded on any website to connect visitors with TMS (Ticket Management System) agents.

## Development Commands

### Building
```bash
npm run build          # Build for development
npm run build:prod     # Build for production with minification
```

### Development Server
```bash
npm run dev            # Start Vite development server
```

### Testing
```bash
./serve-test.sh        # Start local HTTP server on port 3001 for testing
# Alternatively: python3 -m http.server 3001
```

### Preview
```bash
npm run preview        # Preview the built widget
```

## Architecture

### Core Structure
- **Entry Point**: `src/index.ts` - Auto-initializes widget if `TMSChatConfig` is present on window
- **Main Widget Class**: `src/widget.ts` - Contains the TMSChatWidget class with all UI logic and state management
- **API Layer**: `src/api.ts` - Handles all HTTP requests and WebSocket connections to TMS backend
- **Type Definitions**: `src/types.ts` - TypeScript interfaces for all data structures
- **Event System**: `src/events.ts` - Internal event emitter for component communication
- **Storage**: `src/storage.ts` - Session persistence and visitor fingerprinting
- **Themes**: `src/themes.ts` - CSS injection and styling utilities

### Widget Initialization
The widget supports two initialization modes:
1. **Auto-initialization**: When `window.TMSChatConfig` is present, the widget auto-initializes
2. **Manual initialization**: `new window.TMSChatWidget(config)` for programmatic control

### API Integration
The widget communicates with these TMS backend endpoints:
- `GET /v1/public/chat/widgets/{widgetId}` - Get widget configuration
- `POST /v1/public/chat/widgets/{widgetId}/sessions` - Start chat session
- `GET /v1/public/chat/sessions/{token}` - Get session details
- `GET /v1/public/chat/sessions/{token}/messages` - Get messages
- `POST /v1/public/chat/sessions/{token}/messages` - Send messages
- `WebSocket /v1/public/chat/sessions/{token}/ws` - Real-time updates

### Session Management
- Sessions persist for 7 days using localStorage
- JWT tokens used for session authentication
- Automatic reconnection logic for WebSocket connections
- Visitor fingerprinting for session continuity

### Build Output
- Builds to `dist/chat-widget.js` as UMD format
- Includes CSS injection for complete styling
- Self-contained with no external runtime dependencies (except jose and mitt)

## Testing Setup

1. Build the widget: `npm run build`
2. Start TMS backend on `localhost:8080`
3. Run test server: `./serve-test.sh`
4. Open `http://localhost:3001/test.html`

The test page includes a sample configuration that connects to a local TMS backend instance.

## Key Configuration Options

```javascript
window.TMSChatConfig = {
  widgetId: 'your-widget-id',     // Required: Widget identifier
  apiUrl: 'https://api.url',      // Optional: TMS API base URL
  enableSessionPersistence: true, // Optional: Enable session storage
  debugMode: false               // Optional: Enable debug logging
}
```