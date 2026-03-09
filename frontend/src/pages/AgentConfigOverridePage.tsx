import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, Trash2, Eye, ChevronDown, ChevronRight, X } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import { useConfigStore } from '@/store/config-store';
import { useAgent } from '@/hooks/useAgents';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';

interface ConfigField {
  key: string;
  label: string;
  type: string;
  placeholder?: string;
  options?: string[];
  description?: string;
}

interface ConfigSection {
  key: string;
  label: string;
  description?: string;
  nested?: string;
  fields: ConfigField[];
}

// Agent-level sections include GitHub owner/repo/token for per-repo targeting
const AGENT_OVERRIDE_SECTIONS: ConfigSection[] = [
  {
    key: 'github',
    label: 'GitHub',
    description: 'Repository-specific GitHub settings for this agent.',
    fields: [
      { key: 'token', label: 'GitHub Token', type: 'text', placeholder: 'ghp_...', description: 'Personal access token (repo scope)' },
      { key: 'owner', label: 'Repository Owner', type: 'text', description: 'GitHub org or username (e.g. myorg)' },
      { key: 'repo', label: 'Repository Name', type: 'text', description: 'Repository name (e.g. my-project)' },
      { key: 'poll_interval', label: 'Poll Interval', type: 'text', placeholder: '30s', description: 'Override polling interval' },
      { key: 'watch_labels', label: 'Watch Labels', type: 'text', placeholder: 'agent:ready', description: 'Override watch labels (comma-separated)' },
    ],
  },
  {
    key: 'claude',
    label: 'Claude AI',
    description: 'Override Claude model settings for this agent.',
    fields: [
      { key: 'api_key', label: 'API Key', type: 'text', placeholder: 'sk-ant-...', description: 'Override Anthropic API key' },
      { key: 'model', label: 'Model', type: 'text', placeholder: 'claude-sonnet-4-20250514', description: 'Override Claude model' },
      { key: 'max_tokens', label: 'Max Tokens', type: 'number', placeholder: '8192', description: 'Override max tokens' },
    ],
  },
  {
    key: 'agents',
    label: 'Agent Settings',
    description: 'Override core agent behavior.',
    nested: 'developer',
    fields: [
      { key: 'enabled', label: 'Enabled', type: 'boolean', description: 'Override enabled state' },
      { key: 'max_concurrent', label: 'Max Concurrent', type: 'number', placeholder: '1' },
      { key: 'workspace_dir', label: 'Workspace Directory', type: 'text', placeholder: './workspaces' },
      { key: 'allow_pr_merging', label: 'Allow PR Merging', type: 'boolean' },
      { key: 'allow_auto_issue_processing', label: 'Auto Issue Processing', type: 'boolean' },
    ],
  },
  {
    key: 'creativity',
    label: 'Creativity Engine',
    fields: [
      { key: 'enabled', label: 'Enabled', type: 'boolean' },
      { key: 'idle_threshold_seconds', label: 'Idle Threshold (seconds)', type: 'number', placeholder: '120' },
      { key: 'suggestion_cooldown_seconds', label: 'Suggestion Cooldown (seconds)', type: 'number', placeholder: '300' },
      { key: 'max_pending_suggestions', label: 'Max Pending Suggestions', type: 'number', placeholder: '5' },
      { key: 'max_rejection_history', label: 'Max Rejection History', type: 'number', placeholder: '50' },
    ],
  },
  {
    key: 'decomposition',
    label: 'Issue Decomposition',
    fields: [
      { key: 'enabled', label: 'Enabled', type: 'boolean' },
      { key: 'max_iteration_budget', label: 'Max Iteration Budget', type: 'number', placeholder: '250' },
      { key: 'max_subtasks', label: 'Max Subtasks', type: 'number', placeholder: '5' },
    ],
  },
  {
    key: 'memory',
    label: 'Repository Memory',
    fields: [
      { key: 'enabled', label: 'Enabled', type: 'boolean' },
      { key: 'max_entries', label: 'Max Entries', type: 'number', placeholder: '100' },
      { key: 'max_prompt_size', label: 'Max Prompt Size', type: 'number', placeholder: '8000' },
      { key: 'extract_on_complete', label: 'Extract on Complete', type: 'boolean' },
    ],
  },
  {
    key: 'logging',
    label: 'Logging',
    fields: [
      { key: 'level', label: 'Log Level', type: 'select', options: ['debug', 'info', 'warn', 'error'] },
      { key: 'file_path', label: 'Log File Path', type: 'text', placeholder: './logs/agent.log' },
    ],
  },
  {
    key: 'shutdown',
    label: 'Shutdown',
    fields: [
      { key: 'timeout', label: 'Timeout', type: 'text', placeholder: '30s' },
      { key: 'cleanup_workspaces', label: 'Cleanup Workspaces', type: 'boolean' },
      { key: 'reset_claims', label: 'Reset Claims', type: 'boolean' },
    ],
  },
  {
    key: 'error_handling',
    label: 'Error Handling',
    fields: [
      { key: 'retry_enabled', label: 'Retry Enabled', type: 'boolean' },
      { key: 'retry_max_attempts', label: 'Retry Max Attempts', type: 'number', placeholder: '3' },
      { key: 'retry_base_delay', label: 'Retry Base Delay', type: 'text', placeholder: '1s' },
      { key: 'retry_max_delay', label: 'Retry Max Delay', type: 'text', placeholder: '30s' },
      { key: 'retry_backoff_factor', label: 'Retry Backoff Factor', type: 'number', placeholder: '2.0' },
      { key: 'circuit_breaker_enabled', label: 'Circuit Breaker Enabled', type: 'boolean' },
      { key: 'circuit_breaker_max_failures', label: 'CB Max Failures', type: 'number', placeholder: '5' },
      { key: 'circuit_breaker_timeout', label: 'CB Timeout', type: 'text', placeholder: '60s' },
      { key: 'circuit_breaker_failure_ratio', label: 'CB Failure Ratio', type: 'number', placeholder: '0.5' },
    ],
  },
];

