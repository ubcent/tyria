/** @type {import('next').NextConfig} */
const nextConfig = {
  env: {
    ADMIN_API_URL: process.env.ADMIN_API_URL || 'http://localhost:3001',
    NEXT_PUBLIC_ADMIN_API_URL: process.env.NEXT_PUBLIC_ADMIN_API_URL || process.env.ADMIN_API_URL || 'http://localhost:3001',
  },
  // Configure to export static files for better CI/CD performance
  output: process.env.NODE_ENV === 'production' ? 'standalone' : undefined,
  
  // Improve build performance
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
  },
}

module.exports = nextConfig