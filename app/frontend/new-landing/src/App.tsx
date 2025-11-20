import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { useEffect } from 'react'
import HomePage from './pages/HomePage'
import BuildPage from './pages/BuildPage'
import PreviewPage from './pages/PreviewPage'

function App() {
  useEffect(() => {
    const script = document.createElement('script')
    script.src = 'https://api.hith.chat/embed/3cee3e02-36ff-4afa-b072-9de2093003c6.js'
    script.async = true
    document.body.appendChild(script)

    return () => {
      document.body.removeChild(script)
    }
  }, [])

  return (
    <Router>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/build" element={<BuildPage />} />
        <Route path="/preview/:projectId" element={<PreviewPage />} />
      </Routes>
    </Router>
  )
}

export default App
