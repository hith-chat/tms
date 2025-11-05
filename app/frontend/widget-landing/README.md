# Chat Widget Landing Page

A beautiful, modern landing page for the AI Chat Widget platform, inspired by Crisp.chat design principles.

## Features

- **Clean, Modern Design**: Simple and professional interface with smooth animations
- **AI Widget Builder**: Enter any website URL to automatically build a custom chat widget
- **Real-time Progress**: Server-Sent Events (SSE) for live progress updates
- **Live Preview**: See your widget in action on your actual website (iframe with fallback)
- **Mobile Responsive**: Works perfectly on all devices
- **Rate Limited**: Protected with 2 requests per 6 hours per IP

## File Structure

```
widget-landing/
├── index.html          # Main landing page
├── styles.css          # Landing page styles
├── script.js           # Landing page functionality & API integration
├── preview.html        # Widget preview/demo page
├── preview-styles.css  # Preview page styles
├── preview-script.js   # Preview page functionality & widget injection
└── README.md          # This file
```

## How It Works

### 1. Landing Page Flow

1. User enters their website URL
2. API call to `/api/public/ai-widget-builder` with SSE streaming
3. Real-time progress updates as the widget is built:
   - Website analysis
   - Theme extraction
   - Knowledge base creation
   - FAQ generation
4. Redirect to preview page on completion

### 2. Preview Page Flow

1. Load widget data from sessionStorage
2. Attempt to load user's website in iframe
3. Check for X-Frame-Options restrictions
4. Inject interactive chat widget demo
5. Provide embed code for installation

## API Integration

The landing page connects to the backend endpoint:

```
POST /api/public/ai-widget-builder?url={websiteUrl}&email={userEmail}
```

**Rate Limiting**: 2 requests per 6 hours per IP address

**Response**: Server-Sent Events (SSE) stream with events:
- `builder_started`
- `widget_stage_started`
- `widget_theme_ready`
- `knowledge_stage_started`
- `scraping_progress`
- `faq_generation_started`
- `completed`
- `error`

## Configuration

Update the `API_BASE_URL` and `USER_EMAIL` constants in `script.js`:

```javascript
const API_BASE_URL = window.location.origin;
const USER_EMAIL = 'sumansaurabh@hith.chat';
```

## Deployment

### Development

Simply open `index.html` in a browser or serve with any static file server:

```bash
# Python
python3 -m http.server 8000

# Node.js (http-server)
npx http-server

# PHP
php -S localhost:8000
```

### Production

1. Copy all files to your web server
2. Ensure CORS is properly configured on the backend
3. Update `API_BASE_URL` if needed
4. Serve via nginx, Apache, or any static hosting

### Nginx Configuration Example

```nginx
location /widget-landing {
    root /var/www/html;
    try_files $uri $uri/ /widget-landing/index.html;
}
```

## Browser Support

- Chrome/Edge: ✅
- Firefox: ✅
- Safari: ✅
- Mobile browsers: ✅

## Features Breakdown

### Landing Page Features
- Hero section with clear value proposition
- Interactive widget builder form
- Real-time progress tracking with SSE
- Loading states with spinner and progress bar
- Error handling with retry functionality
- Trust indicators (free plan, no credit card, etc.)
- Feature showcase grid
- Responsive design

### Preview Page Features
- Desktop/Mobile view toggle
- Live website iframe (with fallback)
- Interactive chat widget demo
- Embed code modal with copy functionality
- Account information display
- Next steps checklist
- Smooth animations and transitions

## Customization

### Colors
Edit CSS variables in `styles.css` and `preview-styles.css`:

```css
:root {
    --primary-color: #4F46E5;
    --secondary-color: #10B981;
    --error-color: #DC2626;
    /* ... more variables */
}
```

### Content
Update text content in HTML files:
- Hero title and description
- Feature descriptions
- Trust indicators
- Footer information

### Widget Appearance
Modify widget styles in `preview-script.js` within the `injectChatWidget()` function.

## Technical Details

### EventSource (SSE) Implementation
The landing page uses EventSource API for real-time updates:

```javascript
const eventSource = new EventSource(sseUrl);
eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    // Handle event
};
```

### Session Storage
Widget data is stored in sessionStorage for the preview page:

```javascript
sessionStorage.setItem('widgetPreview', JSON.stringify({
    websiteUrl: url,
    widgetId: data.widget_id,
    embedCode: data.embed_code,
    email: user@example.com
}));
```

### iframe Security
The preview page handles X-Frame-Options restrictions gracefully:
- Attempts to load user's website in iframe
- Detects iframe blocking
- Shows fallback demo if blocked
- Explains that the widget will work on the actual site

## Future Enhancements

- [ ] Email verification
- [ ] Multi-step onboarding wizard
- [ ] Customization options (colors, position, etc.)
- [ ] Analytics dashboard preview
- [ ] Video tutorial integration
- [ ] Live chat support integration
- [ ] Testimonials section
- [ ] Pricing comparison
- [ ] Integration guides

## Support

For issues or questions, contact: sumansaurabh@hith.chat

## License

Proprietary - All rights reserved
