// Configuration
const API_BASE_URL = window.location.origin;

// State
let widgetData = null;
let isDesktop = true;
let iframeLoaded = false;

// DOM Elements
const websiteIframe = document.getElementById('websiteIframe');
const iframeLoading = document.getElementById('iframeLoading');
const iframeError = document.getElementById('iframeError');
const fallbackPreview = document.getElementById('fallbackPreview');
const deviceFrame = document.getElementById('deviceFrame');
const deviceToggleText = document.getElementById('deviceToggleText');
const websiteUrlElement = document.getElementById('websiteUrl');
const userEmailElement = document.getElementById('userEmail');
const frameUrlElement = document.getElementById('frameUrl');
const embedModal = document.getElementById('embedModal');
const embedCodeText = document.getElementById('embedCodeText');
const widgetContainer = document.getElementById('widgetContainer');

// Initialize
init();

function init() {
    // Load widget data from sessionStorage
    const storedData = sessionStorage.getItem('widgetPreview');

    if (!storedData) {
        // No data, redirect back to landing
        alert('No widget data found. Please build a widget first.');
        window.location.href = 'index.html';
        return;
    }

    try {
        widgetData = JSON.parse(storedData);

        // Update UI with widget data
        websiteUrlElement.textContent = extractDomain(widgetData.websiteUrl);
        userEmailElement.textContent = widgetData.email;
        frameUrlElement.textContent = widgetData.websiteUrl;

        // Load website in iframe
        loadWebsiteInIframe(widgetData.websiteUrl);

        // Inject chat widget
        setTimeout(() => {
            injectChatWidget();
        }, 2000);

    } catch (err) {
        console.error('Error parsing widget data:', err);
        alert('Invalid widget data. Please try again.');
        window.location.href = 'index.html';
    }
}

// Load website in iframe
function loadWebsiteInIframe(url) {
    iframeLoading.style.display = 'flex';
    iframeError.style.display = 'none';
    websiteIframe.style.display = 'none';
    fallbackPreview.style.display = 'none';

    // Set iframe source
    websiteIframe.src = url;

    // Set timeout for loading
    const loadTimeout = setTimeout(() => {
        if (!iframeLoaded) {
            showIframeError();
        }
    }, 10000); // 10 second timeout

    // Listen for iframe load
    websiteIframe.onload = function() {
        iframeLoaded = true;
        clearTimeout(loadTimeout);

        // Check if iframe loaded successfully
        try {
            // Try to access iframe content (will throw if blocked by X-Frame-Options)
            const iframeDoc = websiteIframe.contentDocument || websiteIframe.contentWindow.document;

            // If we can access it, show iframe
            iframeLoading.style.display = 'none';
            websiteIframe.style.display = 'block';
        } catch (err) {
            // Blocked by X-Frame-Options
            console.error('Iframe blocked:', err);
            showIframeError();
        }
    };

    // Listen for iframe errors
    websiteIframe.onerror = function() {
        clearTimeout(loadTimeout);
        showIframeError();
    };
}

// Show iframe error
function showIframeError() {
    iframeLoading.style.display = 'none';
    websiteIframe.style.display = 'none';
    iframeError.style.display = 'flex';
}

// Show fallback preview
function showFallbackPreview() {
    iframeError.style.display = 'none';
    fallbackPreview.style.display = 'block';
}

