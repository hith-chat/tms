# New Landing - Hith Chat Widget Builder

Modern landing page for building and previewing AI-powered chat widgets.

## Features

- 🎨 **Modern Design** - Beautiful, responsive UI inspired by noupe.com and emergent.sh
- ⚡ **Real-time Streaming** - Live progress updates during widget building
- 🔍 **Live Preview** - See your widget on your actual website before deploying
- 📦 **One-Click Deploy** - Get embed code and deploy in seconds
- 🎯 **AI-Powered** - Automatically learns from your website content

## Getting Started

### Development

```bash
# Install dependencies
pnpm install

# Start dev server
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview
```

### Environment Variables

Copy `.env.example` to `.env` and configure:

```env
VITE_API_URL=http://localhost:8080
```

## Architecture

- **Framework**: React 18 + TypeScript
- **Bundler**: Vite
- **Routing**: React Router v6
- **Animations**: Framer Motion
- **Icons**: Lucide React
- **Styling**: Pure CSS with CSS Variables

## Project Structure

```
src/
├── components/       # Reusable components (future)
├── pages/           # Page components
│   ├── HomePage.tsx         # Landing page
│   ├── BuildPage.tsx        # Streaming build progress
│   └── PreviewPage.tsx      # Widget preview with iframe
├── utils/           # Utilities
│   ├── api.ts               # API client
│   └── constants.ts         # App constants
├── App.tsx          # Main app component
└── main.tsx         # Entry point
```

## API Integration

### Build Widget

```typescript
POST /api/public/ai-widget-builder?url={websiteUrl}&depth=3
```

Returns Server-Sent Events (SSE) stream with progress updates.

### Event Types

- `builder_started` - Build process initiated
- `project_created` - Project created
- `widget_stage_started` - Widget configuration started
- `widget_theme_ready` - Widget styling complete
- `knowledge_stage_started` - Learning from website
- `scraping_progress` - Page scraping progress
- `faq_generation_started` - Generating FAQs
- `completed` - Build complete
- `error` - Build failed

## Deployment

### Production Build

```bash
pnpm build
```

Output will be in `dist/` directory.

### Environment Variables for Production

```env
VITE_API_URL=https://api.hith.chat
```

## Features in Detail

### 1. HomePage

- Hero section with URL input
- Features showcase
- How it works guide
- Responsive navigation
- Animated interactions

### 2. BuildPage

- Real-time streaming progress
- Visual step indicators
- Event log display
- Error handling
- Smooth transitions

### 3. PreviewPage

- Iframe integration
- Widget injection
- Embed code generation
- Copy to clipboard
- Fallback for iframe restrictions
- Sign-up flow

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## License

Proprietary - Hith Chat
