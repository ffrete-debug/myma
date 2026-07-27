"use client";

import { useEffect, useRef, useState, useCallback } from 'react';
import Cookies from 'js-cookie';
import { buildWebSocketUrl } from '@/lib/ws-url';

export type WebSocketMessage = { type: string; server_id: number; data: Record<string, unknown> };

type MessageHandler = (data: WebSocketMessage) => void;

/**
 * `idle`        - disabled, or no server selected.
 * `connecting`  - first connection attempt in flight.
 * `open`        - socket is live.
 * `reconnecting`- socket dropped, a backed-off retry is scheduled.
 * `disconnected`- terminal: retries exhausted (or no auth token). No further attempts.
 */
export type WebSocketStatus = 'idle' | 'connecting' | 'open' | 'reconnecting' | 'disconnected';

const INITIAL_RECONNECT_DELAY_MS = 1_000;
const MAX_RECONNECT_DELAY_MS = 30_000;
const MAX_RECONNECT_ATTEMPTS = 8;

// Scheme selection lives in the shared helper so every WS call site derives
// ws:// vs wss:// the same way. See src/lib/ws-url.ts.

export function useWebSocket(
  serverId: string,
  onMessage: MessageHandler,
  enabled: boolean = true,
) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  // Set by the effect cleanup *before* close(). close() settles asynchronously, so
  // onclose runs after teardown and would otherwise queue a reconnect for a dead hook.
  const intentionalCloseRef = useRef(false);
  const onMessageRef = useRef(onMessage);
  const [status, setStatus] = useState<WebSocketStatus>('idle');

  // Keep the ref current so the reconnect loop always uses the latest callback.
  onMessageRef.current = onMessage;

  const connect = useCallback(() => {
    if (!serverId) {
      setStatus('idle');
      return;
    }

    const token = Cookies.get('auth-token');
    if (!token) {
      setStatus('disconnected');
      return;
    }

    // SECURITY / FOLLOW-UP: the JWT is sent as a `?token=` query parameter because that
    // is where the backend currently reads it. Query strings leak into reverse-proxy and
    // CDN access logs, browser history and error reporting, so this is a real credential
    // exposure. Changing it is a backend contract change, hence left alone here. The
    // proper fix is either a `Sec-WebSocket-Protocol` subprotocol carrying the token, or
    // a short-lived single-use ticket minted over HTTPS and redeemed on connect.
    const url = buildWebSocketUrl(`/api/ws/updates/${serverId}`, token);

    setStatus(reconnectAttemptsRef.current === 0 ? 'connecting' : 'reconnecting');

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      reconnectAttemptsRef.current = 0;
      setStatus('open');
    };

    ws.onmessage = (event: MessageEvent<string>) => {
      try {
        onMessageRef.current(JSON.parse(event.data) as WebSocketMessage);
      } catch {
        // ignore parse errors
      }
    };

    ws.onclose = () => {
      // `wsRef.current !== ws` means this socket was superseded (serverId changed, or the
      // effect re-ran) or torn down; either way it must not resurrect itself.
      const isCurrentSocket = wsRef.current === ws;
      if (isCurrentSocket) wsRef.current = null;
      if (intentionalCloseRef.current || !isCurrentSocket) return;

      if (reconnectAttemptsRef.current >= MAX_RECONNECT_ATTEMPTS) {
        setStatus('disconnected');
        return;
      }

      // Exponential backoff with a ceiling: 1s, 2s, 4s ... capped at 30s.
      const delay = Math.min(
        INITIAL_RECONNECT_DELAY_MS * 2 ** reconnectAttemptsRef.current,
        MAX_RECONNECT_DELAY_MS,
      );
      reconnectAttemptsRef.current += 1;
      setStatus('reconnecting');
      reconnectTimerRef.current = setTimeout(connect, delay);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [serverId]);

  useEffect(() => {
    if (!enabled) {
      setStatus('idle');
      return;
    }

    intentionalCloseRef.current = false;
    reconnectAttemptsRef.current = 0;
    connect();

    return () => {
      // Order matters: flag the teardown first, then drop the timer, then close.
      intentionalCloseRef.current = true;
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      const ws = wsRef.current;
      wsRef.current = null;
      ws?.close();
    };
  }, [connect, enabled]);

  return { status };
}
