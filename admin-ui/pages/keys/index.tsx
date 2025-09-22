import { useState, useEffect } from 'react';
import Layout from '@/components/layouts/Layout';
import { ApiKey } from '@/lib/db';
import { PlusIcon, EyeIcon, EyeSlashIcon, TrashIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';

export default function ApiKeysPage() {
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [visibleKeys, setVisibleKeys] = useState<Set<number>>(new Set());

  useEffect(() => {
    fetchApiKeys();
  }, []);

  const fetchApiKeys = async () => {
    try {
      const response = await fetch('/api/keys');
      if (response.ok) {
        const data = await response.json();
        setApiKeys(data);
      }
    } catch (error) {
      console.error('Failed to fetch API keys:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this API key?')) return;

    try {
      const response = await fetch(`/api/keys/${id}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        setApiKeys(apiKeys.filter(key => key.id !== id));
      }
    } catch (error) {
      console.error('Failed to delete API key:', error);
    }
  };

  const toggleKeyEnabled = async (id: number, enabled: boolean) => {
    try {
      const response = await fetch(`/api/keys/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled }),
      });

      if (response.ok) {
        setApiKeys(apiKeys.map(key => 
          key.id === id ? { ...key, enabled } : key
        ));
      }
    } catch (error) {
      console.error('Failed to toggle API key:', error);
    }
  };

  const toggleKeyVisibility = (id: number) => {
    const newVisible = new Set(visibleKeys);
    if (newVisible.has(id)) {
      newVisible.delete(id);
    } else {
      newVisible.add(id);
    }
    setVisibleKeys(newVisible);
  };

  const maskKey = (key: string) => {
    if (key.length <= 8) return '*'.repeat(key.length);
    return key.substring(0, 4) + '*'.repeat(key.length - 8) + key.substring(key.length - 4);
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
            <h1 className="text-2xl font-semibold text-gray-900">API Keys</h1>
            <p className="mt-1 text-sm text-gray-500">
              Manage authentication keys for accessing your proxy API
            </p>
          </div>
          <Link
            href="/keys/new"
            className="btn btn-primary inline-flex items-center"
          >
            <PlusIcon className="h-4 w-4 mr-2" />
            Generate API Key
          </Link>
        </div>

        {/* API Keys Table */}
        <div className="card p-0 overflow-hidden">
          {apiKeys.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="table">
                <thead>
                  <tr>
                    <th className="table-header">Name</th>
                    <th className="table-header">Key</th>
                    <th className="table-header">Permissions</th>
                    <th className="table-header">Rate Limit</th>
                    <th className="table-header">Status</th>
                    <th className="table-header">Created</th>
                    <th className="table-header">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {apiKeys.map((apiKey) => (
                    <tr key={apiKey.id}>
                      <td className="table-cell font-medium">
                        {apiKey.name}
                      </td>
                      <td className="table-cell">
                        <div className="flex items-center space-x-2">
                          <code className="bg-gray-100 px-2 py-1 rounded text-sm font-mono">
                            {visibleKeys.has(apiKey.id) ? apiKey.key_value : maskKey(apiKey.key_value)}
                          </code>
                          <button
                            onClick={() => toggleKeyVisibility(apiKey.id)}
                            className="text-gray-400 hover:text-gray-600"
                          >
                            {visibleKeys.has(apiKey.id) ? (
                              <EyeSlashIcon className="h-4 w-4" />
                            ) : (
                              <EyeIcon className="h-4 w-4" />
                            )}
                          </button>
                        </div>
                      </td>
                      <td className="table-cell">
                        <div className="flex flex-wrap gap-1">
                          {apiKey.permissions.map((permission) => (
                            <span
                              key={permission}
                              className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800"
                            >
                              {permission}
                            </span>
                          ))}
                        </div>
                      </td>
                      <td className="table-cell">
                        <span className="text-sm text-gray-600">
                          {apiKey.rate_limit.toLocaleString()}/hour
                        </span>
                      </td>
                      <td className="table-cell">
                        <button
                          onClick={() => toggleKeyEnabled(apiKey.id, !apiKey.enabled)}
                          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                            apiKey.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {apiKey.enabled ? 'Active' : 'Disabled'}
                        </button>
                      </td>
                      <td className="table-cell">
                        <span className="text-sm text-gray-500">
                          {new Date(apiKey.created_at).toLocaleDateString()}
                        </span>
                      </td>
                      <td className="table-cell">
                        <button
                          onClick={() => handleDelete(apiKey.id)}
                          className="text-red-600 hover:text-red-900"
                        >
                          <TrashIcon className="h-4 w-4" />
                        </button>
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
                  d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
                />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">No API keys</h3>
              <p className="mt-1 text-sm text-gray-500">
                Get started by generating your first API key.
              </p>
              <div className="mt-6">
                <Link href="/keys/new" className="btn btn-primary">
                  <PlusIcon className="h-4 w-4 mr-2" />
                  Generate API Key
                </Link>
              </div>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}