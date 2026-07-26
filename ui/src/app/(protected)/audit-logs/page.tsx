"use client";

import { useEffect, useState, useCallback } from 'react';
import { useTranslations } from 'next-intl';
import { AuditLog } from '@/components/audit/types';
import { AuditLogFilters } from '@/components/audit/AuditLogFilters';
import { Button } from '@/components/ui/button';
import { ClosableAlert } from '@/components/ui/closable-alert';
import { Loader2, Search, RefreshCw } from 'lucide-react';

export default function AuditLogsPage() {
  const t = useTranslations('auditLogs');
  const tCommon = useTranslations('common');

  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);

  const [filters, setFilters] = useState({
    user_id: '',
    action: '',
    start_date: '',
    end_date: '',
  });

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params = new URLSearchParams();
      params.set('page', String(page));
      params.set('limit', String(limit));
      if (filters.user_id) params.set('user_id', filters.user_id);
      if (filters.action) params.set('action', filters.action);
      if (filters.start_date) params.set('start_date', filters.start_date);
      if (filters.end_date) params.set('end_date', filters.end_date);

      const res = await fetch(`/api/audit-logs?${params.toString()}`);
      if (!res.ok) throw new Error(await res.text());
      const data = await res.json();
      setLogs(data.data || []);
      setTotal(data.total || 0);
    } catch {
      setError(t('fetchFailed'));
    } finally {
      setLoading(false);
    }
  }, [page, limit, filters, t]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const handleFilter = (newFilters: typeof filters) => {
    setFilters(newFilters);
    setPage(1);
  };

  const totalPages = Math.ceil(total / limit) || 1;

  return (
    <div className="w-full max-w-none py-8">
      <div className="mb-6">
        <h1 className="text-2xl lg:text-3xl font-bold text-foreground">{t('title')}</h1>
      </div>

      {error && <ClosableAlert variant="destructive" className="mb-4" title={tCommon('error')} onClose={() => setError('')}>{error}</ClosableAlert>}

      <AuditLogFilters filters={filters} onFilter={handleFilter} onRefresh={fetchLogs} />

      {loading && logs.length === 0 ? (
        <div className="text-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-blue-400 mx-auto mb-4" />
          <p className="text-muted-foreground">{tCommon('loading')}</p>
        </div>
      ) : logs.length === 0 ? (
        <div className="text-center py-16 px-4">
          <p className="text-muted-foreground">{t('noLogs')}</p>
        </div>
      ) : (
        <div className="rounded-lg border bg-card">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">#</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('userID')}</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('action')}</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('resource')}</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('detail')}</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('ip')}</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">{t('time')}</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((log, idx) => (
                <tr key={log.id} className="border-b last:border-0 hover:bg-accent/50">
                  <td className="px-4 py-3 text-muted-foreground">{(page - 1) * limit + idx + 1}</td>
                  <td className="px-4 py-3">{log.user_id}</td>
                  <td className="px-4 py-3"><code className="text-xs bg-muted px-1.5 py-0.5 rounded">{log.action}</code></td>
                  <td className="px-4 py-3 text-muted-foreground">{log.resource || '-'}</td>
                  <td className="px-4 py-3 text-muted-foreground max-w-xs truncate" title={log.detail || ''}>{log.detail || '-'}</td>
                  <td className="px-4 py-3 text-muted-foreground">{log.ip || '-'}</td>
                  <td className="px-4 py-3 text-muted-foreground whitespace-nowrap">{new Date(log.created_at).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {totalPages > 1 && !loading && (
        <div className="flex items-center justify-center gap-2 mt-6">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page <= 1}
          >
            {tCommon('previous')}
          </Button>
          <span className="text-sm text-muted-foreground">
            {t('page')} {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(p => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
          >
            {tCommon('next')}
          </Button>
        </div>
      )}
    </div>
  );
}