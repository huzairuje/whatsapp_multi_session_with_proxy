import { useState } from 'react'
import { Save, RefreshCw } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'
import Input from '@/components/common/Input'
import Badge from '@/components/common/Badge'

export default function Settings() {
  const [settings, setSettings] = useState({
    minDelay: 20000,
    maxDelay: 60000,
    batchSize: 8,
    batchPauseMin: 360,
    batchPauseMax: 720,
    dailyLimit: 40,
    typingDelayMin: 2500,
    typingDelayMax: 6000,
    enablePresenceSimulation: true,
    allowedHourStart: 8,
    allowedHourEnd: 22,
    timezone: 'Asia/Jakarta',
    enableTimeRestrictions: true,
    errorBackoffMinutes: 30,
    enableRecipientValidation: true,
    validationCacheDuration: 24,
    enableHealthCheck: true,
    maxErrorRate: 0.3,
  })

  const [isSaving, setIsSaving] = useState(false)
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const handleSave = async () => {
    setIsSaving(true)
    setSaveMessage(null)

    // Simulate API call
    setTimeout(() => {
      setIsSaving(false)
      setSaveMessage({ type: 'success', text: 'Settings saved successfully!' })
      setTimeout(() => setSaveMessage(null), 3000)
    }, 1000)
  }

  const handleReset = () => {
    if (confirm('Are you sure you want to reset to default settings?')) {
      setSettings({
        minDelay: 15000,
        maxDelay: 45000,
        batchSize: 10,
        batchPauseMin: 300,
        batchPauseMax: 600,
        dailyLimit: 50,
        typingDelayMin: 2000,
        typingDelayMax: 5000,
        enablePresenceSimulation: true,
        allowedHourStart: 8,
        allowedHourEnd: 22,
        timezone: 'Local',
        enableTimeRestrictions: true,
        errorBackoffMinutes: 30,
        enableRecipientValidation: true,
        validationCacheDuration: 24,
        enableHealthCheck: true,
        maxErrorRate: 0.3,
      })
      setSaveMessage({ type: 'success', text: 'Settings reset to defaults' })
      setTimeout(() => setSaveMessage(null), 3000)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
          <p className="text-gray-600 mt-1">Configure anti-ban and bulk send parameters</p>
        </div>
        <div className="flex space-x-3">
          <Button variant="secondary" onClick={handleReset}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Reset to Defaults
          </Button>
          <Button onClick={handleSave} isLoading={isSaving}>
            <Save className="w-4 h-4 mr-2" />
            Save Changes
          </Button>
        </div>
      </div>

      {saveMessage && (
        <div
          className={`rounded-lg p-4 ${
            saveMessage.type === 'success'
              ? 'bg-green-50 border border-green-200 text-green-800'
              : 'bg-red-50 border border-red-200 text-red-800'
          }`}
        >
          {saveMessage.text}
        </div>
      )}

      {/* Message Timing */}
      <Card title="Message Timing" subtitle="Control delays between messages">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            label="Min Delay (ms)"
            type="number"
            value={settings.minDelay}
            onChange={(e) => setSettings({ ...settings, minDelay: parseInt(e.target.value) })}
            helperText="Minimum delay between messages (15-30 seconds recommended)"
          />
          <Input
            label="Max Delay (ms)"
            type="number"
            value={settings.maxDelay}
            onChange={(e) => setSettings({ ...settings, maxDelay: parseInt(e.target.value) })}
            helperText="Maximum delay between messages (45-60 seconds recommended)"
          />
          <Input
            label="Typing Delay Min (ms)"
            type="number"
            value={settings.typingDelayMin}
            onChange={(e) => setSettings({ ...settings, typingDelayMin: parseInt(e.target.value) })}
            helperText="Minimum typing simulation duration"
          />
          <Input
            label="Typing Delay Max (ms)"
            type="number"
            value={settings.typingDelayMax}
            onChange={(e) => setSettings({ ...settings, typingDelayMax: parseInt(e.target.value) })}
            helperText="Maximum typing simulation duration"
          />
        </div>
        <div className="mt-4">
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={settings.enablePresenceSimulation}
              onChange={(e) => setSettings({ ...settings, enablePresenceSimulation: e.target.checked })}
              className="rounded border-gray-300 text-primary focus:ring-primary"
            />
            <span className="text-sm text-gray-700">Enable typing presence simulation</span>
          </label>
        </div>
      </Card>

      {/* Batch Settings */}
      <Card title="Batch Settings" subtitle="Configure batch pauses">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            label="Batch Size"
            type="number"
            value={settings.batchSize}
            onChange={(e) => setSettings({ ...settings, batchSize: parseInt(e.target.value) })}
            helperText="Messages per batch before taking a longer pause"
          />
          <Input
            label="Daily Limit"
            type="number"
            value={settings.dailyLimit}
            onChange={(e) => setSettings({ ...settings, dailyLimit: parseInt(e.target.value) })}
            helperText="Maximum messages per sender per day"
          />
          <Input
            label="Batch Pause Min (seconds)"
            type="number"
            value={settings.batchPauseMin}
            onChange={(e) => setSettings({ ...settings, batchPauseMin: parseInt(e.target.value) })}
            helperText="Minimum pause duration after each batch"
          />
          <Input
            label="Batch Pause Max (seconds)"
            type="number"
            value={settings.batchPauseMax}
            onChange={(e) => setSettings({ ...settings, batchPauseMax: parseInt(e.target.value) })}
            helperText="Maximum pause duration after each batch"
          />
        </div>
      </Card>

      {/* Time Restrictions */}
      <Card title="Time Restrictions" subtitle="Control when messages can be sent">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Input
            label="Start Hour (0-23)"
            type="number"
            min="0"
            max="23"
            value={settings.allowedHourStart}
            onChange={(e) => setSettings({ ...settings, allowedHourStart: parseInt(e.target.value) })}
            helperText="Earliest hour to send messages"
          />
          <Input
            label="End Hour (0-23)"
            type="number"
            min="0"
            max="23"
            value={settings.allowedHourEnd}
            onChange={(e) => setSettings({ ...settings, allowedHourEnd: parseInt(e.target.value) })}
            helperText="Latest hour to send messages"
          />
          <Input
            label="Timezone"
            value={settings.timezone}
            onChange={(e) => setSettings({ ...settings, timezone: e.target.value })}
            helperText="e.g., Asia/Jakarta, Local"
          />
        </div>
        <div className="mt-4">
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={settings.enableTimeRestrictions}
              onChange={(e) => setSettings({ ...settings, enableTimeRestrictions: e.target.checked })}
              className="rounded border-gray-300 text-primary focus:ring-primary"
            />
            <span className="text-sm text-gray-700">Enable time-of-day restrictions</span>
          </label>
        </div>
      </Card>

      {/* Error Handling */}
      <Card title="Error Handling" subtitle="Configure error detection and backoff">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            label="Error Backoff (minutes)"
            type="number"
            value={settings.errorBackoffMinutes}
            onChange={(e) => setSettings({ ...settings, errorBackoffMinutes: parseInt(e.target.value) })}
            helperText="Pause duration after rate limit detection"
          />
          <Input
            label="Max Error Rate (0.0-1.0)"
            type="number"
            step="0.1"
            min="0"
            max="1"
            value={settings.maxErrorRate}
            onChange={(e) => setSettings({ ...settings, maxErrorRate: parseFloat(e.target.value) })}
            helperText="Maximum acceptable error rate (e.g., 0.3 = 30%)"
          />
        </div>
        <div className="mt-4">
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={settings.enableHealthCheck}
              onChange={(e) => setSettings({ ...settings, enableHealthCheck: e.target.checked })}
              className="rounded border-gray-300 text-primary focus:ring-primary"
            />
            <span className="text-sm text-gray-700">Enable session health checks before bulk send</span>
          </label>
        </div>
      </Card>

      {/* Validation */}
      <Card title="Recipient Validation" subtitle="Configure recipient validation">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            label="Cache Duration (hours)"
            type="number"
            value={settings.validationCacheDuration}
            onChange={(e) => setSettings({ ...settings, validationCacheDuration: parseInt(e.target.value) })}
            helperText="How long to cache validation results"
          />
        </div>
        <div className="mt-4">
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={settings.enableRecipientValidation}
              onChange={(e) => setSettings({ ...settings, enableRecipientValidation: e.target.checked })}
              className="rounded border-gray-300 text-primary focus:ring-primary"
            />
            <span className="text-sm text-gray-700">Enable recipient validation before bulk send</span>
          </label>
        </div>
      </Card>

      {/* Current Configuration Summary */}
      <Card title="Configuration Summary">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <p className="text-sm font-medium text-gray-700 mb-2">Anti-Ban Features</p>
            <div className="space-y-1">
              <Badge variant={settings.enablePresenceSimulation ? 'success' : 'default'}>
                Typing Simulation: {settings.enablePresenceSimulation ? 'ON' : 'OFF'}
              </Badge>
              <Badge variant={settings.enableTimeRestrictions ? 'success' : 'default'}>
                Time Restrictions: {settings.enableTimeRestrictions ? 'ON' : 'OFF'}
              </Badge>
              <Badge variant={settings.enableHealthCheck ? 'success' : 'default'}>
                Health Checks: {settings.enableHealthCheck ? 'ON' : 'OFF'}
              </Badge>
              <Badge variant={settings.enableRecipientValidation ? 'success' : 'default'}>
                Validation: {settings.enableRecipientValidation ? 'ON' : 'OFF'}
              </Badge>
            </div>
          </div>
          <div>
            <p className="text-sm font-medium text-gray-700 mb-2">Current Limits</p>
            <div className="space-y-1 text-sm text-gray-600">
              <p>• {settings.dailyLimit} messages per day per sender</p>
              <p>• {settings.minDelay / 1000}-{settings.maxDelay / 1000}s between messages</p>
              <p>• {settings.batchSize} messages per batch</p>
              <p>• {settings.batchPauseMin / 60}-{settings.batchPauseMax / 60} min batch pauses</p>
            </div>
          </div>
        </div>
      </Card>
    </div>
  )
}
