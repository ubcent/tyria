import { NextApiRequest, NextApiResponse } from 'next';
import { DB } from '@/lib/db';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  try {
    switch (req.method) {
      case 'GET':
        const routes = await DB.getRoutes();
        res.status(200).json(routes);
        break;

      case 'POST':
        const newRoute = await DB.createRoute(req.body);
        res.status(201).json(newRoute);
        break;

      default:
        res.setHeader('Allow', ['GET', 'POST']);
        res.status(405).end(`Method ${req.method} Not Allowed`);
    }
  } catch (error) {
    console.error('API Error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}