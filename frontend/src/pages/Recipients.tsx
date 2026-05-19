import { useState } from 'react'
import { Upload, UserCheck, Trash2 } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'
import Input from '@/components/common/Input'
import Badge from '@/components/common/Badge'

interface Recipient {
  phone: string
  name?: string
  isValid?: boolean
  lastChecked?: Date
}

export default function Recipients() {
  const [recipients, setRecipients] = useState<Recipient[]>([])
  const [newPhone, setNewPhone] = useState('')
  const [newName, setNewName] = useState('')
  const [isValidating, setIsValidating] = useState(false)

  const handleAddRecipient = () => {
    if (newPhone.trim()) {
      setRecipients([
        ...recipients,
        {
          phone: newPhone.trim(),
          name: newName.trim() || undefined,
        },
      ])
      setNewPhone('')
      setNewName('')
    }
  }

  const handleRemoveRecipient = (phone: string) => {
    setRecipients(recipients.filter(r => r.phone !== phone))
  }

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (e) => {
      const text = e.target?.result as string
      const lines = text.split('\n')
      const newRecipients: Recipient[] = []

      lines.forEach(line => {
        const trimmed = line.trim()
        if (trimmed) {
          // Support CSV format: phone,name
          const parts = trimmed.split(',')
          newRecipients.push({
            phone: parts[0].trim(),
            name: parts[1]?.trim(),
          })
        }
      })

      setRecipients([...recipients, ...newRecipients])
    }
    reader.readAsText(file)
  }

  const handleValidateAll = async () => {
    setIsValidating(true)
    // Simulate validation - in real app, call API
    setTimeout(() => {
      setRecipients(recipients.map(r => ({
        ...r,
        isValid: Math.random() > 0.2, // 80% valid rate
        lastChecked: new Date(),
      })))
      setIsValidating(false)
    }, 2000)
  }

  const handleExport = () => {
    const csv = recipients.map(r => `${r.phone},${r.name || ''}`).join('\n')
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'recipients.csv'
    a.click()
  }

  const validCount = recipients.filter(r => r.isValid).length
  const invalidCount = recipients.filter(r => r.isValid === false).length
  const uncheckedCount = recipients.filter(r => r.isValid === undefined).length

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Recipients</h1>
        <p className="text-gray-600 mt-1">Manage your recipient list and validate phone numbers</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <div className="text-center">
            <p className="text-sm text-gray-600">Total</p>
            <p className="text-2xl font-bold text-gray-900 mt-1">{recipients.length}</p>
          </div>
        </Card>
        <Card>
          <div className="text-center">
            <p className="text-sm text-gray-600">Valid</p>
            <p className="text-2xl font-bold text-green-600 mt-1">{validCount}</p>
          </div>
        </Card>
        <Card>
          <div className="text-center">
            <p className="text-sm text-gray-600">Invalid</p>
            <p className="text-2xl font-bold text-red-600 mt-1">{invalidCount}</p>
          </div>
        </Card>
        <Card>
          <div className="text-center">
            <p className="text-sm text-gray-600">Unchecked</p>
            <p className="text-2xl font-bold text-gray-600 mt-1">{uncheckedCount}</p>
          </div>
        </Card>
      </div>

      {/* Add Recipients */}
      <Card title="Add Recipients">
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Input
              placeholder="Phone number (e.g., 6281234567890)"
              value={newPhone}
              onChange={(e) => setNewPhone(e.target.value)}
            />
            <Input
              placeholder="Name (optional)"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
            />
            <Button onClick={handleAddRecipient} disabled={!newPhone.trim()}>
              Add Recipient
            </Button>
          </div>

          <div className="flex items-center space-x-4">
            <div className="flex-1 border-t border-gray-300"></div>
            <span className="text-sm text-gray-500">OR</span>
            <div className="flex-1 border-t border-gray-300"></div>
          </div>

          <div className="flex space-x-3">
            <label className="flex-1">
              <input
                type="file"
                accept=".txt,.csv"
                onChange={handleFileUpload}
                className="hidden"
              />
              <Button variant="secondary" className="w-full" onClick={() => (document.querySelector('input[type="file"]') as HTMLInputElement)?.click()}>
                <Upload className="w-4 h-4 mr-2" />
                Upload CSV/TXT
              </Button>
            </label>
            <Button variant="secondary" onClick={handleExport} disabled={recipients.length === 0}>
              Export CSV
            </Button>
          </div>
        </div>
      </Card>

      {/* Recipients List */}
      <Card
        title="Recipients List"
        action={
          <Button
            size="sm"
            onClick={handleValidateAll}
            isLoading={isValidating}
            disabled={recipients.length === 0}
          >
            <UserCheck className="w-4 h-4 mr-2" />
            Validate All
          </Button>
        }
      >
        {recipients.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No recipients added yet
          </div>
        ) : (
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {recipients.map((recipient, index) => (
              <div
                key={index}
                className="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <div className="flex-1">
                  <p className="font-medium text-gray-900">{recipient.phone}</p>
                  {recipient.name && (
                    <p className="text-sm text-gray-600">{recipient.name}</p>
                  )}
                </div>
                <div className="flex items-center space-x-3">
                  {recipient.isValid !== undefined && (
                    <Badge variant={recipient.isValid ? 'success' : 'error'}>
                      {recipient.isValid ? 'Valid' : 'Invalid'}
                    </Badge>
                  )}
                  <button
                    onClick={() => handleRemoveRecipient(recipient.phone)}
                    className="text-red-500 hover:text-red-700 transition-colors"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
