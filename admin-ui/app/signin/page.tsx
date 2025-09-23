import { Suspense } from 'react'
import SigninForm from './SigninForm'

export default function SigninPage() {
  return (
    <Suspense fallback={null}>
      <SigninForm />
    </Suspense>
  )
}