import { useEffect } from 'react';
import {
  Bot,
  Wifi,
  WifiOff,
  AlertTriangle,
  GitPullRequest,
  CheckCircle2,
} from 'lucide-react';
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
} from 'recharts';
import { useEventStats } from '@/hooks/useEvents';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useAuthStore } from '@/store/auth-store';
import { useEventStore } from '@/store/event-store';
import { EventFeed } from '@/components/events/EventFeed';

const STATUS_COLORS: Record<string, string> = {
  online: '#22c55e',
  offline: '#6b7280',
  error: '#ef4444',
  idle: '#a1a1aa',
};

function StatCard({
  label,
  value,
  icon: Icon,
  color,
}: {
  label: string;
  value: number | string;
  icon: typeof Bot;
  color: string;
}) {
  return (
    <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-5">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-zinc-400">{label}</p>
          <p className="mt-1 text-2xl font-bold text-zinc-100">{value}</p>
        </div>
        <div
          className="flex h-10 w-10 items-center justify-center rounded-lg"
          style={{ backgroundColor: `${color}20` }}
        >
          <Icon className="h-5 w-5" style={{ color }} />
        </div>
      </div>
    </div>
  );
}

export function DashboardPage() {
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const { stats, refresh } = useEventStats();
  const realtimeEvents = useEventStore((s) => s.realtimeEvents);

  useWebSocket(currentOrg?.slug);

  useEffect(() => {
    const interval = setInterval(refresh, 30000);
    return () => clearInterval(interval);
  }, [refresh]);

  const agentsOnline = stats?.agents_online ?? 0;
  const agentsTotal = stats?.agents_total ?? 0;
  const agentsOffline = agentsTotal - agentsOnline;

  const pieData = [
    { name: 'Online', value: agentsOnline, color: STATUS_COLORS.online },
    { name: 'Offline', value: agentsOffline, color: STATUS_COLORS.offline },
  ].filter((d) => d.value > 0);

  const eventsByType = stats?.events_by_type
    ? Object.entries(stats.events_by_type).map(([name, value]) => ({ name, value }))
    : [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-zinc-100">Dashboard</h1>
        <p className="mt-1 text-sm text-zinc-400">
          Fleet overview for {currentOrg?.name ?? 'your organization'}
        </p>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total Agents" value={agentsTotal} icon={Bot} color="#3b82f6" />
        <StatCard label="Online" value={agentsOnline} icon={Wifi} color="#22c55e" />
        <StatCard label="Offline" value={agentsOffline} icon={WifiOff} color="#6b7280" />
        <StatCard
          label="Errors Today"
          value={stats?.events_by_severity?.['error'] ?? 0}
          icon={AlertTriangle}
          color="#ef4444"
        />
      </div>

      {/* Quick stats */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-5">
          <div className="flex items-center gap-2 text-sm text-zinc-400">
            <CheckCircle2 className="h-4 w-4 text-green-500" />
            Issues Processed Today
          </div>
          <p className="mt-2 text-3xl font-bold text-zinc-100">
            {stats?.issues_processed_today ?? 0}
          </p>
        </div>
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-5">
          <div className="flex items-center gap-2 text-sm text-zinc-400">
            <GitPullRequest className="h-4 w-4 text-blue-500" />
            PRs Created Today
          </div>
          <p className="mt-2 text-3xl font-bold text-zinc-100">
            {stats?.prs_created_today ?? 0}
          </p>
        </div>
      </div>

      {/* Charts and feed */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Agent status distribution */}
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-5">
          <h2 className="mb-4 text-sm font-medium text-zinc-300">Agent Status Distribution</h2>
          {pieData.length > 0 ? (
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={90}
                  paddingAngle={4}
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={index} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#27272a',
                    border: '1px solid #3f3f46',
                    borderRadius: '8px',
                    color: '#fafafa',
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex h-[250px] items-center justify-center text-sm text-zinc-500">
              No agent data available
            </div>
          )}
          <div className="mt-2 flex justify-center gap-4">
            {pieData.map((d) => (
              <div key={d.name} className="flex items-center gap-2 text-xs text-zinc-400">
                <span
                  className="h-2.5 w-2.5 rounded-full"
                  style={{ backgroundColor: d.color }}
                />
                {d.name}: {d.value}
              </div>
            ))}
          </div>
        </div>

        {/* Events by type chart */}
        <div className="rounded-lg border border-zinc-700 bg-zinc-800 p-5">
          <h2 className="mb-4 text-sm font-medium text-zinc-300">Events by Type</h2>
          {eventsByType.length > 0 ? (
            <ResponsiveContainer width="100%" height={280}>
              <BarChart data={eventsByType} layout="vertical">
                <XAxis type="number" tick={{ fill: '#71717a', fontSize: 12 }} />
                <YAxis
                  type="category"
                  dataKey="name"
                  tick={{ fill: '#a1a1aa', fontSize: 11 }}
                  width={140}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#27272a',
                    border: '1px solid #3f3f46',
                    borderRadius: '8px',
                    color: '#fafafa',
                  }}
                />
                <Bar dataKey="value" fill="#3b82f6" radius={[0, 4, 4, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex h-[280px] items-center justify-center text-sm text-zinc-500">
              No event data available
            </div>
          )}
        </div>
      </div>

      {/* Recent events feed */}
      <div>
        <h2 className="mb-3 text-sm font-medium text-zinc-300">Recent Events</h2>
        <EventFeed events={realtimeEvents} compact maxHeight="400px" />
      </div>
    </div>
  );
}
