import { useState, useEffect } from 'react';
import Layout from '@/components/layouts/Layout';

interface MetricData {
  route_path: string;
  total_requests: number;
  avg_response_time: number;
  success_count: number;
  error_count: number;
  cache_hits: number;
}

export default function MetricsPage() {
  const [metrics, setMetrics] = useState<MetricData[]>([]);
  const [timeframe, setTimeframe] = useState('24h');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchMetrics();
  }, [timeframe]);

  const fetchMetrics = async () => {
    setLoading(true);
    try {
      // For MVP demo, we'll show mock data
      // In production, this would fetch from /api/metrics
      const mockMetrics: MetricData[] = [
        {
          route_path: '/api/v1/posts',
          total_requests: 1250,
          avg_response_time: 45,
          success_count: 1200,
          error_count: 50,
          cache_hits: 850,
        },
        {
          route_path: '/api/secure/',
          total_requests: 680,
          avg_response_time: 120,
          success_count: 650,
          error_count: 30,
          cache_hits: 420,
        },
        {
          route_path: '/public/',
          total_requests: 320,
          avg_response_time: 80,
          success_count: 315,
          error_count: 5,
          cache_hits: 280,
        },
      ];
      
      setMetrics(mockMetrics);
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
    } finally {
      setLoading(false);
    }
  };

  const calculateSuccessRate = (success: number, total: number) => {
    return total > 0 ? ((success / total) * 100).toFixed(1) : '0.0';
  };

  const calculateCacheHitRate = (hits: number, total: number) => {
    return total > 0 ? ((hits / total) * 100).toFixed(1) : '0.0';
  };

  return (
    <Layout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-semibold text-gray-900">Metrics & Analytics</h1>
            <p className="mt-1 text-sm text-gray-500">
              Performance metrics and analytics for your proxy routes
            </p>
          </div>
          <div>
            <select
              value={timeframe}
              onChange={(e) => setTimeframe(e.target.value)}
              className="form-input w-auto"
            >
              <option value="1h">Last 1 Hour</option>
              <option value="24h">Last 24 Hours</option>
              <option value="7d">Last 7 Days</option>
            </select>
          </div>
        </div>

        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <div className="card">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <span className="text-2xl">📊</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Total Requests</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {metrics.reduce((sum, m) => sum + m.total_requests, 0).toLocaleString()}
                </p>
              </div>
            </div>
          </div>

          <div className="card">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <span className="text-2xl">⚡</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Avg Response Time</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {metrics.length > 0 
                    ? Math.round(metrics.reduce((sum, m) => sum + m.avg_response_time, 0) / metrics.length)
                    : 0}ms
                </p>
              </div>
            </div>
          </div>

          <div className="card">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <span className="text-2xl">✅</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Success Rate</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {metrics.length > 0 
                    ? calculateSuccessRate(
                        metrics.reduce((sum, m) => sum + m.success_count, 0),
                        metrics.reduce((sum, m) => sum + m.total_requests, 0)
                      )
                    : '100.0'}%
                </p>
              </div>
            </div>
          </div>

          <div className="card">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <span className="text-2xl">💾</span>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Cache Hit Rate</p>
                <p className="text-2xl font-semibold text-gray-900">
                  {metrics.length > 0 
                    ? calculateCacheHitRate(
                        metrics.reduce((sum, m) => sum + m.cache_hits, 0),
                        metrics.reduce((sum, m) => sum + m.total_requests, 0)
                      )
                    : '0.0'}%
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Route Performance Table */}
        <div className="card p-0 overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Route Performance</h2>
          </div>
          
          {loading ? (
            <div className="p-6">
              <div className="animate-pulse space-y-4">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="h-12 bg-gray-200 rounded"></div>
                ))}
              </div>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="table">
                <thead>
                  <tr>
                    <th className="table-header">Route</th>
                    <th className="table-header">Requests</th>
                    <th className="table-header">Avg Response Time</th>
                    <th className="table-header">Success Rate</th>
                    <th className="table-header">Error Rate</th>
                    <th className="table-header">Cache Hit Rate</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {metrics.map((metric) => (
                    <tr key={metric.route_path}>
                      <td className="table-cell font-medium">
                        <code className="bg-gray-100 px-2 py-1 rounded text-sm">
                          {metric.route_path}
                        </code>
                      </td>
                      <td className="table-cell">
                        {metric.total_requests.toLocaleString()}
                      </td>
                      <td className="table-cell">
                        <span className={`${
                          metric.avg_response_time < 100 ? 'text-green-600' :
                          metric.avg_response_time < 500 ? 'text-yellow-600' : 'text-red-600'
                        }`}>
                          {metric.avg_response_time}ms
                        </span>
                      </td>
                      <td className="table-cell">
                        <span className={`${
                          parseFloat(calculateSuccessRate(metric.success_count, metric.total_requests)) > 95 
                            ? 'text-green-600' : 'text-yellow-600'
                        }`}>
                          {calculateSuccessRate(metric.success_count, metric.total_requests)}%
                        </span>
                      </td>
                      <td className="table-cell">
                        <span className={`${
                          metric.error_count < 10 ? 'text-green-600' :
                          metric.error_count < 50 ? 'text-yellow-600' : 'text-red-600'
                        }`}>
                          {((metric.error_count / metric.total_requests) * 100).toFixed(1)}%
                        </span>
                      </td>
                      <td className="table-cell">
                        <span className={`${
                          parseFloat(calculateCacheHitRate(metric.cache_hits, metric.total_requests)) > 70 
                            ? 'text-green-600' : 'text-yellow-600'
                        }`}>
                          {calculateCacheHitRate(metric.cache_hits, metric.total_requests)}%
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}