import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Button } from './Button';
import { Badge } from './Badge';
import { Check, Zap, MessageSquare, Star, Bot, BookOpen, Users, BarChart3, Shield, Clock } from 'lucide-react';
import styles from './PricingSection.module.css';

export const PricingSection = ({ className }: { className?: string }) => {
  return (
    <section id="pricing" className={`${styles.pricing} ${className || ''}`}>
      <div className={styles.container}>
        <div className={styles.header}>
          <Badge variant="outline" className={styles.badge}>
            <Zap size={14} />
            Token-Based Pricing
          </Badge>
          <h2 className={styles.title}>
            Simple, Transparent Pricing
          </h2>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 24, alignItems: 'center', marginBottom: 32 }}>
          {/* New big slider section - slider value is thousands of tokens; step 5 -> 5k increments */}
          <div style={{ width: '100%', maxWidth: 920 }}>
            <div style={{ marginBottom: 12, color: 'var(--muted-foreground, #64748b)', textAlign: 'center' }}>
              Move the slider to select token amount. First <strong>5,000 tokens</strong> are free.
            </div>

            <BigTokenSlider />
          </div>
        </div>

        <div className={styles.creditInfo}>
          <div className={styles.creditCard}>
            <MessageSquare size={24} className={styles.creditIcon} />
            <div className={styles.creditDetails}>
              <h4>How Tokens Work</h4>
              <p>
                Tokens are used across all platform features. Based on industry-standard pricing (~$0.15-0.60 per 1M tokens). Short messages use ~100 tokens, longer conversations use more.
                Tokens never expire and can be used across tickets, chat, and automation features.
              </p>
            </div>
          </div>
        </div>

        <div className={styles.faq}>
          <h3>Frequently Asked Questions</h3>
          <div className={styles.faqGrid}>
            <div className={styles.faqItem}>
              <h4>How are tokens consumed?</h4>
              <p>Token usage varies by content length and complexity. A typical short message uses ~100 tokens, while longer conversations use more. Pricing follows industry standards (~$0.15-0.60 per 1M tokens).</p>
            </div>
            <div className={styles.faqItem}>
              <h4>Do tokens expire?</h4>
              <p>Tokens expire after 1 year. Purchase once and use them whenever you need them across all features.</p>
            </div>
            <div className={styles.faqItem}>
              <h4>Can I upgrade or add more tokens?</h4>
              <p>Absolutely! You can purchase additional token packs anytime or upgrade to a higher tier for better value.</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

function TokenSlider() {
  const [tokens, setTokens] = useState<number>(5000);
  const min = 0;
  const freeThreshold = 5000;
  const max = 2000000; // 2M tokens max for slider

  // simple pricing: $0 for first 5k, then $0.00002 per token (example)
  // we intentionally keep logic simple and proportional as requested
  function computePrice(t: number) {
    const billable = Math.max(0, t - freeThreshold);
    const pricePerToken = 0.00002; // $0.00002 per token => $20 per 1M tokens
    return +(billable * pricePerToken).toFixed(2);
  }

  function estimateMessages(t: number) {
    // assume ~100 tokens per short message
    const tokensPerMessage = 100;
    return Math.max(0, Math.round(t / tokensPerMessage));
  }

  const [theme, setTheme] = useState<'light' | 'dark'>('light');

  useEffect(() => {
    const doc = document.documentElement;
    const attr = doc.getAttribute('data-theme');
    setTheme(attr === 'dark' ? 'dark' : 'light');
  }, []);

  const price = computePrice(tokens);
  const messages = estimateMessages(tokens);

  const trackColor = theme === 'dark' ? '#2b6cb0' : '#0ea5a4';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div style={{ fontSize: 14 }}>
          <strong>{tokens.toLocaleString()}</strong> tokens
        </div>
        <div style={{ textAlign: 'right' }}>
          <div style={{ fontSize: 14 }}><strong>${price}</strong> estimated</div>
          <div style={{ fontSize: 12, color: 'var(--muted-foreground, #64748b)' }}>{messages.toLocaleString()} messages (estimate)</div>
        </div>
      </div>

      <input
        type="range"
        min={min}
        max={max}
        step={1000}
        value={tokens}
        onChange={(e) => setTokens(Number(e.target.value))}
        style={{
          width: '100%',
          appearance: 'none',
          height: 8,
          borderRadius: 9999,
          background: `linear-gradient(90deg, ${trackColor} ${(tokens / max) * 100}%, #e6e6e6 ${(tokens / max) * 100}%)`,
        }}
      />
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, color: 'var(--muted-foreground, #64748b)' }}>
        <span>$0</span>
        <span>{(max / 1000000).toFixed(0)}M tokens</span>
      </div>
    </div>
  );
}

