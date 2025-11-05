import React from 'react';
import { 
  MessageSquare, 
  Bot, 
  BookOpen, 
  Workflow, 
  BarChart3, 
  Users, 
  Zap,
  Shield,
  Clock,
  Bell,
  Search,
  Globe,
  Check,
  ArrowRight,
  Mail,
  Slack,
  Puzzle
} from 'lucide-react';
import { Button } from './Button';
import { Link } from 'react-router-dom';
import styles from './FeaturesSection.module.css';

const mainFeatures = [
  {
    icon: <Bot size={32} />,
    title: 'AI-Powered Conversations',
    description: 'Intelligent chatbot trained on your website content, providing accurate answers instantly.',
    highlights: ['Instant AI responses', 'Trained on your content', 'Natural conversations', '24/7 availability'],
    gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
  },
  {
    icon: <Zap size={32} />,
    title: 'Setup in 60 Seconds',
    description: 'No coding required. Just enter your website URL and get a ready-to-use chat widget.',
    highlights: ['One-click setup', 'Auto theme matching', 'Instant deployment', 'No technical skills needed'],
    gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)'
  },
  {
    icon: <Users size={32} />,
    title: 'Lead Generation',
    description: 'Capture visitor information naturally during conversations and build your pipeline.',
    highlights: ['Smart lead capture', 'Email collection', 'Contact forms', 'CRM integration ready'],
    gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)'
  },
  {
    icon: <MessageSquare size={32} />,
    title: 'Smart Ticketing',
    description: 'Automatically create support tickets for complex issues that need human attention.',
    highlights: ['Auto ticket creation', 'Priority routing', 'Team collaboration', 'Track everything'],
    gradient: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)'
  },
  {
    icon: <BookOpen size={32} />,
    title: 'Meeting Scheduling',
    description: 'Let visitors book meetings directly through chat with seamless calendar integration.',
    highlights: ['Calendar sync', 'Automatic booking', 'Time zone handling', 'Confirmation emails'],
    gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)'
  },
  {
    icon: <BarChart3 size={32} />,
    title: 'Analytics & Insights',
    description: 'Track conversations, measure engagement, and understand what your visitors need.',
    highlights: ['Conversation analytics', 'Visitor insights', 'Performance metrics', 'Custom reports'],
    gradient: 'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)'
  }
];

const additionalFeatures = [
  { icon: <Workflow size={20} />, text: 'Automated responses based on your content' },
  { icon: <Shield size={20} />, text: 'Enterprise-grade security & privacy' },
  { icon: <Clock size={20} />, text: 'Works 24/7 without breaks' },
  { icon: <Search size={20} />, text: 'Smart search through conversations' },
  { icon: <Globe size={20} />, text: 'Mobile-friendly responsive design' },
  { icon: <Puzzle size={20} />, text: 'Easy integration with existing tools' }
];

const integrationsIcons = [
  { id: 'native', name: 'Native Hith', icon: <MessageSquare size={28} />, anchor: 'native' },
  { id: 'zendesk', name: 'Zendesk', icon: <Puzzle size={28} />, anchor: 'zendesk' },
  { id: 'freshdesk', name: 'Freshdesk', icon: <Globe size={28} />, anchor: 'freshdesk' },
  { id: 'slack', name: 'Slack', icon: <Slack size={28} />, anchor: 'slack' },
  { id: 'email', name: 'Email', icon: <Mail size={28} />, anchor: 'email' }
];

