import { NextApiRequest, NextApiResponse } from 'next';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    // Mock recent activity for MVP demo
    const recentActivity = [
      {
        id: 1,
        message: 'New API key "demo-client" generated',
        timestamp: '2 minutes ago',
        type: 'success' as const,
      },
      {
        id: 2,
        message: 'Route /api/v1/posts cache hit rate improved to 85%',
        timestamp: '15 minutes ago',
        type: 'success' as const,
      },
      {
        id: 3,
        message: 'Rate limit exceeded for IP 192.168.1.100',
        timestamp: '1 hour ago',
        type: 'warning' as const,
      },
      {
        id: 4,
        message: 'New route /api/secure/ created and enabled',
        timestamp: '2 hours ago',
        type: 'success' as const,
      },
      {
        id: 5,
        message: 'Temporary connection issue with target API',
        timestamp: '3 hours ago',
        type: 'error' as const,
      },
    ];

    res.status(200).json(recentActivity);
  } catch (error) {
    console.error('Dashboard activity error:', error);
    res.status(500).json({ error: 'Failed to fetch recent activity' });
  }
}