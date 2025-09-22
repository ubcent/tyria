import type { Metadata } from 'next'
import './globals.css'
import { Providers } from './providers'

export const metadata: Metadata = {
  title: 'Edge.link - Proxy-as-a-Service for MACH integrations',
  description: 'High-performance API routing, intelligent caching, rate limiting, and security for modern microservices architectures. Deploy in seconds, scale effortlessly.',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className="font-sans antialiased">
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  )
}