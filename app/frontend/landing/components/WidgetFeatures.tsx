import React from 'react';
import {
  MessageSquare,
  Calendar,
  Users,
  ClipboardList,
  Zap,
  Globe,
  BarChart3,
  Lock,
  Sparkles,
} from 'lucide-react';
import styles from './WidgetFeatures.module.css';

const features = [
  {
    icon: MessageSquare,
    title: 'Answer Questions',
    description:
      'AI-powered responses trained on your website content. Instant, accurate answers 24/7.',
    gradient: 'linear-gradient(135deg, hsl(215 80% 60%), hsl(215 90% 70%))',
  },
  {
    icon: Calendar,
    title: 'Schedule Meetings',
    description:
      'Let visitors book meetings directly through chat. Seamless calendar integration.',
    gradient: 'linear-gradient(135deg, hsl(260 75% 65%), hsl(260 85% 75%))',
  },
  {
    icon: Users,
    title: 'Generate Leads',
    description:
      'Capture contact information naturally during conversations. Build your pipeline effortlessly.',
    gradient: 'linear-gradient(135deg, hsl(160 70% 45%), hsl(160 80% 55%))',
  },
  {
    icon: ClipboardList,
    title: 'Raise Tickets',
    description:
      'Automatically create support tickets for complex issues. Track everything in one place.',
    gradient: 'linear-gradient(135deg, hsl(38 90% 55%), hsl(38 100% 65%))',
  },
  {
    icon: Zap,
    title: 'Lightning Fast',
    description:
      'Instant responses powered by AI. No waiting, no frustration. Pure speed.',
    gradient: 'linear-gradient(135deg, hsl(340 80% 60%), hsl(340 90% 70%))',
  },
  {
    icon: Globe,
    title: 'Works Everywhere',
    description:
      'Mobile, desktop, any browser. One widget, all platforms. Universal compatibility.',
    gradient: 'linear-gradient(135deg, hsl(190 80% 50%), hsl(190 90% 60%))',
  },
  {
    icon: BarChart3,
    title: 'Analytics & Insights',
    description:
      'Track conversations, engagement, and conversions. Data-driven improvements.',
    gradient: 'linear-gradient(135deg, hsl(280 70% 60%), hsl(280 80% 70%))',
  },
  {
    icon: Lock,
    title: 'Secure & Private',
    description:
      'Enterprise-grade security. GDPR compliant. Your data stays protected.',
    gradient: 'linear-gradient(135deg, hsl(145 65% 45%), hsl(145 75% 55%))',
  },
  {
    icon: Sparkles,
    title: 'Smart Customization',
    description:
      'Automatically matches your brand colors. Or customize every detail manually.',
    gradient: 'linear-gradient(135deg, hsl(50 90% 55%), hsl(50 100% 65%))',
  },
];

export const WidgetFeatures = () => {
  return (
    <section className={styles.features}>
      <div className={styles.container}>
        {/* Section Header */}
        <div className={styles.header}>
          <div className={styles.badge}>
            <Sparkles size={16} />
            <span>Features</span>
          </div>
          <h2 className={styles.title}>
            Everything you need to
            <br />
            <span className={styles.gradientText}>delight customers</span>
          </h2>
          <p className={styles.subtitle}>
            Powerful features designed to help you convert visitors and provide amazing support
          </p>
        </div>

        {/* Features Grid */}
        <div className={styles.grid}>
          {features.map((feature, index) => {
            const Icon = feature.icon;
            return (
              <div
                key={index}
                className={styles.card}
                style={{
                  animationDelay: `${index * 0.1}s`,
                }}
              >
                <div
                  className={styles.iconWrapper}
                  style={{ background: feature.gradient }}
                >
                  <Icon size={24} />
                </div>
                <h3 className={styles.cardTitle}>{feature.title}</h3>
                <p className={styles.cardDescription}>{feature.description}</p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
};
