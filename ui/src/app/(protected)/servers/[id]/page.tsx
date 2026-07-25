"use client";

import { useEffect, useState, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loader2, ArrowLeft, Map, Wifi, Lock, Users, RefreshCw, Play, Square, Edit, FileText } from 'lucide-react';
import { serversActions } from '@/stores/servers';
import { Server } from '@/stores/servers';
import { RCONConsole } from '@/components/servers/RCONConsole';

export default function ServerDetailPage() {
  const t = useTranslations('servers');
  const tCommon = useTranslations('common');
  const params = useParams();
  const router = useRouter();
  const serverId = params.id as string;

  const [server, setServer] = useState<Server | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionInProgress, setActionInProgress] = useState<string>('');

  const fetchServer = useCallback(async () => {
    try {
      const srv = await serversActions.getServer(serverId);
      setServer(srv);
      setError('');
    } catch {
      setError(t('loadServerInfoFailed'));
    } finally {
      setLoading(false);
    }
  }, [serverId, t]);

  useEffect(() => {
    fetchServer();
  }, [fetchServer]);

  const handleStart = async () => {
    setActionInProgress('start');
    try { await serversActions.startServer(serverId); } finally { setActionInProgress(''); }
  };
  const handleStop = async () => {
    setActionInProgress('stop');
    try { await serversActions.stopServer(serverId); } finally { setActionInProgress(''); }
  };
  const handleRestart = async () => {
    setActionInProgress('restart');
    try { await serversActions.restartServer(serverId); } finally { setActionInProgress(''); }
  };

  const getStatusVariant = (s: Server['status']): 'default' | 'destructive' | 'secondary' | 'outline' => {
    switch (s) {
      case 'running': return 'default';
      case 'stopped': return 'destructive';
      case 'starting': case 'stopping': case 'restarting': return 'secondary';
      default: return 'outline';
    }
  };

  const getMapDisplayName = (mapName: string) => {
    const mapKey = `edit.maps.${mapName}`;
    const translated = t(mapKey);
    return translated !== mapKey ? translated : mapName;
  };

  if (loading) {
    return (
      <div className="w-full max-w-none py-8">
        <div className="flex justify-center items-center h-64">
          <Loader2 className="h-8 w-8 animate-spin" />
          <span className="ml-2">{tCommon('loading')}</span>
        </div>
      </div>
    );
  }

  if (error && !server) {
    return (
      <div className="w-full max-w-none py-8 space-y-4">
        <Button variant="outline" onClick={() => router.push('/servers')}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          {tCommon('back')}
        </Button>
        <div className="text-red-600">{error}</div>
      </div>
    );
  }

  if (!server) return null;

  const isRunning = server.status === 'running';

  return (
    <div className="w-full max-w-none py-8 space-y-4">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" onClick={() => router.push('/servers')}>
          <ArrowLeft className="h-4 w-4 mr-1" />
          {tCommon('back')}
        </Button>
        <h1 className="text-2xl font-bold text-foreground">{server.session_name}</h1>
        <Badge variant={getStatusVariant(server.status)} className="text-xs">
          {t(`card.${server.status}`)}
        </Badge>
        <div className="ml-auto flex items-center gap-1">
          {isRunning ? (
            <Button variant="outline" size="sm" onClick={handleStop} disabled={actionInProgress !== ''}>
              {actionInProgress === 'stop' ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <Square className="h-4 w-4 mr-1" />}
              {t('stopServer')}
            </Button>
          ) : (
            <Button variant="outline" size="sm" onClick={handleStart} disabled={actionInProgress !== ''}>
              {actionInProgress === 'start' ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <Play className="h-4 w-4 mr-1" />}
              {t('startServer')}
            </Button>
          )}
          <Button variant="outline" size="sm" onClick={handleRestart} disabled={actionInProgress !== '' || !isRunning}>
            {actionInProgress === 'restart' ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <RefreshCw className="h-4 w-4 mr-1" />}
            {t('restartServer')}
          </Button>
          <Button variant="outline" size="sm" onClick={() => router.push(`/servers/${server.id}/edit`)}>
            <Edit className="h-4 w-4 mr-1" />
            {t('editServer')}
          </Button>
          <Button variant="outline" size="sm" onClick={() => router.push(`/servers/${server.id}/logs`)}>
            <FileText className="h-4 w-4 mr-1" />
            {t('serverLogs')}
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader className="py-3">
          <CardTitle className="text-sm font-medium text-gray-700">{t('serverConfig')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div className="flex items-center gap-2">
              <Wifi className="h-4 w-4 text-primary" />
              <div>
                <div className="text-xs text-muted-foreground">{t('card.gamePort')} / {t('card.queryPort')} / RCON</div>
                <div className="font-mono font-semibold">{server.port} / {server.query_port} / {server.rcon_port}</div>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Map className="h-4 w-4 text-emerald-600" />
              <div>
                <div className="text-xs text-muted-foreground">{t('card.map')}</div>
                <div className="font-medium">{getMapDisplayName(server.map)}</div>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Users className="h-4 w-4 text-primary" />
              <div>
                <div className="text-xs text-muted-foreground">{t('card.maxPlayers')}</div>
                <div className="font-medium">{server.max_players}</div>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Lock className="h-4 w-4 text-amber-600" />
              <div>
                <div className="text-xs text-muted-foreground">{t('card.adminPassword')}</div>
                <div className="font-mono">{server.admin_password ? '••••••••' : '—'}</div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <RCONConsole serverId={serverId} serverStatus={server.status} />
    </div>
  );
}
