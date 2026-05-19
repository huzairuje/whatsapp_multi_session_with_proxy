import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/layout/Layout'
import Dashboard from './pages/Dashboard'
import Sessions from './pages/Sessions'
import Messages from './pages/Messages'
import BulkSend from './pages/BulkSend'
import Recipients from './pages/Recipients'
import Templates from './pages/Templates'
import Settings from './pages/Settings'
import Activities from './pages/Activities'
import WarmUp from './pages/WarmUp'
import Contacts from './pages/Contacts'
import Login from './pages/Login'
import { ProtectedRoute } from './components/ProtectedRoute'

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="sessions" element={<Sessions />} />
          <Route path="messages" element={<Messages />} />
          <Route path="bulk-send" element={<BulkSend />} />
          <Route path="recipients" element={<Recipients />} />
          <Route path="templates" element={<Templates />} />
          <Route path="warmup" element={<WarmUp />} />
          <Route path="contacts" element={<Contacts />} />
          <Route path="settings" element={<Settings />} />
          <Route path="activities" element={<Activities />} />
        </Route>
      </Routes>
    </Router>
  )
}

export default App
