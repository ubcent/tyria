'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layouts/Layout';
import RouteForm, { type RouteFormData } from '@/components/forms/RouteForm';
import { routesApi } from '@/lib/api/client';

export default function NewRoutePage() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const createMutation = useMutation({
    mutationFn: routesApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['routes'] });
      router.push('/routes');
    },
    onError: (error) => {
      console.error('Failed to create route:', error);
      alert('Failed to create route');
    },
  });

  const handleSubmit = async (data: RouteFormData) => {
    createMutation.mutate(data);
  };

  return (
    <Layout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Create New Route</h1>
          <p className="mt-1 text-sm text-gray-500">
            Configure a new proxy route for your API
          </p>
        </div>

        <RouteForm onSubmit={handleSubmit} loading={createMutation.isPending} />
      </div>
    </Layout>
  );
}