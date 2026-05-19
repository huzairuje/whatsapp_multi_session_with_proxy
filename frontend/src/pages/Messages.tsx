import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Download, RefreshCw, CheckCircle, AlertCircle, Clock } from 'lucide-react'
import Card from '@/components/common/Card'
import Button from '@/components/common/Button'

interface Message {
  id: number
  sender: string
  recipient: string
  content: string
  status: string
  message_id: string
  error: string
  created_at: string
  updated_at: string
}

export default function Messages() {
  const [limit] = useState(50)
  const [offset] = useState(0)
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [senderFilter, setSenderFilter] = useState<string>('')

  const { data: messages, isLoading, refetch } = useQuery({
    queryKey: ['messages', limit, offset, statusFilter, senderFilter],
    queryFn: async () => {
      const params = new URLSearchParams({
        limit: limit.toString(),
        offset: offset.toString(),
      })
      if (senderFilter) params.append('sender', senderFilter)

      const response = await fetch(`/api/messages?${params.toString()}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      })
      if (!response.ok) throw new Error('Failed to fetch messages')
      const data = await response.json() as Message[]

      // Filter by status on frontend if needed
      if (statusFilter !== 'all') {
        return data.filter(msg => msg.status === statusFilter)
      }
      return data
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
    return <Clock className="w-4 h-4 text-gray-500" />
  }

  const getStatusColor = (status: string) => {
    if (status === 'sent' || status === 'delivered' || status === 'read') {
      return 'text-green-600 bg-green-50'
    }
    if (status === 'failed') {
      return 'text-red-600 bg-red-50'
    }
    return 'text-gray-600 bg-gray-50'
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleString()
  }

  const truncateContent = (content: string, maxLength: number = 50) => {
    return content.length > maxLength ? content.substring(0, maxLength) + '...' : content
  }

  const exportToCSV = () => {
    if (!messages || messages.length === 0) {
      alert('No messages to export')
      return
    }

    // CSV headers
    const headers = ['ID', 'Sender', 'Recipient', 'Message', 'Status', 'Error', 'Created At', 'Updated At']

    // Convert messages to CSV rows
    const rows = messages.map(msg => [
      msg.id,
      msg.sender,
      msg.recipient,
      `"${msg.content.replace(/"/g, '""')}"`, // Escape quotes in content
      msg.status,
      msg.error ? `"${msg.error.replace(/"/g, '""')}"` : '',
      msg.created_at,
      msg.updated_at
    ])

    // Combine headers and rows
    const csvContent = [
      headers.join(','),
      ...rows.map(row => row.join(','))
    ].join('\n')

    // Create blob and download
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `messages_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Messages</h1>
          <p className="text-gray-600 mt-1">View and export message history</p>
        </div>
        <div className="flex space-x-3">
          <Button variant="secondary" onClick={() => refetch()} isLoading={isLoading}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={exportToCSV}>
            <Download className="w-4 h-4 mr-2" />
            Export CSV
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Status
            </label>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="all">All Status</option>
              <option value="pending">Pending</option>
              <option value="sent">Sent</option>
              <option value="delivered">Delivered</option>
              <option value="read">Read</option>
              <option value="failed">Failed</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Sender
            </label>
            <input
              type="text"
              placeholder="Filter by sender (e.g., 6281234567890)"
              value={senderFilter}
              onChange={(e) => setSenderFilter(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
            />
          </div>

          <div className="flex items-end">
            <Button
              variant="secondary"
              onClick={() => {
                setStatusFilter('all')
                setSenderFilter('')
              }}
              className="w-full"
            >
              Clear Filters
            </Button>
          </div>
        </div>
      </Card>

      <Card>
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
          </div>
        ) : !messages || messages.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">No messages found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Sender
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Recipient
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Message
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Date
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {messages.map((message) => (
                  <tr key={message.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center space-x-2">
                        {getStatusIcon(message.status)}
                        <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(message.status)}`}>
                          {message.status}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">+{message.sender}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">+{message.recipient}</div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-900 max-w-md">
                        {truncateContent(message.content)}
                      </div>
                      {message.error && (
                        <div className="text-xs text-red-600 mt-1">
                          Error: {message.error}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-500">
                        {formatDate(message.created_at)}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Pagination will be added here */}
    </div>
  )
}
