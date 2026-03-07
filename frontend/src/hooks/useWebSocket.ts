import { useEffect, useRef } from 'react';
import ReconnectingWebSocket from 'reconnecting-websocket';
import { createEventStream, parseEventMessage } from '@/api/ws';
import { useAuthStore } from '@/store/auth-store';
import { useEventStore } from '@/store/event-store';

export function useWebSocket(orgSlug: string | undefined) {
  const wsRef = useRef<ReconnectingWebSocket | null>(null);
  const token = useAuthStore((s) => s.token);
  const addRealtimeEvent = useEventStore((s) => s.addRealtimeEvent);

  useEffect(() => {
    if (!orgSlug || !token) return;

    const ws = createEventStream(orgSlug, token);
    wsRef.current = ws;

    ws.onmessage = (evt) => {
      const event = parseEventMessage(evt.data);
      if (event) {
        addRealtimeEvent(event);
      }
    };

    ws.onerror = (err) => {
      console.error('WebSocket error:', err);
    };

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [orgSlug, token, addRealtimeEvent]);

  return wsRef;
}
