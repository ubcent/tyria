'use client'

import { useQuery } from '@tanstack/react-query';
import Layout from '@/components/layouts/Layout';
import { dashboardApi, type DashboardStats, type Activity } from '@/lib/api/client';
import { useRouter } from 'next/navigation';

export default function Dashboard() {
  const router = useRouter();

  const { data: stats, isLoading: statsLoading } = useQuery<DashboardStats>({
    queryKey: ['dashboard', 'stats'],
    queryFn: dashboardApi.getStats,
  });

  const { data: activity, isLoading: activityLoading } = useQuery<Activity[]>({
    queryKey: ['dashboard', 'activity'],
    queryFn: dashboardApi.getActivity,
  });

  if (statsLoading || activityLoading) {
    return (
      <Layout>
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-6"></div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="card">
                <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                <div className="h-8 bg-gray-200 rounded w-1/2"></div>
              </div>
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
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-sm text-gray-500">
            Overview of your Edge.link proxy service
          </p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <StatCard
            title="Total Requests"
            value={stats?.total_requests.toLocaleString() || '0'}
            icon="📊"
            trend="+12%"
            trendColor="text-green-600"
          />
          <StatCard
            title="Avg Response Time"
            value={`${stats?.avg_response_time || 0}ms`}
            icon="⚡"
            trend="-5ms"
            trendColor="text-green-600"
          />
          <StatCard
            title="Success Rate"
            value={`${stats?.success_rate.toFixed(1) || '0.0'}%`}
            icon="✅"
            trend="+0.2%"
            trendColor="text-green-600"
          />
          <StatCard
            title="Cache Hit Rate"
            value={`${stats?.cache_hit_rate.toFixed(1) || '0.0'}%`}
            icon="💾"
            trend="+3%"
            trendColor="text-green-600"
          />
          <StatCard
            title="Active Routes"
            value={stats?.active_routes.toString() || '0'}
            icon="🛣️"
            trend="No change"
            trendColor="text-gray-500"
          />
          <StatCard
            title="Active API Keys"
            value={stats?.active_api_keys.toString() || '0'}
            icon="🔑"
            trend="No change"
            trendColor="text-gray-500"
          />
        </div>

        {/* Recent Activity */}
        <div className="card">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Recent Activity</h2>
          {activity && activity.length > 0 ? (
            <div className="space-y-3">
              {activity.map((item) => (
                <div key={item.id} className="flex items-center space-x-3">
                  <div className={`w-2 h-2 rounded-full ${
                    item.type === 'success' ? 'bg-green-400' :
                    item.type === 'warning' ? 'bg-yellow-400' : 'bg-red-400'
                  }`} />
                  <div className="flex-1">
                    <p className="text-sm text-gray-900">{item.message}</p>
                    <p className="text-xs text-gray-500">{item.timestamp}</p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-500">No recent activity</p>
          )}
        </div>

        {/* Quick Actions */}
        <div className="card">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Quick Actions</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <QuickActionButton
              href="/routes/new"
              icon="➕"
              title="Add Route"
              description="Create new proxy route"
              onClick={() => router.push('/routes/new')}
            />
            <QuickActionButton
              href="/keys/new"
              icon="🔐"
              title="Generate API Key"
              description="Create new API key"
              onClick={() => router.push('/keys/new')}
            />
            <QuickActionButton
              href="/metrics"
              icon="📈"
              title="View Metrics"
              description="Detailed analytics"
              onClick={() => router.push('/metrics')}
            />
            <QuickActionButton
              href="/routes"
              icon="📋"
              title="Manage Routes"
              description="Edit proxy routes"
              onClick={() => router.push('/routes')}
            />
          </div>
        </div>
      </div>
    </Layout>
  );
}

interface StatCardProps {
  title: string;
  value: string;
  icon: string;
  trend: string;
  trendColor: string;
}

function StatCard({ title, value, icon, trend, trendColor }: StatCardProps) {
  return (
    <div className="card">
      <div className="flex items-center">
        <div className="flex-shrink-0">
          <span className="text-2xl">{icon}</span>
        </div>
        <div className="ml-4">
          <p className="text-sm font-medium text-gray-500">{title}</p>
          <p className="text-2xl font-semibold text-gray-900">{value}</p>
          <p className={`text-sm ${trendColor}`}>{trend}</p>
        </div>
      </div>
    </div>
  );
}

interface QuickActionButtonProps {
  href: string;
  icon: string;
  title: string;
  description: string;
  onClick: () => void;
}

function QuickActionButton({ icon, title, description, onClick }: QuickActionButtonProps) {
  return (
    <button
      onClick={onClick}
      className="text-left p-4 rounded-lg border border-gray-200 hover:border-primary-300 hover:bg-primary-50 transition-colors"
    >
      <div className="text-xl mb-2">{icon}</div>
      <h3 className="font-medium text-gray-900">{title}</h3>
      <p className="text-sm text-gray-500">{description}</p>
    </button>
  );
}