function getNestedValue(obj: Record<string, unknown>, section: ConfigSection, fieldKey: string): unknown {
  const sectionData = obj[section.key] as Record<string, unknown> | undefined;
  if (!sectionData) return undefined;
  if (section.nested) {
    const nested = sectionData[section.nested] as Record<string, unknown> | undefined;
    return nested?.[fieldKey];
  }
  return sectionData[fieldKey];
}

function setNestedValue(
  obj: Record<string, unknown>,
  section: ConfigSection,
  fieldKey: string,
  value: unknown,
): Record<string, unknown> {
  const result = { ...obj };
  if (section.nested) {
    const sectionData = { ...(result[section.key] as Record<string, unknown> || {}) };
    const nestedData = { ...(sectionData[section.nested] as Record<string, unknown> || {}) };
    if (value === undefined) {
      delete nestedData[fieldKey];
    } else {
      nestedData[fieldKey] = value;
    }
    // Clean up empty nested objects
    if (Object.keys(nestedData).length === 0) {
      delete sectionData[section.nested];
    } else {
      sectionData[section.nested] = nestedData;
    }
    if (Object.keys(sectionData).length === 0) {
      delete result[section.key];
    } else {
      result[section.key] = sectionData;
    }
  } else {
    const sectionData = { ...(result[section.key] as Record<string, unknown> || {}) };
    if (value === undefined) {
      delete sectionData[fieldKey];
    } else {
      sectionData[fieldKey] = value;
    }
    if (Object.keys(sectionData).length === 0) {
      delete result[section.key];
    } else {
      result[section.key] = sectionData;
    }
  }
  return result;
}