// Inject chat widget
function injectChatWidget() {
    if (!widgetData || !widgetData.widgetId) {
        console.error('No widget ID found');
        return;
    }

    // Create widget button
    const widgetButton = document.createElement('div');
    widgetButton.id = 'chatWidgetButton';
    widgetButton.innerHTML = `
        <style>
            #chatWidgetButton {
                position: fixed;
                bottom: 20px;
                right: 20px;
                width: 60px;
                height: 60px;
                border-radius: 50%;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
                cursor: pointer;
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 9999;
                transition: all 0.3s ease;
                animation: slideIn 0.5s ease;
            }

            #chatWidgetButton:hover {
                transform: scale(1.1);
                box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
            }

            #chatWidgetButton svg {
                width: 28px;
                height: 28px;
                color: white;
            }

            #chatWidgetWindow {
                position: fixed;
                bottom: 90px;
                right: 20px;
                width: 380px;
                height: 600px;
                background: white;
                border-radius: 16px;
                box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
                z-index: 9998;
                display: none;
                flex-direction: column;
                overflow: hidden;
                animation: slideUp 0.3s ease;
            }

            #chatWidgetWindow.open {
                display: flex;
            }

            .chat-widget-header {
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                padding: 20px;
                color: white;
                display: flex;
                align-items: center;
                gap: 12px;
            }

            .chat-widget-avatar {
                width: 40px;
                height: 40px;
                border-radius: 50%;
                background: white;
                color: #667eea;
                display: flex;
                align-items: center;
                justify-content: center;
                font-weight: 600;
                font-size: 14px;
            }

            .chat-widget-info {
                flex: 1;
            }

            .chat-widget-title {
                font-weight: 600;
                font-size: 16px;
                margin-bottom: 2px;
            }

            .chat-widget-status {
                font-size: 13px;
                opacity: 0.9;
            }

            .chat-widget-close {
                background: none;
                border: none;
                color: white;
                cursor: pointer;
                padding: 4px;
                opacity: 0.8;
                transition: opacity 0.2s;
            }

            .chat-widget-close:hover {
                opacity: 1;
            }

            .chat-widget-body {
                flex: 1;
                padding: 20px;
                background: #f9fafb;
                overflow-y: auto;
                display: flex;
                flex-direction: column;
                gap: 12px;
            }

            .chat-message {
                display: flex;
                animation: messageSlide 0.3s ease;
            }

            .chat-message.bot {
                justify-content: flex-start;
            }

            .chat-message.user {
                justify-content: flex-end;
            }

            .message-bubble {
                max-width: 75%;
                padding: 12px 16px;
                border-radius: 12px;
                font-size: 14px;
                line-height: 1.5;
            }

            .chat-message.bot .message-bubble {
                background: white;
                color: #111827;
                box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
            }

            .chat-message.user .message-bubble {
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
            }

            .chat-widget-footer {
                padding: 16px;
                background: white;
                border-top: 1px solid #e5e7eb;
            }

            .chat-input-container {
                display: flex;
                gap: 8px;
            }

            .chat-input {
                flex: 1;
                padding: 10px 14px;
                border: 1px solid #e5e7eb;
                border-radius: 8px;
                font-size: 14px;
                font-family: inherit;
                outline: none;
                transition: border-color 0.2s;
            }

            .chat-input:focus {
                border-color: #667eea;
            }

            .chat-send-btn {
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
                border: none;
                padding: 10px 16px;
                border-radius: 8px;
                cursor: pointer;
                font-weight: 500;
                font-size: 14px;
                transition: opacity 0.2s;
            }

            .chat-send-btn:hover {
                opacity: 0.9;
            }

            @keyframes slideIn {
                from {
                    opacity: 0;
                    transform: translateY(20px);
                }
                to {
                    opacity: 1;
                    transform: translateY(0);
                }
            }

            @keyframes slideUp {
                from {
                    opacity: 0;
                    transform: translateY(20px);
                }
                to {
                    opacity: 1;
                    transform: translateY(0);
                }
            }

            @keyframes messageSlide {
                from {
                    opacity: 0;
                    transform: translateY(10px);
                }
                to {
                    opacity: 1;
                    transform: translateY(0);
                }
            }

            @media (max-width: 768px) {
                #chatWidgetWindow {
                    width: calc(100% - 40px);
                    height: calc(100% - 110px);
                }
            }
        </style>
        <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"></path>
        </svg>
    `;

    // Create widget window
    const widgetWindow = document.createElement('div');
    widgetWindow.id = 'chatWidgetWindow';
    widgetWindow.innerHTML = `
        <div class="chat-widget-header">
            <div class="chat-widget-avatar">AI</div>
            <div class="chat-widget-info">
                <div class="chat-widget-title">Support Bot</div>
                <div class="chat-widget-status">● Online now</div>
            </div>
            <button class="chat-widget-close" onclick="toggleChatWidget()">
                <svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                </svg>
            </button>
        </div>
        <div class="chat-widget-body" id="chatWidgetBody">
            <div class="chat-message bot">
                <div class="message-bubble">
                    Hi there! 👋 I'm your AI assistant. How can I help you today?
                </div>
            </div>
        </div>
        <div class="chat-widget-footer">
            <div class="chat-input-container">
                <input type="text" class="chat-input" placeholder="Type your message..." id="chatInput" onkeypress="handleChatKeyPress(event)">
                <button class="chat-send-btn" onclick="sendChatMessage()">Send</button>
            </div>
        </div>
    `;

    // Append to page
    document.body.appendChild(widgetButton);
    document.body.appendChild(widgetWindow);

    // Add click event
    widgetButton.addEventListener('click', toggleChatWidget);

    console.log('Chat widget injected successfully!');
}

