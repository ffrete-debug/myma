"use client";

import { useEffect, useState, useCallback } from 'react';
import { useTranslations } from 'next-intl';
import axios from 'axios';
import Cookies from 'js-cookie';
import { Users, RefreshCw, Loader2, Shield, Clock, MapPin, ChevronDown, ChevronUp } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Alert, AlertDescription } from '@/components/ui/alert';

interface OnlinePlayer {
  name: string;
  steam_id: string;
  character_id?: string;
  ip?: string;
  duration?: string;
}

interface PlayerListResponse {
  server_id: number;
  identifier: string;
  session_name: string;
  online: OnlinePlayer[];
  total_online: number;
  max_players: number;
}

interface DBPlayer {
  id: number;
  server_id: number;
  name: string;
  steam_id: string;
  character_id: string;
  status: string;
  ip: string;
  joined_at: string;
  created_at: string;
  updated_at: string;
}

const getAuthHeaders = () => {
  const token = Cookies.get('auth-token');
  if (!token) throw new Error('Authentication token not found');
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
};

export default function PlayersPage() {
  const t = useTranslations('players');
  const tCommon = useTranslations('common');

  const [serverId, setServerId] = useState('');
  const [players, setPlayers] = useState<PlayerListResponse | null>(null);
  const [history, setHistory] = useState<DBPlayer[]>([]);
  const [loading, setLoading] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [historyExpanded, setHistoryExpanded] = useState(false);

  const fetchPlayers = useCallback(async () => {
    if (!serverId.trim()) return;
    setLoading(true);
    setError('');
    try {
      const response = await axios.get<PlayerListResponse>(
        `/api/servers/${serverId}/players`,
        { headers: getAuthHeaders() }
      );
      setPlayers(response.data);
      setSuccess(t('fetched'));
    } catch {
      setError(t('fetchFailed'));
    } finally {
      setLoading(false);
    }
  }, [serverId, t]);

  const fetchHistory = useCallback(async () => {
    if (!serverId.trim()) return;
    setHistoryLoading(true);
    try {
      const response = await axios.get<DBPlayer[]>(
        `/api/servers/${serverId}/players/history`,
        { headers: getAuthHeaders() }
      );
      setHistory(response.data);
    } catch {
      // silently ignore history fetch errors
    } finally {
      setHistoryLoading(false);
    }
  }, [serverId]);

  useEffect(() => {
    if (serverId.trim()) {
      fetchPlayers();
      fetchHistory();
    }
  }, [serverId, fetchPlayers, fetchHistory]);

  const handleRefresh = () => {
    fetchPlayers();
    fetchHistory();
  };

  return (
    <div className="w-full max-w-none py-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('title')}</h1>
          <p className="text-muted-foreground mt-1">{t('description')}</p>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      {success && (
        <Alert>
          <AlertDescription>{success}</AlertDescription>
        </Alert>
      )}

      {/* Server ID Input */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{t('selectServer')}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-3 items-end">
            <div className="flex-1">
              <label className="text-sm font-medium text-foreground mb-1 block">
                {t('serverIdLabel')}
              </label>
              <input
                type="text"
                value={serverId}
                onChange={(e) => setServerId(e.target.value)}
                placeholder={t('serverIdPlaceholder')}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>
            <Button onClick={handleRefresh} disabled={loading || !serverId.trim()}>
              {loading ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : <RefreshCw className="h-4 w-4 mr-2" />}
              {tCommon('refresh')}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Player Summary */}
      {players && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <Card>
              <CardContent className="p-4 flex items-center gap-3">
                <Users className="h-8 w-8 text-primary" />
                <div>
                  <p className="text-xs text-muted-foreground">{t('onlineNow')}</p>
                  <p className="text-2xl font-bold">{players.total_online}</p>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4 flex items-center gap-3">
                <Shield className="h-8 w-8 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">{t('maxPlayers')}</p>
                  <p className="text-2xl font-bold">{players.max_players}</p>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4 flex items-center gap-3">
                <MapPin className="h-8 w-8 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">{t('server')}</p>
                  <p className="text-sm font-bold truncate">{players.identifier}</p>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4 flex items-center gap-3">
                <Clock className="h-8 w-8 text-muted-foreground" />
                <div>
                  <p className="text-xs text-muted-foreground">{t('session')}</p>
                  <p className="text-sm font-bold truncate">{players.session_name}</p>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Online Players Table */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">{t('onlinePlayers')}</CardTitle>
            </CardHeader>
            <CardContent>
              {players.online.length === 0 ? (
                <p className="text-muted-foreground text-sm">{t('noOnlinePlayers')}</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('playerName')}</TableHead>
                      <TableHead>Steam ID</TableHead>
                      <TableHead>Character ID</TableHead>
                      <TableHead>{t('status')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {players.online.map((p, i) => (
                      <TableRow key={`${p.steam_id}-${i}`}>
                        <TableCell className="font-medium">{p.name || '—'}</TableCell>
                        <TableCell className="font-mono text-xs">{p.steam_id || '—'}</TableCell>
                        <TableCell className="font-mono text-xs">{p.character_id || '—'}</TableCell>
                        <TableCell>
                          <Badge variant="default">{t('online')}</Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>

          {/* Player History */}
          <Card>
            <CardHeader
              className="cursor-pointer select-none"
              onClick={() => setHistoryExpanded(!historyExpanded)}
            >
              <CardTitle className="text-lg flex items-center justify-between">
                <span>{t('playerHistory')}</span>
                {historyExpanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
              </CardTitle>
            </CardHeader>
            {historyExpanded && (
              <CardContent>
                {historyLoading ? (
                  <div className="flex items-center gap-2 text-muted-foreground text-sm">
                    <Loader2 className="h-4 w-4 animate-spin" /> {tCommon('loading')}
                  </div>
                ) : history.length === 0 ? (
                  <p className="text-muted-foreground text-sm">{t('noHistory')}</p>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t('playerName')}</TableHead>
                        <TableHead>Steam ID</TableHead>
                        <TableHead>{t('status')}</TableHead>
                        <TableHead>{t('joinedAt')}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {history.map((p) => (
                        <TableRow key={p.id}>
                          <TableCell className="font-medium">{p.name}</TableCell>
                          <TableCell className="font-mono text-xs">{p.steam_id}</TableCell>
                          <TableCell>
                            <Badge variant={p.status === 'online' ? 'default' : 'secondary'}>
                              {p.status}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-xs text-muted-foreground">
                            {new Date(p.joined_at).toLocaleString()}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            )}
          </Card>
        </>
      )}
    </div>
  );
}