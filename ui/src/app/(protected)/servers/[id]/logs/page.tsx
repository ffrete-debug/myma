"use client";

import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, ArrowLeft, RefreshCw, Pause, Play, Trash2, Map } from 'lucide-react';
import { serversActions } from '@/stores/servers';
import { Server } from '@/stores/servers';
import Cookies from 'js-cookie';
import { buildWebSocketUrl } from '@/lib/ws-url';

const MAX_LOG_CHARS = 500_000;
const MAX_LOG_LINES = 10_000;

export default function ServerLogsPage() {
  const t = useTranslations('servers');
  const tCommon = useTranslations('common');
  const params = useParams();
  const router = useRouter();
  const serverId = params.id as string;

  const [server, setServer] = useState<Server | null>(null);
  const [logs, setLogs] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [tailLines, setTailLines] = useState(200);
  const [error, setError] = useState('');
  const [wsConnected, setWsConnected] = useState(false);
  const logEndRef = useRef<HTMLDivElement>(null);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const isClosingRef = useRef(false);

  const fetchLogs = useCallback(async () => {
    try {
      const token = Cookies.get('auth-token');
      const response = await fetch(`/api/servers/${serverId}/logs?tail=${tailLines}`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      if (!response.ok) throw new Error(`Log request failed with status ${response.status}`);
      const data = await response.json();
      setLogs(data.data || '');
      setError('');
    } catch {
      setError(t('fetchLogsFailed'));
    } finally {
      setLoading(false);
    }
  }, [serverId, tailLines, t]);

  // Keep the polling fallback stable: the interval reads the latest fetcher
  // through a ref instead of being torn down on every tail-size change.
  const fetchLogsRef = useRef(fetchLogs);
  fetchLogsRef.current = fetchLogs;

  const appendLogLine = useCallback((line: string) => {
    setLogs(prev => {
      const next = prev ? prev + '\n' + line : line;
      // Cap log size to prevent unbounded memory growth.
      if (next.length > MAX_LOG_CHARS) {
        const cut = next.length - MAX_LOG_CHARS;
        const idx = next.indexOf('\n', cut);
        return idx >= 0 ? next.slice(idx + 1) : next;
      }
      const lines = next.split('\n');
      if (lines.length > MAX_LOG_LINES) {
        return lines.slice(lines.length - MAX_LOG_LINES).join('\n');
      }
      return next;
    });
  }, []);

  const connectWS = useCallback(() => {
    if (!autoRefresh || isClosingRef.current) return;
    const token = Cookies.get('auth-token');
    if (!token) return;

    // The log socket is served by the Go backend, not the Next server, so the
    // URL has to be derived from the API base like the other WS call sites.
    const url = buildWebSocketUrl(`/api/ws/logs/${serverId}`, token);

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      // Only now is the socket really live; the polling fallback keys off this.
      setWsConnected(true);
      reconnectAttemptsRef.current = 0;
    };

    ws.onmessage = (event) => {
      if (typeof event.data === 'string') {
        appendLogLine(event.data);
        reconnectAttemptsRef.current = 0;
      }
    };

    ws.onclose = () => {
      wsRef.current = null;
      setWsConnected(false);
      // Never reconnect a socket we closed ourselves during cleanup.
      if (isClosingRef.current || !autoRefresh) return;
      // Exponential backoff capped at 30s.
      const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30_000);
      reconnectAttemptsRef.current++;
      reconnectTimerRef.current = setTimeout(connectWS, delay);
    };

    ws.onerror = () => {
      setWsConnected(false);
      ws.close();
    };
  }, [serverId, autoRefresh, appendLogLine]);

  const fetchServer = useCallback(async () => {
    try {
      const srv = await serversActions.getServer(serverId);
      setServer(srv);
    } catch {
      setError(t('loadServerInfoFailed'));
    }
  }, [serverId, t]);

  useEffect(() => {
    fetchServer();
  }, [fetchServer]);

  // Initial log load via HTTP fetch.
  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  // WebSocket for live tailing (after initial load).
  useEffect(() => {
    if (loading || !autoRefresh) return;
    isClosingRef.current = false;
    connectWS();
    return () => {
      // Flag the close as intentional *before* closing, otherwise the onclose
      // handler fires after cleanup and schedules a reconnect for a dead page.
      isClosingRef.current = true;
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      setWsConnected(false);
    };
  }, [connectWS, loading, autoRefresh]);

  // Polling fallback for initial load and reconnection.
  useEffect(() => {
    if (!autoRefresh || wsConnected) return;
    intervalRef.current = setInterval(() => {
      void fetchLogsRef.current();
    }, 3000);
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [autoRefresh, wsConnected]);

  useEffect(() => {
    if (logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs]);

  const clearLogs = () => setLogs('');

  return (
    <div className="w-full max-w-none py-8 space-y-4">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" onClick={() => router.push('/servers')}>
          <ArrowLeft className="h-4 w-4 mr-1" />
          {tCommon('back')}
        </Button>
        <h1 className="text-2xl font-bold text-foreground">
          {t('serverLogs')}{server ? ` - ${server.session_name || server.identifier}` : ''}
        </h1>
        {server && (
          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
            server.status === 'running' ? 'badge-success' :
            server.status === 'stopped' ? 'badge-destructive' :
            'badge-warning'
          }`}>
            {t(`card.${server.status}`)}
          </span>
        )}
        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
          wsConnected ? 'badge-success' : 'badge-warning'
        }`}>
          {wsConnected ? 'Live' : 'Polling'}
        </span>
      </div>

      {server && (
        <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground bg-card rounded-lg px-4 py-2">
          <div className="flex items-center gap-1">
            <Map className="h-4 w-4" />
            <span className="font-medium">{t('card.map')}:</span>
            <span>{server.map}</span>
          </div>
          <div className="flex items-center gap-1">
            <span className="font-medium">{t('card.serverPort')}:</span>
            <span>{server.port}</span>
          </div>
          <div className="flex items-center gap-1">
            <span className="font-medium">{t('edit.queryPort')}:</span>
            <span>{server.query_port}</span>
          </div>
          <div className="flex items-center gap-1">
            <span className="font-medium">{t('card.maxPlayers')}:</span>
            <span>{server.max_players}</span>
          </div>
        </div>
      )}

      <Card>
        <CardHeader className="py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {t('serverLogs')}
              </CardTitle>
              <select
                value={tailLines}
                onChange={(e) => setTailLines(Number(e.target.value))}
                className="text-xs border rounded px-2 py-1"
              >
                <option value={50}>50</option>
                <option value={100}>100</option>
                <option value={200}>200</option>
                <option value={500}>500</option>
                <option value={1000}>1000</option>
              </select>
            </div>
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="sm" onClick={fetchLogs} title={tCommon('refresh')}>
                <RefreshCw className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setAutoRefresh(!autoRefresh)}
                title={autoRefresh ? tCommon('autoRefreshOff') : tCommon('autoRefreshOn')}
              >
                {autoRefresh ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
              </Button>
              <Button variant="ghost" size="sm" onClick={clearLogs} title="Clear">
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="w-6 h-6 animate-spin text-blue-400" />
            </div>
          ) : error ? (
            <div className="text-center py-12 text-red-500">{error}</div>
          ) : (
            <div className="bg-gray-900 text-green-400 rounded-lg p-4 font-mono text-xs leading-relaxed whitespace-pre-wrap overflow-auto max-h-[70vh]">
              {logs || <span className="text-muted-foreground/60">{t('noLogs')}</span>}
              <div ref={logEndRef} />
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
