"use client";

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Server } from '@/stores/servers';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import {
  Play, Square, Loader2, Info, Edit, Trash2, Wifi, Lock, Eye, EyeOff, Copy, RefreshCw, Map,
} from 'lucide-react';

interface ServerCardProps {
  server: Server;
  canStartServer: boolean;
  onStart: (server: Server) => void;
  onStop: (server: Server) => void;
  onRestart: (server: Server) => void;
  onEdit: (server: Server) => void;
  onDelete: (server: Server) => void;
  onViewLogs?: (server: Server) => void;
  mapClickable?: boolean;
}

export function ServerCard({
  server, canStartServer, onStart, onStop, onRestart, onEdit, onDelete, onViewLogs, mapClickable,
}: ServerCardProps) {
  const t = useTranslations('servers');
  const [showPassword, setShowPassword] = useState(false);

  const getMapDisplayName = (mapName: string) => {
    const mapKey = `edit.maps.${mapName}`;
    const translatedName = t(mapKey);
    return translatedName !== mapKey ? translatedName : mapName;
  };

  const getStatusVariant = (s: Server['status']): 'default' | 'destructive' | 'secondary' | 'outline' => {
    switch (s) {
      case 'running': return 'default';
      case 'stopped': return 'destructive';
      case 'starting': case 'stopping': case 'restarting': return 'secondary';
      default: return 'outline';
    }
  };

  const iconBtn = (icon: React.ReactNode, onClick?: () => void, className = '', disabled = false) => (
    <Button variant="ghost" size="sm" className={`h-8 w-8 p-0 ${className}`} onClick={onClick} disabled={disabled}>
      {icon}
    </Button>
  );

  return (
    <Card className="h-full border-0 shadow-sm">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold">{server.session_name}</CardTitle>
        {/* Badge + action buttons row */}
        <div className="flex items-center justify-between mt-1">
          <Badge variant={getStatusVariant(server.status)} className="text-xs px-2 py-0.5">
            {t(`card.${server.status}`)}
          </Badge>
          <div className="flex items-center gap-0.5 flex-shrink-0">
            {server.status === 'running' ? (
              iconBtn(<Square className="h-4 w-4" />, () => onStop(server), 'text-red-600 hover:text-red-700 hover:bg-red-50')
            ) : server.status === 'stopped' ? (
              iconBtn(<Play className="h-4 w-4" />, () => onStart(server), 'text-green-600 hover:text-green-700 hover:bg-green-50', !canStartServer)
            ) : (
              iconBtn(<Loader2 className="h-4 w-4 animate-spin" />, undefined, 'text-blue-600', true)
            )}
            {server.status === 'running' && iconBtn(<RefreshCw className="h-4 w-4" />, () => onRestart(server), 'text-orange-600 hover:text-orange-700 hover:bg-orange-50')}
            {iconBtn(<Edit className="h-4 w-4" />, () => onEdit(server), 'text-blue-600 hover:text-blue-700 hover:bg-blue-50')}
            <Popover>
              <PopoverTrigger asChild>
                {iconBtn(<Trash2 className="h-4 w-4" />, undefined, 'text-red-600 hover:text-red-700 hover:bg-red-50')}
              </PopoverTrigger>
              <PopoverContent>
                <div className="space-y-2">
                  <p className="text-sm">{t('card.confirmDeleteMessage', { identifier: server.session_name })}</p>
                  <Button size="sm" variant="destructive" onClick={() => onDelete(server)}>{t('deleteServer')}</Button>
                </div>
              </PopoverContent>
            </Popover>
            {onViewLogs && mapClickable && iconBtn(<Map className="h-4 w-4" />, () => onViewLogs(server), 'text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50')}
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-2 space-y-2">
        <div className="bg-muted/50 rounded-lg p-2.5">
          <div className="flex items-center mb-1.5">
            <Wifi className="h-3.5 w-3.5 text-primary mr-1.5" />
            <span className="text-xs font-medium text-foreground/80">{t('card.portConfig')}</span>
          </div>
          <div className="grid grid-cols-3 gap-2 text-xs text-center">
            {[
              [t('card.gamePort'), server.port],
              [t('card.queryPort'), server.query_port],
              ['RCON', server.rcon_port],
            ].map(([label, val]) => (
              <div key={String(label)}>
                <div className="text-muted-foreground">{label}</div>
                <div className="font-mono font-semibold">{val}</div>
              </div>
            ))}
          </div>
        </div>
        <div className="space-y-1.5">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground flex items-center"><Info className="h-3.5 w-3.5 mr-1.5 text-green-600" />{t('card.map')}</span>
            {onViewLogs && mapClickable ? (
              <button onClick={() => onViewLogs(server)} className="font-medium truncate ml-2 text-emerald-600 hover:underline flex items-center gap-1">
                <Map className="h-3 w-3" />{getMapDisplayName(server.map)}
              </button>
            ) : (
              <span className="font-medium truncate ml-2">{getMapDisplayName(server.map)}</span>
            )}
          </div>
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">{t('card.maxPlayers')}</span>
            <span className="font-medium">{server.max_players}</span>
          </div>
        </div>
        <div className="bg-amber-50 dark:bg-amber-950/30 rounded-lg p-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <Lock className="h-3.5 w-3.5 text-amber-600 mr-1.5" />
              <span className="text-xs font-medium text-amber-800 dark:text-amber-200">{t('card.adminPassword')}</span>
            </div>
            <div className="flex items-center gap-1">
              <span className="font-mono text-xs">{showPassword ? server.admin_password : '••••••••'}</span>
              <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => setShowPassword(!showPassword)}>
                {showPassword ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
              </Button>
              <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => navigator.clipboard.writeText(server.admin_password)}>
                <Copy className="h-3 w-3" />
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