// Toggle chat widget
function toggleChatWidget() {
    const widgetWindow = document.getElementById('chatWidgetWindow');
    if (widgetWindow) {
        widgetWindow.classList.toggle('open');
    }
}

// Handle chat key press
function handleChatKeyPress(event) {
    if (event.key === 'Enter') {
        sendChatMessage();
    }
}

// Send chat message
function sendChatMessage() {
    const chatInput = document.getElementById('chatInput');
    const chatBody = document.getElementById('chatWidgetBody');

    if (!chatInput || !chatBody) return;

    const message = chatInput.value.trim();
    if (!message) return;

    // Add user message
    const userMessage = document.createElement('div');
    userMessage.className = 'chat-message user';
    userMessage.innerHTML = `
        <div class="message-bubble">
            ${escapeHtml(message)}
        </div>
    `;
    chatBody.appendChild(userMessage);

    // Clear input
    chatInput.value = '';

    // Scroll to bottom
    chatBody.scrollTop = chatBody.scrollHeight;

    // Simulate bot response
    setTimeout(() => {
        const botResponses = [
            "That's a great question! I'd be happy to help you with that.",
            "Let me look into that for you. Can you provide more details?",
            "I understand! Would you like me to schedule a call with our team?",
            "Thanks for reaching out! I can help you with that right away.",
            "Great! Let me connect you with the right person to assist you."
        ];

        const randomResponse = botResponses[Math.floor(Math.random() * botResponses.length)];

        const botMessage = document.createElement('div');
        botMessage.className = 'chat-message bot';
        botMessage.innerHTML = `
            <div class="message-bubble">
                ${randomResponse}
            </div>
        `;
        chatBody.appendChild(botMessage);

        // Scroll to bottom
        chatBody.scrollTop = chatBody.scrollHeight;
    }, 1000);
}

// Toggle device view
function toggleDevice() {
    isDesktop = !isDesktop;
    deviceFrame.className = `device-frame ${isDesktop ? 'desktop' : 'mobile'}`;
    deviceToggleText.textContent = isDesktop ? 'Mobile' : 'Desktop';
}

// Get embed code
function getEmbedCode() {
    if (!widgetData || !widgetData.embedCode) {
        alert('Embed code not available');
        return;
    }

    embedCodeText.textContent = widgetData.embedCode;
    embedModal.style.display = 'flex';
}

// Close embed modal
function closeEmbedModal() {
    embedModal.style.display = 'none';
}

// Copy embed code
function copyEmbedCode() {
    const code = embedCodeText.textContent;

    navigator.clipboard.writeText(code).then(() => {
        // Show success feedback
        const btn = event.target.closest('button');
        const originalText = btn.innerHTML;
        btn.innerHTML = `
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                <circle cx="10" cy="10" r="8" stroke="currentColor" stroke-width="2"/>
                <path d="M7 10L9 12L13 8" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            Copied!
        `;

        setTimeout(() => {
            btn.innerHTML = originalText;
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy:', err);
        alert('Failed to copy code. Please copy manually.');
    });
}

// Utility functions
function extractDomain(url) {
    try {
        const urlObj = new URL(url);
        return urlObj.hostname;
    } catch (err) {
        return url;
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Close modal on escape key
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && embedModal.style.display === 'flex') {
        closeEmbedModal();
    }
});
