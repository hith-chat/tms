import React, { useState } from 'react';
import { Button } from './Button';
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

    const encodedUrl = encodeURIComponent(websiteUrl);
    const email = 'sumansaurabh@hith.chat';
    const sseUrl = `/api/public/ai-widget-builder?url=${encodedUrl}&email=${encodeURIComponent(email)}`;

    const eventSource = new EventSource(sseUrl);
    let widgetData: any = null;

    eventSource.onmessage = (event) => {
      try {
        const data: BuilderEvent = JSON.parse(event.data);
        handleBuilderEvent(data);

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
        setBuildMessage('Analyzing your website theme...');
        break;
      case 'widget_theme_ready':
        setBuildStatus('Widget Created');
        setBuildMessage('Custom branding applied');
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
        setBuildMessage('Creating intelligent responses...');
        break;
      case 'completed':
        setBuildStatus('Complete');
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
      <div className={styles.container}>
        <div className={styles.content}>
          <h1 className={styles.headline}>
            The all-in-one AI chat widget for your website
          </h1>
          <p className={styles.subheadline}>
            Answer questions, schedule meetings, and capture leads automatically.
            Deploy in under 60 seconds.
          </p>

          {!isBuilding ? (
            <div className={styles.formWrapper}>
              <form onSubmit={handleSubmit} className={styles.form}>
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
                  Get Started Free
                </Button>
              </form>
              {error && <div className={styles.error}>{error}</div>}
              <p className={styles.subtext}>
                No credit card required • Free forever plan • 2 minute setup
              </p>
            </div>
          ) : (
            <div className={styles.loadingWrapper}>
              <div className={styles.loadingCard}>
                <div className={styles.spinner}></div>
                <h3>{buildStatus}</h3>
                <p>{buildMessage}</p>
                <div className={styles.progressBar}>
                  <div
                    className={styles.progressFill}
                    style={{ width: `${buildProgress}%` }}
                  ></div>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className={styles.visual}>
          <div className={styles.mockup}>
            <div className={styles.browserWindow}>
              <div className={styles.browserBar}>
                <div className={styles.browserDots}>
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              </div>
              <div className={styles.browserContent}>
                <img
                  src="https://images.unsplash.com/photo-1460925895917-afdab827c52f?w=1200&h=800&fit=crop&q=80"
                  alt="Dashboard Preview"
                />
                <div className={styles.chatBubble}>
                  <div className={styles.chatHeader}>
                    <div className={styles.chatAvatar}></div>
                    <div>
                      <div className={styles.chatName}>Support Assistant</div>
                      <div className={styles.chatStatus}>Online</div>
                    </div>
                  </div>
                  <div className={styles.chatBody}>
                    <div className={styles.message}>
                      Hi! How can I help you today?
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
