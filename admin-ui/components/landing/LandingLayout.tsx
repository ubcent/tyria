'use client'

import { ReactNode, useState } from 'react';
import Link from 'next/link';
import { Bars3Icon, XMarkIcon } from '@heroicons/react/24/outline';

interface LandingLayoutProps {
  children: ReactNode;
}

export default function LandingLayout({ children }: LandingLayoutProps) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <div className="min-h-screen bg-white">
      {/* Navigation */}
      <nav className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <Link href="/" className="text-xl font-bold text-primary-600">
                Edge.link
              </Link>
            </div>
            
            {/* Desktop menu */}
            <div className="hidden md:flex items-center space-x-8">
              <Link href="#features" className="text-gray-500 hover:text-gray-900">
                Features
              </Link>
              <Link href="#pricing" className="text-gray-500 hover:text-gray-900">
                Pricing
              </Link>
              <Link href="#how-it-works" className="text-gray-500 hover:text-gray-900">
                How it Works
              </Link>
              <Link href="/dashboard" className="btn btn-primary">
                Admin Dashboard
              </Link>
            </div>

            {/* Mobile menu button */}
            <div className="md:hidden">
              <button
                onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                className="text-gray-500 hover:text-gray-900"
              >
                {mobileMenuOpen ? (
                  <XMarkIcon className="h-6 w-6" />
                ) : (
                  <Bars3Icon className="h-6 w-6" />
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Mobile menu */}
        {mobileMenuOpen && (
          <div className="md:hidden">
            <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3 bg-white border-t border-gray-200">
              <Link 
                href="#features" 
                className="block px-3 py-2 text-gray-500 hover:text-gray-900"
                onClick={() => setMobileMenuOpen(false)}
              >
                Features
              </Link>
              <Link 
                href="#pricing" 
                className="block px-3 py-2 text-gray-500 hover:text-gray-900"
                onClick={() => setMobileMenuOpen(false)}
              >
                Pricing
              </Link>
              <Link 
                href="#how-it-works" 
                className="block px-3 py-2 text-gray-500 hover:text-gray-900"
                onClick={() => setMobileMenuOpen(false)}
              >
                How it Works
              </Link>
              <Link 
                href="/dashboard" 
                className="block mx-3 my-2 btn btn-primary text-center"
                onClick={() => setMobileMenuOpen(false)}
              >
                Admin Dashboard
              </Link>
            </div>
          </div>
        )}
      </nav>

      {/* Main Content */}
      <main>
        {children}
      </main>

      {/* Footer */}
      <footer className="bg-gray-50 border-t border-gray-200">
        <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            <div className="col-span-1">
              <div className="text-xl font-bold text-primary-600 mb-4">
                Edge.link
              </div>
              <p className="text-gray-600 text-sm">
                Proxy-as-a-Service for MACH integrations. High-performance API routing, caching, and security.
              </p>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-gray-900 mb-4">Product</h3>
              <ul className="space-y-2 text-sm text-gray-600">
                <li><Link href="#features" className="hover:text-primary-600">Features</Link></li>
                <li><Link href="#pricing" className="hover:text-primary-600">Pricing</Link></li>
                <li><Link href="#how-it-works" className="hover:text-primary-600">How it Works</Link></li>
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-gray-900 mb-4">Company</h3>
              <ul className="space-y-2 text-sm text-gray-600">
                <li><a href="#" className="hover:text-primary-600">About</a></li>
                <li><a href="#" className="hover:text-primary-600">Contact</a></li>
                <li><a href="#" className="hover:text-primary-600">Support</a></li>
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-gray-900 mb-4">Resources</h3>
              <ul className="space-y-2 text-sm text-gray-600">
                <li><a href="#" className="hover:text-primary-600">Documentation</a></li>
                <li><a href="#" className="hover:text-primary-600">API Reference</a></li>
                <li><a href="#" className="hover:text-primary-600">GitHub</a></li>
              </ul>
            </div>
          </div>
          <div className="mt-8 pt-8 border-t border-gray-200">
            <p className="text-center text-sm text-gray-500">
              © 2024 Edge.link. All rights reserved.
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}