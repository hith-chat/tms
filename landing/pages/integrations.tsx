import React from "react";
import { Button } from "../components/Button";
import { Footer } from "../components/Footer";
import { 
  MessageSquare, 
  Bot, 
  Mail, 
  Slack, 
  Bell, 
  Globe,
  Zap,
  Check,
  ArrowRight,
  ExternalLink,
  Puzzle,
  Webhook,
  Shield,
  Clock,
} from "lucide-react";
import { Link } from "react-router-dom";

const integrations = [
  {
    category: "Support Platforms",
    description: "Connect with your existing support infrastructure",
    items: [
      {
        name: "Native Hith Support",
        description: "Full-featured native support system with AI chat, ticket management, and knowledge base",
        icon: <MessageSquare size={32} />,
        features: ["AI-powered responses", "Ticket automation", "Knowledge base", "Team collaboration"],
        status: "native",
        gradient: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
      },
      {
        name: "Zendesk Integration",
        description: "Seamlessly sync tickets, customer data, and conversations with your Zendesk instance",
        icon: <Globe size={32} />,
        features: ["Bi-directional sync", "Custom field mapping", "Automated workflows", "Real-time updates"],
        status: "available",
        gradient: "linear-gradient(135deg, #f093fb 0%, #f5576c 100%)"
      },
      {
        name: "Freshdesk Integration",
        description: "Connect with Freshdesk to unify your customer support across platforms",
        icon: <Puzzle size={32} />,
        features: ["Ticket synchronization", "Agent productivity tools", "Custom automations", "Reporting integration"],
        status: "available",
        gradient: "linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)"
      }
    ]
  },
  {
    category: "Notification Channels",
    description: "Get notified through your preferred channels",
    items: [
      {
        name: "Email Notifications",
        description: "Customizable email alerts for tickets, mentions, and important updates",
        icon: <Mail size={32} />,
        features: ["Custom templates", "Priority-based routing", "Digest emails", "Attachment support"],
        status: "available",
        gradient: "linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)"
      },
      {
        name: "Slack Integration",
        description: "Receive notifications and manage tickets directly from Slack channels",
        icon: <Slack size={32} />,
        features: ["Channel notifications", "Ticket creation", "Status updates", "Team mentions"],
        status: "available",
        gradient: "linear-gradient(135deg, #fa709a 0%, #fee140 100%)"
      },
      {
        name: "Browser Notifications",
        description: "Real-time browser notifications for urgent tickets and mentions",
        icon: <Bell size={32} />,
        features: ["Real-time alerts", "Priority filtering", "Sound notifications", "Desktop badges"],
        status: "available",
        gradient: "linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)"
      }
    ]
  },
  {
    category: "Developer Tools",
    description: "Extend functionality with APIs and webhooks",
    items: [
      {
        name: "REST API",
        description: "Full REST API access for custom integrations and automations",
        icon: <Webhook size={32} />,
        features: ["Complete API coverage", "Rate limiting", "Authentication", "Comprehensive docs"],
        status: "available",
        gradient: "linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)"
      },
      {
        name: "Webhooks",
        description: "Real-time event notifications to your custom endpoints",
        icon: <Zap size={32} />,
        features: ["Real-time events", "Custom payloads", "Retry logic", "Security headers"],
        status: "available",
        gradient: "linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)"
      }
    ]
  }
];

const statusConfig = {
  native: { label: "Native", color: "#10b981", bg: "color-mix(in srgb, #10b981 15%, transparent)" },
  available: { label: "Available", color: "#3b82f6", bg: "color-mix(in srgb, #3b82f6 15%, transparent)" },
  coming: { label: "Coming Soon", color: "#f59e0b", bg: "color-mix(in srgb, #f59e0b 15%, transparent)" }
};

