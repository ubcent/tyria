'use client'

import { ReactNode, useState } from 'react';
import { Bars3Icon, XMarkIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import ThemeToggle from '@/components/ui/ThemeToggle';

interface AppDemoLayoutProps {
  children: ReactNode;
}

const navigation = [
  { name: 'Dashboard', href: '/app-demo', icon: '📊' },
  { name: 'Routes', href: '/app-demo/routes', icon: '🛣️' },
  { name: 'API Keys', href: '/app-demo/keys', icon: '🔑' },
  { name: 'Domains', href: '/app-demo/domains', icon: '🌐' },
  { name: 'Metrics', href: '/app-demo/metrics', icon: '📈' },
];

export default function AppDemoLayout({ children }: AppDemoLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const pathname = usePathname();

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Mobile sidebar */}
      <div className={`fixed inset-0 z-40 md:hidden ${sidebarOpen ? 'block' : 'hidden'}`}>
        <div className="fixed inset-0 bg-gray-600 bg-opacity-75" onClick={() => setSidebarOpen(false)} />
        <div className="relative flex-1 flex flex-col max-w-xs w-full bg-white dark:bg-gray-800">
          <div className="absolute top-0 right-0 -mr-12 pt-2">
            <button
              className="ml-1 flex items-center justify-center h-10 w-10 rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
              onClick={() => setSidebarOpen(false)}
            >
              <XMarkIcon className="h-6 w-6 text-white" />
            </button>
          </div>
          <SidebarContent pathname={pathname} />
        </div>
      </div>

      {/* Static sidebar for desktop */}
      <div className="hidden md:flex md:w-64 md:flex-col md:fixed md:inset-y-0">
        <SidebarContent pathname={pathname} />
      </div>

      {/* Main content */}
      <div className="md:pl-64 flex flex-col flex-1">
        {/* Top bar */}
        <div className="sticky top-0 z-10 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between h-16 px-4 sm:px-6 lg:px-8">
            {/* Mobile menu button */}
            <button
              className="md:hidden p-2 rounded-md text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary-500"
              onClick={() => setSidebarOpen(true)}
            >
              <Bars3Icon className="h-6 w-6" />
            </button>

            {/* Right side of top bar */}
            <div className="flex items-center space-x-4">
              <ThemeToggle />
              {/* Demo Tenant Switcher */}
              <div className="flex items-center space-x-2 px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-200">
                <div className="flex-shrink-0 w-6 h-6 bg-primary-600 dark:bg-primary-500 rounded-full flex items-center justify-center">
                  <span className="text-xs font-semibold text-white">D</span>
                </div>
                <div className="min-w-0 flex-1 text-left">
                  <p className="truncate">Demo Tenant</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">Free plan</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Page content */}
        <main className="flex-1 p-6">
          {children}
        </main>
      </div>
    </div>
  );
}

function SidebarContent({ pathname }: { pathname: string }) {
  return (
    <div className="flex flex-col flex-grow pt-5 pb-4 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
      {/* Logo */}
      <div className="flex items-center flex-shrink-0 px-4">
        <Link href="/" className="text-xl font-bold text-primary-600 dark:text-primary-400">
          Edge.link
        </Link>
        <span className="ml-2 text-xs bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200 px-2 py-1 rounded">
          DEMO
        </span>
      </div>

      {/* Navigation */}
      <div className="mt-5 flex-grow flex flex-col">
        <nav className="flex-1 px-2 space-y-1">
          {navigation.map((item) => {
            const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
            return (
              <Link
                key={item.name}
                href={item.href}
                className={`group flex items-center px-2 py-2 text-sm font-medium rounded-md transition-colors ${
                  isActive
                    ? 'bg-primary-100 dark:bg-primary-900 text-primary-900 dark:text-primary-100'
                    : 'text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100'
                }`}
              >
                <span className="mr-3 text-lg">{item.icon}</span>
                {item.name}
              </Link>
            );
          })}
        </nav>

        {/* Demo notice */}
        <div className="flex-shrink-0 border-t border-gray-200 dark:border-gray-700 p-4">
          <div className="bg-blue-50 dark:bg-blue-900 rounded-lg p-3">
            <p className="text-xs text-blue-700 dark:text-blue-200 font-medium">Demo Mode</p>
            <p className="text-xs text-blue-600 dark:text-blue-300 mt-1">
              This is a preview of the app layout
            </p>
            <Link
              href="/demo"
              className="text-xs text-blue-600 dark:text-blue-400 hover:underline mt-1 inline-block"
            >
              ← Back to demo page
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}