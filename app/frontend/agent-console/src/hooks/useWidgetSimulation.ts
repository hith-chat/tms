import { useState, useEffect } from 'react'

interface SimulationMessage {
  id: string
  type: 'system' | 'visitor' | 'agent'
  content: string
  timestamp: Date
}

export function useWidgetSimulation(welcomeMessage?: string, agentName?: string) {
  const [isWidgetOpen, setIsWidgetOpen] = useState(true)
  const [isTyping, setIsTyping] = useState(false)
  const [simulationMessages, setSimulationMessages] = useState<SimulationMessage[]>([])

  // Update simulation messages when welcome message or agent name changes
  useEffect(() => {
    setSimulationMessages([
      {
        id: '1',
        type: 'system',
        content: welcomeMessage || 'Hello! How can we help you today?',
        timestamp: new Date()
      },
      {
        id: '2',
        type: 'visitor',
        content: 'Hello! I have a question about your services.',
        timestamp: new Date()
      },
      {
        id: '3',
        type: 'agent',
        content: 'Hi there! I\'d be happy to help you with that. What specific information are you looking for?',
        timestamp: new Date()
      }
    ])
  }, [welcomeMessage, agentName])

  const toggleWidget = () => {
    setIsWidgetOpen(!isWidgetOpen)
    if (!isWidgetOpen) {
      // Simulate typing when opening
      setTimeout(() => {
        setIsTyping(true)
        setTimeout(() => setIsTyping(false), 2000)
      }, 1000)
    }
  }

  const startTypingDemo = () => {
    setIsTyping(true)
    setTimeout(() => setIsTyping(false), 2000)
  }

  return {
    isWidgetOpen,
    isTyping,
    simulationMessages,
    toggleWidget,
    startTypingDemo
  }
}
