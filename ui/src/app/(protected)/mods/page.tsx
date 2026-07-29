"use client";

import { useCallback, useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  Search, Plus, Trash2, ArrowUp, ArrowDown, Loader2, ExternalLink, Eye, EyeOff,
} from 'lucide-react';
import api from '@/lib/axios';
import { useServers, serversActions } from '@/stores/servers';

interface WorkshopItem {
  workshop_id: string;
  title: string;
  description: string;
  preview_url: string;
  file_size_bytes: number;
  subscriptions: number;
  time_updated: number;
}

interface ServerMod {
  ID: number;
  server_id: number;
  workshop_id: string;
  name: string;
  preview_url: string;
  position: number;
  enabled: boolean;
}

function errorMessage(e: unknown, fallback: string): string {
  const err = e as { response?: { data?: { error?: string } } };
  return err.response?.data?.error || fallback;
}

const WORKSHOP_URL = 'https://steamcommunity.com/sharedfiles/filedetails/?id=';

export default function ModsPage() {
  const t = useTranslations('mods');
  const servers = useServers();

  // Server ids are strings in the store, so keep them strings here too
  // rather than round-tripping through Number and back.
  const [serverId, setServerId] = useState<string | null>(null);
  const [mods, setMods] = useState<ServerMod[]>([]);
  const [results, setResults] = useState<WorkshopItem[]>([]);
  const [query, setQuery] = useState('');
  const [directId, setDirectId] = useState('');
  const [searching, setSearching] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [searchError, setSearchError] = useState('');

  useEffect(() => {
    serversActions.fetchServers().catch(() => {});
  }, []);

  // Preselect a server passed as ?server_id=, matching the plugins page so a
  // card can deep-link straight into its mods.
  useEffect(() => {
    const raw = new URLSearchParams(window.location.search).get('server_id');
    if (raw && /^[0-9]+$/.test(raw)) setServerId(raw);
  }, []);

  useEffect(() => {
    // String(): the store types Server.id as a string but the API returns a
    // NUMBER, so assigning it raw made serverId a number and any later
    // serverId.trim() threw "i.trim is not a function", which crashed the whole
    // page with a client-side exception.
    if (!serverId && servers.length) setServerId(String(servers[0].id));
  }, [servers, serverId]);

  const loadMods = useCallback(async (id: string) => {
    try {
      const res = await api.get<{ data?: ServerMod[] }>(`/api/servers/${id}/mods`);
      setMods(res.data.data ?? []);
      setError('');
    } catch (e) {
      setError(errorMessage(e, t('loadFailed')));
    }
  }, [t]);

  useEffect(() => {
    if (serverId) loadMods(serverId);
  }, [serverId, loadMods]);

  const search = async () => {
    setSearching(true);
    setSearchError('');
    try {
      const params = new URLSearchParams({ q: query, limit: '24' });
      const res = await api.get<{ data?: WorkshopItem[] }>(`/api/mods/search?${params}`);
      setResults(res.data.data ?? []);
    } catch (e) {
      setResults([]);
      setSearchError(errorMessage(e, t('searchFailed')));
    } finally {
      setSearching(false);
    }
  };

  const addMod = async (workshopId: string) => {
    if (!serverId) return;
    setBusy(true);
    setError('');
    try {
      await api.post(`/api/servers/${serverId}/mods`, { workshop_id: workshopId });
      await loadMods(serverId);
      setDirectId('');
    } catch (e) {
      setError(errorMessage(e, t('addFailed')));
    } finally {
      setBusy(false);
    }
  };

  const removeMod = async (workshopId: string) => {
    if (!serverId) return;
    setBusy(true);
    try {
      await api.delete(`/api/servers/${serverId}/mods/${encodeURIComponent(workshopId)}`);
      await loadMods(serverId);
    } catch (e) {
      setError(errorMessage(e, t('removeFailed')));
    } finally {
      setBusy(false);
    }
  };

  const toggleMod = async (mod: ServerMod) => {
    if (!serverId) return;
    setBusy(true);
    try {
      await api.put(`/api/servers/${serverId}/mods/${encodeURIComponent(mod.workshop_id)}/enabled`, {
        enabled: !mod.enabled,
      });
      await loadMods(serverId);
    } catch (e) {
      setError(errorMessage(e, t('updateFailed')));
    } finally {
      setBusy(false);
    }
  };

  // Load order is sent in full rather than as a move instruction, so the
  // request is idempotent and cannot half-apply.
  const move = async (index: number, delta: number) => {
    if (!serverId) return;
    const next = [...mods];
    const target = index + delta;
    if (target < 0 || target >= next.length) return;
    [next[index], next[target]] = [next[target], next[index]];

    setMods(next); // optimistic; reconciled by the reload below
    setBusy(true);
    try {
      await api.put(`/api/servers/${serverId}/mods/order`, {
        workshop_ids: next.map((m) => m.workshop_id),
      });
    } catch (e) {
      setError(errorMessage(e, t('reorderFailed')));
    } finally {
      await loadMods(serverId);
      setBusy(false);
    }
  };

  const attached = new Set(mods.map((m) => m.workshop_id));

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-foreground">{t('title')}</h1>
        <p className="text-sm text-muted-foreground">{t('subtitle')}</p>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <label htmlFor="mod-server" className="text-sm text-muted-foreground">
          {t('server')}
        </label>
        <select
          id="mod-server"
          className="h-9 rounded-md border border-input bg-background px-3 text-sm"
          value={serverId ?? ''}
          onChange={(e) => setServerId(e.target.value)}
        >
          {servers.map((s) => (
            <option key={s.id} value={s.id}>
              {s.session_name || s.identifier}
            </option>
          ))}
        </select>
      </div>

      {error && (
        <div className="rounded-md border border-red-900/50 bg-red-950/30 px-4 py-3 text-sm text-red-300">
          {error}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Attached mods, in load order */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t('installed', { count: mods.length })}</CardTitle>
            <p className="text-xs text-muted-foreground">{t('loadOrderHint')}</p>
          </CardHeader>
          <CardContent className="space-y-2">
            {mods.length === 0 && (
              <p className="py-6 text-center text-sm text-muted-foreground">{t('noMods')}</p>
            )}
            {mods.map((mod, i) => (
              <div
                key={mod.workshop_id}
                className="flex items-center gap-3 rounded-md border border-border px-3 py-2"
              >
                <span className="w-6 shrink-0 text-xs tabular-nums text-muted-foreground">
                  {i + 1}
                </span>
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm text-foreground">
                    {mod.name || mod.workshop_id}
                  </div>
                  <div className="text-xs tabular-nums text-muted-foreground">
                    {mod.workshop_id}
                  </div>
                </div>
                {!mod.enabled && <Badge variant="secondary">{t('disabled')}</Badge>}
                <div className="flex shrink-0 items-center gap-1">
                  <Button type="button" variant="ghost" size="sm" className="h-8 w-8 p-0"
                    disabled={busy || i === 0} onClick={() => move(i, -1)}
                    title={t('moveUp')} aria-label={t('moveUp')}>
                    <ArrowUp className="h-4 w-4" />
                  </Button>
                  <Button type="button" variant="ghost" size="sm" className="h-8 w-8 p-0"
                    disabled={busy || i === mods.length - 1} onClick={() => move(i, 1)}
                    title={t('moveDown')} aria-label={t('moveDown')}>
                    <ArrowDown className="h-4 w-4" />
                  </Button>
                  <Button type="button" variant="ghost" size="sm" className="h-8 w-8 p-0"
                    disabled={busy} onClick={() => toggleMod(mod)}
                    title={mod.enabled ? t('disable') : t('enable')}
                    aria-label={mod.enabled ? t('disable') : t('enable')}>
                    {mod.enabled ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
                  </Button>
                  <Button type="button" variant="ghost" size="sm"
                    className="h-8 w-8 p-0 text-red-400 hover:text-red-300"
                    disabled={busy} onClick={() => removeMod(mod.workshop_id)}
                    title={t('remove')} aria-label={t('remove')}>
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
            <p className="pt-2 text-xs text-muted-foreground">{t('restartHint')}</p>
          </CardContent>
        </Card>

        {/* Workshop browser */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t('browse')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <form
              className="flex gap-2"
              onSubmit={(e) => { e.preventDefault(); search(); }}
            >
              <Input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder={t('searchPlaceholder')}
                aria-label={t('searchPlaceholder')}
              />
              <Button type="submit" disabled={searching}>
                {searching ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
              </Button>
            </form>

            {/* Search needs a Steam API key; adding by ID never does, so it is
                always offered as the fallback path. */}
            {searchError && (
              <div className="space-y-2 rounded-md border border-amber-900/50 bg-amber-950/20 px-3 py-2">
                <p className="text-xs text-amber-200">{searchError}</p>
                <p className="text-xs text-muted-foreground">{t('addByIdHint')}</p>
              </div>
            )}

            <form
              className="flex gap-2"
              onSubmit={(e) => { e.preventDefault(); if (directId.trim()) addMod(directId.trim()); }}
            >
              <Input
                value={directId}
                onChange={(e) => setDirectId(e.target.value)}
                placeholder={t('workshopIdPlaceholder')}
                aria-label={t('workshopIdPlaceholder')}
                inputMode="numeric"
              />
              <Button type="submit" variant="outline" disabled={busy || !directId.trim()}>
                <Plus className="h-4 w-4" />
              </Button>
            </form>

            <div className="space-y-2">
              {results.map((item) => (
                <div
                  key={item.workshop_id}
                  className="flex items-center gap-3 rounded-md border border-border px-3 py-2"
                >
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm text-foreground">{item.title}</div>
                    <div className="text-xs tabular-nums text-muted-foreground">
                      {item.workshop_id}
                    </div>
                  </div>
                  <a
                    href={`${WORKSHOP_URL}${encodeURIComponent(item.workshop_id)}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-muted-foreground hover:text-foreground"
                    title={t('viewOnSteam')}
                    aria-label={t('viewOnSteam')}
                  >
                    <ExternalLink className="h-4 w-4" />
                  </a>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={busy || attached.has(item.workshop_id)}
                    onClick={() => addMod(item.workshop_id)}
                  >
                    {attached.has(item.workshop_id) ? t('added') : t('add')}
                  </Button>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
