import { useEffect, useCallback } from 'react';
import { useEventStore } from '@/store/event-store';
import { useAuthStore } from '@/store/auth-store';
import type { EventFilters } from '@/types';

export function useEvents(filters?: EventFilters) {
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const { events, loading, error, total, page, fetchEvents } = useEventStore();

  const refresh = useCallback(() => {
    if (currentOrg) {
      fetchEvents(currentOrg.slug, filters);
    }
  }, [currentOrg, filters, fetchEvents]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { events, loading, error, total, page, refresh };
}

export function useEventStats() {
  const currentOrg = useAuthStore((s) => s.currentOrg);
  const { stats, fetchStats } = useEventStore();

  const refresh = useCallback(() => {
    if (currentOrg) {
      fetchStats(currentOrg.slug);
    }
  }, [currentOrg, fetchStats]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { stats, refresh };
}
