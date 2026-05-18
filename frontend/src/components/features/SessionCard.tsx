import { useState } from 'react'
import { Smartphone, Power, LogOut, QrCode, Send, Activity } from 'lucide-react'
import Badge from '../common/Badge'
import Button from '../common/Button'
import Card from '../common/Card'
import type { Device } from '@/types'

interface SessionCardProps {
  device: Device
  onConnect: () => void
  onDisconnect: () => void
  onLogout: () => void
  onShowQR: () => void
  onSend: () => void
  dailyCount?: number
  dailyLimit?: number
}

export default function SessionCard({
  device,
  onConnect,
  onDisconnect,
  onLogout,
  onShowQR,
  onSend,
  dailyCount = 0,
  dailyLimit = 50,
}: SessionCardProps) {
  const [isLoading, setIsLoading] = useState(false)

  const handleAction = async (action: () => void) => {
    setIsLoading(true)
    try {
      await action()
    } finally {
      setIsLoading(false)
    }
  }

  const getHealthStatus = () => {
    const percentage = (dailyCount / dailyLimit) * 100
    if (percentage >= 90) return { variant: 'error' as const, label: 'Critical' }
    if (percentage >= 70) return { variant: 'warning' as const, label: 'Warning' }
    return { variant: 'success' as const, label: 'Healthy' }
  }

  const healthStatus = getHealthStatus()

  return (
    <Card className="hover:shadow-md transition-shadow">
      <div className="space-y-4">
        {/* Header */}
        <div className="flex items-start justify-between">
          <div className="flex items-center space-x-3">
            <div className={`p-2 rounded-lg ${device.isLoggedIn ? 'bg-green-100' : 'bg-gray-100'}`}>
              <Smartphone className={`w-5 h-5 ${device.isLoggedIn ? 'text-green-600' : 'text-gray-400'}`} />
            </div>
            <div>
              <h3 className="font-semibold text-gray-900">+{device.user}</h3>
              <p className="text-sm text-gray-500">{device.pushName || 'No name'}</p>
            </div>
          </div>
          <div className="flex flex-col items-end space-y-2">
            <Badge variant={device.isLoggedIn ? 'success' : 'error'}>
              {device.isLoggedIn ? 'Connected' : 'Disconnected'}
            </Badge>
            <Badge variant={healthStatus.variant} size="sm">
              {healthStatus.label}
            </Badge>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 gap-4 py-3 border-t border-b border-gray-100">
          <div>
            <p className="text-xs text-gray-500">Daily Sends</p>
            <p className="text-lg font-semibold text-gray-900">
              {dailyCount} / {dailyLimit}
            </p>
          </div>
          <div>
            <p className="text-xs text-gray-500">Platform</p>
            <p className="text-lg font-semibold text-gray-900">{device.platform || 'Unknown'}</p>
          </div>
        </div>

        {/* Progress Bar */}
        <div>
          <div className="flex justify-between text-xs text-gray-600 mb-1">
            <span>Usage</span>
            <span>{Math.round((dailyCount / dailyLimit) * 100)}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`h-2 rounded-full transition-all ${
                dailyCount >= dailyLimit * 0.9
                  ? 'bg-red-500'
                  : dailyCount >= dailyLimit * 0.7
                  ? 'bg-yellow-500'
                  : 'bg-green-500'
              }`}
              style={{ width: `${Math.min((dailyCount / dailyLimit) * 100, 100)}%` }}
            />
          </div>
        </div>

        {/* Actions */}
        <div className="flex flex-wrap gap-2">
          {!device.isLoggedIn ? (
            <>
              <Button
                size="sm"
                onClick={() => handleAction(onConnect)}
                isLoading={isLoading}
              >
                <Power className="w-4 h-4 mr-1" />
                Connect
              </Button>
              <Button
                size="sm"
                variant="secondary"
                onClick={() => handleAction(onShowQR)}
                isLoading={isLoading}
              >
                <QrCode className="w-4 h-4 mr-1" />
                QR Code
              </Button>
            </>
          ) : (
            <>
              <Button
                size="sm"
                onClick={() => handleAction(onSend)}
                isLoading={isLoading}
              >
                <Send className="w-4 h-4 mr-1" />
                Send
              </Button>
              <Button
                size="sm"
                variant="secondary"
                onClick={() => handleAction(onDisconnect)}
                isLoading={isLoading}
              >
                <Power className="w-4 h-4 mr-1" />
                Disconnect
              </Button>
              <Button
                size="sm"
                variant="danger"
                onClick={() => handleAction(onLogout)}
                isLoading={isLoading}
              >
                <LogOut className="w-4 h-4 mr-1" />
                Logout
              </Button>
            </>
          )}
        </div>
      </div>
    </Card>
  )
}
