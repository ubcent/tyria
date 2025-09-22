import { NextApiRequest, NextApiResponse } from 'next';
import { DB } from '@/lib/db';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  const { id } = req.query;
  const keyId = parseInt(id as string);

  if (isNaN(keyId)) {
    return res.status(400).json({ error: 'Invalid key ID' });
  }

  try {
    switch (req.method) {
      case 'GET':
        const key = await DB.getApiKey(keyId);
        if (!key) {
          return res.status(404).json({ error: 'API key not found' });
        }
        res.status(200).json(key);
        break;

      case 'PUT':
        const updatedKey = await DB.updateApiKey(keyId, req.body);
        if (!updatedKey) {
          return res.status(404).json({ error: 'API key not found' });
        }
        res.status(200).json(updatedKey);
        break;

      case 'DELETE':
        const deleted = await DB.deleteApiKey(keyId);
        if (!deleted) {
          return res.status(404).json({ error: 'API key not found' });
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