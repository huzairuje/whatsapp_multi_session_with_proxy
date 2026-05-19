import { useQuery } from '@tanstack/react-query'
import { BarChart3, TrendingUp, AlertTriangle, Users } from 'lucide-react'
import Card from '@/components/common/Card'
import StatsCard from '@/components/features/StatsCard'

interface AdminStats {
  total_activities: number
  activities_by_type: Record<string, number>
  sessions_connected: number
  messages_sent: number
  messages_failed: number
  rate_limit_events: number
}

export default function AdminDashboard() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: async () => {
      const response = await fetch('/api/activities/stats', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
        }
      })
      if (!response.ok) throw new Error('Failed to fetch stats')
      return response.json() as Promise<AdminStats>
    },
    refetchInterval: 15000,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (!stats) {
    return null
  }

  const successRate = stats.messages_sent + stats.messages_failed > 0
    ? ((stats.messages_sent / (stats.messages_sent + stats.messages_failed)) * 100).toFixed(1)
    : '0'

  const errorRate = stats.messages_sent + stats.messages_failed > 0
    ? ((stats.messages_failed / (stats.messages_sent + stats.messages_failed)) * 100).toFixed(1)
    : '0'

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Admin Dashboard</h1>
        <p className="text-gray-600 mt-1">System-wide statistics and monitoring</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatsCard
          title="Total Activities"
          value={stats.total_activities}
          icon={BarChart3}
          color="primary"
        />
        <StatsCard
          title="Active Sessions"
          value={stats.sessions_connected}
          icon={Users}
          color="success"
        />
        <StatsCard
          title="Success Rate"
          value={`${successRate}%`}
          icon={TrendingUp}
          color="success"
        />
        <StatsCard
          title="Error Rate"
          value={`${errorRate}%`}
          icon={AlertTriangle}
          color={parseFloat(errorRate) > 10 ? 'warning' : 'success'}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card title="Message Statistics">
          <div className="space-y-4">
            <div className="flex justify-between items-center p-3 bg-green-50 rounded-lg">
              <span className="text-sm font-medium text-gray-700">Messages Sent</span>
              <span className="text-2xl font-bold text-green-600">{stats.messages_sent}</span>
            </div>
            <div className="flex justify-between items-center p-3 bg-red-50 rounded-lg">
              <span className="text-sm font-medium text-gray-700">Messages Failed</span>
              <span className="text-2xl font-bold text-red-600">{stats.messages_failed}</span>
            </div>
            <div className="flex justify-between items-center p-3 bg-blue-50 rounded-lg">
              <span className="text-sm font-medium text-gray-700">Rate Limit Events</span>
              <span className="text-2xl font-bold text-blue-600">{stats.rate_limit_events}</span>
            </div>
          </div>
        </Card>

        <Card title="Activity Breakdown">
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {Object.entries(stats.activities_by_type)
              .sort(([, a], [, b]) => b - a)
              .map(([type, count]) => (
                <div key={type} className="flex justify-between items-center p-2 bg-gray-50 rounded">
                  <span className="text-sm text-gray-700 capitalize">
                    {type.replace(/_/g, ' ')}
                  </span>
                  <span className="text-sm font-semibold text-gray-900">{count}</span>
                </div>
              ))}
          </div>
        </Card>
      </div>

      <Card title="System Health">
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-700">Overall Status</span>
            <span className="px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm font-medium">
              Healthy
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-700">Last Activity</span>
            <span className="text-sm text-gray-600">
              {stats.total_activities > 0 ? 'Active' : 'No activity'}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-700">Connected Sessions</span>
            <span className="text-sm font-semibold text-gray-900">
              {stats.sessions_connected}
            </span>
          </div>
        </div>
      </Card>
    </div>
  )
}
