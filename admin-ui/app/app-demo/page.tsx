import Link from 'next/link'

export default function AppDemoDashboard() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-gradient-to-r from-primary-600 to-blue-600 dark:from-primary-700 dark:to-blue-700 rounded-lg p-6 text-white">
        <h1 className="text-3xl font-bold">Welcome to Edge.link</h1>
        <p className="mt-2 text-primary-100">
          This is a demo of the app shell layout with sidebar navigation, topbar, tenant switcher, and dark mode support.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center">
                  <span className="text-white text-sm font-bold">📊</span>
                </div>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                    Total Requests
                  </dt>
                  <dd className="text-lg font-medium text-gray-900 dark:text-gray-100">
                    42,543
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center">
                  <span className="text-white text-sm font-bold">⚡</span>
                </div>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                    Avg Response Time
                  </dt>
                  <dd className="text-lg font-medium text-gray-900 dark:text-gray-100">
                    86ms
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-8 h-8 bg-yellow-500 rounded-full flex items-center justify-center">
                  <span className="text-white text-sm font-bold">🛣️</span>
                </div>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                    Active Routes
                  </dt>
                  <dd className="text-lg font-medium text-gray-900 dark:text-gray-100">
                    12
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-8 h-8 bg-purple-500 rounded-full flex items-center justify-center">
                  <span className="text-white text-sm font-bold">🔑</span>
                </div>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                    API Keys
                  </dt>
                  <dd className="text-lg font-medium text-gray-900 dark:text-gray-100">
                    5
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Features Showcase */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100">
              App Shell Features
            </h3>
            <div className="mt-4 space-y-3">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-6 h-6 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center">
                    <span className="text-green-600 dark:text-green-400 text-sm">✓</span>
                  </div>
                </div>
                <p className="ml-3 text-sm text-gray-600 dark:text-gray-300">Responsive sidebar navigation</p>
              </div>
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-6 h-6 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center">
                    <span className="text-green-600 dark:text-green-400 text-sm">✓</span>
                  </div>
                </div>
                <p className="ml-3 text-sm text-gray-600 dark:text-gray-300">Dark mode toggle with SSR support</p>
              </div>
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-6 h-6 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center">
                    <span className="text-green-600 dark:text-green-400 text-sm">✓</span>
                  </div>
                </div>
                <p className="ml-3 text-sm text-gray-600 dark:text-gray-300">Tenant switcher component</p>
              </div>
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-6 h-6 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center">
                    <span className="text-green-600 dark:text-green-400 text-sm">✓</span>
                  </div>
                </div>
                <p className="ml-3 text-sm text-gray-600 dark:text-gray-300">Mobile-friendly hamburger menu</p>
              </div>
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-6 h-6 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center">
                    <span className="text-green-600 dark:text-green-400 text-sm">✓</span>
                  </div>
                </div>
                <p className="ml-3 text-sm text-gray-600 dark:text-gray-300">Protected route structure under /app/*</p>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100">
              Quick Navigation
            </h3>
            <div className="mt-4 space-y-2">
              <Link 
                href="/app-demo/routes" 
                className="block p-3 rounded-md bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors"
              >
                <div className="flex items-center">
                  <span className="mr-3 text-lg">🛣️</span>
                  <span className="text-sm font-medium text-gray-900 dark:text-gray-100">Routes Management</span>
                </div>
              </Link>
              <Link 
                href="/app-demo/keys" 
                className="block p-3 rounded-md bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors"
              >
                <div className="flex items-center">
                  <span className="mr-3 text-lg">🔑</span>
                  <span className="text-sm font-medium text-gray-900 dark:text-gray-100">API Keys</span>
                </div>
              </Link>
              <Link 
                href="/app-demo/domains" 
                className="block p-3 rounded-md bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors"
              >
                <div className="flex items-center">
                  <span className="mr-3 text-lg">🌐</span>
                  <span className="text-sm font-medium text-gray-900 dark:text-gray-100">Custom Domains</span>
                </div>
              </Link>
              <Link 
                href="/app-demo/metrics" 
                className="block p-3 rounded-md bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors"
              >
                <div className="flex items-center">
                  <span className="mr-3 text-lg">📈</span>
                  <span className="text-sm font-medium text-gray-900 dark:text-gray-100">Analytics</span>
                </div>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}