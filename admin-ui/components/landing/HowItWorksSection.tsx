export default function HowItWorksSection() {
  const steps = [
    {
      step: "1",
      title: "Enter Your Domain",
      description: "Simply enter your API domain and we'll instantly generate a unique proxy URL for you.",
      icon: (
        <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9v-9m0-9v9" />
        </svg>
      ),
      example: `Input: api.yourcompany.com
↓
Generated: api-yourcompany-com.edge.link`
    },
    {
      step: "2", 
      title: "Configure Target",
      description: "Point your proxy to your existing API. That's it - your proxy is now live and protecting your API.",
      icon: (
        <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      ),
      example: `Target: https://your-api-server.com
↓
✅ Proxy is live and ready!`
    },
    {
      step: "3",
      title: "Start Using", 
      description: "Use your new proxy URL in your applications. All requests are automatically cached, rate-limited, and secured.",
      icon: (
        <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      ),
      example: `curl api-yourcompany-com.edge.link/users
↓
✨ Fast, cached, secure responses`
    }
  ];

  return (
    <section id="how-it-works" className="py-20 bg-white dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl lg:text-4xl font-bold text-gray-900 dark:text-gray-100 mb-4">
            Get Your Proxy in 2 Clicks
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            No complex setup, no Docker, no configuration files. Just enter your domain and start proxying.
          </p>
        </div>

        <div className="space-y-16">
          {steps.map((step, index) => (
            <div key={index} className="grid grid-cols-1 lg:grid-cols-2 gap-8 items-center">
              <div className={`${index % 2 === 1 ? 'lg:order-2' : ''}`}>
                <div className="flex items-center mb-4">
                  <div className="flex items-center justify-center w-10 h-10 bg-primary-100 dark:bg-primary-900 rounded-full mr-4">
                    <span className="text-lg font-bold text-primary-600 dark:text-primary-400">{step.step}</span>
                  </div>
                  <div className="text-primary-600 dark:text-primary-400">
                    {step.icon}
                  </div>
                </div>
                <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
                  {step.title}
                </h3>
                <p className="text-gray-600 dark:text-gray-300 mb-6 text-lg">
                  {step.description}
                </p>
                <ul className="space-y-3">
                  {step.step === "1" && (
                    <>
                      <li className="flex items-center text-gray-600 dark:text-gray-300">
                        <svg className="w-5 h-5 text-green-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        Instant proxy URL generation
                      </li>
                      <li className="flex items-center text-gray-600 dark:text-gray-300">
                        <svg className="w-5 h-5 text-green-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        SSL certificates included
                      </li>
                      <li className="flex items-center text-gray-600 dark:text-gray-300">
                        <svg className="w-5 h-5 text-green-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        No DNS changes required
                      </li>
                    </>
                  )}
                  {step.step === "2" && (
                    <>
                      <li className="flex items-center text-gray-600 dark:text-gray-300">
                        <svg className="w-5 h-5 text-green-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        Automatic health monitoring
                      </li>
                      <li className="flex items-center text-gray-600 dark:text-gray-300">
                        <svg className="w-5 h-5 text-green-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        Intelligent caching enabled
                      </li>
                      <li className="flex items-center text-gray-600 dark:text-gray-300">
                        <svg className="w-5 h-5 text-green-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        Rate limiting protection
                      </li>
                    </>
                  )}
                  {step.step === "3" && (
                    <>
                      <li className="flex items-center text-gray-600 dark:text-gray-300">
                        <svg className="w-5 h-5 text-green-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        Real-time analytics dashboard
                      </li>
                      <li className="flex items-center text-gray-600 dark:text-gray-300">
                        <svg className="w-5 h-5 text-green-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        API key management
                      </li>
                      <li className="flex items-center text-gray-600 dark:text-gray-300">
                        <svg className="w-5 h-5 text-green-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        24/7 monitoring and alerts
                      </li>
                    </>
                  )}
                </ul>
              </div>
              
              <div className={`${index % 2 === 1 ? 'lg:order-1' : ''}`}>
                <div className="bg-gray-900 dark:bg-gray-800 rounded-lg p-6 border border-gray-200 dark:border-gray-700">
                  <div className="flex items-center mb-4">
                    <div className="flex space-x-2">
                      <div className="w-3 h-3 bg-red-500 rounded-full"></div>
                      <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
                      <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                    </div>
                  </div>
                  <pre className="text-sm text-green-400 font-mono whitespace-pre-wrap">
                    {step.example}
                  </pre>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* CTA */}
        <div className="mt-20 text-center">
          <div className="bg-gradient-to-r from-primary-50 to-blue-50 dark:from-primary-900 dark:to-blue-900 rounded-2xl p-8">
            <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
              Ready to get started?
            </h3>
            <p className="text-gray-600 dark:text-gray-300 mb-6">
              Join thousands of developers who chose the simple way to proxy their APIs.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <a href="/signup" className="btn btn-primary">
                Get Started Free
              </a>
              <a href="#pricing" className="btn btn-secondary">
                View Pricing
              </a>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}