import { useState, useEffect, useCallback } from 'react';
import { ChevronDown, ChevronRight } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import apiClient from '@/api/client';
import type { AuditLog, AuditFilters, PaginatedResponse } from '@/types';
import { DataTable, type Column } from '@/components/common/DataTable';
import { TimeAgo } from '@/components/common/TimeAgo';


export function AuditLogPage() {
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [filters, setFilters] = useState<AuditFilters>({ page: 1, per_page: 25 });
  const [expandedRow, setExpandedRow] = useState<string | null>(null);

  const fetchLogs = useCallback(async () => {
    if (!currentOrg) return;
    setLoading(true);
    try {
      const { data } = await apiClient.get<PaginatedResponse<AuditLog>>(
        `/orgs/${currentOrg.slug}/audit`,
        { params: filters },
      );
      setLogs(data.data ?? []);
      setTotal(data.total ?? 0);
    } finally {
      setLoading(false);
    }
  }, [currentOrg, filters]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const totalPages = Math.ceil(total / (filters.per_page || 25));

  const columns: Column<AuditLog>[] = [
    {
      key: 'expand',
      header: '',
      className: 'w-8',
      render: (log) => (
        <button
          onClick={() => setExpandedRow(expandedRow === log.id ? null : log.id)}
          className="text-zinc-500 hover:text-zinc-300"
        >
          {expandedRow === log.id ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </button>
      ),
    },
    {
      key: 'action',
      header: 'Action',
      sortable: true,
      render: (log) => (
        <span className="font-medium text-zinc-200">{log.action}</span>
      ),
    },
    {
      key: 'user',
      header: 'User',
      render: (log) => <span className="text-zinc-400">{log.user_email}</span>,
    },
    {
      key: 'resource',
      header: 'Resource',
      render: (log) => (
        <span className="text-zinc-400">
          {log.resource_type}
          {log.resource_id && (
            <span className="ml-1 font-mono text-xs text-zinc-600">
              {log.resource_id.substring(0, 8)}
            </span>
          )}
        </span>
      ),
    },
    {
      key: 'ip',
      header: 'IP',
      render: (log) => (
        <span className="font-mono text-xs text-zinc-500">{log.ip_address}</span>
      ),
    },
    {
      key: 'created_at',
      header: 'Time',
      sortable: true,
      render: (log) => <TimeAgo date={log.created_at} className="text-zinc-500" />,
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-zinc-100">Audit Log</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Track all actions taken in your organization
        </p>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <input
          type="text"
          placeholder="Filter by action..."
          value={filters.action ?? ''}
          onChange={(e) =>
            setFilters({ ...filters, action: e.target.value || undefined, page: 1 })
          }
          className="rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-sm text-zinc-300 placeholder-zinc-600 outline-none focus:border-blue-500"
        />
        <input
          type="text"
          placeholder="Filter by resource type..."
          value={filters.resource_type ?? ''}
          onChange={(e) =>
            setFilters({ ...filters, resource_type: e.target.value || undefined, page: 1 })
          }
          className="rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-sm text-zinc-300 placeholder-zinc-600 outline-none focus:border-blue-500"
        />
      </div>

      {/* Table */}
      <div>
        <DataTable
          columns={columns}
          data={logs}
          loading={loading}
          rowKey={(l) => l.id}
          emptyMessage="No audit log entries found."
          page={filters.page}
          totalPages={totalPages}
          onPageChange={(p) => setFilters({ ...filters, page: p })}
        />

        {/* Expanded row details */}
        {expandedRow && (
          <div className="mt-2 rounded-lg border border-zinc-700 bg-zinc-800 p-4">
            <h3 className="mb-2 text-xs font-medium uppercase tracking-wider text-zinc-500">
              Details
            </h3>
            <pre className="overflow-x-auto rounded-lg bg-zinc-900 p-4 text-xs text-zinc-400">
              {JSON.stringify(
                logs.find((l) => l.id === expandedRow)?.details ?? {},
                null,
                2,
              )}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}
