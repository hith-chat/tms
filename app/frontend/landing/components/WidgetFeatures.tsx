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
      'AI-powered responses trained on your website content. Instant answers 24/7.',
  },
  {
    icon: Calendar,
    title: 'Schedule Meetings',
    description:
      'Let visitors book meetings directly through chat. Seamless calendar integration.',
  },
  {
    icon: Users,
    title: 'Generate Leads',
    description:
      'Capture contact information naturally during conversations.',
  },
  {
    icon: ClipboardList,
    title: 'Raise Tickets',
    description:
      'Automatically create support tickets for complex issues.',
  },
  {
    icon: Zap,
    title: 'Lightning Fast',
    description:
      'Instant responses powered by AI. No waiting.',
  },
  {
    icon: Globe,
    title: 'Works Everywhere',
    description:
      'Mobile, desktop, any browser. Universal compatibility.',
  },
];

export const WidgetFeatures = () => {
  return (
    <section className={styles.features}>
      <div className={styles.container}>
        <div className={styles.header}>
          <h2 className={styles.title}>
            Everything you need to delight customers
          </h2>
          <p className={styles.subtitle}>
            Powerful features to help you convert visitors and provide support
          </p>
        </div>

        <div className={styles.grid}>
          {features.map((feature, index) => {
            const Icon = feature.icon;
            return (
              <div key={index} className={styles.card}>
                <div className={styles.iconWrapper}>
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
