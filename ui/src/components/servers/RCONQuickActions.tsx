"use client";

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Loader2, Megaphone, Save, Clock, Skull, AlertTriangle } from 'lucide-react';
import api from '@/lib/axios';

/**
 * Common admin operations, sent as structured actions rather than free text.
 *
 * The command string is assembled on the server from the action name and its
 * parameters. ARK's RCON is line-oriented, so building it here from user input
 * would let a newline in a broadcast message append a second command.
 */

interface Props {
  serverId: string;
  disabled?: boolean;
  /** Called with the resolved command and its output so the console can echo it. */
  onResult?: (command: string, output: string) => void;
}

type ActionName =
  | 'broadcast' | 'saveworld' | 'settimeofday' | 'destroywilddinos';

// Mirrors Action.Destructive() on the server. Advisory only — the server does
// not rely on the client to gate anything.
const DESTRUCTIVE: ReadonlySet<ActionName> = new Set<ActionName>(['destroywilddinos']);

function errorMessage(e: unknown, fallback: string): string {
  const err = e as { response?: { data?: { error?: string; message?: string } } };
  return err.response?.data?.error || err.response?.data?.message || fallback;
}

export function RCONQuickActions({ serverId, disabled = false, onResult }: Props) {
  const t = useTranslations('rcon');

  const [message, setMessage] = useState('');
  const [timeOfDay, setTimeOfDay] = useState('12:00');
  const [running, setRunning] = useState<ActionName | null>(null);
  const [error, setError] = useState('');
  const [confirming, setConfirming] = useState<ActionName | null>(null);

  const run = async (action: ActionName, params?: Record<string, string>) => {
    setRunning(action);
    setError('');
    try {
      const res = await api.post<{ data?: { command: string; output: string } }>(
        `/api/servers/${serverId}/rcon/action`,
        { action, params: params ?? {} },
      );
      const data = res.data.data;
      if (data && onResult) onResult(data.command, data.output);
      if (action === 'broadcast') setMessage('');
    } catch (e) {
      setError(errorMessage(e, t('actionFailed')));
    } finally {
      setRunning(null);
      setConfirming(null);
    }
  };

  const request = (action: ActionName, params?: Record<string, string>) => {
    if (DESTRUCTIVE.has(action) && confirming !== action) {
      setConfirming(action);
      return;
    }
    run(action, params);
  };

  const busy = (a: ActionName) => running === a;

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">{t('quickActions')}</CardTitle>
      </CardHeader>

      <CardContent className="space-y-4">
        {error && (
          <div className="rounded-md border border-red-900/50 bg-red-950/30 px-3 py-2 text-xs text-red-300">
            {error}
          </div>
        )}

        <form
          className="flex gap-2"
          onSubmit={(e) => { e.preventDefault(); if (message.trim()) request('broadcast', { message }); }}
        >
          <Input
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            placeholder={t('broadcastPlaceholder')}
            aria-label={t('broadcastPlaceholder')}
            maxLength={256}
            disabled={disabled}
          />
          <Button type="submit" disabled={disabled || !message.trim() || busy('broadcast')}>
            {busy('broadcast')
              ? <Loader2 className="h-4 w-4 animate-spin" />
              : <Megaphone className="h-4 w-4" />}
          </Button>
        </form>

        <div className="flex flex-wrap gap-2">
          <Button
            type="button" variant="outline" size="sm"
            disabled={disabled || busy('saveworld')}
            onClick={() => request('saveworld')}
          >
            {busy('saveworld')
              ? <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              : <Save className="mr-2 h-4 w-4" />}
            {t('saveWorld')}
          </Button>

          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-muted-foreground" />
            <Input
              type="time"
              className="w-28"
              value={timeOfDay}
              onChange={(e) => setTimeOfDay(e.target.value)}
              aria-label={t('setTime')}
              disabled={disabled}
            />
            <Button
              type="button" variant="outline" size="sm"
              disabled={disabled || busy('settimeofday')}
              onClick={() => request('settimeofday', { time: timeOfDay })}
            >
              {t('setTime')}
            </Button>
          </div>
        </div>

        {/* Destructive actions require a second, explicit click. */}
        <div className="border-t border-border pt-3">
          {confirming === 'destroywilddinos' ? (
            <div className="flex flex-wrap items-center gap-2">
              <span className="flex items-center gap-2 text-xs text-amber-300">
                <AlertTriangle className="h-4 w-4" />
                {t('confirmDestroyDinos')}
              </span>
              <Button
                type="button" variant="destructive" size="sm"
                disabled={busy('destroywilddinos')}
                onClick={() => run('destroywilddinos')}
              >
                {busy('destroywilddinos')
                  ? <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  : null}
                {t('confirm')}
              </Button>
              <Button type="button" variant="ghost" size="sm" onClick={() => setConfirming(null)}>
                {t('cancel')}
              </Button>
            </div>
          ) : (
            <Button
              type="button" variant="outline" size="sm"
              className="text-red-400 hover:text-red-300"
              disabled={disabled}
              onClick={() => request('destroywilddinos')}
            >
              <Skull className="mr-2 h-4 w-4" />
              {t('destroyWildDinos')}
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
