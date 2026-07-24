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
  const logEndRef = useRef<HTMLDivElement>(null);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  const fetchLogs = useCallback(async () => {
    try {
      const token = Cookies.get('auth-token');
      const response = await fetch(`/api/servers/${serverId}/logs?tail=${tailLines}`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      if (!response.ok) throw new Error('Failed to fetch logs');
      const data = await response.json();
      setLogs(data.data || '');
      setError('');
    } catch {
      setError('Failed to fetch logs');
    } finally {
      setLoading(false);
    }
  }, [serverId, tailLines]);

  const fetchServer = useCallback(async () => {
    try {
      const srv = await serversActions.getServer(serverId);
      setServer(srv);
    } catch {
      setError('Failed to get server info');
    }
  }, [serverId]);

  useEffect(() => {
    fetchServer();
  }, [fetchServer]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  useEffect(() => {
    if (autoRefresh) {
      intervalRef.current = setInterval(fetchLogs, 3000);
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [autoRefresh, fetchLogs]);

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
              <CardTitle className="text-sm font-medium text-gray-700">
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
              <Loader2 className="w-6 h-6 animate-spin text-blue-600" />
            </div>
          ) : error ? (
            <div className="text-center py-12 text-red-500">{error}</div>
          ) : (
            <div className="bg-gray-900 text-green-400 rounded-lg p-4 font-mono text-xs leading-relaxed whitespace-pre-wrap overflow-auto max-h-[70vh]">
              {logs || <span className="text-gray-500">{t('noLogs')}</span>}
              <div ref={logEndRef} />
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
