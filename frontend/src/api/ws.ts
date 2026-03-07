import ReconnectingWebSocket from 'reconnecting-websocket';
import type { AgentEvent } from '@/types';

export function createEventStream(
  orgSlug: string,
  token: string,
): ReconnectingWebSocket {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = import.meta.env.VITE_WS_HOST || window.location.host;
  const url = `${protocol}//${host}/ws/orgs/${orgSlug}/events?token=${encodeURIComponent(token)}`;

  const ws = new ReconnectingWebSocket(url, [], {
    maxRetries: 10,
    connectionTimeout: 5000,
    maxReconnectionDelay: 30000,
  });

  return ws;
}

export function parseEventMessage(data: string): AgentEvent | null {
  try {
    return JSON.parse(data) as AgentEvent;
  } catch {
    console.error('Failed to parse WebSocket message:', data);
    return null;
  }
}
