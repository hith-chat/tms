import React from 'react';
import { Link } from 'react-router-dom';
import { MessageSquare, Ticket } from 'lucide-react';
import styles from './Footer.module.css';

export const Footer = ({ className }: { className?: string }) => {
  return (
    <footer className={`${styles.footer} ${className || ''}`}>
      <div className={styles.container}>
        <div className={styles.main}>
          <div className={styles.brand}>
            <Link to="/" className={styles.logo}>
              <img src="/assets/hith-logo-expanded-filled-with-slogan.svg" alt="hith logo" width={200} height={100} />
            </Link>
          </div>
          <div className={styles.linksGrid}>
            <div className={styles.linkColumn}>
              <h4>Product</h4>
              <ul>
                <li><a href="/#features">Features</a></li>
                <li><a href="/#pricing">Pricing</a></li>
                <li><Link to="/integrations">Integrations</Link></li>
                <li><a href="https://calendly.com/sumansaurabh-1/hith" target="_blank" rel="noreferrer">Book a Demo</a></li>
              </ul>
            </div>
            <div className={styles.linkColumn}>
              <h4>Company</h4>
              <ul>
                <li><Link to="/about">About Us</Link></li>
                {/* <li><Link to="/careers">Careers</Link></li> */}
                {/* <li><Link to="/press">Press</Link></li> */}
                <li><Link to="/contact">Contact</Link></li>
              </ul>
            </div>
            
            <div className={styles.linkColumn}>
              <h4>Legal</h4>
              <ul>
                <li><Link to="/privacy">Privacy Policy</Link></li>
                <li><Link to="/terms">Terms of Service</Link></li>
              </ul>
            </div>
          </div>
        </div>
        <div className={styles.bottom}>
          <p>&copy; {new Date().getFullYear()} Penify Technologies Pvt Ltd. All rights reserved.</p>
          {/* Social media links can be added here */}
        </div>
      </div>
    </footer>
  );
};