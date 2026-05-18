import { useState, useEffect } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { RefreshCw, X } from 'lucide-react'
import Modal from '../common/Modal'
import Button from '../common/Button'
import { sessionApi } from '@/services/api'

interface QRCodeDisplayProps {
  sender: string
  isOpen: boolean
  onClose: () => void
}

export default function QRCodeDisplay({ sender, isOpen, onClose }: QRCodeDisplayProps) {
  const [qrCode, setQrCode] = useState<string>('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string>('')

  const fetchQRCode = async () => {
    setIsLoading(true)
    setError('')
    try {
      const response = await sessionApi.getQRJson(sender)
      if (response.data.data) {
        setQrCode(response.data.data)
      } else {
        setError('Already logged in or QR code not available')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch QR code')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (isOpen && sender) {
      fetchQRCode()
    }
  }, [isOpen, sender])

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Scan QR Code"
      size="md"
      footer={
        <>
          <Button variant="secondary" onClick={fetchQRCode} isLoading={isLoading}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button variant="ghost" onClick={onClose}>
            Close
          </Button>
        </>
      }
    >
      <div className="flex flex-col items-center space-y-4">
        {isLoading && (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
          </div>
        )}

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 w-full">
            <p className="text-red-800 text-sm">{error}</p>
          </div>
        )}

        {qrCode && !isLoading && !error && (
          <>
            <div className="bg-white p-4 rounded-lg border-2 border-gray-200">
              <QRCodeSVG value={qrCode} size={256} level="H" />
            </div>
            <div className="text-center">
              <p className="text-sm text-gray-600">
                Open WhatsApp on your phone and scan this QR code
              </p>
              <p className="text-xs text-gray-500 mt-2">
                Phone: {sender}
              </p>
            </div>
          </>
        )}
      </div>
    </Modal>
  )
}
