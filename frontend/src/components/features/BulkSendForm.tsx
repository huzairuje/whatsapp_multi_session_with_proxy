import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Send, Plus, X } from 'lucide-react'
import Button from '../common/Button'
import Input from '../common/Input'
import Card from '../common/Card'
import { messageApi } from '@/services/api'
import type { BulkSendRequest } from '@/types'

interface BulkSendFormProps {
  sender: string
  onSuccess?: () => void
}

interface FormData {
  message: string
  recipients: string
}

export default function BulkSendForm({ sender, onSuccess }: BulkSendFormProps) {
  const [variables, setVariables] = useState<Record<string, string>>({})
  const [newVarKey, setNewVarKey] = useState('')
  const [newVarValue, setNewVarValue] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [result, setResult] = useState<{ success: boolean; message: string } | null>(null)

  const { register, handleSubmit, formState: { errors }, watch, reset } = useForm<FormData>()

  const message = watch('message', '')
  const recipients = watch('recipients', '')

  const addVariable = () => {
    if (newVarKey && newVarValue) {
      setVariables({ ...variables, [newVarKey]: newVarValue })
      setNewVarKey('')
      setNewVarValue('')
    }
  }

  const removeVariable = (key: string) => {
    const newVars = { ...variables }
    delete newVars[key]
    setVariables(newVars)
  }

  const getRecipientCount = () => {
    if (!recipients) return 0
    return recipients.split('\n').filter(r => r.trim()).length
  }

  const getPreviewMessage = () => {
    let preview = message
    Object.entries(variables).forEach(([key, value]) => {
      preview = preview.replace(new RegExp(`{{${key}}}`, 'g'), value)
    })
    return preview
  }

  const onSubmit = async (data: FormData) => {
    setIsSubmitting(true)
    setResult(null)

    try {
      const recipientList = data.recipients
        .split('\n')
        .map(r => r.trim())
        .filter(r => r)

      const request: BulkSendRequest = {
        recipients: recipientList,
        message: data.message,
        variables: Object.keys(variables).length > 0 ? variables : undefined,
      }

      await messageApi.sendBulk(sender, request)

      setResult({
        success: true,
        message: `Bulk send started for ${recipientList.length} recipients. Messages are being sent sequentially with anti-ban delays.`,
      })

      reset()
      setVariables({})
      onSuccess?.()
    } catch (error) {
      setResult({
        success: false,
        message: error instanceof Error ? error.message : 'Failed to send bulk messages',
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Message Input */}
      <Card title="Message Content">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Message Template
            </label>
            <textarea
              {...register('message', { required: 'Message is required' })}
              rows={6}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
              placeholder="Enter your message here. Use {{variable}} for dynamic content."
            />
            {errors.message && (
              <p className="mt-1 text-sm text-red-600">{errors.message.message}</p>
            )}
            <p className="mt-1 text-xs text-gray-500">
              Use variables like {'{{name}}'}, {'{{phone}}'} for personalization
            </p>
          </div>

          {/* Variables */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Variables (Optional)
            </label>
            <div className="space-y-2">
              {Object.entries(variables).map(([key, value]) => (
                <div key={key} className="flex items-center space-x-2">
                  <div className="flex-1 bg-gray-50 px-3 py-2 rounded-lg flex items-center justify-between">
                    <span className="text-sm">
                      <span className="font-medium">{'{{' + key + '}}'}</span> = {value}
                    </span>
                    <button
                      type="button"
                      onClick={() => removeVariable(key)}
                      className="text-red-500 hover:text-red-700"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}

              <div className="flex space-x-2">
                <Input
                  placeholder="Variable name"
                  value={newVarKey}
                  onChange={(e) => setNewVarKey(e.target.value)}
                />
                <Input
                  placeholder="Value"
                  value={newVarValue}
                  onChange={(e) => setNewVarValue(e.target.value)}
                />
                <Button
                  type="button"
                  variant="secondary"
                  onClick={addVariable}
                  disabled={!newVarKey || !newVarValue}
                >
                  <Plus className="w-4 h-4" />
                </Button>
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* Recipients Input */}
      <Card title="Recipients">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Phone Numbers (one per line)
            </label>
            <textarea
              {...register('recipients', { required: 'At least one recipient is required' })}
              rows={8}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent font-mono text-sm"
              placeholder="6281234567890&#10;6289876543210&#10;6285555555555"
            />
            {errors.recipients && (
              <p className="mt-1 text-sm text-red-600">{errors.recipients.message}</p>
            )}
            <p className="mt-1 text-sm text-gray-600">
              Total recipients: <span className="font-semibold">{getRecipientCount()}</span>
            </p>
          </div>
        </div>
      </Card>

      {/* Preview */}
      {message && (
        <Card title="Message Preview">
          <div className="bg-gray-50 rounded-lg p-4">
            <p className="text-sm text-gray-700 whitespace-pre-wrap">{getPreviewMessage()}</p>
          </div>
        </Card>
      )}

      {/* Result */}
      {result && (
        <div
          className={`rounded-lg p-4 ${
            result.success ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'
          }`}
        >
          <p className={`text-sm ${result.success ? 'text-green-800' : 'text-red-800'}`}>
            {result.message}
          </p>
        </div>
      )}

      {/* Submit */}
      <div className="flex justify-end">
        <Button type="submit" isLoading={isSubmitting} disabled={getRecipientCount() === 0}>
          <Send className="w-4 h-4 mr-2" />
          Send to {getRecipientCount()} Recipients
        </Button>
      </div>
    </form>
  )
}