export const FeaturesSection = ({ className }: { className?: string }) => {
  return (
    <section id="features" className={`${styles.features} ${className || ''}`}>
      <div className={styles.container}>
        {/* Header */}
        <div className={styles.header}>
          <h2 className={styles.title}>
            Everything you need to engage and convert visitors
          </h2>
          <p className={styles.subtitle}>
            From AI-powered conversations to automatic lead capture, every feature is designed to help you grow your business.
          </p>
        </div>

        {/* Main Features Grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(380px, 1fr))', gap: 32, marginBottom: 48 }}>
          {mainFeatures.map((feature, index) => (
            <div key={index} style={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              borderRadius: 16,
              padding: 28,
              position: 'relative',
              overflow: 'hidden',
              transition: 'all 0.3s ease',
              cursor: 'pointer'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-4px)';
              e.currentTarget.style.boxShadow = '0 20px 40px rgba(16,24,40,0.15)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)';
              e.currentTarget.style.boxShadow = 'none';
            }}>
              {/* Gradient overlay */}
              <div style={{
                position: 'absolute',
                top: 0,
                right: 0,
                width: '50%',
                height: '100%',
                background: feature.gradient,
                opacity: 0.1,
                pointerEvents: 'none'
              }}></div>
              
              <div style={{ position: 'relative', zIndex: 1 }}>
                {/* Icon */}
                <div style={{
                  background: feature.gradient,
                  padding: 12,
                  borderRadius: 12,
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  marginBottom: 16,
                  color: 'white'
                }}>
                  {feature.icon}
                </div>

                {/* Content */}
                <h3 style={{ fontSize: 20, fontWeight: 700, margin: 0, marginBottom: 8 }}>
                  {feature.title}
                </h3>
                <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, marginBottom: 16, lineHeight: 1.6 }}>
                  {feature.description}
                </p>

                {/* Highlights */}
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 8 }}>
                  {feature.highlights.map((highlight, idx) => (
                    <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <Check size={16} style={{ color: 'var(--primary)', flexShrink: 0 }} />
                      <span style={{ fontSize: 13, color: 'var(--foreground)' }}>{highlight}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>


        {/* Integrations icons row (icon-first, SEO-friendly) */}
        <div style={{ display: 'flex', justifyContent: 'center', gap: 20, alignItems: 'center', marginBottom: 28, flexWrap: 'wrap' }}>
          {integrationsIcons.map((it) => (
            <a key={it.id} href={`/integrations#${it.anchor}`} title={it.name} style={{ textDecoration: 'none', color: 'inherit', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6 }}>
              <div style={{ width: 72, height: 72, borderRadius: 12, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'var(--surface)', border: '1px solid var(--border)' }}>
                {it.icon}
              </div>
              {/* visible label optional - keep small visually for clarity */}
              <span style={{ fontSize: 12, color: 'var(--muted-foreground)' }}>{it.name}</span>
              {/* hidden text for SEO and accessibility */}
              <span style={{position:'absolute',width:1,height:1,padding:0,margin:-1,overflow:'hidden',clip:'rect(0,0,0,0)',whiteSpace:'nowrap',border:0}}>{it.name}</span>
            </a>
          ))}
        </div>

        {/* Additional Features Grid */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
          gap: 16,
          marginTop: 32,
          marginBottom: 48
        }}>
          {additionalFeatures.map((feature, index) => (
            <div key={index} style={{
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              padding: '16px 20px',
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              borderRadius: 12,
              transition: 'all 0.2s ease'
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.borderColor = 'var(--primary)';
              e.currentTarget.style.transform = 'translateY(-2px)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.borderColor = 'var(--border)';
              e.currentTarget.style.transform = 'translateY(0)';
            }}>
              <div style={{
                color: 'var(--primary)',
                flexShrink: 0
              }}>
                {feature.icon}
              </div>
              <span style={{ fontSize: 14, fontWeight: 500 }}>{feature.text}</span>
            </div>
          ))}
        </div>

        {/* CTA Section */}
        <div style={{ textAlign: 'center' }}>
          <h3 style={{ fontSize: 28, fontWeight: 800, margin: 0, marginBottom: 12 }}>
            Ready to transform your website?
          </h3>
          <p style={{ fontSize: 18, color: 'var(--muted-foreground)', margin: 0, marginBottom: 24 }}>
            Join hundreds of businesses converting more visitors with AI-powered chat
          </p>
          <div style={{ display: 'flex', gap: 16, justifyContent: 'center', flexWrap: 'wrap' }}>
            <Button size="lg" style={{ padding: '14px 36px', fontSize: 16, fontWeight: 700, background: 'linear-gradient(135deg, var(--primary) 0%, hsl(215 80% 50%) 100%)', boxShadow: '0 4px 12px hsl(215 80% 60% / 0.3)' }}>
              <a href="#" onClick={(e) => { e.preventDefault(); window.scrollTo({ top: 0, behavior: 'smooth' }); }}>
                Build Your Widget Now
              </a>
              <ArrowRight size={20} style={{ marginLeft: 8 }} />
            </Button>
            <a href="https://calendly.com/sumansaurabh-1/hith" target="_blank" rel="noreferrer" style={{ textDecoration: 'none' }}>
              <Button variant="outline" size="lg" style={{ padding: '14px 36px', fontSize: 16, fontWeight: 700 }}>
                Book a Demo
              </Button>
            </a>
          </div>
        </div>
      </div>
    </section>
  );
};