export const USER_EMAIL = 'sumansaurabh@hith.chat'

export const FEATURES = [
  {
    title: 'AI-Powered Responses',
    description: 'Automatically answer customer questions using your website content',
    icon: '🤖',
  },
  {
    title: 'Lead Generation',
    description: 'Capture leads by engaging visitors in meaningful conversations',
    icon: '📈',
  },
  {
    title: 'Schedule Meetings',
    description: 'Let customers book meetings directly through the chat widget',
    icon: '📅',
  },
  {
    title: 'Ticket Management',
    description: 'Automatically create support tickets for complex issues',
    icon: '🎫',
  },
  {
    title: '24/7 Availability',
    description: 'Your AI assistant never sleeps - always available for customers',
    icon: '⏰',
  },
  {
    title: 'Easy Integration',
    description: 'Add to your website with just one line of code',
    icon: '⚡',
  },
]

export const STEPS = [
  {
    number: 1,
    title: 'Enter Your Website',
    description: 'Simply provide your website URL to get started',
  },
  {
    number: 2,
    title: 'AI Analyzes Content',
    description: 'Our AI scans and learns from your website content',
  },
  {
    number: 3,
    title: 'Preview & Test',
    description: 'See how the widget looks on your actual website',
  },
  {
    number: 4,
    title: 'Deploy in Seconds',
    description: 'Copy one line of code and paste it in your website',
  },
]

export const BUILD_STAGE_LABELS: Record<string, string> = {
  initialization: 'Initializing...',
  widget: 'Creating Widget...',
  knowledge: 'Building Knowledge Base...',
  completed: 'Complete!',
  error: 'Error',
}

export const BUILD_EVENT_MESSAGES: Record<string, string> = {
  builder_started: 'Starting AI widget builder',
  project_creation_started: 'Creating project',
  project_created: 'Project created successfully',
  widget_stage_started: 'Setting up widget',
  widget_theme_ready: 'Widget theme configured',
  knowledge_stage_started: 'Learning from your website',
  scraping_progress: 'Analyzing pages',
  faq_generation_started: 'Generating FAQs',
  completed: 'Widget is ready!',
  error: 'Something went wrong',
}
