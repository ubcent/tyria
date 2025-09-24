import type { Metadata } from 'next'
import { cookies } from 'next/headers'
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
  // Get theme from cookies for SSR
  const cookieStore = cookies()
  const theme = cookieStore.get('theme')?.value || 'light'
  
  return (
    <html lang="en" className={theme}>
      <body className="font-sans antialiased">
        <script
          dangerouslySetInnerHTML={{
            __html: `
              // Prevent flash of unstyled content by setting theme immediately
              (function() {
                const theme = localStorage.getItem('theme') || 
                             document.cookie.split('; ').find(row => row.startsWith('theme='))?.split('=')[1] ||
                             (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
                document.documentElement.className = theme;
              })();
            `,
          }}
        />
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  )
}