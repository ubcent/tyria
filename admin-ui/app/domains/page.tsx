'use client'

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Layout from '@/components/layouts/Layout';
import { customDomainsApi, type CustomDomain } from '@/lib/api/client';
import { 
  PlusIcon, 
  TrashIcon, 
  XMarkIcon, 
  CheckCircleIcon,
  ClockIcon,
  ExclamationCircleIcon,
  LinkIcon,
  DocumentDuplicateIcon
} from '@heroicons/react/24/outline';

interface DomainFormData {
  hostname: string;
}

export default function DomainsPage() {
  const queryClient = useQueryClient();
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [formData, setFormData] = useState<DomainFormData>({ hostname: '' });
  const [verifyingDomains, setVerifyingDomains] = useState<Set<number>>(new Set());

  const { data: domains, isLoading } = useQuery<CustomDomain[]>({
    queryKey: ['domains'],
    queryFn: customDomainsApi.getAll,
  });

  const createMutation = useMutation({
    mutationFn: customDomainsApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['domains'] });
      setIsDrawerOpen(false);
      setFormData({ hostname: '' });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: customDomainsApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['domains'] });
    },
  });

  const verifyMutation = useMutation({
    mutationFn: customDomainsApi.verify,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['domains'] });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.hostname.trim()) return;

    createMutation.mutate({ hostname: formData.hostname.trim() });
  };

  const handleVerify = (id: number) => {
    setVerifyingDomains(prev => new Set(prev).add(id));
    verifyMutation.mutate(id, {
      onSettled: () => {
        setVerifyingDomains(prev => {
          const next = new Set(prev);
          next.delete(id);
          return next;
        });
      },
    });
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'verified':
        return (
          <div className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
            <CheckCircleIcon className="w-3 h-3 mr-1" />
            Verified
          </div>
        );
      case 'pending':
        return (
          <div className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
            <ClockIcon className="w-3 h-3 mr-1" />
            Pending
          </div>
        );
      case 'failed':
        return (
          <div className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
            <ExclamationCircleIcon className="w-3 h-3 mr-1" />
            Failed
          </div>
        );
      default:
        return (
          <div className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
            {status}
          </div>
        );
    }
  };

  const getVerificationInstructions = (domain: CustomDomain) => {
    return (
      <div className="mt-4 p-4 bg-gray-50 rounded-lg text-sm">
        <h4 className="font-medium mb-2">Verification Instructions:</h4>
        <div className="space-y-2">
          <div>
            <p className="font-medium">Method 1: DNS TXT Record</p>
            <p className="text-gray-600">Add a TXT record at <code>_edgelink.{domain.hostname}</code></p>
            <div className="flex items-center space-x-2 mt-1">
              <code className="bg-gray-200 px-2 py-1 rounded text-xs">
                edgelink-verify={domain.verification_token}
              </code>
              <button
                onClick={() => copyToClipboard(`edgelink-verify=${domain.verification_token}`)}
                className="text-blue-600 hover:text-blue-700"
              >
                <DocumentDuplicateIcon className="w-4 h-4" />
              </button>
            </div>
          </div>
          <div>
            <p className="font-medium">Method 2: HTTP Challenge</p>
            <p className="text-gray-600">
              Serve the verification content at <code>http://{domain.hostname}/.well-known/edge-link.txt</code>
            </p>
            <div className="flex items-center space-x-2 mt-1">
              <a 
                href={`http://${domain.hostname}/.well-known/edge-link.txt`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:text-blue-700 inline-flex items-center"
              >
                <LinkIcon className="w-4 h-4 mr-1" />
                Test URL
              </a>
            </div>
          </div>
        </div>
      </div>
    );
  };

  return (
    <Layout>
      <div className="space-y-6">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h1 className="text-2xl font-semibold text-gray-900">Custom Domains</h1>
            <p className="mt-2 text-sm text-gray-700">
              Add custom domains to your tenant for branded access to your APIs.
            </p>
          </div>
          <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
            <button
              type="button"
              onClick={() => setIsDrawerOpen(true)}
              className="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto"
            >
              <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
              Add Domain
            </button>
          </div>
        </div>

        {/* Domain List */}
        <div className="bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl">
          <div className="px-4 py-6 sm:p-8">
            {isLoading ? (
              <div className="animate-pulse">
                <div className="space-y-3">
                  <div className="h-4 bg-gray-200 rounded w-1/4"></div>
                  <div className="h-4 bg-gray-200 rounded w-1/3"></div>
                  <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                </div>
              </div>
            ) : domains && domains.length > 0 ? (
              <div className="space-y-6">
                {domains.map((domain) => (
                  <div key={domain.id} className="border border-gray-200 rounded-lg p-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <h3 className="text-lg font-medium text-gray-900">{domain.hostname}</h3>
                        <div className="mt-2 flex items-center space-x-4">
                          {getStatusBadge(domain.status)}
                          <span className="text-sm text-gray-500">
                            Created {new Date(domain.created_at).toLocaleDateString()}
                          </span>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        {domain.status !== 'verified' && (
                          <button
                            onClick={() => handleVerify(domain.id)}
                            disabled={verifyingDomains.has(domain.id)}
                            className="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded-md text-indigo-700 bg-indigo-100 hover:bg-indigo-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
                          >
                            {verifyingDomains.has(domain.id) ? 'Verifying...' : 'Verify'}
                          </button>
                        )}
                        <button
                          onClick={() => deleteMutation.mutate(domain.id)}
                          disabled={deleteMutation.isPending}
                          className="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                        >
                          <TrashIcon className="h-3 w-3" />
                        </button>
                      </div>
                    </div>

                    {domain.status !== 'verified' && getVerificationInstructions(domain)}
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-12">
                <LinkIcon className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">No custom domains</h3>
                <p className="mt-1 text-sm text-gray-500">
                  Get started by adding a custom domain for your tenant.
                </p>
                <div className="mt-6">
                  <button
                    type="button"
                    onClick={() => setIsDrawerOpen(true)}
                    className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                  >
                    <PlusIcon className="-ml-0.5 mr-2 h-4 w-4" aria-hidden="true" />
                    Add Domain
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Add Domain Drawer */}
      {isDrawerOpen && (
        <div className="fixed inset-0 z-50 overflow-hidden">
          <div className="absolute inset-0 overflow-hidden">
            <div className="absolute inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
            <section className="absolute inset-y-0 right-0 max-w-full flex pl-10">
              <div className="w-screen max-w-md">
                <form onSubmit={handleSubmit} className="h-full flex flex-col bg-white shadow-xl">
                  <div className="flex-1 overflow-y-auto">
                    <div className="px-4 py-6 bg-gray-50 sm:px-6">
                      <div className="flex items-center justify-between">
                        <h2 className="text-lg font-medium text-gray-900">Add Custom Domain</h2>
                        <button
                          type="button"
                          onClick={() => setIsDrawerOpen(false)}
                          className="rounded-md bg-gray-50 text-gray-400 hover:text-gray-600"
                        >
                          <XMarkIcon className="h-6 w-6" />
                        </button>
                      </div>
                    </div>

                    <div className="px-4 py-6 sm:px-6">
                      <div className="space-y-6">
                        <div>
                          <label htmlFor="hostname" className="block text-sm font-medium text-gray-700">
                            Domain Name
                          </label>
                          <div className="mt-1">
                            <input
                              type="text"
                              id="hostname"
                              name="hostname"
                              placeholder="api.example.com"
                              value={formData.hostname}
                              onChange={(e) => setFormData({ ...formData, hostname: e.target.value })}
                              className="block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                              required
                            />
                          </div>
                          <p className="mt-2 text-sm text-gray-500">
                            Enter the domain name you want to use for your API endpoints.
                          </p>
                        </div>

                        <div className="rounded-md bg-blue-50 p-4">
                          <div className="text-sm text-blue-700">
                            <p className="font-medium">What happens next?</p>
                            <ol className="mt-2 list-decimal list-inside space-y-1">
                              <li>We&apos;ll generate a verification token</li>
                              <li>You&apos;ll need to verify domain ownership via DNS or HTTP</li>
                              <li>Once verified, your domain will be ready to use</li>
                            </ol>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="px-4 py-6 bg-gray-50 sm:px-6">
                    <div className="flex justify-end space-x-3">
                      <button
                        type="button"
                        onClick={() => setIsDrawerOpen(false)}
                        className="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        disabled={createMutation.isPending}
                        className="bg-indigo-600 py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
                      >
                        {createMutation.isPending ? 'Adding...' : 'Add Domain'}
                      </button>
                    </div>
                  </div>
                </form>
              </div>
            </section>
          </div>
        </div>
      )}
    </Layout>
  );
}