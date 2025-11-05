import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import HomePage from './pages/HomePage'
import BuildPage from './pages/BuildPage'
import PreviewPage from './pages/PreviewPage'

function App() {
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
