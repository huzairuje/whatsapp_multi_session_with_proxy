import { useQuery } from '@tanstack/react-query'
import { Activity, MessageSquare, AlertTriangle, TrendingUp } from 'lucide-react'
import Card from '@/components/common/Card'
import StatsCard from '@/components/features/StatsCard'

interface ActivityStats {
  total_activities: number
  activities_by_type: Record<string, number>
  sessions_connected: number
  messages_sent: number
  messages_failed: number
  rate_limit_events: number
}

export default function ActivityStats() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['activity-stats'],
    queryFn: async () => {
      const response = await fetch('/api/activities/stats', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      })
      if (!response.ok) throw new Error('Failed to fetch stats')
      return response.json() as Promise<ActivityStats>
    },
    refetchInterval: 10000,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (!stats) {
    return null
  }

  const successRate = stats.messages_sent + stats.messages_failed > 0
    ? ((stats.messages_sent / (stats.messages_sent + stats.messages_failed)) * 100).toFixed(1)
    : '0'

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatsCard
          title="Total Activities"
          value={stats.total_activities}
          icon={Activity}
          color="primary"
        />
        <StatsCard
          title="Messages Sent"
          value={stats.messages_sent}
          icon={MessageSquare}
          color="success"
        />
        <StatsCard
          title="Success Rate"
          value={`${successRate}%`}
          icon={TrendingUp}
          color="primary"
        />
        <StatsCard
          title="Rate Limits"
          value={stats.rate_limit_events}
          icon={AlertTriangle}
          color={stats.rate_limit_events > 0 ? 'warning' : 'success'}
        />
      </div>

      <Card title="Activity Breakdown">
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {Object.entries(stats.activities_by_type).map(([type, count]) => (
            <div key={type} className="p-4 bg-gray-50 rounded-lg">
              <p className="text-sm text-gray-600 capitalize">
                {type.replace(/_/g, ' ')}
              </p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{count}</p>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}
