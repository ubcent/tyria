import Link from 'next/link';

export default function CTASection() {
  return (
    <section className="py-20 bg-gradient-to-r from-primary-600 to-blue-600">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <h2 className="text-3xl lg:text-4xl font-bold text-white mb-6">
          Ready to supercharge your API infrastructure?
        </h2>
        <p className="text-xl text-blue-100 mb-8 max-w-3xl mx-auto">
          Join thousands of developers who trust Edge.link for their MACH integrations. 
          Start free, scale seamlessly, and never worry about API performance again.
        </p>
        
        <div className="flex flex-col sm:flex-row gap-4 justify-center mb-12">
          <Link 
            href="/dashboard" 
            className="bg-white text-primary-600 hover:bg-gray-50 px-8 py-3 rounded-md font-medium text-lg inline-flex items-center justify-center transition-colors"
          >
            Start Free Trial
            <svg className="ml-2 w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </Link>
          <a 
            href="#contact" 
            className="border-2 border-white text-white hover:bg-white hover:text-primary-600 px-8 py-3 rounded-md font-medium text-lg inline-flex items-center justify-center transition-colors"
          >
            Talk to Sales
          </a>
        </div>

        {/* Trust indicators */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 items-center opacity-80">
          <div className="text-white text-center">
            <div className="text-2xl font-bold">99.9%</div>
            <div className="text-sm text-blue-100">Uptime SLA</div>
          </div>
          <div className="text-white text-center">
            <div className="text-2xl font-bold">10M+</div>
            <div className="text-sm text-blue-100">Requests Served</div>
          </div>
          <div className="text-white text-center">
            <div className="text-2xl font-bold">500+</div>
            <div className="text-sm text-blue-100">Happy Customers</div>
          </div>
          <div className="text-white text-center">
            <div className="text-2xl font-bold">24/7</div>
            <div className="text-sm text-blue-100">Support</div>
          </div>
        </div>
      </div>
    </section>
  );
}