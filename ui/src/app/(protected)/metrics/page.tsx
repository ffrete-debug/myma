"use client";

import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { RefreshCw, Users, Server as ServerIcon, Cpu } from 'lucide-react';
import api from '@/lib/axios';
import { UsageMeter, severityFor } from '@/components/metrics/UsageMeter';

/** Matches ServerMetrics in server/service/metrics. */
interface ServerMetrics {
  server_id: number;
  server_identifier: string;
  session_name: string;
  status: string;
  cpu_percent: number;
  cpu_cores: number;
  memory_usage_mb: number;
  memory_limit_mb: number;
  memory_percent: number;
  network_rx_mb: number;
  network_tx_mb: number;
  players_online: number;
  max_players: number;
  sampled_at: number;
  error?: string;
}

/** The backend sends -1 when the player count could not be read at all, which
 *  is meaningfully different from a server that answered with zero players. */
const PLAYERS_UNKNOWN = -1;

const REFRESH_MS = 10_000;

function formatGB(mb: number): string {
  if (!mb) return '0 GB';
  return `${(mb / 1024).toFixed(1)} GB`;
}

export default function MetricsPage() {
  const t = useTranslations('metrics');
  const [metrics, setMetrics] = useState<ServerMetrics[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [lastUpdated, setLastUpdated] = useState<number | null>(null);

  // Held in a ref so the polling effect does not need `loading` in its deps,
  // which would tear down and recreate the interval on every tick.
  const inFlight = useRef(false);

  const load = useCallback(async () => {
    if (inFlight.current) return;
    inFlight.current = true;
    try {
      const res = await api.get<{ data?: ServerMetrics[] }>('/api/metrics');
      setMetrics(res.data.data ?? []);
      setError('');
      setLastUpdated(Date.now());
    } catch {
      setError(t('loadFailed'));
    } finally {
      inFlight.current = false;
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    load();
    const id = setInterval(load, REFRESH_MS);
    return () => clearInterval(id);
  }, [load]);

  const running = metrics.filter((m) => m.status === 'running');

  // Only count servers that actually reported a number; summing -1 sentinels
  // would silently understate the total.
  const knownPlayers = running.filter((m) => m.players_online !== PLAYERS_UNKNOWN);
  const totalPlayers = knownPlayers.reduce((sum, m) => sum + m.players_online, 0);
  const totalCapacity = running.reduce((sum, m) => sum + (m.max_players || 0), 0);
  const anyPlayerCountMissing = running.length !== knownPlayers.length;

  const avgCpu = running.length
    ? running.reduce((sum, m) => sum + m.cpu_percent, 0) / running.length
    : 0;
  const totalMemMB = running.reduce((sum, m) => sum + m.memory_usage_mb, 0);

  const severityWord = (percent: number) => t(`severity.${severityFor(percent)}`);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-foreground">{t('title')}</h1>
          <p className="text-sm text-muted-foreground">{t('subtitle')}</p>
        </div>
        <div className="flex items-center gap-3">
          {lastUpdated && (
            <span className="text-xs text-muted-foreground">
              {t('updatedAt', { time: new Date(lastUpdated).toLocaleTimeString() })}
            </span>
          )}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={load}
            aria-label={t('refresh')}
            title={t('refresh')}
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {error && (
        <div className="rounded-md border border-red-900/50 bg-red-950/30 px-4 py-3 text-sm text-red-300">
          {error}
        </div>
      )}

      {/* Hero figure — exactly one per view. Players online is the number an
          operator actually leads with; CPU and memory are supporting detail. */}
      <Card>
        <CardContent className="flex flex-wrap items-end justify-between gap-6 pt-6">
          <div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Users className="h-4 w-4" />
              {t('playersOnline')}
            </div>
            <div className="mt-1 text-5xl font-semibold leading-none text-foreground">
              {loading && !metrics.length ? '—' : totalPlayers}
              {totalCapacity > 0 && (
                <span className="ml-2 text-xl font-normal text-muted-foreground">
                  / {totalCapacity}
                </span>
              )}
            </div>
            {anyPlayerCountMissing && (
              <p className="mt-2 text-xs text-muted-foreground">{t('partialPlayerCount')}</p>
            )}
          </div>

          {/* Supporting stat tiles */}
          <div className="flex flex-wrap gap-8">
            <div>
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <ServerIcon className="h-4 w-4" />
                {t('serversRunning')}
              </div>
              <div className="mt-1 text-2xl font-semibold text-foreground">
                {running.length}
                <span className="ml-1 text-base font-normal text-muted-foreground">
                  / {metrics.length}
                </span>
              </div>
            </div>
            <div>
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Cpu className="h-4 w-4" />
                {t('averageCpu')}
              </div>
              <div className="mt-1 text-2xl font-semibold tabular-nums text-foreground">
                {avgCpu.toFixed(1)}%
              </div>
            </div>
            <div>
              <div className="text-sm text-muted-foreground">{t('totalMemory')}</div>
              <div className="mt-1 text-2xl font-semibold tabular-nums text-foreground">
                {formatGB(totalMemMB)}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {!loading && metrics.length === 0 && (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            {t('noServers')}
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {metrics.map((m) => {
          const isRunning = m.status === 'running';
          const playersKnown = m.players_online !== PLAYERS_UNKNOWN;

          return (
            <Card key={m.server_id}>
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between gap-2">
                  <CardTitle className="text-base">
                    {m.session_name || m.server_identifier}
                  </CardTitle>
                  <Badge variant={isRunning ? 'default' : 'secondary'}>
                    {t(`status.${isRunning ? 'running' : 'stopped'}`)}
                  </Badge>
                </div>
                {m.error && (
                  <p className="text-xs text-muted-foreground">{t(`errors.${m.error}`)}</p>
                )}
              </CardHeader>

              <CardContent className="space-y-4">
                <UsageMeter
                  label={t('cpu')}
                  percent={m.cpu_percent}
                  detail={m.cpu_cores ? t('acrossCores', { count: m.cpu_cores }) : undefined}
                  severityLabel={severityWord(m.cpu_percent)}
                  unavailable={!isRunning}
                  unavailableLabel={t('notRunning')}
                />

                <UsageMeter
                  label={t('memory')}
                  percent={m.memory_percent}
                  detail={
                    m.memory_limit_mb
                      ? `${formatGB(m.memory_usage_mb)} / ${formatGB(m.memory_limit_mb)}`
                      : formatGB(m.memory_usage_mb)
                  }
                  severityLabel={severityWord(m.memory_percent)}
                  unavailable={!isRunning}
                  unavailableLabel={t('notRunning')}
                />

                <div className="flex items-center justify-between border-t border-border pt-3">
                  <span className="flex items-center gap-2 text-xs text-muted-foreground">
                    <Users className="h-3.5 w-3.5" />
                    {t('players')}
                  </span>
                  <span className="text-sm font-semibold tabular-nums text-foreground">
                    {!isRunning || !playersKnown
                      ? t('unknown')
                      : `${m.players_online}${m.max_players ? ` / ${m.max_players}` : ''}`}
                  </span>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