export function AgentConfigOverridePage() {
  const { agentId } = useParams<{ agentId: string }>();
  const navigate = useNavigate();
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const { agent } = useAgent(agentId);
  const {
    agentOverride,
    selectedTypeConfig,
    mergedPreview,
    mergedVersion,
    loading,
    saving,
    error,
    fetchAgentOverride,
    fetchTypeConfig,
    saveAgentOverride,
    deleteAgentOverride,
    fetchMergedPreview,
    clearError,
  } = useConfigStore();

  const [overrideData, setOverrideData] = useState<Record<string, unknown>>({});
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['github']));
  const [jsonMode, setJsonMode] = useState(false);
  const [jsonText, setJsonText] = useState('{}');
  const [showMerged, setShowMerged] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    if (currentOrg && agentId) {
      fetchAgentOverride(currentOrg.slug, agentId);
      // Also fetch the type config to show inherited values
      if (agent?.agent_type) {
        fetchTypeConfig(currentOrg.slug, agent.agent_type);
      }
    }
  }, [currentOrg, agentId, agent?.agent_type, fetchAgentOverride, fetchTypeConfig]);

  useEffect(() => {
    if (agentOverride) {
      const cfg = agentOverride.config ?? {};
      setOverrideData(cfg);
      setJsonText(JSON.stringify(cfg, null, 2));
      setHasChanges(false);
    }
  }, [agentOverride]);

  const toggleSection = (key: string) => {
    setExpandedSections((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  };

  const globalConfig = selectedTypeConfig?.config ?? {};

  const getInheritedValue = useCallback((section: ConfigSection, fieldKey: string): unknown => {
    return getNestedValue(globalConfig, section, fieldKey);
  }, [globalConfig]);

  const handleFieldChange = (section: ConfigSection, fieldKey: string, value: unknown) => {
    const updated = setNestedValue(overrideData, section, fieldKey, value);
    setOverrideData(updated);
    setJsonText(JSON.stringify(updated, null, 2));
    setHasChanges(true);
  };

  const handleClearField = (section: ConfigSection, fieldKey: string) => {
    const updated = setNestedValue(overrideData, section, fieldKey, undefined);
    setOverrideData(updated);
    setJsonText(JSON.stringify(updated, null, 2));
    setHasChanges(true);
  };

  const handleJsonChange = (text: string) => {
    setJsonText(text);
    try {
      const parsed = JSON.parse(text);
      setOverrideData(parsed);
      setHasChanges(true);
    } catch {
      // Invalid JSON
    }
  };

  const handleSave = async () => {
    if (!currentOrg || !agentId) return;
    let dataToSave = overrideData;
    if (jsonMode) {
      try {
        dataToSave = JSON.parse(jsonText);
      } catch {
        return;
      }
    }
    await saveAgentOverride(currentOrg.slug, agentId, dataToSave);
    setHasChanges(false);
  };

  const handleDelete = async () => {
    if (!currentOrg || !agentId) return;
    if (!window.confirm('Remove all configuration overrides for this agent? It will inherit all settings from the global configuration.')) return;
    await deleteAgentOverride(currentOrg.slug, agentId);
    setOverrideData({});
    setJsonText('{}');
    setHasChanges(false);
  };

  const handlePreview = () => {
    if (currentOrg && agentId) {
      fetchMergedPreview(currentOrg.slug, agentId);
      setShowMerged(true);
    }
  };

  if (loading && !agentOverride) {
    return <LoadingSpinner size="lg" className="py-20" />;
  }

  const overrideCount = countOverrides(overrideData);

  return (
    <div className="space-y-6">
      <button
        onClick={() => navigate(`/agents/${agentId}`)}
        className="flex items-center gap-1.5 text-sm text-zinc-400 transition-colors hover:text-zinc-200"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to agent
      </button>

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-zinc-100">
            Agent Configuration {agent ? `- ${agent.name}` : ''}
          </h1>
          <p className="mt-1 text-sm text-zinc-400">
            Override settings from the global <span className="text-zinc-300">{agent?.agent_type ?? 'developer'}</span> configuration.
            Empty fields inherit from the global config. {overrideCount > 0 && (
              <span className="text-amber-400">{overrideCount} override{overrideCount !== 1 ? 's' : ''} active.</span>
            )}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handlePreview}
            className="flex items-center gap-1.5 rounded-lg border border-zinc-600 px-3 py-1.5 text-sm text-zinc-300 transition-colors hover:bg-zinc-700"
          >
            <Eye className="h-4 w-4" />
            Preview Merged
          </button>
          <button
            onClick={() => setJsonMode(!jsonMode)}
            className="rounded-lg border border-zinc-600 px-3 py-1.5 text-sm text-zinc-300 transition-colors hover:bg-zinc-700"
          >
            {jsonMode ? 'Form View' : 'JSON View'}
          </button>
          <button
            onClick={handleDelete}
            disabled={saving || (!agentOverride?.id)}
            className="flex items-center gap-1.5 rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-1.5 text-sm text-red-400 transition-colors hover:bg-red-500/20 disabled:opacity-40"
          >
            <Trash2 className="h-4 w-4" />
            Clear All
          </button>
          <button
            onClick={handleSave}
            disabled={!hasChanges || saving}
            className="flex items-center gap-1.5 rounded-lg bg-blue-600 px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-blue-500 disabled:opacity-40"
          >
            <Save className="h-4 w-4" />
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </div>

      {error && (
        <div className="flex items-center justify-between rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-2 text-sm text-red-400">
          <span>{error}</span>
          <button onClick={clearError} className="text-red-400 hover:text-red-300">Dismiss</button>
        </div>
      )}

      {jsonMode ? (
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-4">
          <h2 className="mb-3 text-sm font-medium text-zinc-300">Override Configuration (JSON)</h2>
          <p className="mb-3 text-xs text-zinc-500">Only include fields you want to override. Empty object = fully inherit global config.</p>
          <textarea
            value={jsonText}
            onChange={(e) => handleJsonChange(e.target.value)}
            className="h-[500px] w-full resize-none rounded-lg bg-zinc-900 p-4 font-mono text-sm text-zinc-300 focus:outline-none focus:ring-1 focus:ring-blue-500"
            spellCheck={false}
          />
        </div>
      ) : (
        <div className="space-y-2">
          {AGENT_OVERRIDE_SECTIONS.map((section) => {
            const sectionHasOverrides = section.fields.some(
              (f) => getNestedValue(overrideData, section, f.key) !== undefined
            );
            return (
              <div key={section.key} className={`rounded-lg border bg-zinc-800 ${sectionHasOverrides ? 'border-amber-500/30' : 'border-zinc-700'}`}>
                <button
                  onClick={() => toggleSection(section.key)}
                  className="flex w-full items-center justify-between px-4 py-3 text-left"
                >
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-zinc-200">{section.label}</span>
                    {sectionHasOverrides && (
                      <span className="rounded-full bg-amber-500/20 px-2 py-0.5 text-xs text-amber-400">overridden</span>
                    )}
                  </div>
                  {expandedSections.has(section.key) ? (
                    <ChevronDown className="h-4 w-4 text-zinc-400 shrink-0" />
                  ) : (
                    <ChevronRight className="h-4 w-4 text-zinc-400 shrink-0" />
                  )}
                </button>
                {expandedSections.has(section.key) && (
                  <div className="border-t border-zinc-700 px-4 py-3 space-y-3">
                    {section.fields.map((field) => {
                      const overrideValue = getNestedValue(overrideData, section, field.key);
                      const inheritedValue = getInheritedValue(section, field.key);
                      const isOverridden = overrideValue !== undefined;
                      const displayValue = isOverridden ? overrideValue : undefined;
                      const inheritedDisplay = inheritedValue !== undefined ? String(inheritedValue) : '';

                      return (
                        <div key={field.key} className="flex items-start justify-between gap-4">
                          <div className="min-w-[220px]">
                            <label className={`text-sm ${isOverridden ? 'text-amber-300' : 'text-zinc-400'}`}>
                              {field.label}
                            </label>
                            {field.description && (
                              <p className="text-xs text-zinc-500 mt-0.5">{field.description}</p>
                            )}
                            {!isOverridden && inheritedDisplay && (
                              <p className="text-xs text-zinc-600 mt-0.5">
                                Inherited: {field.type === 'boolean' ? (inheritedValue ? 'On' : 'Off') : inheritedDisplay}
                              </p>
                            )}
                          </div>
                          <div className="flex items-center gap-1.5">
                            {field.type === 'boolean' ? (
                              <div className="flex items-center gap-2">
                                {!isOverridden && (
                                  <span className="text-xs text-zinc-600">
                                    {inheritedValue ? 'On' : 'Off'}
                                  </span>
                                )}
                                <button
                                  onClick={() => {
                                    if (!isOverridden) {
                                      // First click: set to opposite of inherited
                                      handleFieldChange(section, field.key, !inheritedValue);
                                    } else {
                                      handleFieldChange(section, field.key, !displayValue);
                                    }
                                  }}
                                  className={`relative h-6 w-11 shrink-0 rounded-full transition-colors ${
                                    isOverridden
                                      ? (displayValue ? 'bg-blue-600 ring-1 ring-amber-400/50' : 'bg-zinc-600 ring-1 ring-amber-400/50')
                                      : (inheritedValue ? 'bg-blue-600/40' : 'bg-zinc-700')
                                  }`}
                                >
                                  <span
                                    className={`absolute left-0.5 top-0.5 h-5 w-5 rounded-full transition-transform ${
                                      isOverridden ? 'bg-white' : 'bg-zinc-400'
                                    } ${
                                      (isOverridden ? displayValue : inheritedValue) ? 'translate-x-5' : ''
                                    }`}
                                  />
                                </button>
                              </div>
                            ) : field.type === 'select' ? (
                              <select
                                value={isOverridden ? (displayValue as string) : ''}
                                onChange={(e) => handleFieldChange(section, field.key, e.target.value || undefined)}
                                className={`rounded-lg border bg-zinc-900 px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500 ${
                                  isOverridden ? 'border-amber-500/50 text-zinc-300' : 'border-zinc-600 text-zinc-500'
                                }`}
                              >
                                <option value="">{inheritedDisplay ? `Inherited: ${inheritedDisplay}` : 'Not set'}</option>
                                {field.options?.map((opt) => (
                                  <option key={opt} value={opt}>{opt}</option>
                                ))}
                              </select>
                            ) : field.type === 'number' ? (
                              <input
                                type="number"
                                step="any"
                                value={displayValue !== undefined && displayValue !== null ? String(displayValue) : ''}
                                onChange={(e) => handleFieldChange(section, field.key, e.target.value ? Number(e.target.value) : undefined)}
                                placeholder={inheritedDisplay || field.placeholder}
                                className={`w-48 rounded-lg border bg-zinc-900 px-3 py-1.5 text-sm placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-blue-500 ${
                                  isOverridden ? 'border-amber-500/50 text-zinc-300' : 'border-zinc-600 text-zinc-500'
                                }`}
                              />
                            ) : (
                              <input
                                type={field.key.includes('token') || field.key.includes('api_key') ? 'password' : 'text'}
                                value={isOverridden ? (displayValue as string) : ''}
                                onChange={(e) => handleFieldChange(section, field.key, e.target.value || undefined)}
                                placeholder={inheritedDisplay || field.placeholder}
                                className={`w-64 rounded-lg border bg-zinc-900 px-3 py-1.5 text-sm placeholder:text-zinc-600 focus:outline-none focus:ring-1 focus:ring-blue-500 ${
                                  isOverridden ? 'border-amber-500/50 text-zinc-300' : 'border-zinc-600 text-zinc-500'
                                }`}
                              />
                            )}
                            {isOverridden && (
                              <button
                                onClick={() => handleClearField(section, field.key)}
                                className="rounded p-1 text-zinc-500 hover:bg-zinc-700 hover:text-zinc-300"
                                title="Clear override (inherit from global)"
                              >
                                <X className="h-3.5 w-3.5" />
                              </button>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Override version info */}
      {agentOverride && agentOverride.version > 0 && (
        <div className="text-xs text-zinc-500">
          Override version: {agentOverride.version}
        </div>
      )}

      {/* Merged preview */}
      {showMerged && mergedPreview && (
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-4">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-sm font-medium text-zinc-300">Effective Configuration (Global + Overrides)</h2>
            <div className="flex items-center gap-3">
              <span className="text-xs text-zinc-500">Effective version: {mergedVersion}</span>
              <button onClick={() => setShowMerged(false)} className="text-zinc-500 hover:text-zinc-300">
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>
          <pre className="max-h-[400px] overflow-auto rounded-lg bg-zinc-900 p-4 text-xs text-zinc-400">
            {JSON.stringify(mergedPreview, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}

function countOverrides(obj: Record<string, unknown>): number {
  let count = 0;
  for (const val of Object.values(obj)) {
    if (val && typeof val === 'object' && !Array.isArray(val)) {
      count += countOverrides(val as Record<string, unknown>);
    } else {
      count++;
    }
  }
  return count;
}
