"use client";

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Search, RefreshCw } from 'lucide-react';

export interface AuditLogFiltersData {
  user_id: string;
  action: string;
  start_date: string;
  end_date: string;
}

interface Props {
  filters: AuditLogFiltersData;
  onFilter: (filters: AuditLogFiltersData) => void;
  onRefresh: () => void;
}

export function AuditLogFilters({ filters, onFilter, onRefresh }: Props) {
  const t = useTranslations('auditLogs');
  const tCommon = useTranslations('common');
  const [local, setLocal] = useState(filters);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onFilter(local);
  };

  const handleReset = () => {
    const empty = { user_id: '', action: '', start_date: '', end_date: '' };
    setLocal(empty);
    onFilter(empty);
  };

  return (
    <form onSubmit={handleSubmit} className="mb-4 flex flex-wrap gap-3 items-end">
      <div className="flex flex-col gap-1">
        <Label htmlFor="filter-user-id" className="text-xs">{t('filterUserID')}</Label>
        <Input
          id="filter-user-id"
          type="number"
          min={1}
          placeholder={t('filterUserIDPlaceholder')}
          value={local.user_id}
          onChange={e => setLocal({ ...local, user_id: e.target.value })}
          className="w-32"
        />
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="filter-action" className="text-xs">{t('filterAction')}</Label>
        <Input
          id="filter-action"
          type="text"
          placeholder={t('filterActionPlaceholder')}
          value={local.action}
          onChange={e => setLocal({ ...local, action: e.target.value })}
          className="w-44"
        />
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="filter-start" className="text-xs">{t('filterStartDate')}</Label>
        <Input
          id="filter-start"
          type="datetime-local"
          value={local.start_date}
          onChange={e => setLocal({ ...local, start_date: e.target.value })}
          className="w-44"
        />
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="filter-end" className="text-xs">{t('filterEndDate')}</Label>
        <Input
          id="filter-end"
          type="datetime-local"
          value={local.end_date}
          onChange={e => setLocal({ ...local, end_date: e.target.value })}
          className="w-44"
        />
      </div>
      <div className="flex gap-2 items-end">
        <Button type="submit" size="sm">
          <Search className="mr-1.5 h-3.5 w-3.5" />
          {tCommon('filter')}
        </Button>
        <Button type="button" variant="outline" size="sm" onClick={handleReset}>
          {tCommon('reset')}
        </Button>
        <Button type="button" variant="outline" size="sm" onClick={onRefresh}>
          <RefreshCw className="mr-1.5 h-3.5 w-3.5" />
          {tCommon('refresh')}
        </Button>
      </div>
    </form>
  );
}