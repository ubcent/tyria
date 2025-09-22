import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { ProxyRoute } from '@/lib/db';

interface RouteFormProps {
  route?: ProxyRoute;
  onSubmit: (data: RouteFormData) => void;
  loading?: boolean;
}

export interface RouteFormData {
  path: string;
  target: string;
  methods: string[];
  cache_enabled: boolean;
  cache_ttl: number;
  rate_limit_enabled: boolean;
  rate_limit_rate: number;
  rate_limit_burst: number;
  rate_limit_period: number;
  rate_limit_per_client: boolean;
  auth_required: boolean;
  auth_keys: string[];
  validation_enabled: boolean;
  validation_request_schema?: string;
  validation_response_schema?: string;
  enabled: boolean;
}

const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'];

export default function RouteForm({ route, onSubmit, loading = false }: RouteFormProps) {
  const { register, handleSubmit, watch, setValue, formState: { errors } } = useForm<RouteFormData>({
    defaultValues: {
      path: route?.path || '',
      target: route?.target || '',
      methods: route?.methods || ['GET'],
      cache_enabled: route?.cache_enabled || false,
      cache_ttl: route?.cache_ttl || 300,
      rate_limit_enabled: route?.rate_limit_enabled || false,
      rate_limit_rate: route?.rate_limit_rate || 100,
      rate_limit_burst: route?.rate_limit_burst || 10,
      rate_limit_period: route?.rate_limit_period || 60,
      rate_limit_per_client: route?.rate_limit_per_client || true,
      auth_required: route?.auth_required || false,
      auth_keys: route?.auth_keys || [],
      validation_enabled: route?.validation_enabled || false,
      validation_request_schema: route?.validation_request_schema || '',
      validation_response_schema: route?.validation_response_schema || '',
      enabled: route?.enabled !== undefined ? route.enabled : true,
    },
  });

  const watchCacheEnabled = watch('cache_enabled');
  const watchRateLimitEnabled = watch('rate_limit_enabled');
  const watchAuthRequired = watch('auth_required');
  const watchValidationEnabled = watch('validation_enabled');

  const handleMethodChange = (method: string, checked: boolean) => {
    const currentMethods = watch('methods');
    if (checked) {
      setValue('methods', [...currentMethods, method]);
    } else {
      setValue('methods', currentMethods.filter(m => m !== method));
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Basic Configuration */}
      <div className="card">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Basic Configuration</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="form-label">Path *</label>
            <input
              type="text"
              className="form-input"
              placeholder="/api/v1/"
              {...register('path', { required: 'Path is required' })}
            />
            {errors.path && (
              <p className="mt-1 text-sm text-red-600">{errors.path.message}</p>
            )}
          </div>
          <div>
            <label className="form-label">Target URL *</label>
            <input
              type="url"
              className="form-input"
              placeholder="https://api.example.com"
              {...register('target', { required: 'Target URL is required' })}
            />
            {errors.target && (
              <p className="mt-1 text-sm text-red-600">{errors.target.message}</p>
            )}
          </div>
        </div>

        <div>
          <label className="form-label">HTTP Methods</label>
          <div className="mt-2 grid grid-cols-2 md:grid-cols-4 gap-3">
            {HTTP_METHODS.map((method) => (
              <label key={method} className="flex items-center">
                <input
                  type="checkbox"
                  className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                  checked={watch('methods').includes(method)}
                  onChange={(e) => handleMethodChange(method, e.target.checked)}
                />
                <span className="ml-2 text-sm text-gray-700">{method}</span>
              </label>
            ))}
          </div>
        </div>

        <div>
          <label className="flex items-center">
            <input
              type="checkbox"
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              {...register('enabled')}
            />
            <span className="ml-2 text-sm text-gray-700">Enable this route</span>
          </label>
        </div>
      </div>

      {/* Cache Configuration */}
      <div className="card">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Cache Configuration</h3>
        <div>
          <label className="flex items-center mb-4">
            <input
              type="checkbox"
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              {...register('cache_enabled')}
            />
            <span className="ml-2 text-sm text-gray-700">Enable caching</span>
          </label>

          {watchCacheEnabled && (
            <div>
              <label className="form-label">Cache TTL (seconds)</label>
              <input
                type="number"
                min="1"
                className="form-input"
                {...register('cache_ttl', { 
                  min: { value: 1, message: 'TTL must be at least 1 second' }
                })}
              />
              {errors.cache_ttl && (
                <p className="mt-1 text-sm text-red-600">{errors.cache_ttl.message}</p>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Rate Limiting */}
      <div className="card">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Rate Limiting</h3>
        <div>
          <label className="flex items-center mb-4">
            <input
              type="checkbox"
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              {...register('rate_limit_enabled')}
            />
            <span className="ml-2 text-sm text-gray-700">Enable rate limiting</span>
          </label>

          {watchRateLimitEnabled && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="form-label">Rate (requests per period)</label>
                <input
                  type="number"
                  min="1"
                  className="form-input"
                  {...register('rate_limit_rate', { 
                    min: { value: 1, message: 'Rate must be at least 1' }
                  })}
                />
              </div>
              <div>
                <label className="form-label">Burst</label>
                <input
                  type="number"
                  min="1"
                  className="form-input"
                  {...register('rate_limit_burst', { 
                    min: { value: 1, message: 'Burst must be at least 1' }
                  })}
                />
              </div>
              <div>
                <label className="form-label">Period (seconds)</label>
                <input
                  type="number"
                  min="1"
                  className="form-input"
                  {...register('rate_limit_period', { 
                    min: { value: 1, message: 'Period must be at least 1 second' }
                  })}
                />
              </div>
              <div className="md:col-span-3">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                    {...register('rate_limit_per_client')}
                  />
                  <span className="ml-2 text-sm text-gray-700">Per-client rate limiting</span>
                </label>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Authentication */}
      <div className="card">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Authentication</h3>
        <div>
          <label className="flex items-center mb-4">
            <input
              type="checkbox"
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              {...register('auth_required')}
            />
            <span className="ml-2 text-sm text-gray-700">Require authentication</span>
          </label>

          {watchAuthRequired && (
            <div>
              <label className="form-label">Allowed API Keys (comma-separated)</label>
              <input
                type="text"
                className="form-input"
                placeholder="key1, key2, key3"
                value={watch('auth_keys').join(', ')}
                onChange={(e) => {
                  const keys = e.target.value.split(',').map(k => k.trim()).filter(k => k);
                  setValue('auth_keys', keys);
                }}
              />
              <p className="mt-1 text-sm text-gray-500">
                Leave empty to allow all valid API keys
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Validation */}
      <div className="card">
        <h3 className="text-lg font-medium text-gray-900 mb-4">JSON Schema Validation</h3>
        <div>
          <label className="flex items-center mb-4">
            <input
              type="checkbox"
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              {...register('validation_enabled')}
            />
            <span className="ml-2 text-sm text-gray-700">Enable JSON validation</span>
          </label>

          {watchValidationEnabled && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="form-label">Request Schema Name</label>
                <input
                  type="text"
                  className="form-input"
                  placeholder="user_create"
                  {...register('validation_request_schema')}
                />
              </div>
              <div>
                <label className="form-label">Response Schema Name</label>
                <input
                  type="text"
                  className="form-input"
                  placeholder="user_response"
                  {...register('validation_response_schema')}
                />
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Submit */}
      <div className="flex justify-end space-x-3">
        <button
          type="button"
          className="btn btn-secondary"
          onClick={() => window.history.back()}
        >
          Cancel
        </button>
        <button
          type="submit"
          className="btn btn-primary"
          disabled={loading}
        >
          {loading ? 'Saving...' : (route ? 'Update Route' : 'Create Route')}
        </button>
      </div>
    </form>
  );
}