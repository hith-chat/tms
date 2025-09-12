import { Routes, Route } from 'react-router-dom'
import { PublicTicketView } from './pages/PublicTicketView'
import { NotFound } from './pages/NotFound'

function App() {
  return (
    <div className="min-h-screen bg-background">
      <Routes>
        <Route path="/" element={<NotFound />} />
        <Route path="/tickets/:ticketId" element={<PublicTicketView />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </div>
  )
}

export default App
