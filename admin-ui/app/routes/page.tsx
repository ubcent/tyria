'use client'

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Layout from '@/components/layouts/Layout';
import { routesApi, type Route } from '@/lib/api/client';
import { PlusIcon, PencilIcon, TrashIcon, XMarkIcon } from '@heroicons/react/24/outline';
import CodeMirror from '@uiw/react-codemirror';
import { json } from '@codemirror/lang-json';
import { oneDark } from '@codemirror/theme-one-dark';

interface RouteFormData {
  name: string;
  match_path: string;
  upstream_url: string;
  headers_json: string;
  auth_mode: 'none' | 'api_key' | 'basic';
  caching_policy_json: string;
  rate_limit_policy_json: string;
  enabled: boolean;
}

export default function RoutesPage() {
  const queryClient = useQueryClient();
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [editingRoute, setEditingRoute] = useState<Route | null>(null);
  const [formData, setFormData] = useState<RouteFormData>({
    name: '',
    match_path: '',
    upstream_url: '',
    headers_json: '{}',
    auth_mode: 'none',
    caching_policy_json: '{"enabled": false, "ttl_seconds": 300}',
    rate_limit_policy_json: '{"enabled": false, "requests_per_minute": 100}',
    enabled: true,
  });
  const [jsonErrors, setJsonErrors] = useState<Record<string, string>>({});

  const { data: routes, isLoading } = useQuery<Route[]>({
    queryKey: ['routes'],
    queryFn: routesApi.getAll,
  });

  const deleteMutation = useMutation({
    mutationFn: routesApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['routes'] });
    },
  });

  const createMutation = useMutation({
    mutationFn: routesApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['routes'] });
      setIsDrawerOpen(false);
      resetForm();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Route> }) =>
      routesApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['routes'] });
      setIsDrawerOpen(false);
      resetForm();
    },
  });

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: number; enabled: boolean }) =>
      routesApi.update(id, { enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['routes'] });
    },
  });

  const validateJSON = (jsonString: string): string | null => {
    try {
      JSON.parse(jsonString);
      return null;
    } catch (error) {
      return (error as Error).message;
    }
  };

  const handleJsonChange = (field: keyof RouteFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    const error = validateJSON(value);
    setJsonErrors(prev => ({
      ...prev,
      [field]: error || ''
    }));
  };

  const resetForm = () => {
    setFormData({
      name: '',
      match_path: '',
      upstream_url: '',
      headers_json: '{}',
      auth_mode: 'none',
      caching_policy_json: '{"enabled": false, "ttl_seconds": 300}',
      rate_limit_policy_json: '{"enabled": false, "requests_per_minute": 100}',
      enabled: true,
    });
    setEditingRoute(null);
    setJsonErrors({});
  };

  const openDrawer = (route?: Route) => {
    if (route) {
      setEditingRoute(route);
      setFormData({
        name: route.name,
        match_path: route.match_path,
        upstream_url: route.upstream_url,
        headers_json: JSON.stringify(route.headers_json, null, 2),
        auth_mode: route.auth_mode,
        caching_policy_json: JSON.stringify(route.caching_policy_json, null, 2),
        rate_limit_policy_json: JSON.stringify(route.rate_limit_policy_json, null, 2),
        enabled: route.enabled,
      });
    } else {
      resetForm();
    }
    setIsDrawerOpen(true);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate all JSON fields
    const errors: Record<string, string> = {};
    ['headers_json', 'caching_policy_json', 'rate_limit_policy_json'].forEach(field => {
      const error = validateJSON(formData[field as keyof RouteFormData] as string);
      if (error) {
        errors[field] = error;
      }
    });

    if (Object.keys(errors).length > 0) {
      setJsonErrors(errors);
      return;
    }

    try {
      const routeData = {
        name: formData.name,
        match_path: formData.match_path,
        upstream_url: formData.upstream_url,
        headers_json: JSON.parse(formData.headers_json),
        auth_mode: formData.auth_mode,
        caching_policy_json: JSON.parse(formData.caching_policy_json),
        rate_limit_policy_json: JSON.parse(formData.rate_limit_policy_json),
        enabled: formData.enabled,
      };

      if (editingRoute) {
        updateMutation.mutate({ id: editingRoute.id, data: routeData });
      } else {
        createMutation.mutate(routeData);
      }
    } catch (error) {
      console.error('Error submitting form:', error);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this route?')) return;
    deleteMutation.mutate(id);
  };

  const toggleRoute = async (id: number, enabled: boolean) => {
    toggleMutation.mutate({ id, enabled });
  };

  if (isLoading) {
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
            <h1 className="text-2xl font-semibold text-gray-900">Routes</h1>
            <p className="mt-1 text-sm text-gray-500">
              Manage your API proxy route configurations
            </p>
          </div>
          <button
            onClick={() => openDrawer()}
            className="btn btn-primary inline-flex items-center"
          >
            <PlusIcon className="h-4 w-4 mr-2" />
            Add Route
          </button>
        </div>

        {/* Routes Table */}
        <div className="card p-0 overflow-hidden">
          {routes && routes.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="table">
                <thead>
                  <tr>
                    <th className="table-header">Name</th>
                    <th className="table-header">Path</th>
                    <th className="table-header">Upstream URL</th>
                    <th className="table-header">Auth Mode</th>
                    <th className="table-header">Status</th>
                    <th className="table-header">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {routes.map((route) => (
                    <tr key={route.id}>
                      <td className="table-cell font-medium">
                        {route.name}
                      </td>
                      <td className="table-cell">
                        <code className="bg-gray-100 px-2 py-1 rounded text-sm">
                          {route.match_path}
                        </code>
                      </td>
                      <td className="table-cell">
                        <span className="text-blue-600 truncate max-w-xs inline-block">
                          {route.upstream_url}
                        </span>
                      </td>
                      <td className="table-cell">
                        <span className={`badge ${
                          route.auth_mode === 'none' ? 'badge-gray' :
                          route.auth_mode === 'api_key' ? 'badge-blue' :
                          'badge-green'
                        }`}>
                          {route.auth_mode}
                        </span>
                      </td>
                      <td className="table-cell">
                        <label className="inline-flex items-center">
                          <input
                            type="checkbox"
                            checked={route.enabled}
                            onChange={(e) => toggleRoute(route.id, e.target.checked)}
                            className="form-checkbox h-4 w-4 text-primary-600"
                          />
                          <span className="ml-2 text-sm">
                            {route.enabled ? 'Enabled' : 'Disabled'}
                          </span>
                        </label>
                      </td>
                      <td className="table-cell">
                        <div className="flex items-center space-x-3">
                          <button
                            onClick={() => openDrawer(route)}
                            className="text-primary-600 hover:text-primary-700"
                          >
                            <PencilIcon className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => handleDelete(route.id)}
                            className="text-red-600 hover:text-red-700"
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
              <div className="inline-flex items-center justify-center w-16 h-16 bg-gray-100 rounded-full mb-4">
                <PlusIcon className="h-8 w-8 text-gray-400" />
              </div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">No routes yet</h3>
              <p className="text-gray-500 mb-4">
                Get started by creating your first proxy route.
              </p>
              <button onClick={() => openDrawer()} className="btn btn-primary">
                <PlusIcon className="h-4 w-4 mr-2" />
                Add Route
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Drawer */}
      {isDrawerOpen && (
        <div className="fixed inset-0 z-50 overflow-hidden">
          <div className="absolute inset-0 bg-black bg-opacity-50" onClick={() => setIsDrawerOpen(false)} />
          <div className="absolute right-0 top-0 h-full w-full max-w-2xl bg-white shadow-xl">
            <div className="flex h-full flex-col">
              {/* Header */}
              <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
                <h2 className="text-lg font-medium text-gray-900">
                  {editingRoute ? 'Edit Route' : 'Add Route'}
                </h2>
                <button
                  onClick={() => setIsDrawerOpen(false)}
                  className="text-gray-400 hover:text-gray-500"
                >
                  <XMarkIcon className="h-6 w-6" />
                </button>
              </div>

              {/* Form */}
              <div className="flex-1 overflow-y-auto px-6 py-6">
                <form onSubmit={handleSubmit} className="space-y-6">
                  {/* Basic Fields */}
                  <div className="grid grid-cols-1 gap-6">
                    <div>
                      <label className="form-label">Name</label>
                      <input
                        type="text"
                        value={formData.name}
                        onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                        className="form-input"
                        placeholder="Route name"
                        required
                      />
                    </div>

                    <div>
                      <label className="form-label">Match Path</label>
                      <input
                        type="text"
                        value={formData.match_path}
                        onChange={(e) => setFormData(prev => ({ ...prev, match_path: e.target.value }))}
                        className="form-input"
                        placeholder="/api/v1/users"
                        required
                      />
                    </div>

                    <div>
                      <label className="form-label">Upstream URL</label>
                      <input
                        type="url"
                        value={formData.upstream_url}
                        onChange={(e) => setFormData(prev => ({ ...prev, upstream_url: e.target.value }))}
                        className="form-input"
                        placeholder="https://api.example.com"
                        required
                      />
                    </div>

                    <div>
                      <label className="form-label">Auth Mode</label>
                      <select
                        value={formData.auth_mode}
                        onChange={(e) => setFormData(prev => ({ ...prev, auth_mode: e.target.value as 'none' | 'api_key' | 'basic' }))}
                        className="form-input"
                      >
                        <option value="none">None</option>
                        <option value="api_key">API Key</option>
                        <option value="basic">Basic Auth</option>
                      </select>
                    </div>

                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={formData.enabled}
                        onChange={(e) => setFormData(prev => ({ ...prev, enabled: e.target.checked }))}
                        className="form-checkbox h-4 w-4 text-primary-600"
                      />
                      <label className="ml-2 text-sm font-medium text-gray-700">Enabled</label>
                    </div>
                  </div>

                  {/* JSON Fields with Code Editors */}
                  <div className="space-y-6">
                    <div>
                      <label className="form-label">Headers JSON</label>
                      <div className="mt-1">
                        <CodeMirror
                          value={formData.headers_json}
                          onChange={(value) => handleJsonChange('headers_json', value)}
                          extensions={[json()]}
                          theme={oneDark}
                          className="border border-gray-300 rounded-md"
                          basicSetup={{
                            lineNumbers: true,
                            foldGutter: true,
                            bracketMatching: true,
                            indentOnInput: true,
                            highlightSelectionMatches: false,
                          }}
                        />
                        {jsonErrors.headers_json && (
                          <p className="mt-1 text-sm text-red-600">{jsonErrors.headers_json}</p>
                        )}
                      </div>
                    </div>

                    <div>
                      <label className="form-label">Caching Policy JSON</label>
                      <div className="mt-1">
                        <CodeMirror
                          value={formData.caching_policy_json}
                          onChange={(value) => handleJsonChange('caching_policy_json', value)}
                          extensions={[json()]}
                          theme={oneDark}
                          className="border border-gray-300 rounded-md"
                          basicSetup={{
                            lineNumbers: true,
                            foldGutter: true,
                            bracketMatching: true,
                            indentOnInput: true,
                            highlightSelectionMatches: false,
                          }}
                        />
                        {jsonErrors.caching_policy_json && (
                          <p className="mt-1 text-sm text-red-600">{jsonErrors.caching_policy_json}</p>
                        )}
                      </div>
                    </div>

                    <div>
                      <label className="form-label">Rate Limit Policy JSON</label>
                      <div className="mt-1">
                        <CodeMirror
                          value={formData.rate_limit_policy_json}
                          onChange={(value) => handleJsonChange('rate_limit_policy_json', value)}
                          extensions={[json()]}
                          theme={oneDark}
                          className="border border-gray-300 rounded-md"
                          basicSetup={{
                            lineNumbers: true,
                            foldGutter: true,
                            bracketMatching: true,
                            indentOnInput: true,
                            highlightSelectionMatches: false,
                          }}
                        />
                        {jsonErrors.rate_limit_policy_json && (
                          <p className="mt-1 text-sm text-red-600">{jsonErrors.rate_limit_policy_json}</p>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Submit Buttons */}
                  <div className="flex justify-end space-x-3 pt-6 border-t border-gray-200">
                    <button
                      type="button"
                      onClick={() => setIsDrawerOpen(false)}
                      className="btn btn-secondary"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={createMutation.isPending || updateMutation.isPending}
                      className="btn btn-primary"
                    >
                      {createMutation.isPending || updateMutation.isPending ? 'Saving...' : 
                       editingRoute ? 'Update Route' : 'Create Route'}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      )}
    </Layout>
  );
}