/**
 * API Client Demo Page
 * Demonstrates the enhanced API client features
 */
'use client';

import React, { useState } from 'react';
import { toast } from '../../lib/toast';
import { apiClient, routesApi, authApi } from '../../lib/api/client';
import ToastContainer from '../../components/ToastContainer';

export default function ApiClientDemoPage() {
  const [loading, setLoading] = useState(false);

  const testErrorHandling = async () => {
    setLoading(true);
    try {
      // This will likely trigger a 401/403/5xx error and show a toast
      await apiClient.get('/api/test/nonexistent-endpoint');
      toast.success('Unexpected success!');
    } catch (error) {
      console.log('Expected error caught:', error);
    } finally {
      setLoading(false);
    }
  };

  const testRetryLogic = async () => {
    setLoading(true);
    try {
      // Simulate a GET request that might fail and retry
      await apiClient.get('/api/test/retry-test');
      toast.success('Request succeeded (or retries exhausted)');
    } catch (error) {
      console.log('Request failed after retries:', error);
    } finally {
      setLoading(false);
    }
  };

  const testTypedRequests = async () => {
    setLoading(true);
    try {
      // Test getting user profile (should be typed)
      const profile = await authApi.getProfile();
      console.log('User profile:', profile);
      toast.success('Profile loaded successfully');
    } catch (error) {
      console.log('Profile request failed:', error);
    } finally {
      setLoading(false);
    }
  };

  const testToastTypes = () => {
    toast.success('Success message', 'This is a success notification');
    setTimeout(() => {
      toast.error('Error message', 'This is an error notification');
    }, 500);
    setTimeout(() => {
      toast.warning('Warning message', 'This is a warning notification');
    }, 1000);
    setTimeout(() => {
      toast.info('Info message', 'This is an info notification');
    }, 1500);
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 py-8">
      <ToastContainer />
      
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <h1 className="text-2xl font-semibold text-gray-900 dark:text-gray-100">
              Enhanced API Client Demo
            </h1>
            <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
              Test the enhanced API client features including error handling, retry logic, and toast notifications.
            </p>
          </div>
          
          <div className="p-6">
            <div className="grid gap-6 md:grid-cols-2">
              {/* Features Overview */}
              <div className="space-y-4">
                <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                  ✨ Enhanced Features
                </h2>
                <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
                  <li className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    Environment-based base URL configuration
                  </li>
                  <li className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    Automatic JWT cookie inclusion
                  </li>
                  <li className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    Toast notifications for errors
                  </li>
                  <li className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    Retry logic for GET requests
                  </li>
                  <li className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    Type-safe HTTP helpers
                  </li>
                  <li className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    Comprehensive error handling
                  </li>
                </ul>
              </div>

              {/* Test Actions */}
              <div className="space-y-4">
                <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100">
                  🧪 Test Features
                </h2>
                <div className="space-y-3">
                  <button
                    onClick={testToastTypes}
                    className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                    disabled={loading}
                  >
                    Test Toast Notifications
                  </button>
                  
                  <button
                    onClick={testErrorHandling}
                    className="w-full px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                    disabled={loading}
                  >
                    {loading ? 'Testing...' : 'Test Error Handling'}
                  </button>
                  
                  <button
                    onClick={testRetryLogic}
                    className="w-full px-4 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700 focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:ring-offset-2"
                    disabled={loading}
                  >
                    {loading ? 'Testing...' : 'Test Retry Logic'}
                  </button>
                  
                  <button
                    onClick={testTypedRequests}
                    className="w-full px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
                    disabled={loading}
                  >
                    {loading ? 'Loading...' : 'Test Typed API Call'}
                  </button>
                </div>
              </div>
            </div>

            {/* Implementation Details */}
            <div className="mt-8 pt-8 border-t border-gray-200 dark:border-gray-700">
              <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
                📋 Implementation Details
              </h2>
              <div className="grid gap-6 md:grid-cols-2">
                <div>
                  <h3 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
                    HTTP Client Methods
                  </h3>
                  <code className="text-sm bg-gray-100 dark:bg-gray-700 p-3 rounded block">
                    {`// Type-safe API calls
await apiClient.get<Route[]>('/api/v1/routes');
await apiClient.post<Route>('/api/v1/routes', data);
await apiClient.patch<Route>('/api/v1/routes/1', updates);
await apiClient.delete('/api/v1/routes/1');`}
                  </code>
                </div>
                <div>
                  <h3 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
                    Error Handling
                  </h3>
                  <code className="text-sm bg-gray-100 dark:bg-gray-700 p-3 rounded block">
                    {`// Automatic error handling
// 401 → Redirect to signin + toast
// 403 → Permission error toast
// 5xx → Server error toast
// Network → Connection error toast`}
                  </code>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}