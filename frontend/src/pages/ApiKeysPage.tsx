import { useState, useEffect, useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Plus, Trash2, Copy, Check, Key, AlertTriangle } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import * as apiKeysApi from '@/api/apikeys';
import type { APIKey } from '@/types';
import { TimeAgo } from '@/components/common/TimeAgo';

const AGENT_TYPES = ['developer', 'reviewer', 'monitor'] as const;

const createSchema = z.object({
  agent_type: z.string().min(1, 'Agent type is required'),
  name: z.string().max(255).optional(),
});

type CreateFormData = z.infer<typeof createSchema>;

export function ApiKeysPage() {
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [newRawKey, setNewRawKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<CreateFormData>({
    resolver: zodResolver(createSchema),
    defaultValues: { agent_type: 'developer' },
  });

  const loadKeys = useCallback(async () => {
    if (!currentOrg) return;
    const data = await apiKeysApi.listAPIKeys(currentOrg.slug);
    setKeys(data);
  }, [currentOrg]);

  useEffect(() => {
    loadKeys();
  }, [loadKeys]);

  const handleCreate = async (data: CreateFormData) => {
    if (!currentOrg) return;
    const result = await apiKeysApi.createAPIKey(
      currentOrg.slug,
      data.agent_type,
      data.name || undefined,
    );
    setNewRawKey(result.key);
    reset();
    loadKeys();
  };

  const handleRevoke = async (keyId: string) => {
    if (!currentOrg) return;
    if (!window.confirm('Revoke this API key? This cannot be undone.')) return;
    await apiKeysApi.revokeAPIKey(currentOrg.slug, keyId);
    loadKeys();
  };

  const copyToClipboard = async (text: string) => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-zinc-100">API Keys</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Manage API keys for agent authentication. Each key is bound to an agent type and auto-generates an agent name.
        </p>
      </div>

      {/* Create form */}
      <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
        <h2 className="mb-4 text-sm font-medium text-zinc-300">Create New Key</h2>
        <form
          onSubmit={handleSubmit(handleCreate)}
          className="flex flex-col gap-3"
        >
          <div className="flex flex-col gap-3 sm:flex-row">
            <select
              {...register('agent_type')}
              className="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 outline-none focus:border-blue-500"
            >
              {AGENT_TYPES.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
            <input
              {...register('name')}
              placeholder="Custom name (optional — auto-generated if empty)"
              className="flex-1 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-600 outline-none focus:border-blue-500"
            />
            <button
              type="submit"
              disabled={isSubmitting}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
            >
              <Plus className="h-4 w-4" />
              Create
            </button>
          </div>
          {errors.agent_type && (
            <p className="text-xs text-red-400">{errors.agent_type.message}</p>
          )}
          <p className="text-xs text-zinc-500">
            Agent name defaults to <code className="text-zinc-400">{'{agent_type}-{XX}'}</code> where XX increments per type.
          </p>
        </form>
      </div>

      {/* Modal: show new key once */}
      {newRawKey && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="mx-4 w-full max-w-lg rounded-xl border border-zinc-700 bg-zinc-800 p-6 shadow-2xl">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-green-500/20">
                <Key className="h-5 w-5 text-green-400" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-zinc-100">API Key Created</h2>
                <p className="text-sm text-zinc-400">
                  Copy your key now — it will not be shown again.
                </p>
              </div>
            </div>

            <div className="mt-5 rounded-lg border border-green-500/30 bg-green-500/10 p-4">
              <label className="mb-2 block text-xs font-medium text-green-400">
                Your API Key
              </label>
              <div className="flex items-center gap-2">
                <code className="flex-1 select-all overflow-x-auto rounded-lg bg-zinc-900 px-3 py-2.5 font-mono text-sm text-green-300">
                  {newRawKey}
                </code>
                <button
                  onClick={() => copyToClipboard(newRawKey)}
                  className="flex items-center gap-1.5 rounded-lg border border-green-500/30 bg-green-500/20 px-3 py-2.5 text-sm font-medium text-green-300 transition-colors hover:bg-green-500/30"
                >
                  {copied ? (
                    <>
                      <Check className="h-4 w-4" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="h-4 w-4" />
                      Copy
                    </>
                  )}
                </button>
              </div>
            </div>

            <div className="mt-4 flex items-start gap-2 rounded-lg bg-zinc-900 p-3">
              <AlertTriangle className="mt-0.5 h-4 w-4 flex-shrink-0 text-yellow-400" />
              <p className="text-xs text-zinc-400">
                Store this key securely. For security, we only display it once. If you lose it, you'll need to create a new key.
              </p>
            </div>

            <div className="mt-5 flex justify-end">
              <button
                onClick={() => setNewRawKey(null)}
                className="rounded-lg bg-zinc-700 px-4 py-2 text-sm font-medium text-zinc-200 transition-colors hover:bg-zinc-600"
              >
                Done
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Key list */}
      <div className="rounded-lg border border-zinc-700 bg-zinc-800">
        <div className="border-b border-zinc-700 px-6 py-4">
          <h2 className="text-sm font-medium text-zinc-300">Active Keys</h2>
        </div>
        <div className="divide-y divide-zinc-700/50">
          {keys.length === 0 ? (
            <div className="px-6 py-12 text-center text-sm text-zinc-500">
              No API keys created yet.
            </div>
          ) : (
            keys.map((key) => (
              <div key={key.id} className="flex items-center justify-between px-6 py-4">
                <div className="flex items-center gap-3">
                  <Key className="h-4 w-4 text-zinc-600" />
                  <div>
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium text-zinc-200">
                        {key.agent_name || key.name}
                      </p>
                      <span className="rounded bg-zinc-700 px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider text-zinc-400">
                        {key.agent_type || 'developer'}
                      </span>
                    </div>
                    <div className="mt-0.5 flex items-center gap-3 text-xs text-zinc-500">
                      <span className="font-mono">{key.key_prefix}...</span>
                      <span>
                        Created <TimeAgo date={key.created_at} />
                      </span>
                      {key.last_used_at && (
                        <span>
                          Last used <TimeAgo date={key.last_used_at} />
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => handleRevoke(key.id)}
                  className="flex items-center gap-1.5 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-1.5 text-xs text-red-400 transition-colors hover:bg-red-500/20"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                  Revoke
                </button>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
