import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Smartphone, Copy, CheckCircle, AlertCircle, Loader2 } from 'lucide-react'
import Card from '@/components/common/Card'
import Input from '@/components/common/Input'
import Button from '@/components/common/Button'

interface PairCodeResponse {
  pair_code?: string
  message?: string
}

export default function PairCodeSession() {
  const [phoneNumber, setPhoneNumber] = useState('')
  const [pairCode, setPairCode] = useState('')
  const [copied, setCopied] = useState(false)
  const queryClient = useQueryClient()

  const generatePairCode = useMutation({
    mutationFn: async (phone: string) => {
      const response = await fetch(`/api/pair-code?sender=${phone}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      })
      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.message || 'Failed to generate pair code')
      }
      return response.json() as Promise<PairCodeResponse>
    },
    onSuccess: (data) => {
      if (data.pair_code) {
        setPairCode(data.pair_code)
        // Invalidate activities to show the new pair code generation activity
        queryClient.invalidateQueries({ queryKey: ['activities'] })
      }
    },
  })

  const handleGeneratePairCode = (e: React.FormEvent) => {
    e.preventDefault()
    if (!phoneNumber.trim()) {
      return
    }
    setPairCode('')
    setCopied(false)
    generatePairCode.mutate(phoneNumber.trim())
  }

  const handleCopyPairCode = async () => {
    if (pairCode) {
      await navigator.clipboard.writeText(pairCode)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const formatPairCode = (code: string) => {
    // Format as XXX-XXX for better readability
    if (code.length === 8) {
      return `${code.slice(0, 4)}-${code.slice(4)}`
    }
    return code
  }

  return (
    <Card
      title="Pair Code Session"
      subtitle="Generate pair code for WhatsApp Web linking"
    >
      <div className="space-y-6">
        {/* Generate Pair Code Form */}
        <form onSubmit={handleGeneratePairCode} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Phone Number
            </label>
            <Input
              type="text"
              placeholder="e.g., 6281234567890"
              value={phoneNumber}
              onChange={(e) => setPhoneNumber(e.target.value)}
              disabled={generatePairCode.isPending}
              className="w-full"
            />
            <p className="text-xs text-gray-500 mt-1">
              Enter phone number with country code (without +)
            </p>
          </div>

          <Button
            type="submit"
            disabled={!phoneNumber.trim() || generatePairCode.isPending}
            className="w-full"
          >
            {generatePairCode.isPending ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Generating...
              </>
            ) : (
              <>
                <Smartphone className="w-4 h-4 mr-2" />
                Generate Pair Code
              </>
            )}
          </Button>
        </form>

        {/* Error Display */}
        {generatePairCode.isError && (
          <div className="p-4 bg-red-50 border border-red-200 rounded-lg flex items-start space-x-3">
            <AlertCircle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm font-medium text-red-800">Error</p>
              <p className="text-sm text-red-600 mt-1">
                {generatePairCode.error?.message || 'Failed to generate pair code'}
              </p>
            </div>
          </div>
        )}

        {/* Pair Code Display */}
        {pairCode && (
          <div className="p-6 bg-gradient-to-br from-blue-50 to-indigo-50 border-2 border-blue-200 rounded-lg">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-2">
                <CheckCircle className="w-5 h-5 text-green-500" />
                <span className="text-sm font-medium text-gray-700">
                  Pair Code Generated
                </span>
              </div>
              <span className="text-xs text-gray-500">
                Valid for 20 seconds
              </span>
            </div>

            <div className="bg-white rounded-lg p-4 mb-4">
              <div className="text-center">
                <p className="text-xs text-gray-500 mb-2">Your Pair Code</p>
                <p className="text-4xl font-bold text-gray-900 tracking-wider font-mono">
                  {formatPairCode(pairCode)}
                </p>
              </div>
            </div>

            <div className="flex space-x-2">
              <Button
                onClick={handleCopyPairCode}
                variant="secondary"
                className="flex-1"
              >
                {copied ? (
                  <>
                    <CheckCircle className="w-4 h-4 mr-2" />
                    Copied!
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4 mr-2" />
                    Copy Code
                  </>
                )}
              </Button>
              <Button
                onClick={() => handleGeneratePairCode(new Event('submit') as any)}
                variant="secondary"
                className="flex-1"
              >
                Generate New
              </Button>
            </div>

            {/* Instructions */}
            <div className="mt-4 p-3 bg-blue-50 rounded-lg">
              <p className="text-xs font-medium text-blue-900 mb-2">
                How to use:
              </p>
              <ol className="text-xs text-blue-800 space-y-1 list-decimal list-inside">
                <li>Open WhatsApp on your phone</li>
                <li>Go to Settings → Linked Devices</li>
                <li>Tap "Link a Device"</li>
                <li>Tap "Link with phone number instead"</li>
                <li>Enter the pair code above</li>
              </ol>
            </div>
          </div>
        )}

        {/* Info Box */}
        {!pairCode && !generatePairCode.isError && (
          <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
            <p className="text-sm text-gray-600">
              <strong>Note:</strong> Pair codes are an alternative to QR codes for linking
              WhatsApp Web. Each code is valid for 20 seconds and can only be used once.
            </p>
          </div>
        )}
      </div>
    </Card>
  )
}
