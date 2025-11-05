// Configuration
const API_BASE_URL = window.location.origin;
const USER_EMAIL = 'sumansaurabh@hith.chat';

// State
let eventSource = null;
let widgetData = null;

// DOM Elements
const form = document.getElementById('widgetBuilderForm');
const buildBtn = document.getElementById('buildBtn');
const websiteUrlInput = document.getElementById('websiteUrl');
const loadingState = document.getElementById('loadingState');
const errorState = document.getElementById('errorState');
const loadingTitle = document.getElementById('loadingTitle');
const loadingMessage = document.getElementById('loadingMessage');
const progressFill = document.getElementById('progressFill');
const errorMessage = document.getElementById('errorMessage');

// Event Listeners
form.addEventListener('submit', handleFormSubmit);

// Form submission handler
async function handleFormSubmit(e) {
    e.preventDefault();

    const websiteUrl = websiteUrlInput.value.trim();
    if (!websiteUrl) {
        showError('Please enter a valid website URL');
        return;
    }

    // Validate URL format
    try {
        new URL(websiteUrl);
    } catch (err) {
        showError('Please enter a valid URL (e.g., https://example.com)');
        return;
    }

    // Start building
    startBuilding(websiteUrl);
}

// Start the widget building process
function startBuilding(websiteUrl) {
    // Hide form and error, show loading
    form.style.display = 'none';
    errorState.style.display = 'none';
    loadingState.style.display = 'flex';

    // Reset progress
    progressFill.style.width = '0%';
    loadingTitle.textContent = 'Building your AI widget...';
    loadingMessage.textContent = 'Analyzing your website';

    // Create EventSource for SSE
    const encodedUrl = encodeURIComponent(websiteUrl);
    const sseUrl = `${API_BASE_URL}/api/public/ai-widget-builder?url=${encodedUrl}&email=${encodeURIComponent(USER_EMAIL)}`;

    eventSource = new EventSource(sseUrl);

    // Track progress
    let currentProgress = 0;

    eventSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleBuilderEvent(data);

            // Update progress based on event type
            if (data.type === 'builder_started') {
                currentProgress = 10;
            } else if (data.type === 'widget_stage_started') {
                currentProgress = 20;
            } else if (data.type === 'widget_theme_ready') {
                currentProgress = 40;
            } else if (data.type === 'knowledge_stage_started') {
                currentProgress = 50;
            } else if (data.type === 'scraping_progress') {
                currentProgress = 50 + (data.data?.progress || 0) * 0.3; // 50-80%
            } else if (data.type === 'faq_generation_started') {
                currentProgress = 80;
            } else if (data.type === 'completed') {
                currentProgress = 100;
                widgetData = data.data;
            }

            progressFill.style.width = `${currentProgress}%`;

        } catch (err) {
            console.error('Error parsing SSE data:', err);
        }
    };

    eventSource.onerror = function(err) {
        console.error('EventSource error:', err);
        eventSource.close();

        // Check if we got widget data before error (completed successfully)
        if (widgetData && widgetData.widget_id) {
            // Success! Redirect to preview
            redirectToPreview(websiteUrl, widgetData);
        } else {
            showError('Connection to server lost. Please try again.');
        }
    };
}

// Handle different builder events
function handleBuilderEvent(event) {
    console.log('Builder event:', event);

    switch (event.type) {
        case 'builder_started':
            loadingTitle.textContent = 'Starting AI widget builder';
            loadingMessage.textContent = event.message || 'Initializing...';
            break;

        case 'widget_stage_started':
            loadingTitle.textContent = 'Creating your widget';
            loadingMessage.textContent = 'Analyzing website theme and branding';
            break;

        case 'widget_theme_ready':
            loadingTitle.textContent = 'Widget created!';
            loadingMessage.textContent = 'Your widget is ready with custom branding';
            break;

        case 'knowledge_stage_started':
            loadingTitle.textContent = 'Building knowledge base';
            loadingMessage.textContent = 'Scraping and analyzing website content';
            break;

        case 'scraping_progress':
            const progress = event.data?.progress || 0;
            const pagesCount = event.data?.pages_count || 0;
            loadingTitle.textContent = 'Scanning your website';
            loadingMessage.textContent = `Analyzed ${pagesCount} pages (${Math.round(progress * 100)}%)`;
            break;

        case 'faq_generation_started':
            loadingTitle.textContent = 'Generating FAQs';
            loadingMessage.textContent = 'AI is creating helpful responses';
            break;

        case 'faq_generated':
            const faqCount = event.data?.faq_count || 0;
            loadingMessage.textContent = `Generated ${faqCount} FAQ items`;
            break;

        case 'completed':
            loadingTitle.textContent = 'All done! 🎉';
            loadingMessage.textContent = 'Preparing your preview...';

            // Store widget data and redirect
            widgetData = event.data;

            // Close connection and redirect
            setTimeout(() => {
                if (eventSource) {
                    eventSource.close();
                }
                redirectToPreview(websiteUrlInput.value, widgetData);
            }, 1000);
            break;

        case 'error':
            eventSource.close();
            showError(event.message || 'An error occurred while building the widget');
            break;

        default:
            console.log('Unknown event type:', event.type);
    }
}

// Redirect to preview page with widget data
function redirectToPreview(websiteUrl, data) {
    // Store data in sessionStorage for preview page
    sessionStorage.setItem('widgetPreview', JSON.stringify({
        websiteUrl: websiteUrl,
        widgetId: data.widget_id,
        embedCode: data.embed_code,
        email: USER_EMAIL,
        timestamp: Date.now()
    }));

    // Redirect to preview page
    window.location.href = 'preview.html';
}

// Show error state
function showError(message) {
    form.style.display = 'flex';
    loadingState.style.display = 'none';
    errorState.style.display = 'flex';
    errorMessage.textContent = message;

    // Close EventSource if open
    if (eventSource) {
        eventSource.close();
        eventSource = null;
    }
}

// Reset form to initial state
function resetForm() {
    form.style.display = 'flex';
    loadingState.style.display = 'none';
    errorState.style.display = 'none';
    websiteUrlInput.value = '';
    websiteUrlInput.focus();
    widgetData = null;

    // Close EventSource if open
    if (eventSource) {
        eventSource.close();
        eventSource = null;
    }
}

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    if (eventSource) {
        eventSource.close();
    }
});

// Auto-fill from URL parameter (for testing)
const urlParams = new URLSearchParams(window.location.search);
const prefilledUrl = urlParams.get('url');
if (prefilledUrl) {
    websiteUrlInput.value = prefilledUrl;
}
