import Link from 'next/link'

export default function DemoPage() {
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <h2 className="mt-6 text-3xl font-extrabold text-gray-900 dark:text-gray-100">
            App Shell Demo
          </h2>
          <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
            Preview the protected app layout without authentication
          </p>
        </div>
        <div className="space-y-4">
          <Link
            href="/app-demo"
            className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
          >
            View App Layout Demo
          </Link>
          <Link
            href="/"
            className="group relative w-full flex justify-center py-2 px-4 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Back to Home
          </Link>
        </div>
      </div>
    </div>
  )
}