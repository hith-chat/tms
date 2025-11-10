import { useEffect, useState, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Eye,
  Code2,
  Copy,
  CheckCircle2,
  ExternalLink,
  AlertTriangle,
  X,
  RefreshCw,
  LogIn,
  ArrowLeft,
  Sparkles,
} from 'lucide-react'
import { WidgetData } from '../utils/api'
import './PreviewPage.css'

const PreviewPage = () => {
  const location = useLocation()
  const navigate = useNavigate()
  const iframeRef = useRef<HTMLIFrameElement>(null)

  const [url, setUrl] = useState('')
  const [widgetData, setWidgetData] = useState<WidgetData | null>(null)
  const [iframeError, setIframeError] = useState(false)
  const [iframeLoaded, setIframeLoaded] = useState(false)
  const [showCode, setShowCode] = useState(false)
  const [copied, setCopied] = useState(false)
  const [widgetInjected, setWidgetInjected] = useState(false)

  useEffect(() => {
    // Get data from navigation state
    const state = location.state as { url: string; widgetData: WidgetData }
    if (state?.url && state?.widgetData) {
      setUrl(state.url)
      setWidgetData(state.widgetData)
    } else {
      // Redirect back if no data
      navigate('/')
    }
  }, [location, navigate])

  useEffect(() => {
    // Inject widget directly in the page for preview
    if (widgetData?.widget_id && !widgetInjected) {
      injectWidgetInPage()
    }
  }, [widgetData, widgetInjected])

  const handleIframeLoad = () => {
    setIframeLoaded(true)
    setIframeError(false)
  }

  const handleIframeError = () => {
    setIframeError(true)
    setIframeLoaded(false)
  }

  const injectWidgetInPage = () => {
    if (!widgetData?.widget_id) return

    try {
      // Check if widget script is already loaded
      const existingScript = document.querySelector(
        `script[src*="${widgetData.widget_id}"]`
      )
      if (existingScript) {
        setWidgetInjected(true)
        return
      }

      // Create script element to inject widget directly in the page
      // This loads a tiny loader script that sets config and loads the main widget from CDN
      const script = document.createElement('script')
      script.src = `https://api.hith.chat/api/public/chat/widgets/${widgetData.widget_id}/embed.js`
      script.async = true
      script.onload = () => {
        console.log('Widget script loaded successfully')
        setWidgetInjected(true)
      }
      script.onerror = () => {
        console.error('Failed to load widget script')
      }

      document.head.appendChild(script)
    } catch (error) {
      console.error('Error injecting widget:', error)
    }
  }

  const getEmbedCode = () => {
    if (!widgetData?.widget_id) return ''

    // Single-line script embed - loads tiny loader that fetches main widget from CDN
    return `<script src="https://api.hith.chat/api/public/chat/widgets/${widgetData.widget_id}/embed.js" async></script>`
  }

  const copyCode = () => {
    const code = getEmbedCode()
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleSignUp = () => {
    // In production, this would redirect to the sign-up page
    // For now, show the modal
    setShowCode(true)
  }

  return (
    <div className="preview-page">
      {/* Header */}
      <div className="preview-header">
        <div className="container">
          <div className="preview-header-content">
            <div className="header-left">
              <button
                onClick={() => navigate('/')}
                className="btn-icon"
                title="Back to home"
              >
                <ArrowLeft />
              </button>
              <div className="header-info">
                <h1 className="preview-title">
                  <Eye className="title-icon" />
                  Preview Your Widget
                </h1>
                <p className="preview-subtitle">
                  {url ? new URL(url).hostname : 'Loading...'}
                </p>
              </div>
            </div>

            <div className="header-actions">
              <button
                onClick={() => setShowCode(!showCode)}
                className="btn btn-secondary"
              >
                <Code2 size={18} />
                Get Embed Code
              </button>
              <button
                onClick={handleSignUp}
                className="btn btn-primary"
              >
                <LogIn size={18} />
                Sign Up to Deploy
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Preview Container */}
      <div className="preview-container">
        <div className="preview-wrapper">
          {/* Loading State */}
          {!iframeLoaded && !iframeError && (
            <div className="preview-loading">
              <RefreshCw className="loading-icon spinning" size={48} />
              <p>Loading preview...</p>
            </div>
          )}

          {/* Error State */}
          {iframeError && (
            <div className="preview-error">
              <AlertTriangle className="error-icon" size={48} />
              <h3>Unable to Load Preview</h3>
              <p>
                This website doesn't allow embedding in iframes for security reasons.
                <br />
                Don't worry! Your widget will work perfectly when you install it on your actual site.
              </p>
              <div className="error-actions">
                <a
                  href={url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn btn-secondary"
                >
                  <ExternalLink size={18} />
                  Open Website in New Tab
                </a>
                <button
                  onClick={() => setShowCode(true)}
                  className="btn btn-primary"
                >
                  <Code2 size={18} />
                  Get Embed Code
                </button>
              </div>
            </div>
          )}

          {/* Iframe */}
          <iframe
            ref={iframeRef}
            src={url}
            className={`preview-iframe ${iframeLoaded ? 'loaded' : ''}`}
            onLoad={handleIframeLoad}
            onError={handleIframeError}
            title="Website Preview"
            sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
          />

          {/* Widget Badge */}
          {widgetInjected && (
            <motion.div
              className="widget-badge"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.5 }}
            >
              <CheckCircle2 size={16} />
              <span>Widget Active</span>
            </motion.div>
          )}
        </div>

        {/* Instructions Panel */}
        <motion.div
          className="instructions-panel card"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
        >
          <h3 className="instructions-title">
            <Sparkles size={24} />
            Your Widget is Ready!
          </h3>
          <div className="instructions-content">
            <div className="instruction-item">
              <div className="instruction-number">1</div>
              <div className="instruction-text">
                <strong>Test it out:</strong> Try clicking the chat bubble in the preview above to see your AI assistant in action
              </div>
            </div>
            <div className="instruction-item">
              <div className="instruction-number">2</div>
              <div className="instruction-text">
                <strong>Sign up:</strong> Create a free account to deploy your widget and get your unique embed code
              </div>
            </div>
            <div className="instruction-item">
              <div className="instruction-number">3</div>
              <div className="instruction-text">
                <strong>Deploy:</strong> Copy the single-line script tag and paste it in your website's HTML
              </div>
            </div>
          </div>
          <button
            onClick={handleSignUp}
            className="btn btn-primary btn-full"
          >
            <LogIn size={18} />
            Sign Up to Get Your Code
          </button>
        </motion.div>
      </div>

      {/* Code Modal */}
      <AnimatePresence>
        {showCode && (
          <motion.div
            className="modal-overlay"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setShowCode(false)}
          >
            <motion.div
              className="modal-content"
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.9, opacity: 0 }}
              onClick={(e) => e.stopPropagation()}
            >
              <div className="modal-header">
                <h2 className="modal-title">
                  <Code2 size={24} />
                  Embed Code
                </h2>
                <button
                  onClick={() => setShowCode(false)}
                  className="btn-icon"
                >
                  <X />
                </button>
              </div>

              <div className="modal-body">
                <p className="modal-description">
                  Copy this single-line script tag and paste it before the closing <code>&lt;/body&gt;</code> tag in your website's HTML:
                </p>

                <div className="code-box">
                  <pre className="code-content">{getEmbedCode()}</pre>
                  <button
                    onClick={copyCode}
                    className={`copy-button ${copied ? 'copied' : ''}`}
                  >
                    {copied ? (
                      <>
                        <CheckCircle2 size={16} />
                        Copied!
                      </>
                    ) : (
                      <>
                        <Copy size={16} />
                        Copy
                      </>
                    )}
                  </button>
                </div>

                <div className="modal-note">
                  <AlertTriangle size={18} />
                  <p>
                    <strong>Note:</strong> This is a preview embed code.
                    Sign up to get your permanent widget code that will work after your free trial expires.
                  </p>
                </div>
              </div>

              <div className="modal-footer">
                <button
                  onClick={() => setShowCode(false)}
                  className="btn btn-secondary"
                >
                  Close
                </button>
                <button
                  onClick={handleSignUp}
                  className="btn btn-primary"
                >
                  <LogIn size={18} />
                  Sign Up Now
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

export default PreviewPage
