import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/layout/Layout'
import Dashboard from './pages/Dashboard'
import Sessions from './pages/Sessions'
import BulkSend from './pages/BulkSend'
import Recipients from './pages/Recipients'
import Templates from './pages/Templates'
import Settings from './pages/Settings'

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="sessions" element={<Sessions />} />
          <Route path="bulk-send" element={<BulkSend />} />
          <Route path="recipients" element={<Recipients />} />
          <Route path="templates" element={<Templates />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </Router>
  )
}

export default App
