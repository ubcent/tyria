import Link from 'next/link';

export default function PricingSection() {
  const plans = [
    {
      name: "Starter",
      price: "$0",
      period: "Free forever",
      description: "Perfect for development and small projects",
      features: [
        "Up to 10,000 requests/month",
        "Basic caching and rate limiting",
        "API key authentication",
        "Community support",
        "1 proxy route"
      ],
      buttonText: "Start Free",
      buttonClass: "btn btn-secondary",
      popular: false
    },
    {
      name: "Professional",
      price: "$29",
      period: "per month",
      description: "For growing teams and production workloads",
      features: [
        "Up to 1M requests/month",
        "Advanced caching strategies",
        "Schema validation",
        "Priority support",
        "Unlimited routes",
        "Real-time analytics",
        "Custom rate limits"
      ],
      buttonText: "Start Trial",
      buttonClass: "btn btn-primary",
      popular: true
    },
    {
      name: "Enterprise",
      price: "Custom",
      period: "Contact us",
      description: "For high-scale applications and custom needs",
      features: [
        "Unlimited requests",
        "Dedicated infrastructure",
        "24/7 premium support",
        "Custom integrations",
        "SLA guarantees",
        "Advanced security",
        "On-premise deployment"
      ],
      buttonText: "Contact Sales",
      buttonClass: "btn btn-secondary",
      popular: false
    }
  ];

  return (
    <section id="pricing" className="py-20 bg-gray-50 dark:bg-gray-800">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl lg:text-4xl font-bold text-gray-900 dark:text-gray-100 mb-4">
            Simple, transparent pricing
          </h2>
          <p className="text-xl text-gray-600 dark:text-gray-300 max-w-3xl mx-auto">
            Start free, scale as you grow. No hidden fees, no vendor lock-in.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {plans.map((plan, index) => (
            <div 
              key={index} 
              className={`bg-white dark:bg-gray-800 rounded-lg shadow-lg border-2 p-8 relative ${
                plan.popular ? 'border-primary-500 scale-105' : 'border-gray-200 dark:border-gray-700'
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                  <span className="bg-primary-500 text-white px-4 py-1 rounded-full text-sm font-medium">
                    Most Popular
                  </span>
                </div>
              )}
              
              <div className="text-center mb-8">
                <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">{plan.name}</h3>
                <div className="mb-2">
                  <span className="text-4xl font-bold text-gray-900 dark:text-gray-100">{plan.price}</span>
                  {plan.period !== "Contact us" && (
                    <span className="text-gray-500 dark:text-gray-400 ml-1">/{plan.period}</span>
                  )}
                </div>
                <p className="text-gray-600 dark:text-gray-300">{plan.description}</p>
              </div>

              <ul className="space-y-4 mb-8">
                {plan.features.map((feature, featureIndex) => (
                  <li key={featureIndex} className="flex items-start">
                    <svg className="w-5 h-5 text-primary-500 mt-0.5 mr-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    <span className="text-gray-700 dark:text-gray-300">{feature}</span>
                  </li>
                ))}
              </ul>

              <div className="text-center">
                <Link 
                  href={plan.name === "Enterprise" ? "#contact" : "/signup"} 
                  className={`${plan.buttonClass} w-full justify-center`}
                >
                  {plan.buttonText}
                </Link>
              </div>
            </div>
          ))}
        </div>

        {/* FAQ */}
        <div className="mt-20">
          <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 text-center mb-8">
            Frequently Asked Questions
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl mx-auto">
            <div>
              <h4 className="font-semibold text-gray-900 dark:text-gray-100 mb-2">Can I upgrade or downgrade anytime?</h4>
              <p className="text-gray-600 dark:text-gray-300 text-sm">Yes, you can change your plan at any time. Changes take effect immediately with prorated billing.</p>
            </div>
            <div>
              <h4 className="font-semibold text-gray-900 dark:text-gray-100 mb-2">What happens if I exceed my plan limits?</h4>
              <p className="text-gray-600 dark:text-gray-300 text-sm">We'll notify you before hitting limits. Overage is billed at $0.05 per 1K requests.</p>
            </div>
            <div>
              <h4 className="font-semibold text-gray-900 dark:text-gray-100 mb-2">Do you offer custom enterprise plans?</h4>
              <p className="text-gray-600 dark:text-gray-300 text-sm">Yes, we offer custom enterprise solutions with dedicated support and infrastructure.</p>
            </div>
            <div>
              <h4 className="font-semibold text-gray-900 dark:text-gray-100 mb-2">Is there a free trial for paid plans?</h4>
              <p className="text-gray-600 dark:text-gray-300 text-sm">Yes, all paid plans include a 14-day free trial. No credit card required to start.</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}