function BigTokenSlider() {
  // slider steps in $5 price increments
  const minPrice = 0;
  const maxPrice = 100; // $100 max
  const priceStep = 5; // move in $5 increments
  const [price, setPrice] = useState<number>(0); // default $0 (free tier)

  function tokensFromPrice(priceAmount: number) {
    const freeThreshold = 5000;
    const pricePerToken = 0.00002; // $0.00002 per token => $20 per 1M tokens
    
    if (priceAmount === 0) {
      return freeThreshold; // free 5k tokens
    }
    
    // calculate tokens for the price: price / pricePerToken + free tokens
    const billableTokens = priceAmount / pricePerToken;
    return Math.round(billableTokens + freeThreshold);
  }

  function estimateMessages(tokensAmount: number) {
    const tokensPerMessage = 100; // simple proportional estimate
    return Math.max(0, Math.round(tokensAmount / tokensPerMessage));
  }

  const tokens = tokensFromPrice(price);
  const messages = estimateMessages(tokens);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
      {/* Slider Section */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ fontSize: 18, fontWeight: 700 }}>{tokens.toLocaleString()} tokens</div>
          <div style={{ textAlign: 'right' }}>
            <div style={{ fontSize: 18, fontWeight: 800 }}>${price}</div>
            <div style={{ fontSize: 13, color: 'var(--muted-foreground, #64748b)' }}>{messages.toLocaleString()} messages</div>
          </div>
        </div>

        <div style={{ padding: '8px 0' }}>
          <input
            aria-label="Price selector"
            type="range"
            min={minPrice}
            max={maxPrice}
            step={priceStep}
            value={price}
            onChange={(e) => setPrice(Number(e.target.value))}
            className={styles.tokenSlider}
            style={{
              background: `linear-gradient(90deg, var(--primary) ${(price / maxPrice) * 100}%, #e6e6e6 ${(price / maxPrice) * 100}%)`
            }}
          />
          <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 8, fontSize: 13, color: 'var(--muted-foreground, #64748b)' }}>
            <span>Free ($0)</span>
            <span>${maxPrice}</span>
          </div>
        </div>
      </div>

      {/* Enhanced Feature Card */}
      <div style={{ 
        background: 'linear-gradient(135deg, var(--surface) 0%, color-mix(in srgb, var(--primary) 5%, var(--surface)) 100%)', 
        border: '1px solid var(--border)', 
        padding: 32, 
        borderRadius: 16, 
        width: '100%', 
        boxShadow: '0 20px 40px rgba(16,24,40,0.12)', 
        position: 'relative',
        overflow: 'hidden'
      }}>
        {/* Decorative gradient overlay */}
        <div style={{ 
          position: 'absolute', 
          top: 0, 
          right: 0, 
          width: '50%', 
          height: '100%', 
          background: 'linear-gradient(135deg, transparent 0%, color-mix(in srgb, var(--primary) 8%, transparent) 100%)',
          pointerEvents: 'none'
        }}></div>
        
        <div style={{ position: 'relative', zIndex: 1 }}>
          {/* Header */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 24 }}>
            <div style={{ 
              background: 'var(--primary)', 
              color: 'var(--primary-foreground)', 
              padding: 12, 
              borderRadius: 12,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}>
              <Bot size={24} />
            </div>
            <div>
              <h3 style={{ fontSize: 24, fontWeight: 800, margin: 0, marginBottom: 4 }}>AI-Powered Platform</h3>
              <p style={{ fontSize: 16, color: 'var(--muted-foreground)', margin: 0 }}>
                Everything you need for intelligent customer support
              </p>
            </div>
          </div>

          {/* Features Grid */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 20, marginBottom: 28 }}>
            <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
              <div style={{ background: 'color-mix(in srgb, var(--primary) 15%, transparent)', padding: 8, borderRadius: 8 }}>
                <MessageSquare size={20} style={{ color: 'var(--primary)' }} />
              </div>
              <div>
                <h4 style={{ fontSize: 16, fontWeight: 700, margin: 0, marginBottom: 4 }}>Smart Chat Support</h4>
                <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, lineHeight: 1.5 }}>
                  AI-powered responses with {messages.toLocaleString()} message capacity
                </p>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
              <div style={{ background: 'color-mix(in srgb, var(--primary) 15%, transparent)', padding: 8, borderRadius: 8 }}>
                <BookOpen size={20} style={{ color: 'var(--primary)' }} />
              </div>
              <div>
                <h4 style={{ fontSize: 16, fontWeight: 700, margin: 0, marginBottom: 4 }}>Knowledge Base Creation</h4>
                <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, lineHeight: 1.5 }}>
                  Build comprehensive help docs & FAQs
                </p>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
              <div style={{ background: 'color-mix(in srgb, var(--primary) 15%, transparent)', padding: 8, borderRadius: 8 }}>
                <Users size={20} style={{ color: 'var(--primary)' }} />
              </div>
              <div>
                <h4 style={{ fontSize: 16, fontWeight: 700, margin: 0, marginBottom: 4 }}>Team Collaboration</h4>
                <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, lineHeight: 1.5 }}>
                  Unlimited agents & ticket management
                </p>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
              <div style={{ background: 'color-mix(in srgb, var(--primary) 15%, transparent)', padding: 8, borderRadius: 8 }}>
                <BarChart3 size={20} style={{ color: 'var(--primary)' }} />
              </div>
              <div>
                <h4 style={{ fontSize: 16, fontWeight: 700, margin: 0, marginBottom: 4 }}>Advanced Analytics</h4>
                <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, lineHeight: 1.5 }}>
                  Performance insights & reporting
                </p>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
              <div style={{ background: 'color-mix(in srgb, var(--primary) 15%, transparent)', padding: 8, borderRadius: 8 }}>
                <Shield size={20} style={{ color: 'var(--primary)' }} />
              </div>
              <div>
                <h4 style={{ fontSize: 16, fontWeight: 700, margin: 0, marginBottom: 4 }}>Enterprise Security</h4>
                <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, lineHeight: 1.5 }}>
                  SOC2 compliance & data protection
                </p>
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
              <div style={{ background: 'color-mix(in srgb, var(--primary) 15%, transparent)', padding: 8, borderRadius: 8 }}>
                <Clock size={20} style={{ color: 'var(--primary)' }} />
              </div>
              <div>
                <h4 style={{ fontSize: 16, fontWeight: 700, margin: 0, marginBottom: 4 }}>24/7 Automation</h4>
                <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, lineHeight: 1.5 }}>
                  Smart routing & auto-responses
                </p>
              </div>
            </div>
          </div>

          {/* Pricing Summary & CTA */}
          <div style={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center', 
            padding: 20, 
            background: 'var(--surface)', 
            borderRadius: 12,
            border: '1px solid var(--border)'
          }}>
            <div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 4 }}>
                <span style={{ fontSize: 36, fontWeight: 900, color: 'var(--primary)' }}>${price}</span>
                <span style={{ fontSize: 16, color: 'var(--muted-foreground)' }}>for {tokens.toLocaleString()} tokens</span>
              </div>
              <div style={{ fontSize: 14, color: 'var(--muted-foreground)' }}>
                {price === 0 ? 'Free tier includes all features' : `Supports ${messages.toLocaleString()} messages • No expiry • Pay as you scale`}
              </div>
            </div>
            <Button size="lg" style={{ padding: '12px 32px', fontSize: 16, fontWeight: 700 }}>
              <Link to="https://app.hith.chat/signup" target='_blank'>Get Started</Link>
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}