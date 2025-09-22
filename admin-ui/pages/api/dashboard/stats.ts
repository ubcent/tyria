import { NextApiRequest, NextApiResponse } from 'next';
import { DB } from '@/lib/db';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    // Get stats from various sources
    const [routes, apiKeys, metrics] = await Promise.all([
      DB.getRoutes(),
      DB.getApiKeys(),
      DB.getMetrics('24h'),
    ]);

    const activeRoutes = routes.filter(r => r.enabled).length;
    const activeApiKeys = apiKeys.filter(k => k.enabled).length;

    // Calculate aggregated metrics
    const totalRequests = metrics.reduce((sum, m) => sum + parseInt(m.total_requests || '0'), 0);
    const avgResponseTime = metrics.length > 0 
      ? metrics.reduce((sum, m) => sum + parseFloat(m.avg_response_time || '0'), 0) / metrics.length 
      : 0;
    
    const successCount = metrics.reduce((sum, m) => sum + parseInt(m.success_count || '0'), 0);
    const errorCount = metrics.reduce((sum, m) => sum + parseInt(m.error_count || '0'), 0);
    const successRate = totalRequests > 0 ? (successCount / totalRequests) * 100 : 100;
    
    const cacheHits = metrics.reduce((sum, m) => sum + parseInt(m.cache_hits || '0'), 0);
    const cacheHitRate = totalRequests > 0 ? (cacheHits / totalRequests) * 100 : 0;

    const stats = {
      totalRequests,
      avgResponseTime: Math.round(avgResponseTime),
      successRate,
      cacheHitRate,
      activeRoutes,
      activeApiKeys,
    };

    res.status(200).json(stats);
  } catch (error) {
    console.error('Dashboard stats error:', error);
    res.status(500).json({ error: 'Failed to fetch dashboard stats' });
  }
}