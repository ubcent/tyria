import Link from 'next/link'

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      {/* Welcome Section */}
      <div className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100">
            Welcome to your tenant dashboard
          </h3>
          <div className="mt-2 max-w-xl text-sm text-gray-500 dark:text-gray-400">
            <p>
              You are successfully authenticated and can access your tenant resources.
              Get started by managing your routes and API keys.
            </p>
          </div>
          <div className="mt-4">
            <span className="inline-flex items-center px-3 py-0.5 rounded-full text-sm font-medium bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200">
              ✅ Dashboard Active
            </span>
          </div>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
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
                    0
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
                    0ms
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
                  <span className="text-white text-sm font-bold">🛣️</span>
                </div>
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                    Active Routes
                  </dt>
                  <dd className="text-lg font-medium text-gray-900 dark:text-gray-100">
                    0
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100 mb-4">
            Quick Actions
          </h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <Link
              href="/app/routes"
              className="relative rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-6 py-5 shadow-sm flex flex-col space-y-3 hover:border-gray-400 dark:hover:border-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <span className="text-2xl">🛣️</span>
              <span className="text-sm font-medium text-gray-900 dark:text-gray-100">Manage Routes</span>
              <span className="text-sm text-gray-500 dark:text-gray-400">Configure proxy routes</span>
            </Link>

            <Link
              href="/app/keys"
              className="relative rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-6 py-5 shadow-sm flex flex-col space-y-3 hover:border-gray-400 dark:hover:border-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <span className="text-2xl">🔑</span>
              <span className="text-sm font-medium text-gray-900 dark:text-gray-100">API Keys</span>
              <span className="text-sm text-gray-500 dark:text-gray-400">Manage authentication</span>
            </Link>

            <Link
              href="/app/metrics"
              className="relative rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-6 py-5 shadow-sm flex flex-col space-y-3 hover:border-gray-400 dark:hover:border-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <span className="text-2xl">📈</span>
              <span className="text-sm font-medium text-gray-900 dark:text-gray-100">Analytics</span>
              <span className="text-sm text-gray-500 dark:text-gray-400">View performance metrics</span>
            </Link>

            <div className="relative rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-6 py-5 shadow-sm flex flex-col space-y-3 opacity-50">
              <span className="text-2xl">⚙️</span>
              <span className="text-sm font-medium text-gray-900 dark:text-gray-100">Settings</span>
              <span className="text-sm text-gray-500 dark:text-gray-400">Configure your account</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}