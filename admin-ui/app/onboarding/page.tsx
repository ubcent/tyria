'use client'

import { useState } from 'react';
import Link from 'next/link';
import { CheckIcon, ClipboardIcon, GlobeAltIcon, CogIcon, RocketLaunchIcon } from '@heroicons/react/24/outline';

export default function OnboardingPage() {
  const [currentStep, setCurrentStep] = useState(1);
  const [domain, setDomain] = useState('');
  const [targetUrl, setTargetUrl] = useState('');
  const [proxyUrl, setProxyUrl] = useState('');
  const [copied, setCopied] = useState(false);

  const handleStep1Complete = () => {
    // Generate a proxy URL based on domain
    const generatedProxy = `https://${domain.replace(/\./g, '-')}.edge.link`;
    setProxyUrl(generatedProxy);
    setCurrentStep(2);
  };

  const handleStep2Complete = () => {
    setCurrentStep(3);
  };

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(proxyUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Header */}
      <nav className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <div className="text-xl font-bold text-primary-600 dark:text-primary-400">
                Edge.link
              </div>
            </div>
          </div>
        </div>
      </nav>

      {/* Progress Bar */}
      <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              {[1, 2, 3].map((step) => (
                <div key={step} className="flex items-center">
                  <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                    currentStep >= step 
                      ? 'bg-primary-600 text-white' 
                      : 'bg-gray-200 dark:bg-gray-600 text-gray-600 dark:text-gray-300'
                  }`}>
                    {currentStep > step ? (
                      <CheckIcon className="w-5 h-5" />
                    ) : (
                      step
                    )}
                  </div>
                  {step < 3 && (
                    <div className={`w-12 h-0.5 mx-2 ${
                      currentStep > step ? 'bg-primary-600' : 'bg-gray-200 dark:bg-gray-600'
                    }`} />
                  )}
                </div>
              ))}
            </div>
            <div className="text-sm text-gray-500 dark:text-gray-400">
              Step {currentStep} of 3
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {currentStep === 1 && (
          <div className="text-center">
            <GlobeAltIcon className="mx-auto h-12 w-12 text-primary-600 dark:text-primary-400" />
            <h1 className="mt-4 text-3xl font-bold text-gray-900 dark:text-gray-100">
              Connect Your Domain
            </h1>
            <p className="mt-2 text-lg text-gray-600 dark:text-gray-400">
              Enter your domain to get started with Edge.link proxy
            </p>

            <div className="mt-8 max-w-md mx-auto">
              <div className="space-y-4">
                <div>
                  <label htmlFor="domain" className="block text-sm font-medium text-gray-700 dark:text-gray-300 text-left">
                    Your domain
                  </label>
                  <input
                    type="text"
                    id="domain"
                    value={domain}
                    onChange={(e) => setDomain(e.target.value)}
                    placeholder="api.yourcompany.com"
                    className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 px-3 py-2 shadow-sm focus:border-primary-500 focus:ring-primary-500 dark:bg-gray-700 dark:text-gray-100 sm:text-sm"
                  />
                </div>
                <button
                  onClick={handleStep1Complete}
                  disabled={!domain.trim()}
                  className="w-full btn btn-primary py-3 text-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Create Proxy URL
                </button>
              </div>
            </div>

            <div className="mt-8 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
              <p className="text-sm text-blue-800 dark:text-blue-200">
                <strong>Step 1 of 2:</strong> We&apos;ll generate a unique proxy URL for your domain that you can use immediately.
              </p>
            </div>
          </div>
        )}

        {currentStep === 2 && (
          <div className="text-center">
            <CogIcon className="mx-auto h-12 w-12 text-primary-600 dark:text-primary-400" />
            <h1 className="mt-4 text-3xl font-bold text-gray-900 dark:text-gray-100">
              Configure Target URL
            </h1>
            <p className="mt-2 text-lg text-gray-600 dark:text-gray-400">
              Where should we proxy requests to?
            </p>

            <div className="mt-8 max-w-md mx-auto">
              <div className="space-y-4">
                <div>
                  <label htmlFor="targetUrl" className="block text-sm font-medium text-gray-700 dark:text-gray-300 text-left">
                    Target API URL
                  </label>
                  <input
                    type="url"
                    id="targetUrl"
                    value={targetUrl}
                    onChange={(e) => setTargetUrl(e.target.value)}
                    placeholder="https://your-api-server.com"
                    className="mt-1 block w-full rounded-md border-gray-300 dark:border-gray-600 px-3 py-2 shadow-sm focus:border-primary-500 focus:ring-primary-500 dark:bg-gray-700 dark:text-gray-100 sm:text-sm"
                  />
                </div>
                
                <div className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg">
                  <h3 className="text-sm font-medium text-gray-900 dark:text-gray-100 mb-2">Your proxy URL:</h3>
                  <div className="flex items-center space-x-2">
                    <code className="flex-1 text-sm bg-white dark:bg-gray-700 p-2 rounded border text-primary-600 dark:text-primary-400">
                      {proxyUrl}
                    </code>
                    <button
                      onClick={copyToClipboard}
                      className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                    >
                      <ClipboardIcon className="w-5 h-5" />
                    </button>
                  </div>
                  {copied && (
                    <p className="text-xs text-green-600 dark:text-green-400 mt-1">Copied to clipboard!</p>
                  )}
                </div>

                <button
                  onClick={handleStep2Complete}
                  disabled={!targetUrl}
                  className="w-full btn btn-primary py-3 text-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Complete Setup
                </button>
              </div>
            </div>

            <div className="mt-8 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
              <p className="text-sm text-blue-800 dark:text-blue-200">
                <strong>Step 2 of 2:</strong> Your proxy is ready! All requests to your proxy URL will be forwarded to your target API.
              </p>
            </div>
          </div>
        )}

        {currentStep === 3 && (
          <div className="text-center">
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 dark:bg-green-900">
              <CheckIcon className="h-6 w-6 text-green-600 dark:text-green-400" />
            </div>
            <h1 className="mt-4 text-3xl font-bold text-gray-900 dark:text-gray-100">
              🎉 Your Proxy is Live!
            </h1>
            <p className="mt-2 text-lg text-gray-600 dark:text-gray-400">
              Congratulations! Your Edge.link proxy is now active and ready to use.
            </p>

            <div className="mt-8 space-y-6">
              {/* Proxy URL Display */}
              <div className="bg-white dark:bg-gray-800 p-6 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
                <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-3">Your Proxy URL</h3>
                <div className="flex items-center space-x-2">
                  <code className="flex-1 text-sm bg-gray-50 dark:bg-gray-700 p-3 rounded border text-primary-600 dark:text-primary-400 font-mono">
                    {proxyUrl}
                  </code>
                  <button
                    onClick={copyToClipboard}
                    className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                  >
                    <ClipboardIcon className="w-5 h-5" />
                  </button>
                </div>
                {copied && (
                  <p className="text-xs text-green-600 dark:text-green-400 mt-1">Copied to clipboard!</p>
                )}
              </div>

              {/* Quick Start */}
              <div className="bg-white dark:bg-gray-800 p-6 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
                <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-3">Quick Start</h3>
                <div className="text-left space-y-2">
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Test your proxy with curl:
                  </p>
                  <code className="block text-xs bg-gray-50 dark:bg-gray-700 p-3 rounded border font-mono text-gray-800 dark:text-gray-200">
                    curl {proxyUrl}/your-endpoint
                  </code>
                </div>
              </div>

              {/* Next Steps */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Link 
                  href="/dashboard"
                  className="flex items-center p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm hover:shadow-md transition-shadow"
                >
                  <CogIcon className="h-8 w-8 text-primary-600 dark:text-primary-400 mr-3" />
                  <div className="text-left">
                    <h4 className="text-sm font-medium text-gray-900 dark:text-gray-100">Configure Settings</h4>
                    <p className="text-xs text-gray-600 dark:text-gray-400">Add caching, rate limiting, and more</p>
                  </div>
                </Link>
                
                <a 
                  href="#"
                  className="flex items-center p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm hover:shadow-md transition-shadow"
                >
                  <RocketLaunchIcon className="h-8 w-8 text-primary-600 dark:text-primary-400 mr-3" />
                  <div className="text-left">
                    <h4 className="text-sm font-medium text-gray-900 dark:text-gray-100">View Documentation</h4>
                    <p className="text-xs text-gray-600 dark:text-gray-400">Learn about advanced features</p>
                  </div>
                </a>
              </div>

              <Link
                href="/dashboard"
                className="w-full btn btn-primary py-3 text-lg font-medium"
              >
                Go to Dashboard
              </Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}