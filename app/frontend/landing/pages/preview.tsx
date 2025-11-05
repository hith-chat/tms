import React, { useState, useEffect } from 'react';
import { Helmet } from 'react-helmet-async';
import { Button } from '../components/Button';
import { Copy, Check, ExternalLink, AlertCircle, Sparkles, X } from 'lucide-react';
import styles from './preview.module.css';

interface WidgetPreviewData {
  websiteUrl: string;
  widgetId: string;
  embedCode: string;
  email: string;
  timestamp: number;
}

export default function PreviewPage() {
  const [previewData, setPreviewData] = useState<WidgetPreviewData | null>(null);
  const [iframeError, setIframeError] = useState(false);
  const [copied, setCopied] = useState(false);
  const [showEmbedCode, setShowEmbedCode] = useState(false);

  useEffect(() => {
    // Load preview data from sessionStorage
    const data = sessionStorage.getItem('widgetPreview');
    if (data) {
      try {
        const parsed = JSON.parse(data);
        setPreviewData(parsed);
      } catch (err) {
        console.error('Failed to parse preview data:', err);
        window.location.href = '/';
      }
    } else {
      // No data, redirect back to home
      window.location.href = '/';
    }
  }, []);

  useEffect(() => {
    if (!previewData) return;

    // Try to load the website in iframe and inject widget
    const iframe = document.getElementById('websitePreview') as HTMLIFrameElement;
    if (!iframe) return;

    iframe.addEventListener('load', handleIframeLoad);
    iframe.addEventListener('error', handleIframeError);

    // Set iframe src
    try {
      iframe.src = previewData.websiteUrl;
    } catch (err) {
      setIframeError(true);
    }

    return () => {
      iframe.removeEventListener('load', handleIframeLoad);
      iframe.removeEventListener('error', handleIframeError);
    };
  }, [previewData]);

  function handleIframeLoad() {
    const iframe = document.getElementById('websitePreview') as HTMLIFrameElement;
    if (!iframe || !previewData) return;

    try {
      // Try to access iframe content to check CORS
      const iframeDoc = iframe.contentDocument || iframe.contentWindow?.document;
      if (iframeDoc) {
        // Successfully accessed - inject widget script
        injectWidget(iframeDoc);
      }
    } catch (err) {
      // CORS error - can't access iframe content
      console.log('Cannot access iframe content due to CORS policy');
      // Widget will still work when user installs it on their site
    }
  }

  function handleIframeError() {
    setIframeError(true);
  }

  function injectWidget(doc: Document) {
    if (!previewData) return;

    try {
      // Create script element
      const script = doc.createElement('script');
      script.innerHTML = previewData.embedCode;
      doc.body.appendChild(script);
    } catch (err) {
      console.error('Failed to inject widget:', err);
    }
  }

  async function copyEmbedCode() {
    if (!previewData) return;

    try {
      await navigator.clipboard.writeText(previewData.embedCode);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  }

  if (!previewData) {
    return (
      <div className={styles.loading}>
        <div className={styles.spinner}></div>
        <p>Loading preview...</p>
      </div>
    );
  }

  return (
    <>
      <Helmet>
        <title>Widget Preview | Hith Chat Widget</title>
        <meta name="description" content="Preview your AI chat widget" />
      </Helmet>

      <div className={styles.previewPage}>
        {/* Header */}
        <header className={styles.header}>
          <div className={styles.headerContent}>
            <div className={styles.headerLeft}>
              <div className={styles.logo}>
                <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
                  <rect width="32" height="32" rx="8" fill="url(#gradient)" />
                  <path
                    d="M8 12C8 10.8954 8.89543 10 10 10H22C23.1046 10 24 10.8954 24 12V18C24 19.1046 23.1046 20 22 20H16L12 24V20H10C8.89543 20 8 19.1046 8 18V12Z"
                    fill="white"
                  />
                  <defs>
                    <linearGradient id="gradient" x1="0" y1="0" x2="32" y2="32">
                      <stop offset="0%" stopColor="hsl(215 80% 60%)" />
                      <stop offset="100%" stopColor="hsl(260 75% 65%)" />
                    </linearGradient>
                  </defs>
                </svg>
                <span>Hith Chat Widget</span>
              </div>
              <div className={styles.divider}></div>
              <div className={styles.urlDisplay}>
                <span className={styles.urlLabel}>Preview:</span>
                <span className={styles.url}>{new URL(previewData.websiteUrl).hostname}</span>
              </div>
            </div>

            <div className={styles.headerRight}>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowEmbedCode(true)}
              >
                <Copy size={16} />
                Get Embed Code
              </Button>
              <Button
                size="sm"
                asChild
              >
                <a href={previewData.websiteUrl} target="_blank" rel="noopener noreferrer">
                  <ExternalLink size={16} />
                  Open Original
                </a>
              </Button>
            </div>
          </div>
        </header>

        {/* Main Content */}
        <main className={styles.main}>
          {/* Success Banner */}
          <div className={styles.successBanner}>
            <div className={styles.bannerIcon}>
              <Sparkles size={20} />
            </div>
            <div className={styles.bannerContent}>
              <h2>Your AI widget is ready! 🎉</h2>
              <p>
                {iframeError
                  ? 'Your website prevents iframe embedding, but the widget will work perfectly when installed on your actual site.'
                  : 'Preview how the chat widget looks on your website. Click "Get Embed Code" to install it.'}
              </p>
            </div>
          </div>

          {/* Preview Container */}
          <div className={styles.previewContainer}>
            {iframeError ? (
              <div className={styles.iframeError}>
                <AlertCircle size={48} />
                <h3>Cannot display iframe preview</h3>
                <p>
                  Your website has security settings (X-Frame-Options) that prevent embedding in iframes.
                  <br />
                  <strong>Don't worry!</strong> The widget will work perfectly when you install the embed code on your actual website.
                </p>
                <Button onClick={() => setShowEmbedCode(true)}>
                  Get Embed Code
                </Button>
              </div>
            ) : (
              <iframe
                id="websitePreview"
                className={styles.iframe}
                title="Website Preview"
                sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
              />
            )}
          </div>

          {/* Installation Instructions */}
          <div className={styles.instructions}>
            <h3>Next Steps</h3>
            <ol>
              <li>
                <strong>Copy the embed code</strong> by clicking the "Get Embed Code" button above
              </li>
              <li>
                <strong>Paste it in your website</strong> just before the closing <code>&lt;/body&gt;</code> tag
              </li>
              <li>
                <strong>Done!</strong> Your AI chat widget will appear on all pages
              </li>
            </ol>
          </div>
        </main>

        {/* Embed Code Modal */}
        {showEmbedCode && (
          <div className={styles.modal} onClick={() => setShowEmbedCode(false)}>
            <div className={styles.modalContent} onClick={(e) => e.stopPropagation()}>
              <div className={styles.modalHeader}>
                <h2>Install Your Widget</h2>
                <button
                  className={styles.closeBtn}
                  onClick={() => setShowEmbedCode(false)}
                >
                  <X size={20} />
                </button>
              </div>

              <div className={styles.modalBody}>
                <p>Copy this code and paste it before the closing <code>&lt;/body&gt;</code> tag on your website:</p>

                <div className={styles.codeBlock}>
                  <pre>
                    <code>{previewData.embedCode}</code>
                  </pre>
                  <button
                    className={styles.copyBtn}
                    onClick={copyEmbedCode}
                  >
                    {copied ? (
                      <>
                        <Check size={16} />
                        Copied!
                      </>
                    ) : (
                      <>
                        <Copy size={16} />
                        Copy Code
                      </>
                    )}
                  </button>
                </div>

                <div className={styles.platformGuides}>
                  <h4>Installation Guides:</h4>
                  <div className={styles.guideLinks}>
                    <a href="#" className={styles.guideLink}>WordPress</a>
                    <a href="#" className={styles.guideLink}>Shopify</a>
                    <a href="#" className={styles.guideLink}>Webflow</a>
                    <a href="#" className={styles.guideLink}>Wix</a>
                    <a href="#" className={styles.guideLink}>Custom HTML</a>
                  </div>
                </div>
              </div>

              <div className={styles.modalFooter}>
                <Button variant="outline" onClick={() => setShowEmbedCode(false)}>
                  Close
                </Button>
                <Button onClick={copyEmbedCode}>
                  {copied ? (
                    <>
                      <Check size={16} />
                      Copied!
                    </>
                  ) : (
                    <>
                      <Copy size={16} />
                      Copy Code
                    </>
                  )}
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
