"use client";

import { useEffect, useRef, useCallback } from 'react';
import Cookies from 'js-cookie';

type MessageHandler = (data: { type: string; server_id: number; data: unknown }) => void;

export function useWebSocket(
  serverId: string,
  onMessage: MessageHandler,
  enabled: boolean = true,
) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const onMessageRef = useRef(onMessage);
  const enabledRef = useRef(enabled);

  // Keep the ref current so the reconnect loop always uses the latest callback.
  onMessageRef.current = onMessage;
  enabledRef.current = enabled;

  const connect = useCallback(() => {
    if (!enabledRef.current || !serverId) return;

    const token = Cookies.get('auth-token');
    if (!token) return;

    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const url = `${proto}://${window.location.host}/api/ws/updates/${serverId}?token=${encodeURIComponent(token)}`;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data);
        onMessageRef.current(parsed);
      } catch {
        // ignore parse errors
      }
    };

    ws.onclose = () => {
      // Auto-reconnect after 3 seconds if still enabled.
      if (enabledRef.current) {
        reconnectTimerRef.current = setTimeout(connect, 3000);
      }
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [serverId]);

  useEffect(() => {
    connect();
    return () => {
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [connect]);
}
