export default function HowItWorksSection() {
  const steps = [
    {
      step: "1",
      title: "Configure Routes",
      description: "Set up your API endpoints with simple YAML configuration or our web interface. Define targets, methods, and policies.",
      icon: (
        <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
      ),
      code: `routes:
  - path: "/api/v1/users"
    target: "https://api.example.com"
    methods: ["GET", "POST"]
    cache:
      enabled: true
      ttl: "5m"`
    },
    {
      step: "2", 
      title: "Deploy & Scale",
      description: "Deploy with Docker or our managed service. Auto-scaling handles traffic spikes while maintaining low latency.",
      icon: (
        <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      ),
      code: `docker run -p 8080:8080 \\
  -v ./config.yaml:/config.yaml \\
  edgelink/proxy:latest`
    },
    {
      step: "3",
      title: "Monitor & Optimize", 
      description: "Real-time metrics show performance, cache hit rates, and errors. Adjust policies based on actual usage patterns.",
      icon: (
        <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      ),
      code: `{
  "total_requests": 125000,
  "cache_hit_rate": 0.85,
  "avg_response_time": "45ms",
  "error_rate": 0.001
}`
    }
  ];

  return (
    <section id="how-it-works" className="py-20 bg-white dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl lg:text-4xl font-bold text-gray-900 dark:text-gray-100 mb-4">
            How Edge.link Works
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            Get up and running in minutes with our simple three-step process.
          </p>
        </div>

        <div className="space-y-16">
          {steps.map((step, index) => (
            <div key={index} className={`flex flex-col ${index % 2 === 1 ? 'lg:flex-row-reverse' : 'lg:flex-row'} items-center gap-12`}>
              <div className="flex-1">
                <div className="flex items-center mb-4">
                  <div className="w-12 h-12 bg-primary-100 dark:bg-primary-900 rounded-full flex items-center justify-center mr-4">
                    <span className="text-xl font-bold text-primary-600 dark:text-primary-400">{step.step}</span>
                  </div>
                  <div className="text-primary-600 dark:text-primary-400">
                    {step.icon}
                  </div>
                </div>
                <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">{step.title}</h3>
                <p className="text-lg text-gray-600 dark:text-gray-300 mb-6">{step.description}</p>
                
                {/* Benefits list for each step */}
                {index === 0 && (
                  <ul className="space-y-2 text-gray-600 dark:text-gray-300">
                    <li className="flex items-center">
                      <svg className="w-4 h-4 text-primary-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Visual route builder or YAML configuration
                    </li>
                    <li className="flex items-center">
                      <svg className="w-4 h-4 text-primary-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Hot reload without downtime
                    </li>
                    <li className="flex items-center">
                      <svg className="w-4 h-4 text-primary-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Validation and testing tools
                    </li>
                  </ul>
                )}
                
                {index === 1 && (
                  <ul className="space-y-2 text-gray-600 dark:text-gray-300">
                    <li className="flex items-center">
                      <svg className="w-4 h-4 text-primary-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Docker, Kubernetes, or managed service
                    </li>
                    <li className="flex items-center">
                      <svg className="w-4 h-4 text-primary-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Auto-scaling based on demand
                    </li>
                    <li className="flex items-center">
                      <svg className="w-4 h-4 text-primary-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Health checks and self-healing
                    </li>
                  </ul>
                )}
                
                {index === 2 && (
                  <ul className="space-y-2 text-gray-600 dark:text-gray-300">
                    <li className="flex items-center">
                      <svg className="w-4 h-4 text-primary-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Real-time performance metrics
                    </li>
                    <li className="flex items-center">
                      <svg className="w-4 h-4 text-primary-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Alerting and notifications
                    </li>
                    <li className="flex items-center">
                      <svg className="w-4 h-4 text-primary-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Performance optimization insights
                    </li>
                  </ul>
                )}
              </div>
              
              <div className="flex-1">
                <div className="bg-gray-900 dark:bg-gray-800 rounded-lg p-6 text-green-400 dark:text-green-300 font-mono text-sm overflow-x-auto">
                  <pre>{step.code}</pre>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Quick Start CTA */}
        <div className="mt-20 text-center">
          <div className="bg-gradient-to-r from-primary-50 to-blue-50 dark:from-primary-900 dark:to-blue-900 rounded-2xl p-8">
            <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
              Ready to get started?
            </h3>
            <p className="text-gray-600 dark:text-gray-300 mb-6">
              Join thousands of developers already using Edge.link to power their MACH integrations.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <a href="/dashboard" className="btn btn-primary">
                Start Free Trial
              </a>
              <a href="#" className="btn btn-secondary">
                View Documentation
              </a>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}