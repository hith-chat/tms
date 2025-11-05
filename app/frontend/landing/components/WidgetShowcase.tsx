import React, { useState } from 'react';
import { Code, Palette, Zap, Check } from 'lucide-react';
import styles from './WidgetShowcase.module.css';

const showcaseItems = [
  {
    id: 'easy-setup',
    icon: Zap,
    title: 'Easy Setup',
    description: 'Just paste one line of code before </body> tag',
    features: [
      'No coding required',
      'Works on any platform',
      'Instant activation',
      'Zero configuration',
    ],
    demoImage: 'https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=800&h=600&fit=crop&q=80',
  },
  {
    id: 'smart-design',
    icon: Palette,
    title: 'Smart Design',
    description: 'Automatically adapts to your brand',
    features: [
      'Auto color detection',
      'Custom branding',
      'Responsive layout',
      'Dark mode support',
    ],
    demoImage: 'https://images.unsplash.com/photo-1561070791-2526d30994b5?w=800&h=600&fit=crop&q=80',
  },
  {
    id: 'powerful-ai',
    icon: Code,
    title: 'Powerful AI',
    description: 'Trained on your website content',
    features: [
      'Contextual responses',
      'Multi-language support',
      'Learning system',
      'Custom FAQs',
    ],
    demoImage: 'https://images.unsplash.com/photo-1677442136019-21780ecad995?w=800&h=600&fit=crop&q=80',
  },
];

export const WidgetShowcase = () => {
  const [activeTab, setActiveTab] = useState('easy-setup');

  const activeItem = showcaseItems.find((item) => item.id === activeTab);

  return (
    <section className={styles.showcase}>
      <div className={styles.container}>
        {/* Tabs */}
        <div className={styles.tabs}>
          {showcaseItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;

            return (
              <button
                key={item.id}
                className={`${styles.tab} ${isActive ? styles.tabActive : ''}`}
                onClick={() => setActiveTab(item.id)}
              >
                <Icon size={20} />
                <span>{item.title}</span>
              </button>
            );
          })}
        </div>

        {/* Content */}
        {activeItem && (
          <div className={styles.content}>
            <div className={styles.contentLeft}>
              <h3 className={styles.contentTitle}>{activeItem.title}</h3>
              <p className={styles.contentDescription}>{activeItem.description}</p>

              <ul className={styles.featureList}>
                {activeItem.features.map((feature, index) => (
                  <li key={index} className={styles.featureItem}>
                    <Check size={20} className={styles.checkIcon} />
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>
            </div>

            <div className={styles.contentRight}>
              <div className={styles.demoCard}>
                <div className={styles.demoImageWrapper}>
                  <img
                    src={activeItem.demoImage}
                    alt={activeItem.title}
                    className={styles.demoImage}
                  />
                  <div className={styles.demoOverlay}>
                    <div className={styles.demoChatWidget}>
                      <div className={styles.demoChatHeader}>
                        <div className={styles.demoChatAvatar}>AI</div>
                        <div>
                          <div className={styles.demoChatName}>AI Assistant</div>
                          <div className={styles.demoChatStatus}>Online</div>
                        </div>
                      </div>
                      <div className={styles.demoChatMessage}>
                        <div className={styles.demoChatBubble}>
                          Hi! How can I help you today? 👋
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </section>
  );
};
