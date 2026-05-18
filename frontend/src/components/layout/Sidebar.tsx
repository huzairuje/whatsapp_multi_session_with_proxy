import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard,
  Smartphone,
  Send,
  Users,
  FileText,
  Settings,
} from 'lucide-react'
import { clsx } from 'clsx'

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Sessions', href: '/sessions', icon: Smartphone },
  { name: 'Bulk Send', href: '/bulk-send', icon: Send },
  { name: 'Recipients', href: '/recipients', icon: Users },
  { name: 'Templates', href: '/templates', icon: FileText },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export default function Sidebar() {
  return (
    <aside className="w-64 bg-white border-r border-gray-200 h-screen sticky top-0 flex flex-col">
      <nav className="p-4 space-y-1 flex-1 overflow-y-auto">
        {navigation.map((item) => (
          <NavLink
            key={item.name}
            to={item.href}
            className={({ isActive }) =>
              clsx(
                'flex items-center space-x-3 px-4 py-3 rounded-lg transition-colors',
                isActive
                  ? 'bg-primary text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              )
            }
          >
            {({ isActive }) => (
              <>
                <item.icon className={clsx('w-5 h-5', isActive ? 'text-white' : 'text-gray-500')} />
                <span className="font-medium">{item.name}</span>
              </>
            )}
          </NavLink>
        ))}
      </nav>

      <div className="p-4 border-t border-gray-200">
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-xs font-semibold text-gray-700 mb-1">Anti-Ban Protection</p>
          <div className="flex items-center space-x-2">
            <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            <span className="text-xs text-gray-600">All systems active</span>
          </div>
        </div>
      </div>
    </aside>
  )
}
