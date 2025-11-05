import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from './Button';
import { Avatar, AvatarImage, AvatarFallback } from './Avatar';
import { MoveRight, Sparkles, Loader2 } from 'lucide-react';
import styles from './HeroSection.module.css';

const USER_EMAIL = 'sumansaurabh@hith.chat';

export const HeroSection = ({ className }: { className?: string }) => {
  const [websiteUrl, setWebsiteUrl] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isVisible, setIsVisible] = useState(false);
  const [isBuilding, setIsBuilding] = useState(false);
  const [buildProgress, setBuildProgress] = useState(0);
  const [buildMessage, setBuildMessage] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    setIsVisible(true);
  }, []);

  function validateUrl(url?: string) {
    if (!url) return false;
    try {
      const parsed = new URL(url);
      return parsed.protocol === 'http:' || parsed.protocol === 'https:';
    } catch {
      return false;
    }
  }

  async function handleBuildWidget(url: string) {
    setIsBuilding(true);
    setError(null);
    setBuildProgress(0);
    setBuildMessage('Initializing AI widget builder...');

    try {
      const encodedUrl = encodeURIComponent(url);
      const sseUrl = `${window.location.origin}/api/public/ai-widget-builder?url=${encodedUrl}`;

      const eventSource = new EventSource(sseUrl);
      let widgetData: any = null;

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);

          switch (data.type) {
            case 'builder_started':
              setBuildProgress(10);
              setBuildMessage('Starting AI widget builder...');
              break;
            case 'widget_stage_started':
              setBuildProgress(20);
              setBuildMessage('Analyzing your website theme...');
              break;
            case 'widget_theme_ready':
              setBuildProgress(40);
              setBuildMessage('Custom branding ready!');
              break;
            case 'knowledge_stage_started':
              setBuildProgress(50);
              setBuildMessage('Building knowledge base...');
              break;
            case 'scraping_progress':
              const progress = data.data?.progress || 0;
              const pagesCount = data.data?.pages_count || 0;
              setBuildProgress(50 + progress * 30);
              setBuildMessage(`Analyzed ${pagesCount} pages...`);
              break;
            case 'faq_generation_started':
              setBuildProgress(85);
              setBuildMessage('Generating FAQs with AI...');
              break;
            case 'completed':
              setBuildProgress(100);
              setBuildMessage('Widget ready! Redirecting...');
              widgetData = data.data;

              // Store data and redirect to preview
              setTimeout(() => {
                eventSource.close();
                sessionStorage.setItem('widgetPreview', JSON.stringify({
                  websiteUrl: url,
                  widgetId: widgetData.widget_id,
                  embedCode: widgetData.embed_code,
                  email: USER_EMAIL,
                  timestamp: Date.now()
                }));
                navigate('/preview');
              }, 1000);
              break;
            case 'error':
              eventSource.close();
              setIsBuilding(false);
              setError(data.message || 'Failed to build widget. Please try again.');
              break;
          }
        } catch (err) {
          console.error('Error parsing SSE data:', err);
        }
      };

      eventSource.onerror = (err) => {
        console.error('EventSource error:', err);
        eventSource.close();

        if (widgetData && widgetData.widget_id) {
          // Success - redirect
          sessionStorage.setItem('widgetPreview', JSON.stringify({
            websiteUrl: url,
            widgetId: widgetData.widget_id,
            embedCode: widgetData.embed_code,
            email: USER_EMAIL,
            timestamp: Date.now()
          }));
          navigate('/preview');
        } else {
          setIsBuilding(false);
          setError('Connection lost. Please try again.');
        }
      };
    } catch (err) {
      setIsBuilding(false);
      setError('Failed to start widget builder. Please try again.');
      console.error('Build error:', err);
    }
  }

  function onSubmit(evt?: React.FormEvent) {
    evt?.preventDefault();

    if (!validateUrl(websiteUrl)) {
      setError('Please enter a valid website URL (e.g., https://example.com)');
      return;
    }

    handleBuildWidget(websiteUrl);
  }

  return (
    <section className={`${styles.hero} ${className || ''}`}>
      <div className={styles.container}>
        {/* Left column: headline + form */}
        <div className={styles.content} style={{
          transform: isVisible ? 'translateY(0)' : 'translateY(30px)',
          opacity: isVisible ? 1 : 0,
          transition: 'all 0.8s ease-out',
          zIndex: 1,
          position: 'relative'
        }}>
          <div style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 8,
            background: 'linear-gradient(135deg, hsl(215 80% 60% / 0.1) 0%, hsl(160 70% 45% / 0.1) 100%)',
            border: '1px solid hsl(215 80% 60% / 0.2)',
            borderRadius: 999,
            padding: '6px 16px',
            marginBottom: 24,
            fontSize: 14,
            fontWeight: 500,
            color: 'var(--primary)'
          }}>
            <Sparkles size={16} />
            <span>AI-Powered Chat Widget for Your Website</span>
          </div>

          <h1 className={styles.headline}>
            Turn Website Visitors
            <br />
            into Happy Customers
          </h1>
          <p className={styles.subheadline}>
            Add an AI chat widget to your website in 60 seconds. Answer questions, schedule meetings, and generate leads automatically.
          </p>

          {!isBuilding ? (
            <>
              <form onSubmit={onSubmit} style={{
                display: 'flex',
                gap: 12,
                alignItems: 'flex-start',
                marginTop: 32,
                flexWrap: 'wrap',
                flexDirection: 'column',
                maxWidth: 600
              }}>
                <div style={{
                  display: 'flex',
                  width: '100%',
                  gap: 12,
                  flexWrap: 'wrap'
                }}>
                  <input
                    aria-label="Website URL"
                    placeholder="https://yourwebsite.com"
                    value={websiteUrl}
                    onChange={(e) => {
                      setWebsiteUrl(e.target.value);
                      setError(null);
                    }}
                    style={{
                      padding: '14px 18px',
                      borderRadius: 12,
                      border: '2px solid var(--border)',
                      background: 'var(--surface)',
                      minWidth: 280,
                      flex: '1 1 300px',
                      fontSize: 15,
                      fontWeight: 500,
                      transition: 'all 0.2s ease',
                      outline: 'none'
                    }}
                    onFocus={(e) => {
                      e.currentTarget.style.borderColor = 'var(--primary)';
                      e.currentTarget.style.boxShadow = 'var(--shadow-focus)';
                    }}
                    onBlur={(e) => {
                      e.currentTarget.style.borderColor = 'var(--border)';
                      e.currentTarget.style.boxShadow = 'none';
                    }}
                  />
                  <Button type="submit" size="lg" style={{
                    padding: '14px 28px',
                    fontSize: 15,
                    fontWeight: 600,
                    borderRadius: 12,
                    background: 'linear-gradient(135deg, var(--primary) 0%, hsl(215 80% 50%) 100%)',
                    boxShadow: '0 4px 12px hsl(215 80% 60% / 0.3)',
                    transition: 'all 0.2s ease'
                  }}>
                    Build My Widget
                    <MoveRight size={18} style={{ marginLeft: 6 }} />
                  </Button>
                </div>
                <p style={{
                  fontSize: 13,
                  color: 'var(--muted-foreground)',
                  margin: 0
                }}>
                  Free forever • No credit card required • Setup in 60 seconds
                </p>
              </form>
              {error && (
                <div style={{
                  color: 'var(--error)',
                  marginTop: 16,
                  fontSize: 14,
                  padding: '12px 16px',
                  background: 'hsl(0 75% 55% / 0.1)',
                  borderRadius: 8,
                  border: '1px solid hsl(0 75% 55% / 0.2)'
                }}>
                  {error}
                </div>
              )}
            </>
          ) : (
            <div style={{
              marginTop: 32,
              padding: 24,
              background: 'var(--surface)',
              border: '2px solid var(--border)',
              borderRadius: 16,
              maxWidth: 600
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
                <Loader2 size={24} style={{ animation: 'spin 1s linear infinite', color: 'var(--primary)' }} />
                <div>
                  <h3 style={{ fontSize: 16, fontWeight: 700, margin: 0 }}>Building Your AI Widget</h3>
                  <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, marginTop: 4 }}>
                    {buildMessage}
                  </p>
                </div>
              </div>
              <div style={{
                width: '100%',
                height: 8,
                background: 'var(--muted)',
                borderRadius: 999,
                overflow: 'hidden'
              }}>
                <div style={{
                  width: `${buildProgress}%`,
                  height: '100%',
                  background: 'linear-gradient(90deg, var(--primary), var(--secondary))',
                  transition: 'width 0.3s ease',
                  borderRadius: 999
                }} />
              </div>
              <p style={{
                fontSize: 13,
                color: 'var(--muted-foreground)',
                marginTop: 12,
                textAlign: 'center'
              }}>
                {buildProgress}% complete
              </p>
            </div>
          )}

          <div className={styles.socialProof} style={{ marginTop: 48 }}>
            <div className={styles.avatars}>
              <Avatar style={{ width: 44, height: 44 }}>
                <AvatarImage src="https://randomuser.me/api/portraits/women/44.jpg" alt="User 1" />
                <AvatarFallback>U1</AvatarFallback>
              </Avatar>
              <Avatar style={{ width: 44, height: 44 }}>
                <AvatarImage src="https://randomuser.me/api/portraits/men/32.jpg" alt="User 2" />
                <AvatarFallback>U2</AvatarFallback>
              </Avatar>
              <Avatar style={{ width: 44, height: 44 }}>
                <AvatarImage src="https://randomuser.me/api/portraits/women/67.jpg" alt="User 3" />
                <AvatarFallback>U3</AvatarFallback>
              </Avatar>
              <Avatar style={{ width: 44, height: 44 }}>
                <AvatarImage src="https://randomuser.me/api/portraits/men/55.jpg" alt="User 4" />
                <AvatarFallback>U4</AvatarFallback>
              </Avatar>
            </div>
            <div>
              <p style={{ margin: 0, fontWeight: 700, fontSize: 15 }}>Trusted by 500+ businesses</p>
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: 4,
                marginTop: 6,
                fontSize: 13,
                color: 'var(--muted-foreground)'
              }}>
                {[...Array(5)].map((_, i) => (
                  <span key={i} style={{ color: '#fbbf24', fontSize: 16 }}>★</span>
                ))}
                <span style={{ marginLeft: 6, fontWeight: 600 }}>4.9/5 rating</span>
              </div>
            </div>
          </div>
        </div>

        {/* Right column: enhanced image with chat widget preview */}
        <div className={styles.imageContainer} style={{
          transform: isVisible ? 'translateY(0)' : 'translateY(30px)',
          opacity: isVisible ? 1 : 0,
          transition: 'all 0.8s ease-out 0.2s'
        }}>
          <div style={{
            borderRadius: 24,
            overflow: 'hidden',
            position: 'relative',
            background: 'linear-gradient(135deg, hsl(215 80% 60% / 0.08) 0%, hsl(160 70% 45% / 0.08) 100%)',
            padding: '3px',
            boxShadow: '0 20px 60px hsl(220 15% 15% / 0.15)'
          }}>
            <div style={{
              background: 'var(--surface)',
              borderRadius: 22,
              overflow: 'hidden',
              position: 'relative'
            }}>
              {/* Browser mockup */}
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
                  textAlign: 'center'
                }}>
                  yourwebsite.com
                </div>
              </div>

              <img
                src="https://images.unsplash.com/photo-1460925895917-afdab827c52f?w=800&h=600&fit=crop"
                alt="AI chat widget preview"
                className={styles.heroImage}
                style={{
                  display: 'block',
                  width: '100%',
                  height: 'auto',
                  objectFit: 'cover'
                }}
              />

              {/* Chat widget overlay */}
              <div style={{
                position: 'absolute',
                bottom: 24,
                right: 24,
                width: 320,
                background: 'var(--surface)',
                borderRadius: 16,
                boxShadow: '0 12px 40px hsl(220 15% 15% / 0.25)',
                border: '1px solid var(--border)',
                overflow: 'hidden',
                animation: 'float 3s ease-in-out infinite'
              }}>
                <div style={{
                  background: 'linear-gradient(135deg, var(--primary) 0%, hsl(215 80% 50%) 100%)',
                  padding: '16px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  color: 'white'
                }}>
                  <div style={{
                    width: 40,
                    height: 40,
                    borderRadius: '50%',
                    background: 'white',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    color: 'var(--primary)',
                    fontSize: 18
                  }}>
                    AI
                  </div>
                  <div>
                    <div style={{ fontWeight: 700, fontSize: 15 }}>Support Assistant</div>
                    <div style={{ fontSize: 12, opacity: 0.9 }}>Online • Instant replies</div>
                  </div>
                </div>
                <div style={{ padding: '16px' }}>
                  <div style={{
                    background: 'hsl(215 80% 60% / 0.1)',
                    padding: '12px 14px',
                    borderRadius: 12,
                    fontSize: 14,
                    lineHeight: 1.5
                  }}>
                    👋 Hi! I'm your AI assistant. I can help you with questions, schedule meetings, or raise a ticket. How can I help you today?
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Add animations */}
      <style>{`
        @keyframes float {
          0%, 100% { transform: translateY(0px); }
          50% { transform: translateY(-10px); }
        }
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </section>
  );
};