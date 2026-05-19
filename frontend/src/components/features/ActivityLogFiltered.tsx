import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { AlertCircle, CheckCircle, Clock, Search } from 'lucide-react'
import Card from '@/components/common/Card'
import Input from '@/components/common/Input'

interface ActivityItem {
  id: number
  type: string
  sender?: string
  user?: string
  message: string
  details?: string
  status?: string
  error?: string
  created_at: string
}

export default function ActivityLogFiltered() {
  const [searchTerm, setSearchTerm] = useState('')
  const [filterType, setFilterType] = useState('all')

  const { data: activities, isLoading } = useQuery({
    queryKey: ['activities', filterType],
    queryFn: async () => {
      let url = '/api/activities?limit=100'
      if (filterType !== 'all') {
        url = `/api/activities/type?type=${filterType}&limit=100`
      }
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      })
      if (!response.ok) throw new Error('Failed to fetch activities')
      return response.json() as Promise<ActivityItem[]>
    },
    refetchInterval: 5000,
  })

  const filteredActivities = activities?.filter(activity =>
    activity.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
    activity.sender?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    activity.user?.toLowerCase().includes(searchTerm.toLowerCase())
  ) || []

  const activityTypes = [
    { value: 'all', label: 'All Activities' },
    { value: 'session_connect', label: 'Session Connect' },
    { value: 'session_disconnect', label: 'Session Disconnect' },
    { value: 'message_sent', label: 'Message Sent' },
    { value: 'message_failed', label: 'Message Failed' },
    { value: 'bulk_send_start', label: 'Bulk Send' },
    { value: 'rate_limit', label: 'Rate Limit' },
    { value: 'auto_login', label: 'Auto Login' },
    { value: 'health_check', label: 'Health Check' },
  ]

  const getActivityIcon = (type: string) => {
    if (type.includes('success') || type.includes('sent') || type.includes('connect')) {
      return <CheckCircle className="w-4 h-4 text-green-500" />
    }
    if (type.includes('error') || type.includes('failed') || type.includes('disconnect')) {
      return <AlertCircle className="w-4 h-4 text-red-500" />
    }
    return <Clock className="w-4 h-4 text-blue-500" />
  }

  const formatTime = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleTimeString()
  }

  return (
    <Card title="Activity Log with Filters">
      <div className="space-y-4">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1">
            <div className="relative">
              <Search className="absolute left-3 top-3 w-4 h-4 text-gray-400" />
              <Input
                type="text"
                placeholder="Search activities..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
          >
            {activityTypes.map(type => (
              <option key={type.value} value={type.value}>
                {type.label}
              </option>
            ))}
          </select>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        ) : filteredActivities.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            No activities found
          </div>
        ) : (
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {filteredActivities.map((activity) => (
              <div
                key={activity.id}
                className="p-3 rounded-lg border bg-gray-50 flex items-start space-x-3"
              >
                <div className="flex-shrink-0 mt-1">
                  {getActivityIcon(activity.type)}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900">{activity.message}</p>
                  {activity.details && (
                    <p className="text-xs text-gray-600 mt-1">{activity.details}</p>
                  )}
                  <div className="flex items-center space-x-2 mt-1">
                    {activity.sender && (
                      <span className="text-xs bg-gray-200 px-2 py-1 rounded">
                        {activity.sender}
                      </span>
                    )}
                    <span className="text-xs text-gray-500 ml-auto">
                      {formatTime(activity.created_at)}
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Card>
  )
}
