import { Navigate } from 'react-router-dom';
import { Bot } from 'lucide-react';
import { LoginForm } from '@/components/auth/LoginForm';
import { useAuthStore } from '@/store/auth-store';

export function LoginPage() {
  const token = useAuthStore((s) => s.token);

  if (token) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-900 px-4">
      <div className="w-full max-w-md">
        <div className="mb-8 flex justify-center">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-600">
              <Bot className="h-7 w-7 text-white" />
            </div>
            <span className="text-xl font-bold text-zinc-100">OpenAgent</span>
          </div>
        </div>
        <div className="rounded-xl border border-zinc-700 bg-zinc-800 p-8">
          <LoginForm />
        </div>
      </div>
    </div>
  );
}
