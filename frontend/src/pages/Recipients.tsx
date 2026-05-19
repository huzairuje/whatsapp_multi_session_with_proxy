import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Upload, UserCheck, Trash2, Users } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'
import Input from '@/components/common/Input'
import Badge from '@/components/common/Badge'
import Modal from '@/components/common/Modal'
import { sessionApi, userApi, contactsApi } from '@/services/api'

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
  const [selectedSender, setSelectedSender] = useState<string>('')
  const [showContactPicker, setShowContactPicker] = useState(false)
  const [contacts, setContacts] = useState<Contact[]>([])
  const [loadingContacts, setLoadingContacts] = useState(false)
  const [contactSearch, setContactSearch] = useState('')
  const [pickerPage, setPickerPage] = useState(1)
  const pickerPageSize = 40
  const [totalPickerContacts, setTotalPickerContacts] = useState(0)

  const { data: devices } = useQuery({
    queryKey: ['devices'],
    queryFn: async () => {
      const response = await sessionApi.getAll()
      return response.data
    },
    refetchInterval: 10000,
  })

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

  const loadContactsForPicker = async (page: number = 1) => {
    if (!selectedSender) {
      alert('Please select a WhatsApp session first')
      return
    }

    try {
      setLoadingContacts(true)
      const senderJID = `${selectedSender}@s.whatsapp.net`
      const offset = (page - 1) * pickerPageSize
      const response = await contactsApi.getContacts(senderJID, pickerPageSize, offset)
      setContacts(response.data.contacts || [])
      setTotalPickerContacts(response.data.total || 0)
      setPickerPage(page)
      setShowContactPicker(true)
    } catch (error) {
      console.error('Failed to load contacts:', error)
      alert('Failed to load contacts. Make sure you have synced contacts first.')
    } finally {
      setLoadingContacts(false)
    }
  }

  const handleAddContactAsRecipient = (contact: Contact) => {
    const phone = contact.contact_jid.split('@')[0]
    const name = contact.contact_name || contact.push_name || contact.full_name

    if (!recipients.find(r => r.phone === phone)) {
      setRecipients([
        ...recipients,
        {
          phone,
          name,
        },
      ])
    }
    setShowContactPicker(false)
    setContactSearch('')
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
    if (!selectedSender) {
      alert('Please select a WhatsApp session first')
      return
    }

    if (recipients.length === 0) {
      alert('No recipients to validate')
      return
    }

    setIsValidating(true)
    try {
      // Filter out LID numbers using the helper function
      const validPhones = recipients.filter(r => !isLIDNumber(r.phone))
      const lidPhones = recipients.filter(r => isLIDNumber(r.phone))

      if (validPhones.length === 0) {
        alert('No regular phone numbers to validate. All numbers appear to be LID/business contacts.')
        return
      }

      if (lidPhones.length > 0) {
        alert(`Skipping ${lidPhones.length} LID/business contacts (cannot validate). Validating ${validPhones.length} regular numbers.`)
      }

      const phoneNumbers = validPhones.map(r => r.phone)
      const response = await userApi.checkUser(selectedSender, phoneNumbers)

      // Update recipients with validation results
      const resultsMap = new Map(response.data.map((result: any) => [result.JID.split('@')[0], result]))

      setRecipients(recipients.map(r => {
        if (lidPhones.find(lid => lid.phone === r.phone)) {
          return { ...r, isValid: undefined, lastChecked: new Date() }
        }
        const result = resultsMap.get(r.phone)
        return {
          ...r,
          isValid: result?.IsIn || false,
          lastChecked: new Date(),
        }
      }))
    } catch (error: any) {
      console.error('Validation error:', error)
      const errorMsg = error?.response?.data?.error || error?.message || JSON.stringify(error)
      alert(`Validation failed: ${errorMsg}`)
    } finally {
      setIsValidating(false)
    }
  }

  const isLIDNumber = (phone: string): boolean => {
    // LID numbers: contain @, or >13 digits, or don't start with valid country code
    if (phone.includes('@')) return true
    if (phone.length > 13) return true
    // Check if starts with common country codes (62 for Indonesia, 1 for US, etc.)
    const validPrefixes = ['1', '7', '20', '27', '30', '31', '32', '33', '34', '36', '39', '40', '41', '43', '44', '45', '46', '47', '48', '49', '51', '52', '53', '54', '55', '56', '57', '58', '60', '61', '62', '63', '64', '65', '66', '81', '82', '84', '86', '90', '91', '92', '93', '94', '95', '98']
    const hasValidPrefix = validPrefixes.some(prefix => phone.startsWith(prefix))
    return !hasValidPrefix
  }

  const handleValidateSingle = async (phone: string) => {
    if (!selectedSender) {
      alert('Please select a WhatsApp session first')
      return
    }

    // Check if it's a LID number
    if (isLIDNumber(phone)) {
      alert('Cannot validate LID/business numbers. This appears to be a linked device ID, not a regular phone number.')
      return
    }

    try {
      const response = await userApi.checkUserSingle(selectedSender, phone)

      // Update the specific recipient with validation result
      setRecipients(recipients.map(r => {
        if (r.phone === phone) {
          return {
            ...r,
            isValid: response.data.IsIn || false,
            lastChecked: new Date(),
          }
        }
        return r
      }))
    } catch (error: any) {
      console.error('Validation error:', error)
      const errorMsg = error?.response?.data?.error || error?.message || JSON.stringify(error)
      alert(`Validation failed: ${errorMsg}`)
    }
  }

  const handleExport = () => {
    if (recipients.length === 0) {
      alert('No recipients to export')
      return
    }

    // CSV headers
    const headers = 'number,name,is_on_whatsapp'

    // CSV rows
    const rows = recipients.map(r => {
      const isOnWhatsapp = r.isValid !== undefined ? (r.isValid ? 'true' : 'false') : 'not_checked'
      return `${r.phone},${r.name || ''},${isOnWhatsapp}`
    })

    // Combine headers and rows
    const csv = [headers, ...rows].join('\n')

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `recipients_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    URL.revokeObjectURL(url)
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
          {/* Sender Selection */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Select WhatsApp Session for Validation
            </label>
            <select
              value={selectedSender}
              onChange={(e) => setSelectedSender(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="">Select a session...</option>
              {devices?.filter(d => d.isLoggedIn).map((device) => (
                <option key={device.user} value={device.user}>
                  +{device.user} {device.pushName ? `(${device.pushName})` : ''}
                </option>
              ))}
            </select>
            {(!devices || devices.filter(d => d.isLoggedIn).length === 0) && (
              <p className="text-xs text-red-600 mt-1">
                No active sessions available. Please connect a session first.
              </p>
            )}
          </div>

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
            <Button 
              variant="secondary" 
              onClick={() => loadContactsForPicker(1)}
              isLoading={loadingContacts}
              disabled={!selectedSender}
              className="flex-1"
            >
              <Users className="w-4 h-4 mr-2" />
              Pick from Contacts
            </Button>
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
                    onClick={() => handleValidateSingle(recipient.phone)}
                    disabled={!selectedSender}
                    className="text-blue-500 hover:text-blue-700 transition-colors disabled:text-gray-400 disabled:cursor-not-allowed"
                    title={!selectedSender ? 'Select a session first' : 'Validate this number'}
                  >
                    <UserCheck className="w-4 h-4" />
                  </button>
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

      {/* Contact Picker Modal */}
      <Modal
        isOpen={showContactPicker}
        onClose={() => {
          setShowContactPicker(false)
          setContactSearch('')
          setPickerPage(1)
        }}
        title="Select Contacts"
      >
        <div className="space-y-3">
          <input
            type="text"
            placeholder="Search contacts..."
            value={contactSearch}
            onChange={(e) => setContactSearch(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
          />

          <div className="flex items-center justify-between text-xs text-gray-600">
            <div>
              {totalPickerContacts} contacts (Page {pickerPage} of {Math.ceil(totalPickerContacts / pickerPageSize)})
            </div>
            <div className="flex gap-1">
              <Button
                size="sm"
                variant="secondary"
                onClick={() => loadContactsForPicker(pickerPage - 1)}
                disabled={pickerPage === 1 || loadingContacts}
              >
                Prev
              </Button>
              <Button
                size="sm"
                variant="secondary"
                onClick={() => loadContactsForPicker(pickerPage + 1)}
                disabled={pickerPage >= Math.ceil(totalPickerContacts / pickerPageSize) || loadingContacts}
              >
                Next
              </Button>
            </div>
          </div>

          <div className="max-h-96 overflow-y-auto">
            <div className="grid grid-cols-4 gap-1">
              {contacts
                .filter(
                  (contact) =>
                    contact.contact_name?.toLowerCase().includes(contactSearch.toLowerCase()) ||
                    contact.push_name?.toLowerCase().includes(contactSearch.toLowerCase()) ||
                    contact.contact_jid?.includes(contactSearch)
                )
                .map((contact) => {
                  const phone = contact.contact_jid.split('@')[0]
                  const isAdded = recipients.some((r) => r.phone === phone)
                  const displayName = contact.contact_name || contact.push_name || contact.full_name || phone

                  return (
                    <button
                      key={contact.id}
                      onClick={() => !isAdded && handleAddContactAsRecipient(contact)}
                      disabled={isAdded}
                      className={`p-1.5 rounded border text-center transition-colors ${
                        isAdded
                          ? 'bg-gray-100 border-gray-300 cursor-not-allowed opacity-50'
                          : 'bg-white border-gray-200 hover:border-blue-500 hover:bg-blue-50 cursor-pointer'
                      }`}
                    >
                      <p className="font-medium text-gray-900 truncate text-xs leading-tight">{displayName}</p>
                      <p className="text-gray-600 truncate text-xs leading-tight">+{phone}</p>
                    </button>
                  )
                })}
            </div>

            {contacts.filter(
              (contact) =>
                contact.contact_name?.toLowerCase().includes(contactSearch.toLowerCase()) ||
                contact.push_name?.toLowerCase().includes(contactSearch.toLowerCase()) ||
                contact.contact_jid?.includes(contactSearch)
            ).length === 0 && (
              <div className="text-center py-6 text-gray-500 text-sm">
                {contacts.length === 0 ? 'No contacts available. Sync contacts first.' : 'No contacts match your search.'}
              </div>
            )}
          </div>
        </div>
      </Modal>
    </div>
  )
}
