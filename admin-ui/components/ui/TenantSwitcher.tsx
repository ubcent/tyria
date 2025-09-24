'use client'

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ChevronDownIcon, CheckIcon } from '@heroicons/react/24/outline';
import { tenantsApi, authApi, type Tenant } from '@/lib/api/client';

export default function TenantSwitcher() {
  const [isOpen, setIsOpen] = useState(false);

  // Get current user context (includes current tenant)
  const { data: userContext } = useQuery({
    queryKey: ['auth', 'profile'],
    queryFn: authApi.getProfile,
  });

  // Get all available tenants (in a real multi-tenant system, this would be tenants the user has access to)
  const { data: tenants } = useQuery({
    queryKey: ['tenants'],
    queryFn: tenantsApi.getAll,
    enabled: !!userContext, // Only fetch if we have user context
  });

  const currentTenant = userContext?.tenant;

  const handleTenantSwitch = (tenant: Tenant) => {
    // In a real implementation, this would:
    // 1. Make an API call to switch tenant context
    // 2. Update the auth token/session
    // 3. Refresh the page or update global state
    console.log('Switching to tenant:', tenant.name);
    setIsOpen(false);
    // For now, just log and close
  };

  if (!currentTenant || !tenants?.length) {
    return null; // Don't show if no tenant data or single tenant
  }

  return (
    <div className="relative">
      <button
        type="button"
        className="flex items-center space-x-2 px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:text-gray-900 dark:hover:text-gray-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
        onClick={() => setIsOpen(!isOpen)}
      >
        <div className="flex-shrink-0 w-6 h-6 bg-primary-600 dark:bg-primary-500 rounded-full flex items-center justify-center">
          <span className="text-xs font-semibold text-white">
            {currentTenant.name.charAt(0).toUpperCase()}
          </span>
        </div>
        <div className="min-w-0 flex-1 text-left">
          <p className="truncate">{currentTenant.name}</p>
          <p className="text-xs text-gray-500 dark:text-gray-400 capitalize">
            {currentTenant.plan} plan
          </p>
        </div>
        <ChevronDownIcon className="h-4 w-4 flex-shrink-0" />
      </button>

      {isOpen && (
        <>
          {/* Overlay */}
          <div 
            className="fixed inset-0 z-10" 
            onClick={() => setIsOpen(false)}
          />
          
          {/* Dropdown menu */}
          <div className="absolute right-0 z-20 mt-2 w-64 bg-white dark:bg-gray-800 rounded-md shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
            <div className="py-1">
              <div className="px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                Switch Organization
              </div>
              {tenants.map((tenant) => (
                <button
                  key={tenant.id}
                  onClick={() => handleTenantSwitch(tenant)}
                  className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center justify-between ${
                    tenant.id === currentTenant.id 
                      ? 'bg-gray-50 dark:bg-gray-700 text-gray-900 dark:text-gray-100' 
                      : 'text-gray-700 dark:text-gray-300'
                  }`}
                >
                  <div className="flex items-center space-x-3">
                    <div className="flex-shrink-0 w-6 h-6 bg-primary-600 dark:bg-primary-500 rounded-full flex items-center justify-center">
                      <span className="text-xs font-semibold text-white">
                        {tenant.name.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-medium">{tenant.name}</p>
                      <p className="text-xs text-gray-500 dark:text-gray-400 capitalize">
                        {tenant.plan} plan • {tenant.status}
                      </p>
                    </div>
                  </div>
                  {tenant.id === currentTenant.id && (
                    <CheckIcon className="h-4 w-4 text-primary-600 dark:text-primary-400" />
                  )}
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}