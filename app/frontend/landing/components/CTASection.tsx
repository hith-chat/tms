import React from 'react';
import { Link } from 'react-router-dom';
import { Button } from './Button';
import styles from './CTASection.module.css';

export const CTASection = ({ className }: { className?: string }) => {
  return (
    <section className={`${styles.cta} ${className || ''}`}>
      <div className={styles.container}>
        <div className={styles.content}>
          <h2 className={styles.title}>Ready to transform your customer support?</h2>
          <p className={styles.subtitle}>
            Join hundreds of businesses already delighting their customers. No credit card required.
          </p>
          <div className={styles.actions}>
            <Button asChild size="lg">
              <Link to="https://app.hith.chat/signup" target='_blank'>Get Started</Link>
            </Button>
            <Button asChild variant="outline" size="lg" className={styles.outlineButton}>
              <Link target='_blank' to="https://calendly.com/sumansaurabh-1/hith?month=2025-09">Talk to Sales</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
};