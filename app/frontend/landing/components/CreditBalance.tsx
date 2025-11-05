import React from 'react';
import { Link } from 'react-router-dom';
import { Button } from './Button';
import { Badge } from './Badge';
import { Zap, Plus, TrendingUp } from 'lucide-react';
import styles from './CreditBalance.module.css';

interface CreditBalanceProps {
  credits: number;
  variant?: 'compact' | 'detailed';
  showAddButton?: boolean;
  className?: string;
}

export const CreditBalance = ({
  credits,
  variant = 'compact',
  showAddButton = true,
  className
}: CreditBalanceProps) => {
  const formatCredits = (credits: number) => {
    if (credits >= 1000000) {
      return `${(credits / 1000000).toFixed(1)}M`;
    }
    if (credits >= 1000) {
      return `${(credits / 1000).toFixed(1)}K`;
    }
    return credits.toString();
  };

  const getCreditStatus = (credits: number) => {
    if (credits < 100) return { status: 'low', color: 'destructive', message: 'Low credits' };
    if (credits < 500) return { status: 'medium', color: 'warning', message: 'Running low' };
    return { status: 'good', color: 'success', message: 'Credits available' };
  };

  const creditStatus = getCreditStatus(credits);

  if (variant === 'compact') {
    return (
      <div className={`${styles.creditBalance} ${styles.compact} ${className || ''}`}>
        <div className={styles.creditInfo}>
          <Zap size={16} className={styles.creditIcon} />
          <span className={styles.creditCount}>{formatCredits(credits)}</span>
          <span className={styles.creditLabel}>credits</span>
        </div>
        {showAddButton && (
          <Button asChild size="sm" variant="outline" className={styles.addButton}>
            <Link to="/pricing">
              <Plus size={14} />
              Add
            </Link>
          </Button>
        )}
      </div>
    );
  }

  return (
    <div className={`${styles.creditBalance} ${styles.detailed} ${className || ''}`}>
      <div className={styles.header}>
        <div className={styles.titleRow}>
          <h3 className={styles.title}>
            <Zap size={20} className={styles.titleIcon} />
            Credit Balance
          </h3>
          <Badge variant={creditStatus.color as any} className={styles.statusBadge}>
            {creditStatus.message}
          </Badge>
        </div>
        <div className={styles.balanceDisplay}>
          <span className={styles.mainBalance}>{credits.toLocaleString()}</span>
          <span className={styles.balanceLabel}>credits available</span>
        </div>
      </div>

      <div className={styles.usageInfo}>
        <div className={styles.usageRow}>
          <span className={styles.usageLabel}>Cost per message:</span>
          <span className={styles.usageValue}>1 credit</span>
        </div>
        <div className={styles.usageRow}>
          <span className={styles.usageLabel}>Estimated messages:</span>
          <span className={styles.usageValue}>{credits.toLocaleString()} messages</span>
        </div>
      </div>

      <div className={styles.actions}>
        <Button asChild className={styles.primaryButton}>
          <Link to="/pricing">
            <Plus size={16} />
            Buy More Credits
          </Link>
        </Button>
        <Button asChild variant="outline" className={styles.secondaryButton}>
          <Link to="/usage">
            <TrendingUp size={16} />
            View Usage
          </Link>
        </Button>
      </div>

      <div className={styles.valueProposition}>
        <p>💡 <strong>Best Value:</strong> Get 5000 credits for just $45 - save $5!</p>
      </div>
    </div>
  );
};