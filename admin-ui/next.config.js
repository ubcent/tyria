/** @type {import('next').NextConfig} */
const nextConfig = {
  env: {
    ADMIN_API_URL: process.env.ADMIN_API_URL || 'http://localhost:3001',
  },
  // Disable Next.js telemetry for restricted network environments
  telemetry: false,
}

module.exports = nextConfig