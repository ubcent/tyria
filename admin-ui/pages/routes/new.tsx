import { useState } from 'react';
import { useRouter } from 'next/router';
import Layout from '@/components/layouts/Layout';
import RouteForm, { RouteFormData } from '@/components/forms/RouteForm';

export default function NewRoutePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (data: RouteFormData) => {
    setLoading(true);
    try {
      const response = await fetch('/api/routes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      if (response.ok) {
        router.push('/routes');
      } else {
        const error = await response.json();
        alert(`Failed to create route: ${error.error || 'Unknown error'}`);
      }
    } catch (error) {
      console.error('Failed to create route:', error);
      alert('Failed to create route');
    } finally {
      setLoading(false);
    }
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

        <RouteForm onSubmit={handleSubmit} loading={loading} />
      </div>
    </Layout>
  );
}