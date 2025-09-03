import { HandoffNotificationsList, HandoffNotificationBadge } from '../components/handoff'
import { useHandoffNotifications } from '../hooks/useHandoffNotifications'

export function HandoffTestPage() {
  const { notifications, unreadCount } = useHandoffNotifications()

  const simulateHandoffRequest = () => {
    // This would normally come from the WebSocket, but for testing we can simulate it
    const testData = {
      session_id: `test-session-${Date.now()}`,
      customer_name: "John Doe",
      customer_email: "john.doe@example.com",
      handoff_reason: "Customer needs help with billing questions that AI cannot answer",
      urgency_level: "high" as const,
      requested_at: new Date().toISOString(),
      timeout_at: new Date(Date.now() + 30 * 60 * 1000).toISOString(), // 30 minutes from now
      session_metadata: {
        messages_count: 15,
        session_duration: 1200, // 20 minutes
        last_ai_response: "I understand you need help with billing. Let me connect you with a human agent who can assist you better."
      }
    }

    // Manually trigger the handoff handler (for testing only)
    window.dispatchEvent(new CustomEvent('test-handoff', { detail: testData }))
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-8">
        <div className="flex items-center gap-4 mb-4">
          <h1 className="text-2xl font-bold text-gray-900">AI Handoff System Test</h1>
          <HandoffNotificationBadge />
        </div>
        
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <h2 className="font-semibold text-blue-900 mb-2">Test Instructions</h2>
          <p className="text-blue-800 mb-3">
            This page demonstrates the AI handoff notification system. In production, 
            handoff requests come from the AI service via WebSocket when customers 
            request human assistance.
          </p>
          <button
            onClick={simulateHandoffRequest}
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded font-medium"
          >
            Simulate Handoff Request
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <div className="bg-gray-50 p-4 rounded-lg">
            <div className="text-2xl font-bold text-gray-900">{notifications.length}</div>
            <div className="text-sm text-gray-600">Total Requests</div>
          </div>
          <div className="bg-red-50 p-4 rounded-lg">
            <div className="text-2xl font-bold text-red-600">{unreadCount}</div>
            <div className="text-sm text-gray-600">Unread</div>
          </div>
          <div className="bg-green-50 p-4 rounded-lg">
            <div className="text-2xl font-bold text-green-600">
              {notifications.filter(n => n.isActive).length}
            </div>
            <div className="text-sm text-gray-600">Active</div>
          </div>
        </div>
      </div>

      <HandoffNotificationsList />

      <div className="mt-8 p-4 bg-gray-50 rounded-lg text-sm text-gray-600">
        <h3 className="font-semibold mb-2">System Features:</h3>
        <ul className="space-y-1">
          <li>• Real-time WebSocket notifications from AI service</li>
          <li>• Audio and browser notifications for new requests</li>
          <li>• Urgency-based prioritization and visual indicators</li>
          <li>• Accept/Decline actions with automatic chat redirection</li>
          <li>• Session metadata display (duration, message count, last AI response)</li>
          <li>• Notification badge for navigation integration</li>
        </ul>
      </div>
    </div>
  )
}
