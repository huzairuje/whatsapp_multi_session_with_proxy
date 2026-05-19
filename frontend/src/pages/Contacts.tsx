import { useState, useEffect } from 'react'
import { Search, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'
import Badge from '@/components/common/Badge'
import { sessionApi, contactsApi } from '@/services/api'
import type { Device } from '@/types'

interface Contact {
  id: number
  sender_jid: string
  contact_jid: string
  contact_name: string
  push_name: string
  business_name: string
  first_name: string
  full_name: string
  is_blocked: boolean
  is_business: boolean
  is_enterprise: boolean
  last_synced_at: string
  created_at: string
  updated_at: string
}

export default function Contacts() {
  const [sessions, setSessions] = useState<Device[]>([])
  const [selectedSession, setSelectedSession] = useState('')
  const [contacts, setContacts] = useState<Contact[]>([])
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [isSessionConnected, setIsSessionConnected] = useState(false)
  const [checkingStatus, setCheckingStatus] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  const pageSize = 50
  const [totalContacts, setTotalContacts] = useState(0)

  useEffect(() => {
    loadSessions()
  }, [])

  useEffect(() => {
    if (selectedSession) {
      setCurrentPage(1)
      loadContacts()
      checkSessionStatus()
    }
  }, [selectedSession])

  useEffect(() => {
    if (selectedSession) {
      loadContacts()
    }
  }, [currentPage, pageSize])

  useEffect(() => {
    if (selectedSession) {
      setCurrentPage(1)
      loadContacts()
    }
  }, [searchQuery])

  const loadSessions = async () => {
    try {
      const response = await sessionApi.getAll()
      const devices = Array.isArray(response.data) ? response.data : []
      setSessions(devices)
      if (devices.length > 0) {
        const firstJid = `${devices[0].user}@${devices[0].server}`
        setSelectedSession(firstJid)
      }
    } catch (error) {
      console.error('Failed to load sessions:', error)
      showMessage('error', 'Failed to load sessions')
    }
  }

  const checkSessionStatus = async () => {
    if (!selectedSession) return

    try {
      setCheckingStatus(true)
      const senderPhone = selectedSession.split('@')[0]
      const response = await sessionApi.getStatus(senderPhone)
      setIsSessionConnected(response.data?.isLogin === true)
    } catch (error) {
      console.error('Failed to check session status:', error)
      setIsSessionConnected(false)
    } finally {
      setCheckingStatus(false)
    }
  }

  const loadContacts = async () => {
    if (!selectedSession) return

    try {
      setLoading(true)
      const offset = (currentPage - 1) * pageSize
      
      let response
      if (searchQuery.trim()) {
        response = await contactsApi.searchContacts(selectedSession, searchQuery)
        setContacts(response.data || [])
        setTotalContacts(response.data?.length || 0)
      } else {
        response = await contactsApi.getContacts(selectedSession, pageSize, offset)
        setContacts(response.data.contacts || [])
        setTotalContacts(response.data.total || 0)
      }
    } catch (error) {
      console.error('Failed to load contacts:', error)
      showMessage('error', 'Failed to load contacts')
      setContacts([])
    } finally {
      setLoading(false)
    }
  }

  const handleSyncContacts = async () => {
    if (!selectedSession) return

    try {
      setLoading(true)
      await contactsApi.syncContacts(selectedSession, false)
      showMessage('success', 'Contact sync started in background')
      setTimeout(() => loadContacts(), 2000)
    } catch (error) {
      console.error('Failed to sync contacts:', error)
      showMessage('error', 'Failed to sync contacts')
    } finally {
      setLoading(false)
    }
  }

  const showMessage = (type: 'success' | 'error', text: string) => {
    setMessage({ type, text })
    setTimeout(() => setMessage(null), 3000)
  }

  const totalPages = Math.ceil(totalContacts / pageSize)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Contacts</h1>
          <p className="text-gray-600 mt-1">Manage WhatsApp contacts for each session</p>
        </div>
        <Button 
          onClick={handleSyncContacts} 
          isLoading={loading || checkingStatus}
          disabled={!isSessionConnected || loading || checkingStatus}
        >
          <RefreshCw className="w-4 h-4 mr-2" />
          Sync Contacts
        </Button>
      </div>

      {!isSessionConnected && selectedSession && !checkingStatus && (
        <div className="rounded-lg p-4 bg-yellow-50 border border-yellow-200 text-yellow-800">
          Session is not connected. Please connect the session first to sync contacts.
        </div>
      )}

      {message && (
        <div
          className={`rounded-lg p-4 ${
            message.type === 'success'
              ? 'bg-green-50 border border-green-200 text-green-800'
              : 'bg-red-50 border border-red-200 text-red-800'
          }`}
        >
          {message.text}
        </div>
      )}

      <Card>
        <div className="p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Select Session
              </label>
              <select
                value={selectedSession}
                onChange={(e) => setSelectedSession(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option key="empty" value="">Select a session...</option>
                {sessions.map((session, index) => {
                  const jid = `${session.user}@${session.server}`
                  return (
                    <option key={jid || `session-${index}`} value={jid}>
                      {jid} {session.pushName ? `(${session.pushName})` : ''}
                    </option>
                  )
                })}
              </select>
            </div>

            <div className="relative">
              <Search className="absolute left-3 top-3 w-4 h-4 text-gray-400" />
              <input
                type="text"
                placeholder="Search contacts by name or number..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div className="flex items-center justify-between text-sm text-gray-600">
              <div>
                {totalContacts} contact{totalContacts !== 1 ? 's' : ''} found
                {!searchQuery && ` (Page ${currentPage} of ${totalPages})`}
              </div>
              {!searchQuery && totalPages > 1 && (
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                    disabled={currentPage === 1 || loading}
                  >
                    <ChevronLeft className="w-4 h-4" />
                  </Button>
                  <span className="text-sm">
                    {currentPage} / {totalPages}
                  </span>
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                    disabled={currentPage === totalPages || loading}
                  >
                    <ChevronRight className="w-4 h-4" />
                  </Button>
                </div>
              )}
            </div>
          </div>
        </div>
      </Card>

      <div className="grid gap-1">
        {loading && contacts.length === 0 ? (
          <Card>
            <div className="p-6 text-center text-gray-500">Loading contacts...</div>
          </Card>
        ) : contacts.length === 0 ? (
          <Card>
            <div className="p-6 text-center text-gray-500">
              {searchQuery ? 'No contacts match your search.' : 'No contacts yet. Click "Sync Contacts" to load.'}
            </div>
          </Card>
        ) : (
          <div className="grid grid-cols-4 gap-1">
            {contacts.map((contact) => (
              <div
                key={contact.id}
                className="p-2 border border-gray-200 rounded hover:bg-blue-50 transition-colors"
              >
                <p className="font-medium text-gray-900 truncate text-xs leading-tight">
                  {contact.contact_name || contact.push_name || contact.full_name}
                </p>
                <p className="text-gray-600 truncate text-xs leading-tight">
                  {contact.contact_jid.split('@')[0]}
                </p>
                <div className="flex gap-1 mt-1">
                  {contact.is_business && <Badge variant="info">B</Badge>}
                  {contact.is_blocked && <Badge variant="error">X</Badge>}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
