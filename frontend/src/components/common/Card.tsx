import { HTMLAttributes, ReactNode } from 'react'
import { clsx } from 'clsx'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode
  title?: string
  subtitle?: string
  action?: ReactNode
}

export default function Card({ children, title, subtitle, action, className, ...props }: CardProps) {
  return (
    <div className={clsx('bg-white rounded-lg shadow-sm border border-gray-200', className)} {...props}>
      {(title || subtitle || action) && (
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <div>
            {title && <h3 className="text-lg font-semibold text-gray-900">{title}</h3>}
            {subtitle && <p className="text-sm text-gray-500 mt-1">{subtitle}</p>}
          </div>
          {action && <div>{action}</div>}
        </div>
      )}
      <div className="p-6">{children}</div>
    </div>
  )
}
