import { useState, useEffect } from 'react'
import { Plus, Trash2, Edit2, TrendingUp } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'
import Input from '@/components/common/Input'
import Badge from '@/components/common/Badge'
import { warmupApi, sessionApi } from '@/services/api'
import type { Device } from '@/types'

interface WarmUpConfig {
  id: number
  sender_jid: string
  enabled: boolean
  current_day: number
  start_date: string
  daily_limit: number
  increment_amount: number
  increment_days: number
  max_daily_limit: number
  created_at: string
  updated_at: string
}

interface WarmUpStatus {
  enabled: boolean
  current_day: number
  current_limit: number
  max_limit: number
  start_date: string
  config?: WarmUpConfig
}

export default function WarmUp() {
  const [configs, setConfigs] = useState<WarmUpConfig[]>([])
  const [sessions, setSessions] = useState<Device[]>([])
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [selectedSender, setSelectedSender] = useState('')
  const [status, setStatus] = useState<WarmUpStatus | null>(null)
  const [useSessionDropdown, setUseSessionDropdown] = useState(true)

  const [formData, setFormData] = useState({
    sender_jid: '',
    enabled: true,
    daily_limit: 5,
    increment_amount: 5,
    increment_days: 3,
    max_daily_limit: 1000,
  })

  useEffect(() => {
    loadConfigs()
    loadSessions()
  }, [])

  const loadSessions = async () => {
    try {
      const response = await sessionApi.getAll()
      const devices = Array.isArray(response.data) ? response.data : []
      setSessions(devices)
      if (devices.length > 0) {
        const firstJid = `${devices[0].user}@${devices[0].server}`
        setFormData({ ...formData, sender_jid: firstJid })
      }
    } catch (error) {
      console.error('Failed to load sessions:', error)
      setSessions([])
    }
  }

  const loadConfigs = async () => {
    try {
      setLoading(true)
      const response = await warmupApi.getAll()
      setConfigs(response.data || [])
    } catch (error) {
      console.error('Failed to load templates:', error)
      setConfigs([])
      showMessage('error', 'Failed to load warm-up configurations')
    } finally {
      setLoading(false)
    }
  }

  const loadStatus = async (sender: string) => {
    try {
      const response = await warmupApi.getStatus(sender)
      setStatus(response.data)
      setSelectedSender(sender)
    } catch (error) {
      showMessage('error', 'Failed to load warm-up status')
    }
  }

  const showMessage = (type: 'success' | 'error', text: string) => {
    setMessage({ type, text })
    setTimeout(() => setMessage(null), 3000)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!formData.sender_jid.trim()) {
      showMessage('error', 'Sender JID is required')
      return
    }

    if (formData.daily_limit > formData.max_daily_limit) {
      showMessage('error', 'Daily limit cannot exceed max daily limit')
      return
    }

    try {
      setLoading(true)
      if (editingId) {
        await warmupApi.update(formData.sender_jid, {
          enabled: formData.enabled,
          daily_limit: formData.daily_limit,
          increment_amount: formData.increment_amount,
          increment_days: formData.increment_days,
          max_daily_limit: formData.max_daily_limit,
        })
        showMessage('success', 'Warm-up configuration updated')
      } else {
        await warmupApi.create(formData)
        showMessage('success', 'Warm-up configuration created')
      }
      resetForm()
      loadConfigs()
    } catch (error: any) {
      showMessage('error', error.message || 'Failed to save configuration')
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (config: WarmUpConfig) => {
    setFormData({
      sender_jid: config.sender_jid,
      enabled: config.enabled,
      daily_limit: config.daily_limit,
      increment_amount: config.increment_amount,
      increment_days: config.increment_days,
      max_daily_limit: config.max_daily_limit,
    })
    setEditingId(config.id)
    setShowForm(true)
  }

  const handleDelete = async (senderJid: string) => {
    if (!confirm('Are you sure you want to delete this warm-up configuration?')) return

    try {
      setLoading(true)
      await warmupApi.delete(senderJid)
      showMessage('success', 'Warm-up configuration deleted')
      loadConfigs()
    } catch (error: any) {
      showMessage('error', error.message || 'Failed to delete configuration')
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setFormData({
      sender_jid: '',
      enabled: true,
      daily_limit: 5,
      increment_amount: 5,
      increment_days: 3,
      max_daily_limit: 1000,
    })
    setEditingId(null)
    setShowForm(false)
  }

  const calculateProjectedLimit = (config: WarmUpConfig, daysFromNow: number) => {
    const daysSinceStart = Math.floor((new Date().getTime() - new Date(config.start_date).getTime()) / (1000 * 60 * 60 * 24))
    const totalDays = daysSinceStart + daysFromNow
    const incrementsApplied = Math.floor(totalDays / config.increment_days)
    const projected = config.daily_limit + (incrementsApplied * config.increment_amount)
    return Math.min(projected, config.max_daily_limit)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Warm-up Manager</h1>
          <p className="text-gray-600 mt-1">Gradually increase daily sending limits to avoid account flagging</p>
        </div>
        <Button onClick={() => setShowForm(!showForm)}>
          <Plus className="w-4 h-4 mr-2" />
          New Warm-up
        </Button>
      </div>

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

      {showForm && (
        <Card>
          <div className="p-6">
            <h2 className="text-lg font-semibold mb-4">
              {editingId ? 'Edit Warm-up Configuration' : 'Create Warm-up Configuration'}
            </h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Sender JID (phone number)
                </label>
                <div className="space-y-2">
                  <div className="flex items-center space-x-4 mb-2">
                    <label className="flex items-center">
                      <input
                        type="radio"
                        checked={useSessionDropdown}
                        onChange={() => setUseSessionDropdown(true)}
                        className="mr-2"
                        disabled={!!editingId}
                      />
                      <span className="text-sm">Select from sessions</span>
                    </label>
                    <label className="flex items-center">
                      <input
                        type="radio"
                        checked={!useSessionDropdown}
                        onChange={() => setUseSessionDropdown(false)}
                        className="mr-2"
                        disabled={!!editingId}
                      />
                      <span className="text-sm">Enter manually</span>
                    </label>
                  </div>
                  
                  {useSessionDropdown ? (
                    <select
                      value={formData.sender_jid}
                      onChange={(e) => setFormData({ ...formData, sender_jid: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                      disabled={!!editingId}
                    >
                      <option value="">Select a session...</option>
                      {sessions.map((session, index) => {
                        const jid = `${session.user}@${session.server}`
                        return (
                          <option key={jid || `session-${index}`} value={jid}>
                            {jid} {session.pushName ? `(${session.pushName})` : ''}
                          </option>
                        )
                      })}
                    </select>
                  ) : (
                    <input
                      type="text"
                      placeholder="6281234567890"
                      value={formData.sender_jid}
                      onChange={(e) => setFormData({ ...formData, sender_jid: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                      disabled={!!editingId}
                    />
                  )}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
              <Input
                label="Starting Daily Limit (1-5 for new accounts)"
                type="number"
                min="1"
                max="10"
                value={formData.daily_limit}
                onChange={(e) => setFormData({ ...formData, daily_limit: parseInt(e.target.value) })}
              />
                <Input
                  label="Increment Amount (per period)"
                  type="number"
                  min="1"
                  value={formData.increment_amount}
                  onChange={(e) => setFormData({ ...formData, increment_amount: parseInt(e.target.value) })}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <Input
                  label="Increment Period (days)"
                  type="number"
                  min="1"
                  value={formData.increment_days}
                  onChange={(e) => setFormData({ ...formData, increment_days: parseInt(e.target.value) })}
                />
                <Input
                  label="Maximum Daily Limit (safety cap)"
                  type="number"
                  min="1"
                  max="1000"
                  value={formData.max_daily_limit}
                  onChange={(e) => setFormData({ ...formData, max_daily_limit: parseInt(e.target.value) })}
                />
              </div>

              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="enabled"
                  checked={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                  className="rounded"
                />
                <label htmlFor="enabled" className="text-sm font-medium text-gray-700">
                  Enable warm-up for this sender
                </label>
              </div>

              <div className="bg-blue-50 border border-blue-200 rounded p-3 text-sm text-blue-800">
                <p className="font-semibold mb-1">Warm-up Schedule:</p>
                <p className="mb-1"><strong>⚠️ Important:</strong> New accounts should start with 1-5 messages/day to avoid flagging</p>
                <p>Day 1-{formData.increment_days}: {formData.daily_limit} messages/day</p>
                <p>Day {formData.increment_days + 1}-{formData.increment_days * 2}: {Math.min(formData.daily_limit + formData.increment_amount, formData.max_daily_limit)} messages/day</p>
                <p>Continues until reaching {formData.max_daily_limit} messages/day (max safety limit)</p>
              </div>

              <div className="flex space-x-3">
                <Button type="submit" isLoading={loading}>
                  {editingId ? 'Update' : 'Create'} Configuration
                </Button>
                <Button variant="secondary" onClick={resetForm}>
                  Cancel
                </Button>
              </div>
            </form>
          </div>
        </Card>
      )}

      {status && selectedSender && (
        <Card>
          <div className="p-6">
            <h2 className="text-lg font-semibold mb-4">Current Status: {selectedSender}</h2>
            <div className="grid grid-cols-4 gap-4">
              <div className="bg-blue-50 p-4 rounded">
                <p className="text-sm text-gray-600">Current Day</p>
                <p className="text-2xl font-bold text-blue-600">{status.current_day}</p>
              </div>
              <div className="bg-green-50 p-4 rounded">
                <p className="text-sm text-gray-600">Current Limit</p>
                <p className="text-2xl font-bold text-green-600">{status.current_limit}/day</p>
              </div>
              <div className="bg-purple-50 p-4 rounded">
                <p className="text-sm text-gray-600">Max Limit</p>
                <p className="text-2xl font-bold text-purple-600">{status.max_limit}/day</p>
              </div>
              <div className="bg-orange-50 p-4 rounded">
                <p className="text-sm text-gray-600">Status</p>
                <Badge variant={status.enabled ? 'success' : 'default'}>
                  {status.enabled ? 'Active' : 'Inactive'}
                </Badge>
              </div>
            </div>
          </div>
        </Card>
      )}

      <div className="grid gap-4">
        {loading && configs.length === 0 ? (
          <Card>
            <div className="p-6 text-center text-gray-500">Loading configurations...</div>
          </Card>
        ) : configs.length === 0 ? (
          <Card>
            <div className="p-6 text-center text-gray-500">No warm-up configurations yet</div>
          </Card>
        ) : (
          configs.map((config) => (
            <Card key={config.id}>
              <div className="p-6">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="font-semibold text-lg">{config.sender_jid}</h3>
                    <p className="text-sm text-gray-600">Started: {new Date(config.start_date).toLocaleDateString()}</p>
                  </div>
                  <Badge variant={config.enabled ? 'success' : 'default'}>
                    {config.enabled ? 'Active' : 'Inactive'}
                  </Badge>
                </div>

                <div className="grid grid-cols-5 gap-4 mb-4">
                  <div>
                    <p className="text-xs text-gray-600">Current Day</p>
                    <p className="text-lg font-semibold">{config.current_day}</p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-600">Starting Limit</p>
                    <p className="text-lg font-semibold">{config.daily_limit}</p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-600">Increment</p>
                    <p className="text-lg font-semibold">+{config.increment_amount}/{config.increment_days}d</p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-600">Max Limit</p>
                    <p className="text-lg font-semibold">{config.max_daily_limit}</p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-600">Projected (30d)</p>
                    <p className="text-lg font-semibold">{calculateProjectedLimit(config, 30)}</p>
                  </div>
                </div>

                <div className="bg-gray-50 p-3 rounded mb-4 text-sm">
                  <div className="flex items-center space-x-2 mb-2">
                    <TrendingUp className="w-4 h-4 text-blue-600" />
                    <span className="font-semibold">Growth Schedule</span>
                  </div>
                  <p className="text-gray-700">
                    Increases by {config.increment_amount} messages every {config.increment_days} days until reaching {config.max_daily_limit} messages/day
                  </p>
                </div>

                <div className="flex space-x-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => loadStatus(config.sender_jid)}
                  >
                    View Status
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => handleEdit(config)}
                  >
                    <Edit2 className="w-4 h-4" />
                  </Button>
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => handleDelete(config.sender_jid)}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  )
}
