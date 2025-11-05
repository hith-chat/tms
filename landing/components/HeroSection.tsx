import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from './Button';
import { Avatar, AvatarImage, AvatarFallback } from './Avatar';
import { MoveRight } from 'lucide-react';
import styles from './HeroSection.module.css';

export const HeroSection = ({ className }: { className?: string }) => {
  const [email, setEmail] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isVisible, setIsVisible] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    setIsVisible(true);
  }, []);

  function validateEmail(e?: string) {
    if (!e) return false;
    // simple email regex
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e);
  }

  function onSubmit(evt?: React.FormEvent) {
    evt?.preventDefault();
    if (!validateEmail(email)) {
      setError('Please enter a valid email address');
      return;
    }
    setError(null);
    // navigate to signup with email pre-filled
    window.location.href = `https://app.hith.chat/signup?email=${encodeURIComponent(email)}`;
  }

  return (
    <section className={`${styles.hero} ${className || ''}`}>
      <div className={styles.container}>
        
        

        {/* Left column: headline + form */}
        <div className={styles.content} style={{
          transform: isVisible ? 'translateY(0)' : 'translateY(30px)',
          opacity: isVisible ? 1 : 0,
          transition: 'all 0.8s ease-out',
          zIndex: 1,
          position: 'relative'
        }}>
          
          <h1 className={styles.headline}>
            The future of
            <br />
            customer communications.
          </h1>
          <p className={styles.subheadline}>
            A modern, AI-powered chat and ticketing platform that feels personal-not robotic.
          </p>

          <form onSubmit={onSubmit} style={{ 
            display: 'flex', 
            gap: 12, 
            alignItems: 'center', 
            marginTop: 32, 
            flexWrap: 'wrap'
          }}>
            <input
              aria-label="Email address"
              placeholder="Enter your email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              style={{
                padding: '12px 16px',
                borderRadius: 8,
                border: '1px solid var(--border)',
                background: 'var(--background)',
                minWidth: 280,
                flex: '1 1 420px',
                fontSize: 14
              }}
            />
            <Button type="submit" size="lg" style={{ 
              padding: '12px 20px', 
              fontWeight: 600
            }}>
              Get Started
            </Button>
            <Button asChild variant="ghost" size="lg" style={{
              padding: '12px 20px'
            }}>
              <a href="https://calendly.com/sumansaurabh-1/hith" target="_blank" rel="noreferrer">
                Book a Demo <MoveRight size={16} style={{ marginLeft: 4 }} />
              </a>
            </Button>
          </form>
          {error && <div style={{ color: 'var(--destructive, #ef4444)', marginTop: 12, fontSize: 14 }}>{error}</div>}

          <div className={styles.socialProof} style={{ marginTop: 32 }}>
            <div className={styles.avatars}>
              <Avatar style={{ width: 40, height: 40 }}>
                <AvatarImage src="https://randomuser.me/api/portraits/women/44.jpg" alt="User 1" />
                <AvatarFallback>U1</AvatarFallback>
              </Avatar>
              <Avatar style={{ width: 40, height: 40 }}>
                <AvatarImage src="https://randomuser.me/api/portraits/men/32.jpg" alt="User 2" />
                <AvatarFallback>U2</AvatarFallback>
              </Avatar>
              <Avatar style={{ width: 40, height: 40 }}>
                <AvatarImage src="https://randomuser.me/api/portraits/women/67.jpg" alt="User 3" />
                <AvatarFallback>U3</AvatarFallback>
              </Avatar>
              <Avatar style={{ width: 40, height: 40 }}>
                <AvatarImage src="https://randomuser.me/api/portraits/men/55.jpg" alt="User 4" />
                <AvatarFallback>U4</AvatarFallback>
              </Avatar>
            </div>
            <div>
              <p style={{ margin: 0, fontWeight: 600, fontSize: 14 }}>Trusted by 500+ teams worldwide</p>
              <div style={{ 
                display: 'flex', 
                alignItems: 'center', 
                gap: 4, 
                marginTop: 4,
                fontSize: 12,
                color: 'var(--muted-foreground)'
              }}>
                {[...Array(5)].map((_, i) => (
                  <span key={i} style={{ color: '#fbbf24' }}>★</span>
                ))}
                <span style={{ marginLeft: 4 }}>4.9/5 rating</span>
              </div>
            </div>
          </div>
        </div>

        {/* Right column: enhanced image */}
        <div className={styles.imageContainer} style={{
          transform: isVisible ? 'translateY(0)' : 'translateY(30px)',
          opacity: isVisible ? 1 : 0,
          transition: 'all 0.8s ease-out 0.2s'
        }}>
          <div style={{ 
            borderRadius: 20, 
            overflow: 'hidden', 
            position: 'relative',
            background: 'linear-gradient(135deg, color-mix(in srgb, var(--primary) 5%, transparent) 0%, color-mix(in srgb, var(--secondary) 3%, transparent) 100%)',
            padding: '2px'
          }}>
            <img
              src="https://images.unsplash.com/photo-1551434678-e076c223a692?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=2070&q=80"
              alt="AI-powered customer support dashboard"
              className={styles.heroImage}
              style={{ 
                display: 'block', 
                width: '100%', 
                height: '100%', 
                objectFit: 'cover',
                borderRadius: 18
              }}
            />
          </div>
        </div>
      </div>
      
      {/* Add floating animation keyframes */}
      <style>{`
        @keyframes float {
          0%, 100% { transform: translateY(0px); }
          50% { transform: translateY(-10px); }
        }
      `}</style>
    </section>
  );
};