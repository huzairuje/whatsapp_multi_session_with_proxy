import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Send } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'
import BulkSendForm from '@/components/features/BulkSendForm'
import { sessionApi, messageApi } from '@/services/api'

export default function BulkSend() {
  const [selectedSender, setSelectedSender] = useState<string>('')

  const { data: devices, isLoading } = useQuery({
    queryKey: ['devices'],
    queryFn: async () => {
      const response = await sessionApi.getAll()
      return response.data
    },
  })

  const { data: bulkStatus } = useQuery({
    queryKey: ['bulk-status', selectedSender],
    queryFn: async () => {
      if (!selectedSender) return null
      const response = await messageApi.getBulkSendStatus(selectedSender)
      return response.data
    },
    enabled: !!selectedSender,
    refetchInterval: 5000,
  })

  const connectedDevices = devices?.filter(d => d.isLoggedIn) || []

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Bulk Send</h1>
        <p className="text-gray-600 mt-1">Send messages to multiple recipients with anti-ban protection</p>
      </div>

      {/* Sender Selection */}
      <Card title="Select Sender">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        ) : connectedDevices.length === 0 ? (
          <div className="text-center py-8">
            <Send className="w-12 h-12 text-gray-400 mx-auto mb-3" />
            <p className="text-gray-600">No connected sessions available</p>
            <p className="text-sm text-gray-500 mt-1">Please connect a session first</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {connectedDevices.map((device) => (
              <button
                key={device.user}
                onClick={() => setSelectedSender(device.user)}
                className={`p-4 rounded-lg border-2 transition-all text-left ${
                  selectedSender === device.user
                    ? 'border-primary bg-primary/5'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <p className="font-semibold text-gray-900">+{device.user}</p>
                <p className="text-sm text-gray-500">{device.pushName || 'No name'}</p>
                {bulkStatus && selectedSender === device.user && (
                  <p className="text-xs text-gray-600 mt-2">
                    {bulkStatus.daily_count} / {bulkStatus.daily_limit} sent today
                  </p>
                )}
              </button>
            ))}
          </div>
        )}
      </Card>

      {/* Bulk Send Form */}
      {selectedSender && (
        <>
          {bulkStatus && (
            <Card>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">Daily Usage</p>
                  <p className="text-2xl font-bold text-gray-900">
                    {bulkStatus.daily_count} / {bulkStatus.daily_limit}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-sm text-gray-600">Remaining</p>
                  <p className="text-2xl font-bold text-primary">{bulkStatus.remaining}</p>
                </div>
              </div>
              <div className="mt-4">
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full transition-all ${
                      bulkStatus.daily_count >= bulkStatus.daily_limit * 0.9
                        ? 'bg-red-500'
                        : bulkStatus.daily_count >= bulkStatus.daily_limit * 0.7
                        ? 'bg-yellow-500'
                        : 'bg-green-500'
                    }`}
                    style={{
                      width: `${Math.min((bulkStatus.daily_count / bulkStatus.daily_limit) * 100, 100)}%`,
                    }}
                  />
                </div>
              </div>
            </Card>
          )}

          <BulkSendForm sender={selectedSender} />
        </>
      )}
    </div>
  )
}
