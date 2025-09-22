import { useState, useEffect } from 'react';
import Layout from '@/components/layouts/Layout';
import { useRouter } from 'next/router';

interface DashboardStats {
  totalRequests: number;
  avgResponseTime: number;
  successRate: number;
  cacheHitRate: number;
  activeRoutes: number;
  activeApiKeys: number;
}

interface RecentActivity {
  id: number;
  message: string;
  timestamp: string;
  type: 'success' | 'warning' | 'error';
}

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats>({
    totalRequests: 0,
    avgResponseTime: 0,
    successRate: 0,
    cacheHitRate: 0,
    activeRoutes: 0,
    activeApiKeys: 0,
  });
  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      const [statsRes, activityRes] = await Promise.all([
        fetch('/api/dashboard/stats'),
        fetch('/api/dashboard/activity'),
      ]);

      if (statsRes.ok) {
        const statsData = await statsRes.json();
        setStats(statsData);
      }

      if (activityRes.ok) {
        const activityData = await activityRes.json();
        setRecentActivity(activityData);
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
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
            value={stats.totalRequests.toLocaleString()}
            icon="📊"
            trend="+12%"
            trendColor="text-green-600"
          />
          <StatCard
            title="Avg Response Time"
            value={`${stats.avgResponseTime}ms`}
            icon="⚡"
            trend="-5ms"
            trendColor="text-green-600"
          />
          <StatCard
            title="Success Rate"
            value={`${stats.successRate.toFixed(1)}%`}
            icon="✅"
            trend="+0.2%"
            trendColor="text-green-600"
          />
          <StatCard
            title="Cache Hit Rate"
            value={`${stats.cacheHitRate.toFixed(1)}%`}
            icon="💾"
            trend="+3%"
            trendColor="text-green-600"
          />
          <StatCard
            title="Active Routes"
            value={stats.activeRoutes.toString()}
            icon="🛣️"
            trend="No change"
            trendColor="text-gray-500"
          />
          <StatCard
            title="Active API Keys"
            value={stats.activeApiKeys.toString()}
            icon="🔑"
            trend="No change"
            trendColor="text-gray-500"
          />
        </div>

        {/* Recent Activity */}
        <div className="card">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Recent Activity</h2>
          {recentActivity.length > 0 ? (
            <div className="space-y-3">
              {recentActivity.map((activity) => (
                <div key={activity.id} className="flex items-center space-x-3">
                  <div className={`w-2 h-2 rounded-full ${
                    activity.type === 'success' ? 'bg-green-400' :
                    activity.type === 'warning' ? 'bg-yellow-400' : 'bg-red-400'
                  }`} />
                  <div className="flex-1">
                    <p className="text-sm text-gray-900">{activity.message}</p>
                    <p className="text-xs text-gray-500">{activity.timestamp}</p>
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
            />
            <QuickActionButton
              href="/keys/new"
              icon="🔐"
              title="Generate API Key"
              description="Create new API key"
            />
            <QuickActionButton
              href="/metrics"
              icon="📈"
              title="View Metrics"
              description="Detailed analytics"
            />
            <QuickActionButton
              href="/logs"
              icon="📋"
              title="View Logs"
              description="Recent request logs"
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
}

function QuickActionButton({ href, icon, title, description }: QuickActionButtonProps) {
  const router = useRouter();

  return (
    <button
      onClick={() => router.push(href)}
      className="text-left p-4 rounded-lg border border-gray-200 hover:border-primary-300 hover:bg-primary-50 transition-colors"
    >
      <div className="text-xl mb-2">{icon}</div>
      <h3 className="font-medium text-gray-900">{title}</h3>
      <p className="text-sm text-gray-500">{description}</p>
    </button>
  );
}