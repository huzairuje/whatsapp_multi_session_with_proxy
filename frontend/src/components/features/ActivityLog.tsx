import { useQuery } from '@tanstack/react-query'
import { AlertCircle, CheckCircle, Clock } from 'lucide-react'
import Card from '@/components/common/Card'

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

export default function ActivityLog() {
  const { data: activities, isLoading } = useQuery({
    queryKey: ['activities'],
    queryFn: async () => {
      const response = await fetch('/api/activities?limit=50', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      })
      if (!response.ok) throw new Error('Failed to fetch activities')
      return response.json() as Promise<ActivityItem[]>
    },
    refetchInterval: 5000,
  })

  const getActivityIcon = (type: string) => {
    if (type.includes('success') || type.includes('sent') || type.includes('connect')) {
      return <CheckCircle className="w-4 h-4 text-green-500" />
    }
    if (type.includes('error') || type.includes('failed') || type.includes('disconnect')) {
      return <AlertCircle className="w-4 h-4 text-red-500" />
    }
    return <Clock className="w-4 h-4 text-blue-500" />
  }

  const getActivityColor = (type: string) => {
    if (type.includes('success') || type.includes('sent') || type.includes('connect')) {
      return 'bg-green-50 border-green-200'
    }
    if (type.includes('error') || type.includes('failed') || type.includes('disconnect')) {
      return 'bg-red-50 border-red-200'
    }
    return 'bg-blue-50 border-blue-200'
  }

  const formatTime = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleTimeString()
  }

  return (
    <Card title="Recent Activities" subtitle="Last 50 activities">
      {isLoading ? (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : !activities || activities.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          No activities recorded yet
        </div>
      ) : (
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {activities.map((activity) => (
            <div
              key={activity.id}
              className={`p-3 rounded-lg border ${getActivityColor(activity.type)} flex items-start space-x-3`}
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
                  {activity.user && (
                    <span className="text-xs bg-gray-200 px-2 py-1 rounded">
                      {activity.user}
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
    </Card>
  )
}
