/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    appDir: true,
  },
  env: {
    POSTGRES_URL: process.env.POSTGRES_URL || 'postgresql://localhost:5432/edgelink',
    NEXTAUTH_SECRET: process.env.NEXTAUTH_SECRET || 'your-secret-key',
    NEXTAUTH_URL: process.env.NEXTAUTH_URL || 'http://localhost:3000',
    PROXY_API_URL: process.env.PROXY_API_URL || 'http://localhost:8080',
  },
}

module.exports = nextConfig