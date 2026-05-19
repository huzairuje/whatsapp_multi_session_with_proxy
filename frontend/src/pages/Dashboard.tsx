import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { Smartphone, Send, CheckCircle, TrendingUp, Activity, AlertTriangle } from 'lucide-react'
import Card from '@/components/common/Card'
import StatsCard from '@/components/features/StatsCard'
import Badge from '@/components/common/Badge'
import MessageHistory from '@/components/features/MessageHistory'
import { sessionApi, messageApi, activityApi } from '@/services/api'

export default function Dashboard() {
  const { data: devices, isLoading: devicesLoading } = useQuery({
    queryKey: ['devices'],
    queryFn: async () => {
      const response = await sessionApi.getAll()
      return response.data
    },
    refetchInterval: 10000,
  })

  const { data: bulkStatuses } = useQuery({
    queryKey: ['bulk-statuses'],
    queryFn: async () => {
      if (!devices) return {}
      const statuses: Record<string, any> = {}
      await Promise.all(
        devices.map(async (device) => {
          try {
            const response = await messageApi.getBulkSendStatus(device.user)
            statuses[device.user] = response.data
          } catch (error) {
            statuses[device.user] = null
          }
        })
      )
      return statuses
    },
    enabled: !!devices && devices.length > 0,
  })

  const { data: messageStats } = useQuery({
    queryKey: ['message-stats-all'],
    queryFn: async () => {
      const response = await messageApi.getAllMessageStats()
      return response.data
    },
    refetchInterval: 10000,
  })

  const connectedSessions = devices?.filter(d => d.isLoggedIn).length || 0
  const totalSessions = devices?.length || 0

  const totalMessagesSent = messageStats?.daily_count || 0

  const successRate = messageStats && (messageStats.total_sent + messageStats.total_failed) > 0
    ? ((messageStats.total_sent / (messageStats.total_sent + messageStats.total_failed)) * 100).toFixed(1)
    : '0'

  const { data: recentActivitiesData } = useQuery({
    queryKey: ['recent-activities'],
    queryFn: async () => {
      const response = await activityApi.getRecent(5)
      return response.data
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  const recentActivity = (recentActivitiesData || []).map((activity: any) => {
    const time = new Date(activity.created_at).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    })

    let type = 'info'
    if (activity.status === 'success' || activity.type.includes('connect')) {
      type = 'success'
    } else if (activity.status === 'error' || activity.status === 'failed' || activity.type.includes('failed')) {
      type = 'warning'
    }

    return {
      time,
      message: activity.message,
      type
    }
  })

  const getHealthStatus = (device: any) => {
    const status = bulkStatuses?.[device.user]
    if (!status) return 'unknown'
    const percentage = (status.daily_count / status.daily_limit) * 100
    if (percentage >= 90) return 'critical'
    if (percentage >= 70) return 'warning'
    return 'healthy'
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600 mt-1">Overview of your WhatsApp bulk sender system</p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatsCard
          title="Total Sessions"
          value={totalSessions}
          icon={Smartphone}
          color="primary"
        />
        <StatsCard
          title="Active Sessions"
          value={connectedSessions}
          icon={Activity}
          color="success"
        />
        <StatsCard
          title="Messages Today"
          value={totalMessagesSent}
          icon={Send}
          color="primary"
        />
        <StatsCard
          title="Success Rate"
          value={`${successRate}%`}
          icon={CheckCircle}
          color="success"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Active Sessions */}
        <Card title="Active Sessions" subtitle={`${connectedSessions} of ${totalSessions} connected`}>
          {devicesLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          ) : !devices || devices.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No sessions available
            </div>
          ) : (
            <div className="space-y-3">
              {devices.slice(0, 5).map((device) => {
                const status = bulkStatuses?.[device.user]
                const healthStatus = getHealthStatus(device)

                return (
                  <div
                    key={device.user}
                    className="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                  >
                    <div className="flex items-center space-x-3">
                      <div className={`w-2 h-2 rounded-full ${device.isLoggedIn ? 'bg-green-500' : 'bg-gray-400'}`} />
                      <div>
                        <p className="font-medium text-gray-900">+{device.user}</p>
                        <p className="text-xs text-gray-500">{device.pushName || 'No name'}</p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      {status && (
                        <span className="text-sm text-gray-600">
                          {status.daily_count}/{status.daily_limit}
                        </span>
                      )}
                      <Badge
                        variant={
                          healthStatus === 'critical' ? 'error' :
                          healthStatus === 'warning' ? 'warning' :
                          healthStatus === 'healthy' ? 'success' : 'default'
                        }
                        size="sm"
                      >
                        {device.isLoggedIn ? 'Online' : 'Offline'}
                      </Badge>
                    </div>
                  </div>
                )
              })}
              {devices.length > 5 && (
                <div className="text-center pt-2">
                  <a href="/sessions" className="text-sm text-primary hover:underline">
                    View all {devices.length} sessions →
                  </a>
                </div>
              )}
            </div>
          )}
        </Card>

        {/* Recent Activity */}
        <Card
          title="Recent Activity"
          action={
            <Link
              to="/activities"
              className="text-sm text-primary hover:text-primary-dark font-medium"
            >
              View All →
            </Link>
          }
        >
          <div className="space-y-3">
            {recentActivity.map((activity: { time: string; message: string; type: string }, index: number) => (
              <div key={index} className="flex items-start space-x-3">
                <div className="flex-shrink-0 mt-1">
                  {activity.type === 'success' && (
                    <CheckCircle className="w-5 h-5 text-green-500" />
                  )}
                  {activity.type === 'warning' && (
                    <AlertTriangle className="w-5 h-5 text-yellow-500" />
                  )}
                  {activity.type === 'info' && (
                    <Activity className="w-5 h-5 text-blue-500" />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-gray-900">{activity.message}</p>
                  <p className="text-xs text-gray-500 mt-1">{activity.time}</p>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>

      {/* System Status */}
      <Card title="System Status">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 bg-green-100 rounded-full mb-3">
              <CheckCircle className="w-6 h-6 text-green-600" />
            </div>
            <p className="text-sm font-medium text-gray-900">Anti-Ban Protection</p>
            <p className="text-xs text-gray-500 mt-1">All systems active</p>
          </div>
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 bg-green-100 rounded-full mb-3">
              <Activity className="w-6 h-6 text-green-600" />
            </div>
            <p className="text-sm font-medium text-gray-900">Health Monitoring</p>
            <p className="text-xs text-gray-500 mt-1">Tracking all sessions</p>
          </div>
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 bg-green-100 rounded-full mb-3">
              <TrendingUp className="w-6 h-6 text-green-600" />
            </div>
            <p className="text-sm font-medium text-gray-900">Performance</p>
            <p className="text-xs text-gray-500 mt-1">Optimal</p>
          </div>
        </div>
      </Card>

      {/* Message History */}
      <MessageHistory />

      {/* Quick Actions */}
      <Card title="Quick Actions">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <a
            href="/sessions"
            className="p-4 border-2 border-gray-200 rounded-lg hover:border-primary hover:bg-primary/5 transition-all text-center"
          >
            <Smartphone className="w-8 h-8 text-primary mx-auto mb-2" />
            <p className="font-medium text-gray-900">Manage Sessions</p>
            <p className="text-xs text-gray-500 mt-1">Connect or disconnect devices</p>
          </a>
          <a
            href="/bulk-send"
            className="p-4 border-2 border-gray-200 rounded-lg hover:border-primary hover:bg-primary/5 transition-all text-center"
          >
            <Send className="w-8 h-8 text-primary mx-auto mb-2" />
            <p className="font-medium text-gray-900">Send Bulk Messages</p>
            <p className="text-xs text-gray-500 mt-1">Send to multiple recipients</p>
          </a>
          <a
            href="/settings"
            className="p-4 border-2 border-gray-200 rounded-lg hover:border-primary hover:bg-primary/5 transition-all text-center"
          >
            <Activity className="w-8 h-8 text-primary mx-auto mb-2" />
            <p className="font-medium text-gray-900">Configure Settings</p>
            <p className="text-xs text-gray-500 mt-1">Adjust anti-ban parameters</p>
          </a>
        </div>
      </Card>
    </div>
  )
}
