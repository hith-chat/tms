import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
  Sparkles,
  Zap,
  Shield,
  MessageSquare,
  ArrowRight,
  CheckCircle2,
  Globe,
  Code2
} from 'lucide-react'
import { FEATURES, STEPS } from '../utils/constants'
import './HomePage.css'

const HomePage = () => {
  const [websiteUrl, setWebsiteUrl] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!websiteUrl) {
      setError('Please enter a website URL')
      return
    }

    // Trim whitespace
    const trimmedUrl = websiteUrl.trim()

    // Check for invalid input (spaces, missing domain, etc.)
    if (trimmedUrl.includes(' ')) {
      setError('URL cannot contain spaces')
      return
    }

    // Basic URL validation
    try {
      const url = new URL(trimmedUrl.startsWith('http') ? trimmedUrl : `https://${trimmedUrl}`)

      // Validate that the URL has a proper host
      if (!url.host || url.host.length < 3 || !url.host.includes('.')) {
        setError('Please enter a valid domain (e.g., example.com)')
        return
      }

      setIsLoading(true)

      // Navigate to build page with URL as query param
      navigate(`/build?url=${encodeURIComponent(url.toString())}`)
    } catch (err) {
      setError('Please enter a valid URL (e.g., example.com or https://example.com)')
      setIsLoading(false)
    }
  }

  return (
    <div className="home-page">
      {/* Navigation */}
      <nav className="nav">
        <div className="container">
          <div className="nav-content">
            <motion.div
              className="logo"
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5 }}
            >
              <MessageSquare className="logo-icon" />
              <span className="logo-text">Hith<span className="gradient-text">Chat</span></span>
            </motion.div>
            <motion.div
              className="nav-links"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5 }}
            >
              <a href="#features">Features</a>
              <a href="#how-it-works">How It Works</a>
              <a href="https://hith.chat" target="_blank" rel="noopener noreferrer" className="btn btn-secondary">
                Sign In
              </a>
            </motion.div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="hero">
        <div className="hero-bg">
          <div className="gradient-orb orb-1"></div>
          <div className="gradient-orb orb-2"></div>
          <div className="gradient-orb orb-3"></div>
        </div>

        <div className="container">
          <motion.div
            className="hero-content"
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
          >
            <motion.div
              className="hero-badge"
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              <Sparkles size={16} />
              <span>AI-Powered Customer Support</span>
            </motion.div>

            <h1 className="hero-title">
              Turn Visitors Into
              <br />
              <span className="gradient-text">Engaged Customers</span>
            </h1>

            <p className="hero-description">
              Add an intelligent AI chat widget to your website in minutes.
              Answer questions, capture leads, schedule meetings, and provide 24/7 support
              — all automatically powered by your website content.
            </p>

            <div className="hero-form">
              <form onSubmit={handleSubmit} className="url-form">
                <div className="form-group">
                  <Globe className="input-icon" />
                  <input
                    type="text"
                    placeholder="Enter your website URL (e.g., example.com)"
                    value={websiteUrl}
                    onChange={(e) => setWebsiteUrl(e.target.value)}
                    className="url-input"
                    disabled={isLoading}
                  />
                  <button
                    type="submit"
                    className="btn btn-primary submit-btn"
                    disabled={isLoading}
                  >
                    {isLoading ? (
                      <>
                        <span className="spinner"></span>
                        Building...
                      </>
                    ) : (
                      <>
                        Build My Widget
                        <ArrowRight size={20} />
                      </>
                    )}
                  </button>
                </div>
                {error && (
                  <motion.p
                    className="error-message"
                    initial={{ opacity: 0, y: -10 }}
                    animate={{ opacity: 1, y: 0 }}
                  >
                    {error}
                  </motion.p>
                )}
              </form>

              <div className="hero-stats">
                <div className="stat">
                  <Zap size={20} />
                  <span>Free to try</span>
                </div>
                <div className="stat">
                  <Shield size={20} />
                  <span>No credit card</span>
                </div>
                <div className="stat">
                  <Code2 size={20} />
                  <span>One line of code</span>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="features-section">
        <div className="container">
          <motion.div
            className="section-header"
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <h2 className="section-title">
              Everything You Need to
              <span className="gradient-text"> Delight Customers</span>
            </h2>
            <p className="section-description">
              Powerful features designed for early-stage founders who want to provide
              exceptional customer experience without the overhead.
            </p>
          </motion.div>

          <div className="features-grid">
            {FEATURES.map((feature, index) => (
              <motion.div
                key={feature.title}
                className="feature-card glass"
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
                whileHover={{ y: -5 }}
              >
                <div className="feature-icon">{feature.icon}</div>
                <h3 className="feature-title">{feature.title}</h3>
                <p className="feature-description">{feature.description}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section id="how-it-works" className="how-it-works-section">
        <div className="container">
          <motion.div
            className="section-header"
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <h2 className="section-title">
              Get Started in
              <span className="gradient-text"> 4 Simple Steps</span>
            </h2>
            <p className="section-description">
              From zero to AI-powered support in minutes, not days.
            </p>
          </motion.div>

          <div className="steps-container">
            {STEPS.map((step, index) => (
              <motion.div
                key={step.number}
                className="step-card"
                initial={{ opacity: 0, x: -20 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.15 }}
              >
                <div className="step-number-wrapper">
                  <div className="step-number">{step.number}</div>
                  {index < STEPS.length - 1 && <div className="step-connector"></div>}
                </div>
                <div className="step-content">
                  <h3 className="step-title">{step.title}</h3>
                  <p className="step-description">{step.description}</p>
                </div>
              </motion.div>
            ))}
          </div>

          <motion.div
            className="cta-box card"
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, delay: 0.4 }}
          >
            <CheckCircle2 className="cta-icon" size={48} />
            <h3 className="cta-title">Ready to Transform Your Customer Support?</h3>
            <p className="cta-description">
              Join hundreds of founders who are already using AI to engage customers,
              generate leads, and grow their business.
            </p>
            <button
              onClick={() => {
                const input = document.querySelector('.url-input') as HTMLInputElement
                input?.focus()
              }}
              className="btn btn-primary btn-large"
            >
              Build Your Widget Now
              <ArrowRight size={20} />
            </button>
          </motion.div>
        </div>
      </section>

      {/* Footer */}
      <footer className="footer">
        <div className="container">
          <div className="footer-content">
            <div className="footer-left">
              <div className="logo">
                <MessageSquare className="logo-icon" />
                <span className="logo-text">Hith<span className="gradient-text">Chat</span></span>
              </div>
              <p className="footer-description">
                Hyper Intelligent Tech Helper
              </p>
            </div>
            <div className="footer-right">
              <div className="footer-links">
                <a href="https://hith.chat" target="_blank" rel="noopener noreferrer">Documentation</a>
                <a href="https://hith.chat" target="_blank" rel="noopener noreferrer">Privacy Policy</a>
                <a href="https://hith.chat" target="_blank" rel="noopener noreferrer">Terms of Service</a>
                <a href="mailto:sumansaurabh@hith.chat">Contact</a>
              </div>
            </div>
          </div>
          <div className="footer-bottom">
            <p>&copy; 2024 HithChat. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default HomePage
