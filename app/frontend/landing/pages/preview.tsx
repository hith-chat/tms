import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Helmet } from 'react-helmet-async';
import { Button } from '../components/Button';
import { ArrowLeft, Copy, Check, ExternalLink, AlertCircle, Sparkles } from 'lucide-react';

interface WidgetPreviewData {
  websiteUrl: string;
  widgetId: string;
  embedCode: string;
  email: string;
  timestamp: number;
}

export default function PreviewPage() {
  const navigate = useNavigate();
  const [previewData, setPreviewData] = useState<WidgetPreviewData | null>(null);
  const [iframeError, setIframeError] = useState(false);
  const [copiedEmbed, setCopiedEmbed] = useState(false);
  const [showCode, setShowCode] = useState(false);

  useEffect(() => {
    // Load preview data from sessionStorage
    const stored = sessionStorage.getItem('widgetPreview');
    if (stored) {
      try {
        const data = JSON.parse(stored);
        setPreviewData(data);
      } catch (err) {
        console.error('Error parsing preview data:', err);
        navigate('/');
      }
    } else {
      navigate('/');
    }
  }, [navigate]);

  const handleCopyEmbed = async () => {
    if (!previewData?.embedCode) return;

    try {
      await navigator.clipboard.writeText(previewData.embedCode);
      setCopiedEmbed(true);
      setTimeout(() => setCopiedEmbed(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  const handleIframeError = () => {
    setIframeError(true);
  };

  if (!previewData) {
    return (
      <div style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--background)'
      }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{
            width: 48,
            height: 48,
            border: '4px solid var(--border)',
            borderTopColor: 'var(--primary)',
            borderRadius: '50%',
            animation: 'spin 1s linear infinite',
            margin: '0 auto 16px'
          }} />
          <p style={{ color: 'var(--muted-foreground)' }}>Loading preview...</p>
        </div>
      </div>
    );
  }

  return (
    <>
      <Helmet>
        <title>Widget Preview | Hith AI Chat Widget</title>
        <meta name="description" content="Preview your AI chat widget" />
      </Helmet>

      <div style={{
        minHeight: '100vh',
        background: 'var(--background)',
        display: 'flex',
        flexDirection: 'column'
      }}>
        {/* Header */}
        <header style={{
          background: 'var(--surface)',
          borderBottom: '1px solid var(--border)',
          padding: '16px 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 16,
          flexWrap: 'wrap'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <Button
              variant="ghost"
              onClick={() => navigate('/')}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '8px 16px'
              }}
            >
              <ArrowLeft size={18} />
              Back to Home
            </Button>
            <div style={{ width: 1, height: 24, background: 'var(--border)' }} />
            <div>
              <h1 style={{
                fontSize: 18,
                fontWeight: 700,
                margin: 0,
                display: 'flex',
                alignItems: 'center',
                gap: 8
              }}>
                <Sparkles size={20} style={{ color: 'var(--primary)' }} />
                Widget Preview
              </h1>
              <p style={{
                fontSize: 13,
                color: 'var(--muted-foreground)',
                margin: 0,
                marginTop: 2
              }}>
                Preview for: {new URL(previewData.websiteUrl).hostname}
              </p>
            </div>
          </div>

          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
            <Button
              variant="outline"
              onClick={() => setShowCode(!showCode)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '10px 20px'
              }}
            >
              {showCode ? 'Hide' : 'Show'} Embed Code
            </Button>
            <Button
              onClick={handleCopyEmbed}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '10px 20px',
                background: copiedEmbed ? 'var(--success)' : 'var(--primary)'
              }}
            >
              {copiedEmbed ? (
                <>
                  <Check size={18} />
                  Copied!
                </>
              ) : (
                <>
                  <Copy size={18} />
                  Copy Embed Code
                </>
              )}
            </Button>
          </div>
        </header>

        {/* Embed Code Section */}
        {showCode && (
          <div style={{
            background: 'var(--surface)',
            borderBottom: '1px solid var(--border)',
            padding: '24px'
          }}>
            <div style={{
              maxWidth: 1200,
              margin: '0 auto'
            }}>
              <h2 style={{
                fontSize: 16,
                fontWeight: 700,
                marginBottom: 12
              }}>
                Installation Instructions
              </h2>
              <p style={{
                fontSize: 14,
                color: 'var(--muted-foreground)',
                marginBottom: 16
              }}>
                Copy and paste this code snippet before the closing &lt;/body&gt; tag in your website's HTML:
              </p>
              <div style={{
                position: 'relative',
                background: 'var(--background)',
                border: '1px solid var(--border)',
                borderRadius: 12,
                padding: 16,
                fontFamily: 'var(--font-family-monospace)',
                fontSize: 13,
                overflow: 'auto',
                maxHeight: 300
              }}>
                <pre style={{
                  margin: 0,
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-all'
                }}>
                  <code>{previewData.embedCode}</code>
                </pre>
              </div>
            </div>
          </div>
        )}

        {/* Preview Area */}
        <div style={{
          flex: 1,
          padding: 24,
          display: 'flex',
          flexDirection: 'column',
          gap: 16,
          maxWidth: 1400,
          width: '100%',
          margin: '0 auto'
        }}>
          {/* Info Banner */}
          <div style={{
            background: 'linear-gradient(135deg, hsl(215 80% 60% / 0.1) 0%, hsl(160 70% 45% / 0.1) 100%)',
            border: '1px solid hsl(215 80% 60% / 0.2)',
            borderRadius: 12,
            padding: '16px 20px',
            display: 'flex',
            alignItems: 'center',
            gap: 12
          }}>
            <Sparkles size={20} style={{ color: 'var(--primary)', flexShrink: 0 }} />
            <div style={{ flex: 1 }}>
              <p style={{
                margin: 0,
                fontSize: 14,
                fontWeight: 600,
                color: 'var(--foreground)'
              }}>
                Your AI chat widget has been created successfully!
              </p>
              <p style={{
                margin: 0,
                marginTop: 4,
                fontSize: 13,
                color: 'var(--muted-foreground)'
              }}>
                Below is a preview of how it will look on your website. The widget is ready to install.
              </p>
            </div>
          </div>

          {/* Preview Frame */}
          <div style={{
            flex: 1,
            background: 'var(--surface)',
            border: '2px solid var(--border)',
            borderRadius: 16,
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            minHeight: 600
          }}>
            {/* Browser Header */}
            <div style={{
              background: 'var(--muted)',
              padding: '12px 16px',
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              borderBottom: '1px solid var(--border)'
            }}>
              <div style={{ display: 'flex', gap: 6 }}>
                <div style={{ width: 12, height: 12, borderRadius: '50%', background: '#ff5f57' }} />
                <div style={{ width: 12, height: 12, borderRadius: '50%', background: '#febc2e' }} />
                <div style={{ width: 12, height: 12, borderRadius: '50%', background: '#28c840' }} />
              </div>
              <div style={{
                flex: 1,
                background: 'var(--background)',
                borderRadius: 6,
                padding: '6px 12px',
                fontSize: 12,
                color: 'var(--muted-foreground)',
                display: 'flex',
                alignItems: 'center',
                gap: 8
              }}>
                <span style={{ flex: 1, textAlign: 'center' }}>
                  {new URL(previewData.websiteUrl).hostname}
                </span>
                <a
                  href={previewData.websiteUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 4,
                    color: 'var(--primary)',
                    textDecoration: 'none',
                    fontSize: 11
                  }}
                >
                  <ExternalLink size={14} />
                  Open
                </a>
              </div>
            </div>

            {/* iframe or error */}
            <div style={{ flex: 1, position: 'relative', background: 'white' }}>
              {iframeError ? (
                <div style={{
                  position: 'absolute',
                  inset: 0,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  flexDirection: 'column',
                  gap: 16,
                  padding: 32,
                  textAlign: 'center'
                }}>
                  <AlertCircle size={48} style={{ color: 'var(--warning)' }} />
                  <div>
                    <h3 style={{ fontSize: 18, fontWeight: 700, marginBottom: 8 }}>
                      Cannot display website in preview
                    </h3>
                    <p style={{
                      fontSize: 14,
                      color: 'var(--muted-foreground)',
                      marginBottom: 16,
                      maxWidth: 500
                    }}>
                      This website cannot be displayed in an iframe due to security restrictions (X-Frame-Options).
                      Your widget will still work when installed on your actual website.
                    </p>
                    <div style={{
                      background: 'var(--muted)',
                      borderRadius: 8,
                      padding: '12px 16px',
                      fontSize: 13,
                      marginBottom: 16
                    }}>
                      <strong>Widget ID:</strong> {previewData.widgetId}
                    </div>
                    <Button
                      onClick={() => window.open(previewData.websiteUrl, '_blank')}
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 8
                      }}
                    >
                      <ExternalLink size={18} />
                      Open Website in New Tab
                    </Button>
                  </div>
                </div>
              ) : (
                <iframe
                  src={previewData.websiteUrl}
                  style={{
                    width: '100%',
                    height: '100%',
                    border: 'none'
                  }}
                  title="Website Preview"
                  onError={handleIframeError}
                  sandbox="allow-same-origin allow-scripts allow-popups allow-forms"
                />
              )}

              {/* Widget Script Injection Notice */}
              {!iframeError && (
                <div style={{
                  position: 'absolute',
                  bottom: 16,
                  left: 16,
                  right: 16,
                  background: 'hsl(215 80% 60% / 0.95)',
                  color: 'white',
                  padding: '12px 16px',
                  borderRadius: 12,
                  fontSize: 13,
                  boxShadow: '0 8px 24px hsl(220 15% 15% / 0.2)'
                }}>
                  <strong>Note:</strong> This is a preview of your original website. To see the widget in action, install the embed code on your actual site.
                </div>
              )}
            </div>
          </div>

          {/* Next Steps */}
          <div style={{
            background: 'var(--surface)',
            border: '1px solid var(--border)',
            borderRadius: 12,
            padding: 24
          }}>
            <h3 style={{
              fontSize: 16,
              fontWeight: 700,
              marginBottom: 16
            }}>
              Next Steps
            </h3>
            <ol style={{
              margin: 0,
              paddingLeft: 20,
              display: 'flex',
              flexDirection: 'column',
              gap: 12
            }}>
              <li style={{ fontSize: 14, lineHeight: 1.6 }}>
                <strong>Copy the embed code</strong> using the button above
              </li>
              <li style={{ fontSize: 14, lineHeight: 1.6 }}>
                <strong>Paste it into your website's HTML</strong> before the closing &lt;/body&gt; tag
              </li>
              <li style={{ fontSize: 14, lineHeight: 1.6 }}>
                <strong>Deploy your changes</strong> and the widget will appear on your site
              </li>
              <li style={{ fontSize: 14, lineHeight: 1.6 }}>
                <strong>Test the widget</strong> to ensure it's working correctly
              </li>
            </ol>

            <div style={{
              marginTop: 24,
              paddingTop: 24,
              borderTop: '1px solid var(--border)',
              display: 'flex',
              gap: 12,
              flexWrap: 'wrap'
            }}>
              <Button
                variant="outline"
                onClick={() => navigate('/')}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8
                }}
              >
                <ArrowLeft size={18} />
                Back to Home
              </Button>
              <Button
                onClick={handleCopyEmbed}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  background: copiedEmbed ? 'var(--success)' : 'var(--primary)'
                }}
              >
                {copiedEmbed ? (
                  <>
                    <Check size={18} />
                    Copied!
                  </>
                ) : (
                  <>
                    <Copy size={18} />
                    Copy Embed Code
                  </>
                )}
              </Button>
            </div>
          </div>
        </div>
      </div>

      <style>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </>
  );
}
