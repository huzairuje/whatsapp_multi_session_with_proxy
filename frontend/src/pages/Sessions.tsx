import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, RefreshCw } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'
import Input from '@/components/common/Input'
import Modal from '@/components/common/Modal'
import SessionCard from '@/components/features/SessionCard'
import QRCodeDisplay from '@/components/features/QRCodeDisplay'
import { sessionApi, messageApi } from '@/services/api'

export default function Sessions() {
  const [qrModalOpen, setQrModalOpen] = useState(false)
  const [selectedSender, setSelectedSender] = useState('')
  const [phoneModalOpen, setPhoneModalOpen] = useState(false)
  const [phoneNumber, setPhoneNumber] = useState('')
  const queryClient = useQueryClient()

  const { data: devices, isLoading, refetch } = useQuery({
    queryKey: ['devices'],
    queryFn: async () => {
      const response = await sessionApi.getAll()
      return response.data
    },
    refetchInterval: 10000,
  })

  const { data: bulkStatuses } = useQuery({
    queryKey: ['bulk-statuses'],
    queryFn: async () => {
      if (!devices) return {}
      const statuses: Record<string, any> = {}
      await Promise.all(
        devices.map(async (device) => {
          try {
            const response = await messageApi.getBulkSendStatus(device.user)
            statuses[device.user] = response.data
          } catch (error) {
            statuses[device.user] = null
          }
        })
      )
      return statuses
    },
    enabled: !!devices && devices.length > 0,
  })

  const connectMutation = useMutation({
    mutationFn: (sender: string) => sessionApi.connect(sender),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] })
    },
  })

  const disconnectMutation = useMutation({
    mutationFn: (sender: string) => sessionApi.disconnect(sender),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] })
    },
  })

  const logoutMutation = useMutation({
    mutationFn: (sender: string) => sessionApi.logout(sender),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] })
    },
  })

  const handleShowQR = (sender: string) => {
    setSelectedSender(sender)
    setQrModalOpen(true)
  }

  const handleSend = (sender: string) => {
    // Navigate to bulk send page with pre-selected sender
    window.location.href = `/bulk-send?sender=${sender}`
  }

  const handleAddSession = () => {
    setPhoneNumber('')
    setPhoneModalOpen(true)
  }

  const handlePhoneSubmit = () => {
    if (phoneNumber.trim()) {
      setPhoneModalOpen(false)
      handleShowQR(phoneNumber.trim())
    }
  }

  const connectedCount = devices?.filter(d => d.isLoggedIn).length || 0
  const totalCount = devices?.length || 0

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Sessions</h1>
          <p className="text-gray-600 mt-1">
            {connectedCount} of {totalCount} sessions connected
          </p>
        </div>
        <div className="flex space-x-3">
          <Button variant="secondary" onClick={() => refetch()} isLoading={isLoading}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={handleAddSession}>
            <Plus className="w-4 h-4 mr-2" />
            Add Session
          </Button>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <div className="text-center">
            <p className="text-sm text-gray-600">Total Sessions</p>
            <p className="text-3xl font-bold text-gray-900 mt-2">{totalCount}</p>
          </div>
        </Card>
        <Card>
          <div className="text-center">
            <p className="text-sm text-gray-600">Connected</p>
            <p className="text-3xl font-bold text-green-600 mt-2">{connectedCount}</p>
          </div>
        </Card>
        <Card>
          <div className="text-center">
            <p className="text-sm text-gray-600">Disconnected</p>
            <p className="text-3xl font-bold text-red-600 mt-2">{totalCount - connectedCount}</p>
          </div>
        </Card>
      </div>

      {/* Sessions List */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      ) : !devices || devices.length === 0 ? (
        <Card>
          <div className="text-center py-12">
            <Plus className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">No sessions yet</h3>
            <p className="text-gray-600 mb-4">Add your first WhatsApp session to get started</p>
            <Button onClick={() => handleShowQR('new')}>
              <Plus className="w-4 h-4 mr-2" />
              Add Session
            </Button>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {devices.map((device) => (
            <SessionCard
              key={device.user}
              device={device}
              dailyCount={bulkStatuses?.[device.user]?.daily_count || 0}
              dailyLimit={bulkStatuses?.[device.user]?.daily_limit || 50}
              onConnect={() => connectMutation.mutate(device.user)}
              onDisconnect={() => disconnectMutation.mutate(device.user)}
              onLogout={() => logoutMutation.mutate(device.user)}
              onShowQR={() => handleShowQR(device.user)}
              onSend={() => handleSend(device.user)}
            />
          ))}
        </div>
      )}

      {/* Phone Number Input Modal */}
      <Modal
        isOpen={phoneModalOpen}
        onClose={() => setPhoneModalOpen(false)}
        title="Add New Session"
        footer={
          <>
            <Button variant="secondary" onClick={() => setPhoneModalOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handlePhoneSubmit}
              disabled={!phoneNumber.trim()}
            >
              Generate QR Code
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600">
            Enter the phone number for the new WhatsApp session (with country code, no + sign)
          </p>
          <Input
            label="Phone Number"
            placeholder="e.g., 6281234567890"
            value={phoneNumber}
            onChange={(e) => setPhoneNumber(e.target.value)}
            onKeyPress={(e) => {
              if (e.key === 'Enter' && phoneNumber.trim()) {
                handlePhoneSubmit()
              }
            }}
            helperText="Format: country code + phone number (e.g., 6281234567890 for Indonesia)"
          />
        </div>
      </Modal>

      {/* QR Code Modal */}
      <QRCodeDisplay
        sender={selectedSender}
        isOpen={qrModalOpen}
        onClose={() => {
          setQrModalOpen(false)
          setSelectedSender('')
        }}
      />
    </div>
  )
}
