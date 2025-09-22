# Hith Chat Widget

The embeddable chat widget for the Hith (Ticket Management System) that allows websites to integrate real-time chat functionality.

## Development Setup

### Building the Widget

```bash
# Install dependencies
npm install

# Build the widget (creates dist/chat-widget.js)
npm run build
```

### Testing Locally

1. **Build the widget**:
   ```bash
   npm run build
   ```

2. **Start your Hith backend** (in another terminal):
   ```bash
   cd ../../backend
   go run cmd/api/main.go
   ```

3. **Serve the test page**:
   ```bash
   npm run serve
   # Or manually: python3 -m http.server 3001
   ```

4. **Open in browser**: `http://localhost:3001/test.html`

## Integration

### Basic Integration

Add this script to any website where you want the chat widget to appear:

```html
<!-- Hith Chat Widget -->
<script>
  (function() {
    window.TMSChatConfig = {
      widgetId: 'your-widget-id-here',
      domain: 'your-domain.com',
      apiUrl: 'https://your-tms-api.com/v1' // Optional, defaults to localhost:8080/v1
    };
    var script = document.createElement('script');
    script.src = 'https://your-cdn.com/chat-widget.js';
    script.async = true;
    document.head.appendChild(script);
  })();
</script>
```

### Configuration Options

The `TMSChatConfig` object supports the following options:

- `widgetId` (required): The unique identifier for your chat widget
- `domain` (required): Your verified domain name
- `apiUrl` (optional): The Hith API base URL (defaults to http://localhost:8080/v1)

## Widget Features

- **Customizable appearance**: Colors, position, and messages
- **Auto-open functionality**: Automatically open after a delay
- **Real-time messaging**: WebSocket-based real-time communication
- **File uploads**: Support for file sharing (if enabled)
- **Agent avatars**: Display agent profile pictures
- **Typing indicators**: Show when agents are typing
- **Offline messaging**: Handle offline scenarios
- **Mobile responsive**: Works on all device sizes

## API Endpoints Used

The widget communicates with these Hith API endpoints:

- `GET /v1/public/chat/widgets/domain/{domain}` - Get widget configuration
- `POST /v1/public/chat/widgets/{widgetId}/sessions` - Start a chat session
- `GET /v1/public/chat/sessions/{sessionToken}` - Get session details
- `GET /v1/public/chat/sessions/{sessionToken}/messages` - Get chat messages
- `POST /v1/public/chat/sessions/{sessionToken}/messages` - Send a message
- WebSocket: `/v1/public/chat/sessions/{sessionToken}/ws` - Real-time updates

## Development Notes

- Built with TypeScript and Vite
- Uses UMD format for maximum compatibility
- No external dependencies (except for event emitter)
- Automatically initializes when `TMSChatConfig` is present
- Can also be manually initialized: `new window.TMSChatWidget(config)`

## Files Structure

```
src/
├── index.ts       # Entry point and auto-initialization
├── widget.ts      # Main widget class and UI logic
├── api.ts         # API communication layer
├── events.ts      # Event emitter for internal communication
└── types.ts       # TypeScript type definitions
```