export default function IntegrationsPage() {
  return (
    <>
      <div style={{ padding: 20, display: 'flex', justifyContent: 'center' }}>
        <div style={{ maxWidth: 1200, width: '100%' }}>
          {/* Header */}
          <div style={{ textAlign: 'center', marginBottom: 64 }}>
            <div style={{ 
              background: 'var(--primary)', 
              color: 'var(--primary-foreground)', 
              padding: 16, 
              borderRadius: 12,
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              marginBottom: 24
            }}>
              <Puzzle size={32} />
            </div>
            <h1 style={{ fontSize: 48, fontWeight: 800, margin: 0, marginBottom: 16 }}>
              Powerful Integrations
            </h1>
            <p style={{ fontSize: 20, color: 'var(--muted-foreground)', lineHeight: 1.6, maxWidth: 700, margin: '0 auto' }}>
              Connect Hith with your existing tools and workflows. From support platforms to notification channels, 
              we integrate with the tools your team already loves.
            </p>
          </div>

          {/* Integration Categories */}
          {integrations.map((category, categoryIndex) => (
            <div key={categoryIndex} style={{ marginBottom: 64 }}>
              {/* Category Header */}
              <div style={{ marginBottom: 32 }}>
                <h2 style={{ fontSize: 32, fontWeight: 700, margin: 0, marginBottom: 8 }}>
                  {category.category}
                </h2>
                <p style={{ fontSize: 18, color: 'var(--muted-foreground)', margin: 0 }}>
                  {category.description}
                </p>
              </div>

              {/* Integration Cards */}
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', gap: 24 }}>
                {category.items.map((integration, index) => (
                  <div key={index} style={{
                    background: 'var(--surface)',
                    border: '1px solid var(--border)',
                    borderRadius: 16,
                    padding: 32,
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
                      width: '40%',
                      height: '100%',
                      background: integration.gradient,
                      opacity: 0.1,
                      pointerEvents: 'none'
                    }}></div>
                    
                    <div style={{ position: 'relative', zIndex: 1 }}>
                      {/* Header */}
                      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 16 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                          <div style={{
                            background: integration.gradient,
                            padding: 12,
                            borderRadius: 12,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            color: 'white'
                          }}>
                            {integration.icon}
                          </div>
                          <div>
                            <h3 style={{ fontSize: 20, fontWeight: 700, margin: 0, marginBottom: 4 }}>
                              {integration.name}
                            </h3>
                            <div style={{
                              background: statusConfig[integration.status as keyof typeof statusConfig].bg,
                              color: statusConfig[integration.status as keyof typeof statusConfig].color,
                              padding: '4px 12px',
                              borderRadius: 6,
                              fontSize: 12,
                              fontWeight: 600,
                              display: 'inline-block'
                            }}>
                              {statusConfig[integration.status as keyof typeof statusConfig].label}
                            </div>
                          </div>
                        </div>
                      </div>

                      {/* Description */}
                      <p style={{ fontSize: 14, color: 'var(--muted-foreground)', margin: 0, marginBottom: 20, lineHeight: 1.6 }}>
                        {integration.description}
                      </p>

                      {/* Features */}
                      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 8, marginBottom: 20 }}>
                        {integration.features.map((feature, idx) => (
                          <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <Check size={16} style={{ color: 'var(--primary)', flexShrink: 0 }} />
                            <span style={{ fontSize: 13, color: 'var(--foreground)' }}>{feature}</span>
                          </div>
                        ))}
                      </div>

                      {/* Action */}
                      {/* <Button 
                        variant="outline" 
                        size="sm"
                        style={{ 
                          display: 'flex', 
                          alignItems: 'center', 
                          gap: 8,
                          fontSize: 14,
                          fontWeight: 600
                        }}
                      >
                        {integration.status === 'native' ? 'Learn More' : 'Learn More'}
                        <ExternalLink size={14} />
                      </Button> */}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}

          {/* Security & Compliance */}
          <div style={{
            background: 'linear-gradient(135deg, var(--surface) 0%, color-mix(in srgb, var(--primary) 3%, var(--surface)) 100%)',
            border: '1px solid var(--border)',
            borderRadius: 20,
            padding: 40,
            marginBottom: 48,
            textAlign: 'center'
          }}>
            <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 20 }}>
              <div style={{
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                padding: 16,
                borderRadius: 12,
                color: 'white'
              }}>
                <Shield size={32} />
              </div>
            </div>
            
            <h3 style={{ fontSize: 28, fontWeight: 800, margin: 0, marginBottom: 12 }}>
              Enterprise-Grade Security
            </h3>
            <p style={{ fontSize: 16, color: 'var(--muted-foreground)', maxWidth: 600, margin: '0 auto 24px' }}>
              All integrations are built with security and compliance in mind. We maintain SOC2 compliance 
              and ensure your data remains protected across all connected platforms.
            </p>
            
            <div style={{ display: 'flex', justifyContent: 'center', gap: 32, flexWrap: 'wrap' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Check size={20} style={{ color: 'var(--primary)' }} />
                <span style={{ fontWeight: 600 }}>SOC2 Compliant</span>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Check size={20} style={{ color: 'var(--primary)' }} />
                <span style={{ fontWeight: 600 }}>End-to-End Encryption</span>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Check size={20} style={{ color: 'var(--primary)' }} />
                <span style={{ fontWeight: 600 }}>GDPR Compliant</span>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Check size={20} style={{ color: 'var(--primary)' }} />
                <span style={{ fontWeight: 600 }}>99.9% Uptime SLA</span>
              </div>
            </div>
          </div>

          {/* CTA Section */}
          <div style={{ textAlign: 'center' }}>
            <h3 style={{ fontSize: 32, fontWeight: 800, margin: 0, marginBottom: 16 }}>
              Ready to connect your tools?
            </h3>
            <p style={{ fontSize: 18, color: 'var(--muted-foreground)', maxWidth: 600, margin: '0 auto 32px' }}>
              Start integrating Hith with your existing workflow today. Our team is here to help with setup and configuration.
            </p>
            <div style={{ display: 'flex', gap: 16, justifyContent: 'center', flexWrap: 'wrap' }}>
              <Button size="lg" style={{ padding: '12px 32px', fontSize: 16, fontWeight: 700 }}>
                
                <Link to="https://app.hith.chat/signup" target='_blank'>Start Integration</Link>
                <ArrowRight size={20} style={{ marginLeft: 8 }} />
              </Button>
              <Button variant="outline" size="lg" style={{ padding: '12px 32px', fontSize: 16, fontWeight: 700 }}>
                <Link to="https://calendly.com/sumansaurabh-1/hith" target='_blank'>Contact Sales</Link>
              </Button>
            </div>
          </div>
        </div>
      </div>
      
      <Footer />
    </>
  );
}