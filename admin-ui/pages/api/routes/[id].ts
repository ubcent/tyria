import { NextApiRequest, NextApiResponse } from 'next';
import { DB } from '@/lib/db';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  const { id } = req.query;
  const routeId = parseInt(id as string);

  if (isNaN(routeId)) {
    return res.status(400).json({ error: 'Invalid route ID' });
  }

  try {
    switch (req.method) {
      case 'GET':
        const route = await DB.getRoute(routeId);
        if (!route) {
          return res.status(404).json({ error: 'Route not found' });
        }
        res.status(200).json(route);
        break;

      case 'PUT':
        const updatedRoute = await DB.updateRoute(routeId, req.body);
        if (!updatedRoute) {
          return res.status(404).json({ error: 'Route not found' });
        }
        res.status(200).json(updatedRoute);
        break;

      case 'DELETE':
        const deleted = await DB.deleteRoute(routeId);
        if (!deleted) {
          return res.status(404).json({ error: 'Route not found' });
        }
        res.status(204).end();
        break;

      default:
        res.setHeader('Allow', ['GET', 'PUT', 'DELETE']);
        res.status(405).end(`Method ${req.method} Not Allowed`);
    }
  } catch (error) {
    console.error('API Error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}