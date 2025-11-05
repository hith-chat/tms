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
    title: 'AI-Powered Chat Widget',
    description: 'Intelligent chat widget that provides instant responses.',
    highlights: ['Instant AI responses', 'Natural language processing', 'Custom training on your data', '24/7 availability'],
    gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
  },
  {
    icon: <MessageSquare size={32} />,
    title: 'Multi-Platform Support',
    description: 'Unified ticket management across platforms.',
    highlights: ['Native Hith support', 'Zendesk integration', 'Freshdesk integration', 'Unified dashboard'],
    gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)'
  },
  {
    icon: <Bell size={32} />,
    title: 'Smart Notifications',
    description: 'Get notified instantly when tickets need attention.',
    highlights: ['Email alerts', 'Slack integration', 'Browser notifications', 'Custom notification rules'],
    gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)'
  },
  {
    icon: <BookOpen size={32} />,
    title: 'Knowledge Base Creation',
    description: 'Build comprehensive help documentation with AI-Agent assistance.',
    highlights: ['AI-assisted writing', 'Smart categorization', 'Search optimization', 'Version control'],
    gradient: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)'
  },
  // {
  //   icon: <BarChart3 size={32} />,
  //   title: 'Advanced Analytics',
  //   description: 'Deep insights into support performance, customer satisfaction, and team productivity.',
  //   highlights: ['Real-time dashboards', 'Performance metrics', 'Customer satisfaction tracking', 'Custom reports'],
  //   gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)'
  // },
  // {
  //   icon: <Shield size={32} />,
  //   title: 'Enterprise Security',
  //   description: 'SOC2 compliant platform with enterprise-grade security and data protection.',
  //   highlights: ['SOC2 compliance', 'Data encryption', 'Role-based access', 'Audit logs'],
  //   gradient: 'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)'
  // }
];

const additionalFeatures = [
  { icon: <Workflow size={20} />, text: 'Automated ticket routing and prioritization' },
  { icon: <Users size={20} />, text: 'Team collaboration with internal notes' },
  { icon: <Clock size={20} />, text: 'SLA tracking and deadline management' },
  { icon: <Search size={20} />, text: 'Advanced search and filtering' },
  { icon: <Globe size={20} />, text: 'Multi-language support' },
  { icon: <Zap size={20} />, text: 'Custom automations and workflows' }
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
            Everything you need for world-class customer support
          </h2>
          <p className={styles.subtitle}>
            From intelligent chat widgets to comprehensive ticket management, every feature is designed to scale with your business.
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

        {/* CTA Section */}
        <div style={{ textAlign: 'center' }}>
          <h3 style={{ fontSize: 28, fontWeight: 800, margin: 0, marginBottom: 12 }}>
            Ready to transform your customer support?
          </h3>
          <p style={{ fontSize: 18, color: 'var(--muted-foreground)', margin: 0, marginBottom: 24 }}>
            Join hundreds of companies delivering exceptional customer experiences with Hith
          </p>
          <div style={{ display: 'flex', gap: 16, justifyContent: 'center', flexWrap: 'wrap' }}>
            <Button size="lg" style={{ padding: '12px 32px', fontSize: 16, fontWeight: 700 }}>
              <Link to="https://app.hith.chat/signup" target='_blank'>Get Started</Link>
              <ArrowRight size={20} style={{ marginLeft: 8 }} />
            </Button>
            <Link to="/integrations" style={{ textDecoration: 'none' }}>
              <Button variant="outline" size="lg" style={{ padding: '12px 32px', fontSize: 16, fontWeight: 700 }}>
                View Integrations
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
};