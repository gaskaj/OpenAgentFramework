import { useState, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Loader2, UserPlus, X, Shield, UserMinus, Globe, ExternalLink, Copy, Check } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import * as orgsApi from '@/api/orgs';
import * as tunnelApi from '@/api/tunnel';
import type { OrgMember, Invitation, OrgRole } from '@/types';
import { StatusBadge } from '@/components/common/StatusBadge';
import { TimeAgo } from '@/components/common/TimeAgo';

const orgSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
});

const inviteSchema = z.object({
  email: z.string().email('Invalid email'),
  role: z.enum(['admin', 'member', 'viewer']),
});

type OrgFormData = z.infer<typeof orgSchema>;
type InviteFormData = z.infer<typeof inviteSchema>;

export function OrgSettingsPage() {
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const setCurrentOrg = useAuthStore((s) => s.setCurrentOrg);
  const [members, setMembers] = useState<OrgMember[]>([]);
  const [invitations, setInvitations] = useState<Invitation[]>([]);
  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [tunnelStatus, setTunnelStatus] = useState<tunnelApi.TunnelStatus>({ enabled: false, has_auth_token: false });
  const [tunnelLoading, setTunnelLoading] = useState(false);
  const [copied, setCopied] = useState(false);
  const [authToken, setAuthToken] = useState('');
  const [tokenSaving, setTokenSaving] = useState(false);
  const [tokenSaved, setTokenSaved] = useState(false);

  const orgForm = useForm<OrgFormData>({
    resolver: zodResolver(orgSchema),
    defaultValues: { name: currentOrg?.name ?? '' },
  });

  const inviteForm = useForm<InviteFormData>({
    resolver: zodResolver(inviteSchema),
    defaultValues: { role: 'member' },
  });

  const loadData = useCallback(async () => {
    if (!currentOrg) return;
    const [m, i] = await Promise.all([
      orgsApi.listMembers(currentOrg.slug),
      orgsApi.listInvitations(currentOrg.slug),
    ]);
    setMembers(m);
    setInvitations(i);
  }, [currentOrg]);

  const loadTunnelStatus = useCallback(async () => {
    try {
      const status = await tunnelApi.getTunnelStatus();
      setTunnelStatus(status);
    } catch {
      // tunnel endpoint may not exist on older backends
    }
  }, []);

  useEffect(() => {
    loadData();
    loadTunnelStatus();
  }, [loadData, loadTunnelStatus]);

  // Poll tunnel status while loading (URL takes a few seconds to appear)
  useEffect(() => {
    if (!tunnelStatus.enabled || tunnelStatus.public_url) return;
    const interval = setInterval(loadTunnelStatus, 2000);
    return () => clearInterval(interval);
  }, [tunnelStatus.enabled, tunnelStatus.public_url, loadTunnelStatus]);

  const handleSaveOrg = async (data: OrgFormData) => {
    if (!currentOrg) return;
    setSaving(true);
    try {
      const updated = await orgsApi.updateOrg(currentOrg.slug, data);
      setCurrentOrg(updated);
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } finally {
      setSaving(false);
    }
  };

  const handleInvite = async (data: InviteFormData) => {
    if (!currentOrg) return;
    await orgsApi.createInvitation(currentOrg.slug, data.email, data.role as OrgRole);
    inviteForm.reset();
    loadData();
  };

  const handleRoleChange = async (userId: string, role: OrgRole) => {
    if (!currentOrg) return;
    await orgsApi.updateMemberRole(currentOrg.slug, userId, role);
    loadData();
  };

  const handleRemoveMember = async (userId: string) => {
    if (!currentOrg) return;
    if (!window.confirm('Remove this member?')) return;
    await orgsApi.removeMember(currentOrg.slug, userId);
    loadData();
  };

  const handleCancelInvite = async (invId: string) => {
    if (!currentOrg) return;
    await orgsApi.cancelInvitation(currentOrg.slug, invId);
    loadData();
  };

  const handleTunnelToggle = async () => {
    setTunnelLoading(true);
    try {
      const status = await tunnelApi.toggleTunnel(!tunnelStatus.enabled);
      setTunnelStatus(status);
    } catch {
      // ignore
    } finally {
      setTunnelLoading(false);
    }
  };

  const handleCopyUrl = () => {
    if (tunnelStatus.public_url) {
      navigator.clipboard.writeText(tunnelStatus.public_url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleSaveToken = async (e: React.FormEvent) => {
    e.preventDefault();
    setTokenSaving(true);
    try {
      const status = await tunnelApi.saveAuthToken(authToken);
      setTunnelStatus(status);
      setAuthToken('');
      setTokenSaved(true);
      setTimeout(() => setTokenSaved(false), 3000);
    } catch {
      // ignore
    } finally {
      setTokenSaving(false);
    }
  };

  const handleClearToken = async () => {
    setTokenSaving(true);
    try {
      const status = await tunnelApi.saveAuthToken('');
      setTunnelStatus(status);
      setTokenSaved(true);
      setTimeout(() => setTokenSaved(false), 3000);
    } catch {
      // ignore
    } finally {
      setTokenSaving(false);
    }
  };

  return (
    <div className="mx-auto max-w-3xl space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-zinc-100">Organization Settings</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Manage your organization profile and team
        </p>
      </div>

      {/* Org name editor */}
      <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
        <h2 className="mb-4 text-sm font-medium text-zinc-300">Organization Details</h2>
        <form onSubmit={orgForm.handleSubmit(handleSaveOrg)} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-zinc-400">Name</label>
            <input
              {...orgForm.register('name')}
              className="mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 outline-none focus:border-blue-500"
            />
            {orgForm.formState.errors.name && (
              <p className="mt-1 text-xs text-red-400">
                {orgForm.formState.errors.name.message}
              </p>
            )}
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-400">Slug</label>
            <input
              value={currentOrg?.slug ?? ''}
              disabled
              className="mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900/50 px-3 py-2 text-sm text-zinc-500 outline-none"
            />
          </div>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={saving}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
            >
              {saving && <Loader2 className="h-4 w-4 animate-spin" />}
              Save
            </button>
            {saveSuccess && (
              <span className="text-xs text-green-400">Saved successfully</span>
            )}
          </div>
        </form>
      </div>

      {/* ngrok Tunnel */}
      <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Globe className="h-4 w-4 text-zinc-400" />
            <h2 className="text-sm font-medium text-zinc-300">Public Tunnel (ngrok)</h2>
          </div>
          {tunnelStatus.has_auth_token && (
            <button
              onClick={handleTunnelToggle}
              disabled={tunnelLoading}
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                tunnelStatus.enabled ? 'bg-blue-600' : 'bg-zinc-600'
              }`}
            >
              {tunnelLoading && (
                <Loader2 className="absolute left-1/2 top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 animate-spin text-white" />
              )}
              <span
                className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                  tunnelStatus.enabled ? 'translate-x-6' : 'translate-x-1'
                } ${tunnelLoading ? 'opacity-0' : ''}`}
              />
            </button>
          )}
        </div>

        {/* Authtoken configuration */}
        <div className="mb-4 space-y-3">
          <div>
            <label className="block text-sm font-medium text-zinc-400 mb-1">
              Authtoken
              {tunnelStatus.has_auth_token && (
                <span className="ml-2 text-xs text-green-400">configured</span>
              )}
            </label>
            <form onSubmit={handleSaveToken} className="flex gap-2">
              <input
                type="password"
                value={authToken}
                onChange={(e) => setAuthToken(e.target.value)}
                placeholder={tunnelStatus.has_auth_token ? '••••••••••••••••' : 'Paste your ngrok authtoken'}
                className="flex-1 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-600 outline-none focus:border-blue-500"
              />
              <button
                type="submit"
                disabled={tokenSaving || !authToken}
                className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
              >
                {tokenSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Save'}
              </button>
              {tunnelStatus.has_auth_token && (
                <button
                  type="button"
                  onClick={handleClearToken}
                  disabled={tokenSaving}
                  className="rounded-lg border border-zinc-700 px-3 py-2 text-sm text-zinc-400 transition-colors hover:bg-zinc-700 hover:text-red-400 disabled:opacity-50"
                >
                  Clear
                </button>
              )}
            </form>
            {tokenSaved && (
              <p className="mt-1 text-xs text-green-400">Token saved successfully</p>
            )}
            <p className="mt-1 text-xs text-zinc-600">
              Get your authtoken from{' '}
              <a
                href="https://dashboard.ngrok.com/get-started/your-authtoken"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-400 hover:underline"
              >
                dashboard.ngrok.com
              </a>
            </p>
          </div>
        </div>

        {/* Tunnel status */}
        {tunnelStatus.enabled && tunnelStatus.public_url ? (
          <div className="space-y-3">
            <div className="flex items-center gap-2 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2">
              <span className="relative flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
                <span className="relative inline-flex h-2 w-2 rounded-full bg-green-500" />
              </span>
              <span className="flex-1 font-mono text-sm text-zinc-200 truncate">
                {tunnelStatus.public_url}
              </span>
              <button
                onClick={handleCopyUrl}
                className="rounded p-1 text-zinc-500 transition-colors hover:bg-zinc-700 hover:text-zinc-300"
                title="Copy URL"
              >
                {copied ? <Check className="h-4 w-4 text-green-400" /> : <Copy className="h-4 w-4" />}
              </button>
              <a
                href={tunnelStatus.public_url}
                target="_blank"
                rel="noopener noreferrer"
                className="rounded p-1 text-zinc-500 transition-colors hover:bg-zinc-700 hover:text-zinc-300"
                title="Open in new tab"
              >
                <ExternalLink className="h-4 w-4" />
              </a>
            </div>
            <p className="text-xs text-zinc-500">
              This URL is publicly accessible and forwards to your local control plane.
            </p>
          </div>
        ) : tunnelStatus.enabled ? (
          <div className="flex items-center gap-2 text-sm text-zinc-400">
            <Loader2 className="h-4 w-4 animate-spin" />
            Establishing tunnel...
          </div>
        ) : tunnelStatus.has_auth_token ? (
          <p className="text-sm text-zinc-500">
            Tunnel is off. Toggle it on to expose your control plane via a public URL.
            {tunnelStatus.error && (
              <span className="mt-1 block text-xs text-red-400">{tunnelStatus.error}</span>
            )}
          </p>
        ) : (
          <p className="text-sm text-zinc-500">
            Enter your ngrok authtoken above to enable public tunnel access.
          </p>
        )}
      </div>

      {/* Members */}
      <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
        <h2 className="mb-4 text-sm font-medium text-zinc-300">Members</h2>
        <div className="divide-y divide-zinc-700/50">
          {members.map((m) => (
            <div key={m.user_id} className="flex items-center justify-between py-3">
              <div>
                <p className="text-sm font-medium text-zinc-200">
                  {m.display_name ?? 'Unknown'}
                </p>
                <p className="text-xs text-zinc-500">{m.email}</p>
              </div>
              <div className="flex items-center gap-3">
                <select
                  value={m.role}
                  onChange={(e) => handleRoleChange(m.user_id, e.target.value as OrgRole)}
                  className="rounded-lg border border-zinc-700 bg-zinc-900 px-2 py-1 text-xs text-zinc-300 outline-none focus:border-blue-500"
                >
                  <option value="owner">Owner</option>
                  <option value="admin">Admin</option>
                  <option value="member">Member</option>
                  <option value="viewer">Viewer</option>
                </select>
                <button
                  onClick={() => handleRemoveMember(m.user_id)}
                  className="rounded p-1 text-zinc-500 transition-colors hover:bg-zinc-700 hover:text-red-400"
                  title="Remove member"
                >
                  <UserMinus className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
          {members.length === 0 && (
            <p className="py-4 text-center text-sm text-zinc-500">No members found.</p>
          )}
        </div>
      </div>

      {/* Invite form */}
      <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
        <h2 className="mb-4 text-sm font-medium text-zinc-300">Invite Member</h2>
        <form
          onSubmit={inviteForm.handleSubmit(handleInvite)}
          className="flex flex-col gap-3 sm:flex-row"
        >
          <input
            {...inviteForm.register('email')}
            placeholder="email@company.com"
            className="flex-1 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-600 outline-none focus:border-blue-500"
          />
          <select
            {...inviteForm.register('role')}
            className="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 outline-none focus:border-blue-500"
          >
            <option value="admin">Admin</option>
            <option value="member">Member</option>
            <option value="viewer">Viewer</option>
          </select>
          <button
            type="submit"
            className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
          >
            <UserPlus className="h-4 w-4" />
            Invite
          </button>
        </form>
        {inviteForm.formState.errors.email && (
          <p className="mt-1 text-xs text-red-400">
            {inviteForm.formState.errors.email.message}
          </p>
        )}
      </div>

      {/* Pending invitations */}
      {invitations.length > 0 && (
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
          <h2 className="mb-4 text-sm font-medium text-zinc-300">Pending Invitations</h2>
          <div className="divide-y divide-zinc-700/50">
            {invitations.map((inv) => (
              <div key={inv.id} className="flex items-center justify-between py-3">
                <div>
                  <p className="text-sm text-zinc-200">{inv.email}</p>
                  <div className="mt-0.5 flex items-center gap-2 text-xs text-zinc-500">
                    <Shield className="h-3 w-3" />
                    {inv.role}
                    <span className="text-zinc-600">|</span>
                    <StatusBadge status={inv.status} showDot={false} />
                    <span className="text-zinc-600">|</span>
                    Expires <TimeAgo date={inv.expires_at} />
                  </div>
                </div>
                <button
                  onClick={() => handleCancelInvite(inv.id)}
                  className="rounded p-1.5 text-zinc-500 transition-colors hover:bg-zinc-700 hover:text-red-400"
                  title="Cancel invitation"
                >
                  <X className="h-4 w-4" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
