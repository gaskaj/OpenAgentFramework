import { Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { useAuthStore } from '@/store/auth-store';
import { AppShell } from '@/components/layout/AppShell';
import { LoginPage } from '@/pages/LoginPage';
import { RegisterPage } from '@/pages/RegisterPage';
import { DashboardPage } from '@/pages/DashboardPage';
import { AgentListPage } from '@/pages/AgentListPage';
import { AgentDetailPage } from '@/pages/AgentDetailPage';
import { EventFeedPage } from '@/pages/EventFeedPage';
import { OrgSettingsPage } from '@/pages/OrgSettingsPage';
import { ApiKeysPage } from '@/pages/ApiKeysPage';
import { AuditLogPage } from '@/pages/AuditLogPage';
import { InviteAcceptPage } from '@/pages/InviteAcceptPage';

function ProtectedRoute() {
  const token = useAuthStore((s) => s.token);
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/invite/:token" element={<InviteAcceptPage />} />

      <Route element={<ProtectedRoute />}>
        <Route element={<AppShell />}>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/agents" element={<AgentListPage />} />
          <Route path="/agents/:agentId" element={<AgentDetailPage />} />
          <Route path="/events" element={<EventFeedPage />} />
          <Route path="/settings" element={<OrgSettingsPage />} />
          <Route path="/settings/api-keys" element={<ApiKeysPage />} />
          <Route path="/audit" element={<AuditLogPage />} />
        </Route>
      </Route>
    </Routes>
  );
}
