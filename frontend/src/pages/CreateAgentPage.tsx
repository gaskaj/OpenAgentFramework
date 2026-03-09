import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { ArrowLeft, Bot, Copy, Check, Key, AlertTriangle } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import { provisionAgent } from '@/api/agents';
import { DownloadInstructions } from '@/components/DownloadInstructions';
import type { Agent } from '@/types';

const AGENT_TYPES = [
  { value: 'developer', label: 'Developer', description: 'Monitors GitHub issues, implements solutions, and creates pull requests' },
  { value: 'reviewer', label: 'Reviewer', description: 'Reviews pull requests and provides feedback' },
  { value: 'monitor', label: 'Monitor', description: 'Monitors system health and reports issues' },
] as const;

const schema = z.object({
  agent_type: z.string().min(1, 'Agent type is required'),
  name: z.string().max(255).optional(),
});

type FormData = z.infer<typeof schema>;

export function CreateAgentPage() {
  const navigate = useNavigate();
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const [createdAgent, setCreatedAgent] = useState<Agent | null>(null);
  const [rawKey, setRawKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { agent_type: 'developer' },
  });

  const selectedType = watch('agent_type');

  const onSubmit = async (data: FormData) => {
    if (!currentOrg) return;
    setError(null);
    try {
      const result = await provisionAgent(
        currentOrg.slug,
        data.agent_type,
        data.name || undefined,
      );
      setCreatedAgent(result.agent);
      setRawKey(result.key);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create agent';
      setError(msg);
    }
  };

  const copyToClipboard = async (text: string) => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // Success view — show agent + API key
  if (createdAgent && rawKey) {
    return (
      <div className="mx-auto max-w-2xl space-y-6">
        <button
          onClick={() => navigate('/agents')}
          className="flex items-center gap-1.5 text-sm text-zinc-400 transition-colors hover:text-zinc-200"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to agents
        </button>

        {/* Success header */}
        <div className="rounded-lg border border-green-500/30 bg-green-500/5 p-6">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-green-500/20">
              <Bot className="h-6 w-6 text-green-400" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-zinc-100">Agent Created</h1>
              <p className="text-sm text-zinc-400">
                <span className="font-medium text-zinc-200">{createdAgent.name}</span>
                {' '}is ready to connect.
              </p>
            </div>
          </div>

          <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-zinc-500">Name</span>
              <p className="font-medium text-zinc-200">{createdAgent.name}</p>
            </div>
            <div>
              <span className="text-zinc-500">Type</span>
              <p className="font-medium text-zinc-200">{createdAgent.agent_type}</p>
            </div>
            <div>
              <span className="text-zinc-500">Status</span>
              <p className="font-medium text-yellow-400">Offline</p>
            </div>
            <div>
              <span className="text-zinc-500">Agent ID</span>
              <p className="font-mono text-xs text-zinc-400">{createdAgent.id}</p>
            </div>
          </div>
        </div>

        {/* API Key */}
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
          <div className="flex items-center gap-3">
            <Key className="h-5 w-5 text-blue-400" />
            <div>
              <h2 className="text-lg font-semibold text-zinc-100">API Key</h2>
              <p className="text-sm text-zinc-400">
                Use this key in your agent's configuration to authenticate.
              </p>
            </div>
          </div>

          <div className="mt-4 rounded-lg border border-green-500/30 bg-green-500/10 p-4">
            <label className="mb-2 block text-xs font-medium text-green-400">
              Your API Key
            </label>
            <div className="flex items-center gap-2">
              <code className="flex-1 select-all overflow-x-auto rounded-lg bg-zinc-900 px-3 py-2.5 font-mono text-sm text-green-300">
                {rawKey}
              </code>
              <button
                onClick={() => copyToClipboard(rawKey)}
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
              Copy this key now — it will not be shown again. Use it in your agent's config file
              as the <code className="text-zinc-300">controlplane.api_key</code> value.
            </p>
          </div>

          {/* Minimal config example */}
          <div className="mt-4">
            <label className="mb-2 block text-xs font-medium text-zinc-400">
              Minimal agent configuration
            </label>
            <pre className="overflow-x-auto rounded-lg bg-zinc-900 p-4 text-xs text-zinc-400">
{`controlplane:
  enabled: true
  url: "${window.location.origin}"
  api_key: "${rawKey}"
  config_mode: "remote"
  config_poll_interval: "30s"`}
            </pre>
          </div>
        </div>

        {/* Download & Run instructions */}
        <DownloadInstructions
          apiKey={rawKey}
          controlPlaneUrl={window.location.origin}
        />

        <div className="flex justify-end gap-3">
          <button
            onClick={() => navigate(`/agents/${createdAgent.id}`)}
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
          >
            View Agent
          </button>
          <button
            onClick={() => navigate('/agents')}
            className="rounded-lg bg-zinc-700 px-4 py-2 text-sm font-medium text-zinc-200 transition-colors hover:bg-zinc-600"
          >
            Back to Agents
          </button>
        </div>
      </div>
    );
  }

  // Create form
  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <button
        onClick={() => navigate('/agents')}
        className="flex items-center gap-1.5 text-sm text-zinc-400 transition-colors hover:text-zinc-200"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to agents
      </button>

      <div>
        <h1 className="text-2xl font-bold text-zinc-100">Create New Agent</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Provision a new agent and generate its API key for authentication.
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* Agent type selection */}
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
          <label className="mb-3 block text-sm font-medium text-zinc-300">
            Agent Type
          </label>
          <div className="space-y-3">
            {AGENT_TYPES.map((type) => (
              <label
                key={type.value}
                className={`flex cursor-pointer items-start gap-3 rounded-lg border p-4 transition-colors ${
                  selectedType === type.value
                    ? 'border-blue-500 bg-blue-500/10'
                    : 'border-zinc-700 bg-zinc-900 hover:border-zinc-600'
                }`}
              >
                <input
                  type="radio"
                  value={type.value}
                  {...register('agent_type')}
                  className="mt-0.5 h-4 w-4 accent-blue-500"
                />
                <div>
                  <p className="text-sm font-medium text-zinc-200">{type.label}</p>
                  <p className="mt-0.5 text-xs text-zinc-500">{type.description}</p>
                </div>
              </label>
            ))}
          </div>
          {errors.agent_type && (
            <p className="mt-2 text-xs text-red-400">{errors.agent_type.message}</p>
          )}
        </div>

        {/* Agent name */}
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-6">
          <label className="mb-1 block text-sm font-medium text-zinc-300">
            Agent Name
          </label>
          <p className="mb-3 text-xs text-zinc-500">
            Leave blank to auto-generate as <code className="text-zinc-400">{selectedType}-XX</code> (e.g., developer-01).
          </p>
          <input
            {...register('name')}
            placeholder={`${selectedType}-01`}
            className="w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-600 outline-none focus:border-blue-500"
          />
          {errors.name && (
            <p className="mt-1 text-xs text-red-400">{errors.name.message}</p>
          )}
        </div>

        {error && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-sm text-red-400">
            {error}
          </div>
        )}

        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate('/agents')}
            className="rounded-lg bg-zinc-700 px-4 py-2 text-sm font-medium text-zinc-200 transition-colors hover:bg-zinc-600"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
          >
            <Bot className="h-4 w-4" />
            {isSubmitting ? 'Creating...' : 'Create Agent'}
          </button>
        </div>
      </form>
    </div>
  );
}
