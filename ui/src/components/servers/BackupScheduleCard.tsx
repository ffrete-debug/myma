"use client";

import { useCallback, useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { CloudUpload, Loader2 } from 'lucide-react';
import api from '@/lib/axios';

interface BackupSchedule {
  enabled: boolean;
  interval_hours: number;
  retain_count: number;
  upload_to_cloud: boolean;
  last_run_at: string | null;
  last_status: string;
  last_error?: string;
  next_run_at: string | null;
}

interface CloudStatus {
  configured: boolean;
  bucket: string;
  endpoint: string;
  prefix: string;
}

function errorMessage(e: unknown, fallback: string): string {
  const err = e as { response?: { data?: { error?: string } } };
  return err.response?.data?.error || fallback;
}

export function BackupScheduleCard({ serverId }: { serverId: string }) {
  const t = useTranslations('backups');

  const [schedule, setSchedule] = useState<BackupSchedule | null>(null);
  const [cloud, setCloud] = useState<CloudStatus | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);

  const load = useCallback(async () => {
    try {
      const [s, c] = await Promise.all([
        api.get<{ data?: BackupSchedule }>(`/api/servers/${serverId}/backup-schedule`),
        api.get<{ data?: CloudStatus }>('/api/backups/cloud-status'),
      ]);
      if (s.data.data) setSchedule(s.data.data);
      if (c.data.data) setCloud(c.data.data);
    } catch (e) {
      setError(errorMessage(e, t('loadFailed')));
    }
  }, [serverId, t]);

  useEffect(() => { load(); }, [load]);

  const save = async () => {
    if (!schedule) return;
    setSaving(true);
    setError('');
    setSaved(false);
    try {
      const res = await api.put<{ data?: BackupSchedule }>(
        `/api/servers/${serverId}/backup-schedule`,
        {
          enabled: schedule.enabled,
          interval_hours: schedule.interval_hours,
          retain_count: schedule.retain_count,
          upload_to_cloud: schedule.upload_to_cloud,
        },
      );
      if (res.data.data) setSchedule(res.data.data);
      setSaved(true);
    } catch (e) {
      setError(errorMessage(e, t('saveFailed')));
    } finally {
      setSaving(false);
    }
  };

  if (!schedule) return null;

  const patch = (p: Partial<BackupSchedule>) => {
    setSchedule({ ...schedule, ...p });
    setSaved(false);
  };

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-2">
          <CardTitle className="text-base">{t('scheduleTitle')}</CardTitle>
          {schedule.enabled && <Badge>{t('on')}</Badge>}
        </div>
        <p className="text-xs text-muted-foreground">{t('scheduleHint')}</p>
      </CardHeader>

      <CardContent className="space-y-4">
        {error && (
          <div className="rounded-md border border-red-900/50 bg-red-950/30 px-3 py-2 text-xs text-red-300">
            {error}
          </div>
        )}

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={schedule.enabled}
            onChange={(e) => patch({ enabled: e.target.checked })}
          />
          {t('enableAutomatic')}
        </label>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1">
            <label htmlFor="interval" className="text-xs text-muted-foreground">
              {t('everyHours')}
            </label>
            <Input
              id="interval"
              type="number"
              min={1}
              max={168}
              value={schedule.interval_hours}
              onChange={(e) => patch({ interval_hours: Number(e.target.value) })}
            />
          </div>
          <div className="space-y-1">
            <label htmlFor="retain" className="text-xs text-muted-foreground">
              {t('keepLast')}
            </label>
            <Input
              id="retain"
              type="number"
              min={0}
              max={365}
              value={schedule.retain_count}
              onChange={(e) => patch({ retain_count: Number(e.target.value) })}
            />
            <p className="text-xs text-muted-foreground">{t('keepZeroHint')}</p>
          </div>
        </div>

        {/* Cloud upload is only offered when a destination exists; otherwise the
            setting would be accepted and then fail silently at 3am. */}
        <div className="space-y-1">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              disabled={!cloud?.configured}
              checked={schedule.upload_to_cloud}
              onChange={(e) => patch({ upload_to_cloud: e.target.checked })}
            />
            <CloudUpload className="h-4 w-4" />
            {t('uploadToCloud')}
          </label>
          {cloud && !cloud.configured && (
            <p className="text-xs text-amber-300">{t('cloudNotConfigured')}</p>
          )}
          {cloud?.configured && (
            <p className="text-xs text-muted-foreground">
              {t('cloudTarget', { bucket: cloud.bucket, endpoint: cloud.endpoint })}
            </p>
          )}
        </div>

        {schedule.last_run_at && (
          <div className="border-t border-border pt-3 text-xs text-muted-foreground">
            <div>
              {t('lastRun', {
                time: new Date(schedule.last_run_at).toLocaleString(),
                status: schedule.last_status || '—',
              })}
            </div>
            {schedule.last_error && (
              <div className="mt-1 text-amber-300">{schedule.last_error}</div>
            )}
            {schedule.next_run_at && schedule.enabled && (
              <div className="mt-1">
                {t('nextRun', { time: new Date(schedule.next_run_at).toLocaleString() })}
              </div>
            )}
          </div>
        )}

        <div className="flex items-center gap-3">
          <Button type="button" onClick={save} disabled={saving}>
            {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            {t('saveSchedule')}
          </Button>
          {saved && <span className="text-xs text-muted-foreground">{t('saved')}</span>}
        </div>
      </CardContent>
    </Card>
  );
}
