import { create } from 'zustand';
import type { AgentEvent, EventStats, EventFilters } from '@/types';
import * as eventsApi from '@/api/events';

interface EventState {
  events: AgentEvent[];
  realtimeEvents: AgentEvent[];
  stats: EventStats | null;
  loading: boolean;
  error: string | null;
  total: number;
  page: number;

  fetchEvents: (orgSlug: string, filters?: EventFilters) => Promise<void>;
  fetchStats: (orgSlug: string) => Promise<void>;
  addRealtimeEvent: (event: AgentEvent) => void;
  clearEvents: () => void;
}

const MAX_REALTIME_EVENTS = 100;

export const useEventStore = create<EventState>()((set) => ({
  events: [],
  realtimeEvents: [],
  stats: null,
  loading: false,
  error: null,
  total: 0,
  page: 1,

  fetchEvents: async (orgSlug, filters) => {
    set({ loading: true, error: null });
    try {
      const result = await eventsApi.listEvents(orgSlug, filters);
      set({
        events: result.data,
        total: result.total,
        page: result.page,
        loading: false,
      });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to fetch events',
        loading: false,
      });
    }
  },

  fetchStats: async (orgSlug) => {
    try {
      const stats = await eventsApi.getEventStats(orgSlug);
      set({ stats });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to fetch event stats',
      });
    }
  },

  addRealtimeEvent: (event) =>
    set((state) => ({
      realtimeEvents: [event, ...state.realtimeEvents].slice(0, MAX_REALTIME_EVENTS),
    })),

  clearEvents: () => set({ events: [], realtimeEvents: [] }),
}));
