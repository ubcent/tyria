/** @type {import('next').NextConfig} */
const nextConfig = {
  env: {
    ADMIN_API_URL: process.env.ADMIN_API_URL || 'http://localhost:3001',
  },
}

module.exports = nextConfig