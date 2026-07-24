"use client";

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import { useRouter, useParams } from 'next/navigation';
import { Server } from '@/stores/servers';
import { serversActions } from '@/stores/servers';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, ArrowLeft, Save, Eye, EyeOff, Download, Upload } from 'lucide-react';
import { GameUserSettingsEditor } from '@/components/servers/GameUserSettingsEditor';
import { GameIniEditor } from '@/components/servers/GameIniEditor';
import { PresetSelector } from '@/components/servers/PresetSelector';
import { ServerArgsEditor } from '@/components/servers/ServerArgsEditor';
import { MapSelector } from '@/components/servers/MapSelector';
import { Alert, AlertDescription } from '@/components/ui/alert';

export default function ServerEditPage() {
  const tServers = useTranslations('servers');
  const tServersEdit = useTranslations('servers.edit');
  const tCommon = useTranslations('common');
  const router = useRouter();
  const params = useParams();
  const serverId = params.id as string;

  const { getServer, updateServer } = serversActions;

  const [server, setServer] = useState<Server | null>(null);
  const [formData, setFormData] = useState<Partial<Server>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    const loadServer = async () => {
      try {
        setLoading(true);
        const serverData = await getServer(serverId);
        setServer(serverData);
        setFormData(serverData);
      } catch {
        setError(tServersEdit('loadServerInfoFailed'));
      } finally {
        setLoading(false);
      }
    };

    if (serverId) {
      loadServer();
    }
  }, [serverId, getServer, tServersEdit]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    const isNumber = type === 'number';
    setFormData((prev) => ({ ...prev, [name]: isNumber ? Number(value) : value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setSaving(true);
      await updateServer(serverId, formData);
      setSuccess(tServers('serverUpdateSuccess'));
      // ：SaveSuccessServer list
      // router.push('/servers');
    } catch {
      setError(tServers('serverUpdateError'));
    } finally {
      setSaving(false);
    }
  };

  const handleDownloadText = (content: string, filename: string) => {
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleExportAll = () => {
    const config = {
      game_user_settings: formData.game_user_settings || '',
      game_ini: formData.game_ini || '',
      server_args: formData.server_args || { query_params: {}, command_line_args: {}, custom_args: [] },
      exported_at: new Date().toISOString(),
    };
    const blob = new Blob([JSON.stringify(config, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `server-config-${formData.identifier || 'export'}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleImportAll = async () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;
      try {
        const text = await file.text();
        const config = JSON.parse(text);
        setFormData((prev) => ({
          ...prev,
          game_user_settings: config.game_user_settings || prev.game_user_settings,
          game_ini: config.game_ini || prev.game_ini,
          server_args: config.server_args || prev.server_args,
        }));
      } catch (err) {
        console.error('Import failed:', err);
      }
    };
    input.click();
  };

  const handleImportFile = (accept: string, callback: (content: string) => void) => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = accept;
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;
      try {
        const text = await file.text();
        callback(text);
      } catch (err) {
        console.error('Import failed:', err);
      }
    };
    input.click();
  };

  const handleBack = () => {
    router.push('/servers');
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
      <div className="w-full max-w-none py-8">
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
        <Button onClick={handleBack} className="mt-4">
          <ArrowLeft className="mr-2 h-4 w-4" />
          {tCommon('back')}
        </Button>
      </div>
    );
  }

  return (
    <div className="w-full max-w-none py-8">
      <div className="mb-6">
        <div className="flex items-center gap-4 mb-4">
          <Button variant="outline" onClick={handleBack}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            {tCommon('back')}
          </Button>
          <div>
            <h1 className="text-2xl lg:text-3xl font-bold text-gray-900">{tServersEdit('editTitle')}</h1>
            <p className="text-gray-600">{server?.session_name}</p>
          </div>
        </div>

        {error && (
          <Alert variant="destructive" className="mb-4">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {success && (
          <Alert className="mb-4">
            <AlertDescription>{success}</AlertDescription>
          </Alert>
        )}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{tServersEdit('editTitle')}</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="basic">
            <TabsList>
              <TabsTrigger value="basic">{tServersEdit('basicParams')}</TabsTrigger>
              <TabsTrigger value="game_user_settings">{tServersEdit('gameUserSettings')}</TabsTrigger>
              <TabsTrigger value="game_ini">{tServersEdit('gameIni')}</TabsTrigger>
              <TabsTrigger value="server_args">{tServersEdit('serverArgs')}</TabsTrigger>
            </TabsList>

            <TabsContent value="basic">
              <form onSubmit={handleSubmit} className="space-y-4 py-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="session_name">{tServersEdit('serverName')}</Label>
                    <Input id="session_name" name="session_name" value={formData.session_name || ''} onChange={handleChange} />
                  </div>
                  <div>
                    <MapSelector
                      value={formData.map || ''}
                      onChange={(value) => setFormData(prev => ({ ...prev, map: value }))}
                      label={tServersEdit('map')}
                    />
                  </div>
                </div>
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <Label htmlFor="port">{tServersEdit('gamePort')}</Label>
                    <Input id="port" name="port" type="number" value={formData.port || ''} onChange={handleChange} />
                  </div>
                  <div>
                    <Label htmlFor="query_port">{tServersEdit('queryPort')}</Label>
                    <Input id="query_port" name="query_port" type="number" value={formData.query_port || ''} onChange={handleChange} />
                  </div>
                  <div>
                    <Label htmlFor="rcon_port">{tServersEdit('rconPort')}</Label>
                    <Input id="rcon_port" name="rcon_port" type="number" value={formData.rcon_port || ''} onChange={handleChange} />
                  </div>
                </div>
                <div>
                  <Label htmlFor="admin_password">{tServersEdit('adminPassword')}</Label>
                  <div className="relative">
                    <Input 
                      id="admin_password" 
                      name="admin_password" 
                      type={showPassword ? 'text' : 'password'} 
                      value={formData.admin_password || ''} 
                      onChange={handleChange} 
                      className="pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    >
                      {showPassword ? (
                        <EyeOff className="h-4 w-4 text-gray-500" />
                      ) : (
                        <Eye className="h-4 w-4 text-gray-500" />
                      )}
                    </button>
                  </div>
                </div>
                <div>
                  <Label htmlFor="max_players">{tServersEdit('maxPlayers')}</Label>
                  <Input id="max_players" name="max_players" type="number" value={formData.max_players || ''} onChange={handleChange} />
                </div>
              </form>
              <div className="flex gap-2 pt-4 border-t mt-4">
                <Button type="button" variant="outline" size="sm" onClick={handleExportAll}>
                  <Download className="h-4 w-4 mr-1" />{tServersEdit('exportAll')}
                </Button>
                <Button type="button" variant="outline" size="sm" onClick={handleImportAll}>
                  <Upload className="h-4 w-4 mr-1" />{tServersEdit('importAll')}
                </Button>
              </div>
            </TabsContent>

            <TabsContent value="game_user_settings">
              <div className="flex justify-end gap-2 mb-2">
                <Button variant="outline" size="sm" onClick={() => handleDownloadText(formData.game_user_settings || '', 'GameUserSettings.ini')}>
                  <Download className="h-3 w-3 mr-1" />{tServersEdit('exportFile')}
                </Button>
                <Button variant="outline" size="sm" onClick={() => handleImportFile('.ini,.txt', (text) => setFormData(p => ({ ...p, game_user_settings: text })))}>
                  <Upload className="h-3 w-3 mr-1" />{tServersEdit('importFile')}
                </Button>
              </div>
              <GameUserSettingsEditor
                value={formData.game_user_settings}
                onChange={(v) => setFormData(p => ({ ...p, game_user_settings: v }))}
              />
            </TabsContent>

            <TabsContent value="game_ini">
              <div className="space-y-4 py-4">
                <details className="group">
                  <summary className="cursor-pointer text-sm font-medium text-muted-foreground hover:text-foreground">
                    Presets (PVE/PVP)
                  </summary>
                  <div className="mt-3">
                    <PresetSelector onSelect={(ini) => {
                      setFormData(p => ({ ...p, game_ini: ini }));
                    }} />
                  </div>
                </details>
                <div className="flex justify-end gap-2">
                  <Button variant="outline" size="sm" onClick={() => handleDownloadText(formData.game_ini || '', 'Game.ini')}>
                    <Download className="h-3 w-3 mr-1" />{tServersEdit('exportFile')}
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => handleImportFile('.ini,.txt', (text) => setFormData(p => ({ ...p, game_ini: text })))}>
                    <Upload className="h-3 w-3 mr-1" />{tServersEdit('importFile')}
                  </Button>
                </div>
                <GameIniEditor />
              </div>
            </TabsContent>

            <TabsContent value="server_args">
              <div className="flex justify-end gap-2 mb-2">
                <Button variant="outline" size="sm" onClick={() => handleDownloadText(JSON.stringify(formData.server_args || { query_params: {}, command_line_args: {}, custom_args: [] }, null, 2), 'SERVER_ARGS.json')}>
                  <Download className="h-3 w-3 mr-1" />{tServersEdit('exportFile')}
                </Button>
                <Button variant="outline" size="sm" onClick={() => handleImportFile('.json,.txt', (text) => { try { setFormData(p => ({ ...p, server_args: JSON.parse(text) })); } catch {} })}>
                  <Upload className="h-3 w-3 mr-1" />{tServersEdit('importFile')}
                </Button>
              </div>
              {/* @ts-expect-error: Prop 'value' is not available on type 'IntrinsicAttributes' */}
              <ServerArgsEditor value={formData.server_args} onChange={(v) => setFormData(p => ({ ...p, server_args: v }))} />
            </TabsContent>
          </Tabs>

          <div className="flex justify-end gap-2 mt-6 pt-6 border-t">
            <Button variant="outline" onClick={handleBack}>
              {tCommon('cancel')}
            </Button>
            <Button onClick={handleSubmit} disabled={saving}>
              {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              <Save className="mr-2 h-4 w-4" />
              {tServersEdit('saveChanges')}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}