import Link from 'next/link';

export default function HeroSection() {
  return (
    <section className="bg-gradient-to-b from-gray-50 to-white dark:from-gray-800 dark:to-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20 lg:py-28">
        <div className="text-center">
          <h1 className="text-4xl lg:text-6xl font-bold text-gray-900 dark:text-gray-100 mb-6">
            <span className="text-primary-600 dark:text-primary-400">API Proxy</span>
            <br />
            in 2 clicks
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-300 mb-8 max-w-3xl mx-auto">
            The fastest way to add caching, rate limiting, and security to your APIs. 
            No setup required. Start proxying in seconds.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link 
              href="/signup" 
              className="btn btn-primary text-lg px-8 py-3 inline-flex items-center justify-center"
            >
              Get Started Free
              <svg className="ml-2 w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
              </svg>
            </Link>
            <a 
              href="#how-it-works" 
              className="btn btn-secondary text-lg px-8 py-3 inline-flex items-center justify-center"
            >
              See How Simple It Is
              <svg className="ml-2 w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </a>
          </div>
        </div>

        {/* Stats */}
        <div className="mt-20 grid grid-cols-1 md:grid-cols-3 gap-8 text-center">
          <div>
            <div className="text-3xl font-bold text-primary-600 dark:text-primary-400">2 Clicks</div>
            <div className="text-gray-600 dark:text-gray-300">To Get Started</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-primary-600 dark:text-primary-400">&lt;10ms</div>
            <div className="text-gray-600 dark:text-gray-300">Average Latency</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-primary-600 dark:text-primary-400">99.9%</div>
            <div className="text-gray-600 dark:text-gray-300">Uptime SLA</div>
          </div>
        </div>

        {/* Architecture Diagram */}
        <div className="mt-20">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-8">
            <h3 className="text-xl font-semibold text-gray-900 dark:text-gray-100 text-center mb-8">
              From Your Domain to Live Proxy
            </h3>
            <div className="flex items-center justify-center space-x-8 text-center">
              <div className="flex flex-col items-center">
                <div className="w-16 h-16 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center mb-3">
                  <svg className="w-8 h-8 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9v-9m0-9v9" />
                  </svg>
                </div>
                <div className="text-sm font-medium text-gray-900 dark:text-gray-100">Your Domain</div>
                <div className="text-xs text-gray-500 dark:text-gray-400">api.yoursite.com</div>
              </div>
              
              <svg className="w-8 h-8 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
              </svg>
              
              <div className="flex flex-col items-center">
                <div className="w-16 h-16 bg-primary-100 dark:bg-primary-900 rounded-lg flex items-center justify-center mb-3">
                  <svg className="w-8 h-8 text-primary-600 dark:text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <div className="text-sm font-medium text-gray-900 dark:text-gray-100">Edge.link</div>
                <div className="text-xs text-gray-500 dark:text-gray-400">Instant Proxy</div>
              </div>
              
              <svg className="w-8 h-8 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
              </svg>
              
              <div className="flex flex-col items-center">
                <div className="w-16 h-16 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center mb-3">
                  <svg className="w-8 h-8 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                  </svg>
                </div>
                <div className="text-sm font-medium text-gray-900 dark:text-gray-100">Your API</div>
                <div className="text-xs text-gray-500 dark:text-gray-400">Protected & Fast</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}