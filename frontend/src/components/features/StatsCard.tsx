import { LucideIcon } from 'lucide-react'
import { clsx } from 'clsx'

interface StatsCardProps {
  title: string
  value: string | number
  icon: LucideIcon
  trend?: {
    value: number
    isPositive: boolean
  }
  color?: 'primary' | 'success' | 'warning' | 'error'
}

export default function StatsCard({ title, value, icon: Icon, trend, color = 'primary' }: StatsCardProps) {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-3xl font-bold text-gray-900 mt-2">{value}</p>
          {trend && (
            <p className={clsx('text-sm mt-2', trend.isPositive ? 'text-green-600' : 'text-red-600')}>
              {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}%
            </p>
          )}
        </div>
        <div
          className={clsx('p-3 rounded-lg', {
            'bg-primary/10': color === 'primary',
            'bg-green-100': color === 'success',
            'bg-yellow-100': color === 'warning',
            'bg-red-100': color === 'error',
          })}
        >
          <Icon
            className={clsx('w-8 h-8', {
              'text-primary': color === 'primary',
              'text-green-600': color === 'success',
              'text-yellow-600': color === 'warning',
              'text-red-600': color === 'error',
            })}
          />
        </div>
      </div>
    </div>
  )
}
