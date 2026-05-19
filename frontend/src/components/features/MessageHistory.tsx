import { useQuery } from '@tanstack/react-query'
import { MessageSquare, CheckCircle, AlertCircle } from 'lucide-react'
import Card from '@/components/common/Card'

interface Message {
  id: number
  sender: string
  recipient: string
  content: string
  status: string
  message_id: string
  created_at: string
}

export default function MessageHistory() {
  const { data: messages, isLoading } = useQuery({
    queryKey: ['messages'],
    queryFn: async () => {
      const response = await fetch('/api/messages?limit=20', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      })
      if (!response.ok) throw new Error('Failed to fetch messages')
      return response.json() as Promise<Message[]>
    },
    refetchInterval: 10000,
  })

  const getStatusIcon = (status: string) => {
    if (status === 'sent' || status === 'delivered' || status === 'read') {
      return <CheckCircle className="w-4 h-4 text-green-500" />
    }
    if (status === 'failed') {
      return <AlertCircle className="w-4 h-4 text-red-500" />
    }
    return <MessageSquare className="w-4 h-4 text-blue-500" />
  }

  const getStatusColor = (status: string) => {
    if (status === 'sent' || status === 'delivered' || status === 'read') {
      return 'bg-green-50'
    }
    if (status === 'failed') {
      return 'bg-red-50'
    }
    return 'bg-blue-50'
  }

  const formatTime = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleTimeString()
  }

  const truncateContent = (content: string, maxLength: number = 50) => {
    return content.length > maxLength ? content.substring(0, maxLength) + '...' : content
  }

  return (
    <Card title="Recent Messages" subtitle="Last 20 messages sent">
      {isLoading ? (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : !messages || messages.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          No messages sent yet
        </div>
      ) : (
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={`p-3 rounded-lg border ${getStatusColor(msg.status)} flex items-start space-x-3`}
            >
              <div className="flex-shrink-0 mt-1">
                {getStatusIcon(msg.status)}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium text-gray-900">
                    {msg.sender} → {msg.recipient}
                  </p>
                  <span className="text-xs text-gray-500">
                    {formatTime(msg.created_at)}
                  </span>
                </div>
                <p className="text-sm text-gray-600 mt-1">
                  {truncateContent(msg.content)}
                </p>
                <div className="flex items-center space-x-2 mt-1">
                  <span className="text-xs bg-gray-200 px-2 py-1 rounded capitalize">
                    {msg.status}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </Card>
  )
}
