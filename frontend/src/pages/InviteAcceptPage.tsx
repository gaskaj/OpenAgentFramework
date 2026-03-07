import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { Bot, CheckCircle2, XCircle, Loader2 } from 'lucide-react';
import { acceptInvitation } from '@/api/orgs';
import { useAuthStore } from '@/store/auth-store';

export function InviteAcceptPage() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const authToken = useAuthStore((s) => s.token);
  const setCurrentOrg = useAuthStore((s) => s.setCurrentOrg);
  const [status, setStatus] = useState<'loading' | 'success' | 'error' | 'login-required'>(
    'loading',
  );
  const [orgName, setOrgName] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    if (!authToken) {
      setStatus('login-required');
      return;
    }

    if (!token) {
      setStatus('error');
      setErrorMsg('Invalid invitation link.');
      return;
    }

    acceptInvitation(token)
      .then((result) => {
        setOrgName(result.org.name);
        setCurrentOrg(result.org);
        setStatus('success');
      })
      .catch((err) => {
        setStatus('error');
        setErrorMsg(
          typeof err === 'object' && err !== null && 'message' in err
            ? (err as { message: string }).message
            : 'Failed to accept invitation.',
        );
      });
  }, [authToken, token, setCurrentOrg]);

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

        <div className="rounded-xl border border-zinc-700 bg-zinc-800 p-8 text-center">
          {status === 'loading' && (
            <div className="space-y-4">
              <Loader2 className="mx-auto h-10 w-10 animate-spin text-blue-500" />
              <p className="text-sm text-zinc-400">Accepting invitation...</p>
            </div>
          )}

          {status === 'success' && (
            <div className="space-y-4">
              <CheckCircle2 className="mx-auto h-10 w-10 text-green-500" />
              <div>
                <h2 className="text-lg font-semibold text-zinc-100">Welcome!</h2>
                <p className="mt-1 text-sm text-zinc-400">
                  You have joined <strong className="text-zinc-200">{orgName}</strong>.
                </p>
              </div>
              <button
                onClick={() => navigate('/dashboard')}
                className="rounded-lg bg-blue-600 px-6 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-700"
              >
                Go to Dashboard
              </button>
            </div>
          )}

          {status === 'error' && (
            <div className="space-y-4">
              <XCircle className="mx-auto h-10 w-10 text-red-500" />
              <div>
                <h2 className="text-lg font-semibold text-zinc-100">
                  Could not accept invitation
                </h2>
                <p className="mt-1 text-sm text-zinc-400">{errorMsg}</p>
              </div>
              <Link
                to="/dashboard"
                className="inline-block rounded-lg border border-zinc-700 px-6 py-2.5 text-sm text-zinc-300 transition-colors hover:bg-zinc-700"
              >
                Go to Dashboard
              </Link>
            </div>
          )}

          {status === 'login-required' && (
            <div className="space-y-4">
              <Bot className="mx-auto h-10 w-10 text-blue-500" />
              <div>
                <h2 className="text-lg font-semibold text-zinc-100">Sign in required</h2>
                <p className="mt-1 text-sm text-zinc-400">
                  Please sign in or create an account to accept this invitation.
                </p>
              </div>
              <div className="flex justify-center gap-3">
                <Link
                  to="/login"
                  className="rounded-lg bg-blue-600 px-6 py-2.5 text-sm font-medium text-white transition-colors hover:bg-blue-700"
                >
                  Sign in
                </Link>
                <Link
                  to="/register"
                  className="rounded-lg border border-zinc-700 px-6 py-2.5 text-sm text-zinc-300 transition-colors hover:bg-zinc-700"
                >
                  Create account
                </Link>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
