import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import {
  CheckCircle2,
  XCircle,
  Loader2,
  Sparkles,
  Globe,
  Database,
  Wand2,
  ArrowRight,
} from 'lucide-react'
import { apiClient } from '../utils/api'
import { AIBuilderEvent, WidgetData } from '../utils/api'
import { BUILD_EVENT_MESSAGES } from '../utils/constants'
import './BuildPage.css'

interface BuildStep {
  id: string
  label: string
  status: 'pending' | 'in-progress' | 'completed' | 'error'
  message?: string
  icon: React.ReactNode
}

const BuildPage = () => {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [url, setUrl] = useState('')
  const [progress, setProgress] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const [widgetData, setWidgetData] = useState<WidgetData | null>(null)
  const [isComplete, setIsComplete] = useState(false)
  const [events, setEvents] = useState<AIBuilderEvent[]>([])

  const [steps, setSteps] = useState<BuildStep[]>([
    {
      id: 'initialization',
      label: 'Initializing Project',
      status: 'pending',
      icon: <Globe className="step-icon" />,
    },
    {
      id: 'widget',
      label: 'Creating Widget',
      status: 'pending',
      icon: <Wand2 className="step-icon" />,
    },
    {
      id: 'knowledge',
      label: 'Building Knowledge Base',
      status: 'pending',
      icon: <Database className="step-icon" />,
    },
    {
      id: 'completed',
      label: 'Finalizing',
      status: 'pending',
      icon: <Sparkles className="step-icon" />,
    },
  ])

  useEffect(() => {
    const urlParam = searchParams.get('url')
    if (!urlParam) {
      navigate('/')
      return
    }

    setUrl(urlParam)
    startBuild(urlParam)
  }, [searchParams, navigate])

  const startBuild = (websiteUrl: string) => {
    const eventSource = apiClient.streamWidgetBuild(
      websiteUrl,
      3,
      (event) => {
        setEvents((prev) => [...prev, event])
        handleEvent(event)
      },
      (err) => {
        setError(err.message)
        updateStepStatus('error', 'error')
      },
      (data) => {
        setWidgetData(data)
        setIsComplete(true)
        setProgress(100)
        updateStepStatus('completed', 'completed')
      }
    )

    return () => {
      eventSource.close()
    }
  }

  const handleEvent = (event: AIBuilderEvent) => {
    // Update progress based on event type
    const progressMap: Record<string, number> = {
      builder_started: 5,
      project_creation_started: 10,
      project_created: 20,
      widget_stage_started: 30,
      widget_theme_ready: 45,
      knowledge_stage_started: 50,
      scraping_progress: 60,
      faq_generation_started: 80,
      completed: 100,
    }

    const newProgress = progressMap[event.type] || progress
    setProgress(newProgress)

    // Update step status based on stage
    if (event.stage && event.stage !== 'internal') {
      updateStepStatus(event.stage, event.type === 'error' ? 'error' : 'in-progress')
    }

    // Mark completed steps
    if (event.type === 'project_created') {
      updateStepStatus('initialization', 'completed')
    } else if (event.type === 'widget_theme_ready') {
      updateStepStatus('widget', 'completed')
    } else if (event.type === 'faq_generation_started') {
      updateStepStatus('knowledge', 'completed')
    }
  }

  const updateStepStatus = (
    stepId: string,
    status: 'pending' | 'in-progress' | 'completed' | 'error',
    message?: string
  ) => {
    setSteps((prev) =>
      prev.map((step) =>
        step.id === stepId ? { ...step, status, message } : step
      )
    )
  }

  const handlePreview = () => {
    console.log('Preview clicked, widgetData:', widgetData)
    if (widgetData?.project_id) {
      navigate(`/preview/${widgetData.project_id}`, {
        state: { url, widgetData },
      })
    } else {
      console.error('Missing project_id in widgetData:', widgetData)
      setError('Unable to preview: missing project information')
    }
  }

  return (
    <div className="build-page">
      <div className="build-container">
        {/* Header */}
        <motion.div
          className="build-header"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <h1 className="build-title">
            {isComplete ? (
              <>
                <CheckCircle2 className="title-icon success" />
                Widget Built Successfully!
              </>
            ) : error ? (
              <>
                <XCircle className="title-icon error" />
                Build Failed
              </>
            ) : (
              <>
                <Loader2 className="title-icon spinning" />
                Building Your AI Widget
              </>
            )}
          </h1>
          <p className="build-subtitle">
            {isComplete
              ? 'Your AI-powered chat widget is ready to preview'
              : error
              ? 'Something went wrong during the build process'
              : url ? `Analyzing ${new URL(url).hostname}...` : 'Preparing to build...'}
          </p>
        </motion.div>

        {/* Progress Bar */}
        <motion.div
          className="progress-container"
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.2 }}
        >
          <div className="progress-bar">
            <motion.div
              className="progress-fill"
              initial={{ width: 0 }}
              animate={{ width: `${progress}%` }}
              transition={{ duration: 0.5 }}
            />
          </div>
          <div className="progress-label">{progress}%</div>
        </motion.div>

        {/* Build Steps */}
        <div className="steps-container">
          {steps.map((step, index) => (
            <motion.div
              key={step.id}
              className={`build-step ${step.status}`}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.1 + 0.3 }}
            >
              <div className="step-indicator">
                {step.status === 'completed' ? (
                  <CheckCircle2 className="step-status-icon completed" />
                ) : step.status === 'error' ? (
                  <XCircle className="step-status-icon error" />
                ) : step.status === 'in-progress' ? (
                  <Loader2 className="step-status-icon in-progress spinning" />
                ) : (
                  <div className="step-status-icon pending" />
                )}
              </div>

              <div className="step-content">
                <div className="step-header">
                  {step.icon}
                  <h3 className="step-label">{step.label}</h3>
                </div>
                {step.message && (
                  <p className="step-message">{step.message}</p>
                )}
              </div>
            </motion.div>
          ))}
        </div>

        {/* Event Log */}
        {events.length > 0 && (
          <motion.div
            className="event-log"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
          >
            <h3 className="event-log-title">Build Log</h3>
            <div className="event-log-container">
              <AnimatePresence>
                {events.slice(-5).reverse().map((event, index) => (
                  <motion.div
                    key={`${event.timestamp}-${index}`}
                    className={`event-item ${event.type}`}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, x: 10 }}
                  >
                    <span className="event-time">
                      {new Date(event.timestamp).toLocaleTimeString()}
                    </span>
                    <span className="event-message">
                      {BUILD_EVENT_MESSAGES[event.type] || event.message}
                    </span>
                  </motion.div>
                ))}
              </AnimatePresence>
            </div>
          </motion.div>
        )}

        {/* Error Display */}
        {error && (
          <motion.div
            className="error-box"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
          >
            <XCircle className="error-icon" />
            <div className="error-content">
              <h3 className="error-title">Build Failed</h3>
              <p className="error-message">{error}</p>
            </div>
            <button
              onClick={() => navigate('/')}
              className="btn btn-secondary"
            >
              Try Again
            </button>
          </motion.div>
        )}

        {/* Success Actions */}
        {isComplete && widgetData && (
          <motion.div
            className="success-actions"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3 }}
          >
            <button
              onClick={handlePreview}
              className="btn btn-primary btn-large"
            >
              Preview Widget on Your Site
              <ArrowRight />
            </button>
            <p className="success-note">
              See how the widget looks on your actual website with live AI responses
            </p>
          </motion.div>
        )}
      </div>
    </div>
  )
}

export default BuildPage
