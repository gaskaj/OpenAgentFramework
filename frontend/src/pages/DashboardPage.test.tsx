import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DashboardPage } from './DashboardPage';
import { useAuthStore } from '@/store/auth-store';
import { TEST_USER, TEST_ORG } from '@/test/mocks';

vi.mock('@/hooks/useEvents', () => ({
  useEventStats: () => ({
    stats: null,
    refresh: vi.fn(),
  }),
}));

vi.mock('@/hooks/useWebSocket', () => ({
  useWebSocket: vi.fn(),
}));

vi.mock('recharts', () => ({
  PieChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Pie: () => null,
  Cell: () => null,
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  BarChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Bar: () => null,
  XAxis: () => null,
  YAxis: () => null,
  Tooltip: () => null,
}));

describe('DashboardPage', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: TEST_USER,
      token: 'test-token',
      refreshToken: 'test-refresh',
      currentOrg: TEST_ORG,
    });
  });

  it('renders without crashing', () => {
    expect(() => render(<DashboardPage />)).not.toThrow();
  });

  it('renders dashboard heading', () => {
    render(<DashboardPage />);
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });

  it('renders all stat cards', () => {
    render(<DashboardPage />);
    expect(screen.getByText('Total Agents')).toBeInTheDocument();
    expect(screen.getByText('Online')).toBeInTheDocument();
    expect(screen.getByText('Offline')).toBeInTheDocument();
    expect(screen.getByText('Errors Today')).toBeInTheDocument();
  });

  it('renders with null stats without crashing', () => {
    render(<DashboardPage />);
    const zeroValues = screen.getAllByText('0');
    expect(zeroValues.length).toBeGreaterThanOrEqual(4);
  });

  it('displays org name in subtitle', () => {
    render(<DashboardPage />);
    expect(screen.getByText('Fleet overview for Test Org')).toBeInTheDocument();
  });

  it('renders without currentOrg', () => {
    useAuthStore.setState({ currentOrg: null });
    expect(() => render(<DashboardPage />)).not.toThrow();
    expect(screen.getByText('Fleet overview for your organization')).toBeInTheDocument();
  });
});
