import { useState, useEffect } from 'react';
import Layout from '@/components/layouts/Layout';
import { ProxyRoute } from '@/lib/db';
import { PlusIcon, PencilIcon, TrashIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';

export default function RoutesPage() {
  const [routes, setRoutes] = useState<ProxyRoute[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchRoutes();
  }, []);

  const fetchRoutes = async () => {
    try {
      const response = await fetch('/api/routes');
      if (response.ok) {
        const data = await response.json();
        setRoutes(data);
      }
    } catch (error) {
      console.error('Failed to fetch routes:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this route?')) return;

    try {
      const response = await fetch(`/api/routes/${id}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        setRoutes(routes.filter(route => route.id !== id));
      }
    } catch (error) {
      console.error('Failed to delete route:', error);
    }
  };

  const toggleRoute = async (id: number, enabled: boolean) => {
    try {
      const response = await fetch(`/api/routes/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled }),
      });

      if (response.ok) {
        setRoutes(routes.map(route => 
          route.id === id ? { ...route, enabled } : route
        ));
      }
    } catch (error) {
      console.error('Failed to toggle route:', error);
    }
  };

  if (loading) {
    return (
      <Layout>
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-6"></div>
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-16 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-semibold text-gray-900">Proxy Routes</h1>
            <p className="mt-1 text-sm text-gray-500">
              Manage your API proxy route configurations
            </p>
          </div>
          <Link
            href="/routes/new"
            className="btn btn-primary inline-flex items-center"
          >
            <PlusIcon className="h-4 w-4 mr-2" />
            Add Route
          </Link>
        </div>

        {/* Routes Table */}
        <div className="card p-0 overflow-hidden">
          {routes.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="table">
                <thead>
                  <tr>
                    <th className="table-header">Path</th>
                    <th className="table-header">Target</th>
                    <th className="table-header">Methods</th>
                    <th className="table-header">Cache</th>
                    <th className="table-header">Rate Limit</th>
                    <th className="table-header">Auth</th>
                    <th className="table-header">Status</th>
                    <th className="table-header">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {routes.map((route) => (
                    <tr key={route.id}>
                      <td className="table-cell font-medium">
                        <code className="bg-gray-100 px-2 py-1 rounded text-sm">
                          {route.path}
                        </code>
                      </td>
                      <td className="table-cell">
                        <span className="text-blue-600 hover:text-blue-900">
                          {route.target}
                        </span>
                      </td>
                      <td className="table-cell">
                        <div className="flex gap-1">
                          {route.methods.map((method) => (
                            <span
                              key={method}
                              className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                method === 'GET' ? 'bg-green-100 text-green-800' :
                                method === 'POST' ? 'bg-blue-100 text-blue-800' :
                                method === 'PUT' ? 'bg-yellow-100 text-yellow-800' :
                                method === 'DELETE' ? 'bg-red-100 text-red-800' :
                                'bg-gray-100 text-gray-800'
                              }`}
                            >
                              {method}
                            </span>
                          ))}
                        </div>
                      </td>
                      <td className="table-cell">
                        <div className="flex items-center">
                          <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            route.cache_enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                          }`}>
                            {route.cache_enabled ? `${route.cache_ttl}s` : 'Disabled'}
                          </span>
                        </div>
                      </td>
                      <td className="table-cell">
                        <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                          route.rate_limit_enabled ? 'bg-orange-100 text-orange-800' : 'bg-gray-100 text-gray-800'
                        }`}>
                          {route.rate_limit_enabled ? `${route.rate_limit_rate}/min` : 'Disabled'}
                        </span>
                      </td>
                      <td className="table-cell">
                        <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                          route.auth_required ? 'bg-purple-100 text-purple-800' : 'bg-gray-100 text-gray-800'
                        }`}>
                          {route.auth_required ? 'Required' : 'None'}
                        </span>
                      </td>
                      <td className="table-cell">
                        <button
                          onClick={() => toggleRoute(route.id, !route.enabled)}
                          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            route.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {route.enabled ? 'Active' : 'Inactive'}
                        </button>
                      </td>
                      <td className="table-cell">
                        <div className="flex items-center space-x-2">
                          <Link
                            href={`/routes/${route.id}/edit`}
                            className="text-primary-600 hover:text-primary-900"
                          >
                            <PencilIcon className="h-4 w-4" />
                          </Link>
                          <button
                            onClick={() => handleDelete(route.id)}
                            className="text-red-600 hover:text-red-900"
                          >
                            <TrashIcon className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-12">
              <svg
                className="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">No routes configured</h3>
              <p className="mt-1 text-sm text-gray-500">
                Get started by creating your first proxy route.
              </p>
              <div className="mt-6">
                <Link href="/routes/new" className="btn btn-primary">
                  <PlusIcon className="h-4 w-4 mr-2" />
                  Add Route
                </Link>
              </div>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}