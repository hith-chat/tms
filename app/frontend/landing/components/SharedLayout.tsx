import React from 'react';
import { Link, NavLink } from 'react-router-dom';
import { MessageSquare, Ticket, LogIn } from 'lucide-react';
import { Button } from './Button';
import styles from './SharedLayout.module.css';

interface SharedLayoutProps {
  children: React.ReactNode;
}

export const SharedLayout = ({ children }: SharedLayoutProps) => {
  return (
    <div className={styles.layout}>
      <header className={styles.header}>
        <div className={styles.headerContent}>
          <Link to="/" className={styles.logo}>
            <img src="/assets/hith-logo-expanded-filled.svg" alt="Hith logo" width={120} height={32} />
          </Link>
          
          {/* Navigation and Sign In */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            <Button asChild variant="ghost" size="sm">
              <a href="https://app.hith.chat/login" target="_blank" rel="noreferrer" style={{
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                textDecoration: 'none',
                color: 'inherit'
              }}>
                <LogIn size={16} />
                Sign In
              </a>
            </Button>
          </div>
        </div>
      </header>
      <main className={styles.main}>{children}</main>
    </div>
  );
};