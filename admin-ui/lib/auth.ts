'use server'

import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:3001'

export interface User {
  id: number
  tenant_id: number
  email: string
  role: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  message: string
  user?: User
  token: string
}

export async function signUp(formData: FormData) {
  const email = formData.get('email') as string
  const password = formData.get('password') as string
  const companyName = formData.get('companyName') as string

  if (!email || !password || !companyName) {
    return { error: 'All fields are required' }
  }

  try {
    const response = await fetch(`${API_BASE_URL}/api/auth/signup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email,
        password,
        company_name: companyName,
      }),
    })

    const data = await response.json()

    if (!response.ok) {
      return { error: data.error || 'Signup failed' }
    }

    // Set the auth cookie
    cookies().set('auth_token', data.token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict',
      maxAge: 24 * 60 * 60, // 24 hours
      path: '/',
    })

    return { success: true, user: data.user }
  } catch (error) {
    console.error('Signup error:', error)
    return { error: 'Network error occurred' }
  }
}

export async function signIn(formData: FormData) {
  const email = formData.get('email') as string
  const password = formData.get('password') as string

  if (!email || !password) {
    return { error: 'Email and password are required' }
  }

  try {
    const response = await fetch(`${API_BASE_URL}/api/auth/signin`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email,
        password,
      }),
    })

    const data = await response.json()

    if (!response.ok) {
      return { error: data.error || 'Invalid credentials' }
    }

    // Set the auth cookie
    cookies().set('auth_token', data.token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict',
      maxAge: 24 * 60 * 60, // 24 hours
      path: '/',
    })

    return { success: true, user: data.user }
  } catch (error) {
    console.error('Signin error:', error)
    return { error: 'Network error occurred' }
  }
}

export async function signOut() {
  try {
    const token = cookies().get('auth_token')?.value

    if (token) {
      await fetch(`${API_BASE_URL}/api/auth/signout`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
      })
    }

    // Clear the auth cookie
    cookies().delete('auth_token')
    
    return { success: true }
  } catch (error) {
    console.error('Signout error:', error)
    // Still clear the cookie even if API call fails
    cookies().delete('auth_token')
    return { success: true }
  }
}

export async function getCurrentUser(): Promise<User | null> {
  try {
    const token = cookies().get('auth_token')?.value

    if (!token) {
      return null
    }

    const response = await fetch(`${API_BASE_URL}/api/auth/profile`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    })

    if (!response.ok) {
      return null
    }

    return await response.json()
  } catch (error) {
    console.error('Get current user error:', error)
    return null
  }
}

export async function requireAuth() {
  const user = await getCurrentUser()
  if (!user) {
    redirect('/signin')
  }
  return user
}

export async function redirectIfAuthenticated() {
  const user = await getCurrentUser()
  if (user) {
    redirect('/dashboard')
  }
}