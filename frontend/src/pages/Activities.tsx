import ActivityStats from '@/components/features/ActivityStats'
import ActivityLog from '@/components/features/ActivityLog'

export default function Activities() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Activities</h1>
        <p className="text-gray-600 mt-1">System activity log and statistics</p>
      </div>

      <ActivityStats />

      <ActivityLog />
    </div>
  )
}
