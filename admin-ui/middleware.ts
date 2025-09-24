import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:3001'

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Allow access to auth pages
  if (pathname.startsWith('/signin') || pathname.startsWith('/signup')) {
    return NextResponse.next()
  }

  // Allow access to public pages and demo pages
  if (pathname === '/' || pathname.startsWith('/public') || pathname.startsWith('/_next') || pathname.startsWith('/demo') || pathname.startsWith('/app-demo')) {
    return NextResponse.next()
  }

  // Check for auth token in protected routes (/app/* and /dashboard/*)
  if (pathname.startsWith('/app') || pathname.startsWith('/dashboard')) {
    const token = request.cookies.get('auth_token')?.value

    if (!token) {
      const url = request.nextUrl.clone()
      url.pathname = '/signin'
      url.searchParams.set('redirect', pathname)
      return NextResponse.redirect(url)
    }

    // Validate token with backend
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/profile`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        // Invalid token, redirect to signin
        const url = request.nextUrl.clone()
        url.pathname = '/signin'
        url.searchParams.set('redirect', pathname)
        const redirectResponse = NextResponse.redirect(url)
        
        // Clear invalid token
        redirectResponse.cookies.delete('auth_token')
        return redirectResponse
      }

      // Token is valid, continue
      return NextResponse.next()
    } catch (error) {
      // Network error or other issues, redirect to signin
      const url = request.nextUrl.clone()
      url.pathname = '/signin'
      url.searchParams.set('redirect', pathname)
      return NextResponse.redirect(url)
    }
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
}