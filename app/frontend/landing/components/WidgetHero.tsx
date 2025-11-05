import React, { useState, useEffect } from 'react';
import { Button } from './Button';
import { Sparkles, MessageSquare, Zap, Users, Check } from 'lucide-react';
import styles from './WidgetHero.module.css';

interface BuilderEvent {
  type: string;
  message?: string;
  data?: any;
}

export const WidgetHero = () => {
  const [websiteUrl, setWebsiteUrl] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isBuilding, setIsBuilding] = useState(false);
  const [buildProgress, setBuildProgress] = useState(0);
  const [buildStatus, setBuildStatus] = useState('');
  const [buildMessage, setBuildMessage] = useState('');
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    setIsVisible(true);
  }, []);

  function validateUrl(url: string): boolean {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    if (!websiteUrl.trim()) {
      setError('Please enter your website URL');
      return;
    }

    if (!validateUrl(websiteUrl)) {
      setError('Please enter a valid URL (e.g., https://example.com)');
      return;
    }

    setError(null);
    setIsBuilding(true);
    setBuildProgress(0);
    setBuildStatus('Initializing');
    setBuildMessage('Connecting to your website...');

    // Start SSE connection
    const encodedUrl = encodeURIComponent(websiteUrl);
    const email = 'sumansaurabh@hith.chat';
    const sseUrl = `/api/public/ai-widget-builder?url=${encodedUrl}&email=${encodeURIComponent(email)}`;

    const eventSource = new EventSource(sseUrl);
    let widgetData: any = null;

    eventSource.onmessage = (event) => {
      try {
        const data: BuilderEvent = JSON.parse(event.data);
        handleBuilderEvent(data);

        // Update progress
        if (data.type === 'builder_started') setBuildProgress(10);
        else if (data.type === 'widget_stage_started') setBuildProgress(20);
        else if (data.type === 'widget_theme_ready') setBuildProgress(40);
        else if (data.type === 'knowledge_stage_started') setBuildProgress(50);
        else if (data.type === 'scraping_progress') {
          setBuildProgress(50 + (data.data?.progress || 0) * 30);
        }
        else if (data.type === 'faq_generation_started') setBuildProgress(80);
        else if (data.type === 'completed') {
          setBuildProgress(100);
          widgetData = data.data;
        }
      } catch (err) {
        console.error('Error parsing SSE data:', err);
      }
    };

    eventSource.onerror = () => {
      eventSource.close();

      if (widgetData && widgetData.widget_id) {
        redirectToPreview(widgetData);
      } else {
        setIsBuilding(false);
        setError('Connection failed. Please try again.');
      }
    };
  }

  function handleBuilderEvent(event: BuilderEvent) {
    switch (event.type) {
      case 'builder_started':
        setBuildStatus('Starting');
        setBuildMessage(event.message || 'Initializing AI widget builder...');
        break;
      case 'widget_stage_started':
        setBuildStatus('Creating Widget');
        setBuildMessage('Analyzing your website theme and branding...');
        break;
      case 'widget_theme_ready':
        setBuildStatus('Widget Created');
        setBuildMessage('Custom branding applied successfully!');
        break;
      case 'knowledge_stage_started':
        setBuildStatus('Building Knowledge Base');
        setBuildMessage('Scanning website content...');
        break;
      case 'scraping_progress':
        const progress = event.data?.progress || 0;
        const pagesCount = event.data?.pages_count || 0;
        setBuildStatus('Analyzing Content');
        setBuildMessage(`Scanned ${pagesCount} pages (${Math.round(progress * 100)}%)`);
        break;
      case 'faq_generation_started':
        setBuildStatus('Generating FAQs');
        setBuildMessage('AI is creating intelligent responses...');
        break;
      case 'completed':
        setBuildStatus('Complete!');
        setBuildMessage('Preparing your preview...');
        setTimeout(() => redirectToPreview(event.data), 1000);
        break;
      case 'error':
        setIsBuilding(false);
        setError(event.message || 'An error occurred');
        break;
    }
  }

  function redirectToPreview(data: any) {
    sessionStorage.setItem('widgetPreview', JSON.stringify({
      websiteUrl,
      widgetId: data.widget_id,
      embedCode: data.embed_code,
      email: 'sumansaurabh@hith.chat',
      timestamp: Date.now()
    }));
    window.location.href = '/preview';
  }

  return (
    <section className={styles.hero}>
      <div className={styles.background}>
        <div className={styles.gradientOrb1}></div>
        <div className={styles.gradientOrb2}></div>
        <div className={styles.gradientOrb3}></div>
      </div>

      <div className={styles.container}>
        <div
          className={styles.content}
          style={{
            transform: isVisible ? 'translateY(0)' : 'translateY(30px)',
            opacity: isVisible ? 1 : 0,
            transition: 'all 0.8s cubic-bezier(0.22, 1, 0.36, 1)',
          }}
        >
          {/* Badge */}
          <div className={styles.badge}>
            <Sparkles size={16} />
            <span>AI-Powered Chat Widget</span>
          </div>

          {/* Main Headline */}
          <h1 className={styles.headline}>
            Turn Website Visitors
            <br />
            into <span className={styles.gradientText}>Happy Customers</span>
          </h1>

          <p className={styles.subheadline}>
            AI chat widget that answers questions, schedules meetings, and generates leads automatically.
            <br />
            <strong>Set up in 60 seconds.</strong>
          </p>

          {/* Widget Builder Form */}
          {!isBuilding ? (
            <div className={styles.builderCard}>
              <div className={styles.cardHeader}>
                <h3>Try it on your website</h3>
                <p>Enter your URL to see the magic happen ✨</p>
              </div>

              <form onSubmit={handleSubmit} className={styles.form}>
                <div className={styles.inputWrapper}>
                  <input
                    type="url"
                    value={websiteUrl}
                    onChange={(e) => setWebsiteUrl(e.target.value)}
                    placeholder="https://yourwebsite.com"
                    className={styles.input}
                    disabled={isBuilding}
                  />
                  <Button
                    type="submit"
                    size="lg"
                    className={styles.submitBtn}
                    disabled={isBuilding}
                  >
                    Build My Widget
                    <Zap size={18} />
                  </Button>
                </div>
                {error && <div className={styles.error}>{error}</div>}
              </form>

              {/* Trust Indicators */}
              <div className={styles.trustIndicators}>
                <div className={styles.trustItem}>
                  <Check size={16} />
                  <span>No credit card required</span>
                </div>
                <div className={styles.trustItem}>
                  <Check size={16} />
                  <span>Free forever plan</span>
                </div>
                <div className={styles.trustItem}>
                  <Check size={16} />
                  <span>Setup in 60 seconds</span>
                </div>
              </div>
            </div>
          ) : (
            <div className={styles.builderCard}>
              <div className={styles.loadingState}>
                <div className={styles.spinner}>
                  <div className={styles.spinnerRing}></div>
                  <div className={styles.spinnerRing}></div>
                  <div className={styles.spinnerRing}></div>
                </div>

                <h3 className={styles.loadingTitle}>{buildStatus}</h3>
                <p className={styles.loadingMessage}>{buildMessage}</p>

                <div className={styles.progressBar}>
                  <div
                    className={styles.progressFill}
                    style={{ width: `${buildProgress}%` }}
                  ></div>
                </div>

                <div className={styles.progressPercent}>
                  {Math.round(buildProgress)}%
                </div>
              </div>
            </div>
          )}

          {/* Feature Pills */}
          <div className={styles.featurePills}>
            <div className={styles.pill}>
              <MessageSquare size={16} />
              <span>24/7 AI Support</span>
            </div>
            <div className={styles.pill}>
              <Users size={16} />
              <span>Lead Generation</span>
            </div>
            <div className={styles.pill}>
              <Zap size={16} />
              <span>Instant Setup</span>
            </div>
          </div>
        </div>

        {/* Hero Visual */}
        <div
          className={styles.visual}
          style={{
            transform: isVisible ? 'translateY(0)' : 'translateY(30px)',
            opacity: isVisible ? 1 : 0,
            transition: 'all 0.8s cubic-bezier(0.22, 1, 0.36, 1) 0.2s',
          }}
        >
          <div className={styles.mockup}>
            <div className={styles.browserFrame}>
              <div className={styles.browserHeader}>
                <div className={styles.browserDots}>
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
                <div className={styles.browserUrl}>yourwebsite.com</div>
              </div>

              <div className={styles.browserContent}>
                <img
                  src="https://images.unsplash.com/photo-1460925895917-afdab827c52f?w=1200&h=800&fit=crop&q=80"
                  alt="Website Preview"
                />

                {/* Animated Chat Widget */}
                <div className={styles.chatWidget}>
                  <div className={styles.chatHeader}>
                    <div className={styles.chatAvatar}>
                      <MessageSquare size={20} />
                    </div>
                    <div className={styles.chatInfo}>
                      <div className={styles.chatName}>AI Assistant</div>
                      <div className={styles.chatStatus}>
                        <span className={styles.statusDot}></span>
                        Online
                      </div>
                    </div>
                  </div>

                  <div className={styles.chatMessages}>
                    <div className={styles.chatMessage}>
                      <div className={styles.messageBubble}>
                        Hi! 👋 How can I help you today?
                      </div>
                    </div>
                    <div className={styles.chatMessage}>
                      <div className={styles.messageBubble}>
                        I can answer questions, schedule meetings, or connect you with our team!
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};
