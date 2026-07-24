"use client";

import { useEffect, useState, useCallback, useRef } from 'react';
import { useTranslations } from 'next-intl';
import { useServers, serversActions } from '@/stores/servers';
import { Button } from '@/components/ui/button';
import Cookies from 'js-cookie';
import axios from 'axios';
import {
  Folder, File, Upload, Download, Trash2, Pencil,
  ChevronRight, Home, Loader2, RefreshCw, FolderPlus, Edit3, X, Save,
  FileArchive, FileDown,
} from 'lucide-react';
import { Server } from '@/stores/servers';

interface FileEntry {
  name: string;
  is_dir: boolean;
  size: number;
  mode: string;
  mod_time: string;
}

function authHeaders(): Record<string, string> {
  const token = Cookies.get('auth-token');
  if (!token) return {};
  return { 'Authorization': `Bearer ${token}` };
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '-';
  const units = ['B', 'KB', 'MB', 'GB'];
  let i = 0;
  let size = bytes;
  while (size >= 1024 && i < units.length - 1) { size /= 1024; i++; }
  return `${size.toFixed(1)} ${units[i]}`;
}

function formatTime(unix: string): string {
  const ts = parseInt(unix);
  if (!ts) return '-';
  return new Date(ts * 1000).toLocaleString();
}

export default function PluginsPage() {
  const t = useTranslations('plugins');
  const tCommon = useTranslations('common');
  const servers = useServers();
  const { fetchServers } = serversActions;

  const [selectedServerId, setSelectedServerId] = useState<number | null>(null);
  const [currentPath, setCurrentPath] = useState('/');
  const [files, setFiles] = useState<FileEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [uploading, setUploading] = useState(false);
  const [renaming, setRenaming] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState('');
  const [creatingFolder, setCreatingFolder] = useState(false);
  const [folderName, setFolderName] = useState('');
  const [editingFile, setEditingFile] = useState<FileEntry | null>(null);
  const [editorContent, setEditorContent] = useState('');
  const [editorLoading, setEditorLoading] = useState(false);
  const [editorError, setEditorError] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);
  const dropRef = useRef<HTMLDivElement>(null);
  const [dragOver, setDragOver] = useState(false);

  useEffect(() => { fetchServers().catch(() => {}); }, [fetchServers]);

  const loadFiles = useCallback(async (serverId: number, path: string) => {
    setLoading(true);
    setError('');
    try {
      const params = new URLSearchParams({ server_id: String(serverId), path });
      const res = await fetch(`/api/plugins?${params}`, { headers: authHeaders() });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'load failed');
      setFiles(data.files || []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (selectedServerId) loadFiles(selectedServerId, currentPath);
  }, [selectedServerId, currentPath, loadFiles]);

  const handleServerChange = (id: number) => {
    setSelectedServerId(id);
    setCurrentPath('/');
  };

  const navigateToDir = (dir: string) => {
    const newPath = currentPath === '/' ? `/${dir}` : `${currentPath}/${dir}`;
    setCurrentPath(newPath.replace(/\/+/g, '/'));
  };

  const navigateBreadcrumb = (idx: number) => {
    const parts = currentPath.split('/').filter(Boolean);
    const newPath = '/' + parts.slice(0, idx).join('/');
    setCurrentPath(newPath || '/');
  };

  const handleUpload = async (fileList: FileList) => {
    if (!selectedServerId || !fileList.length) return;
    setUploading(true);
    setError('');
    try {
      const form = new FormData();
      for (let i = 0; i < fileList.length; i++) form.append('files', fileList[i]);
      const token = Cookies.get('auth-token');
      await axios.post(
        `/api/plugins?server_id=${selectedServerId}&path=${currentPath}&action=upload`,
        form,
        { headers: { Authorization: `Bearer ${token}` } }
      );
      loadFiles(selectedServerId, currentPath);
    } catch (e: unknown) {
      const ae = e as { response?: { data?: { error?: string } } };
      setError(ae?.response?.data?.error || 'upload failed');
    } finally {
      setUploading(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    if (e.dataTransfer.files.length) handleUpload(e.dataTransfer.files);
  };

  const handleDelete = async (entry: FileEntry) => {
    if (!selectedServerId) return;
    if (!confirm(t('confirmDelete', { name: entry.name }))) return;
    try {
      const targetPath = currentPath === '/' ? `/${entry.name}` : `${currentPath}/${entry.name}`;
      const params = new URLSearchParams({ server_id: String(selectedServerId), path: targetPath });
      const res = await fetch(`/api/plugins?${params}`, { method: 'DELETE', headers: authHeaders() });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'delete failed');
      loadFiles(selectedServerId, currentPath);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  const handleRenameStart = (entry: FileEntry) => {
    setRenaming(entry.name);
    setRenameValue(entry.name);
  };

  const handleRenameConfirm = async () => {
    if (!selectedServerId || !renaming || !renameValue) return;
    try {
      const oldPath = currentPath === '/' ? `/${renaming}` : `${currentPath}/${renaming}`;
      const newPath = currentPath === '/' ? `/${renameValue}` : `${currentPath}/${renameValue}`;
      const params = new URLSearchParams({
        server_id: String(selectedServerId), action: 'rename',
        old_path: oldPath, new_path: newPath,
      });
      const res = await fetch(`/api/plugins?${params}`, { method: 'POST', body: '{}', headers: authHeaders() });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'rename failed');
      setRenaming(null);
      loadFiles(selectedServerId, currentPath);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  const isEditableFile = (name: string) => /\.(json|ini|txt|cfg|yaml|yml|xml|conf)$/i.test(name);

  const handleEditFile = async (entry: FileEntry) => {
    if (!selectedServerId) return;
    setEditingFile(entry);
    setEditorLoading(true);
    setEditorError('');
    setEditorContent('');
    try {
      const targetPath = currentPath === '/' ? `/${entry.name}` : `${currentPath}/${entry.name}`;
      const token = Cookies.get('auth-token');
      const res = await axios.get(
        `/api/plugins/read?server_id=${selectedServerId}&path=${targetPath}`,
        { headers: { Authorization: `Bearer ${token}` } }
      );
      let content = res.data;
      if (typeof content !== 'string') content = JSON.stringify(content, null, 2);
      if (/\.json$/i.test(entry.name)) {
        try { content = JSON.stringify(JSON.parse(content), null, 2); } catch {}
      }
      setEditorContent(content);
    } catch (e: unknown) {
      const ae = e as { response?: { data?: { error?: string } } };
      setEditorError(ae?.response?.data?.error || 'read failed');
    } finally {
      setEditorLoading(false);
    }
  };

  const handleSaveFile = async () => {
    if (!selectedServerId || !editingFile) return;
    setEditorLoading(true);
    setEditorError('');
    try {
      const targetPath = currentPath === '/' ? `/${editingFile.name}` : `${currentPath}/${editingFile.name}`;
      const token = Cookies.get('auth-token');

      let content = editorContent;
      if (/\.json$/i.test(editingFile.name)) {
        const parsed = JSON.parse(content);
        content = JSON.stringify(parsed, null, 2);
      }

      await axios.post(
        `/api/plugins/write?server_id=${selectedServerId}&path=${targetPath}`,
        { content },
        { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } }
      );
      setEditingFile(null);
      loadFiles(selectedServerId, currentPath);
    } catch (e: unknown) {
      const ae = e as { response?: { data?: { error?: string } } };
      setEditorError(ae?.response?.data?.error || (e instanceof Error ? e.message : 'save failed'));
    } finally {
      setEditorLoading(false);
    }
  };

  const handleUnzip = async (entry: FileEntry) => {
    if (!selectedServerId) return;
    try {
      const targetPath = currentPath === '/' ? `/${entry.name}` : `${currentPath}/${entry.name}`;
      const token = Cookies.get('auth-token');
      await axios.post(
        `/api/plugins/unzip?server_id=${selectedServerId}&path=${targetPath}`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
      loadFiles(selectedServerId, currentPath);
    } catch (e: unknown) {
      const ae = e as { response?: { data?: { error?: string } } };
      setError(ae?.response?.data?.error || 'unzip failed');
    }
  };

  const handleCreateFolder = async () => {
    if (!selectedServerId || !folderName) return;
    try {
      const targetPath = currentPath === '/' ? `/${folderName}` : `${currentPath}/${folderName}`;
      const params = new URLSearchParams({ server_id: String(selectedServerId), action: 'mkdir', path: targetPath });
      const res = await fetch(`/api/plugins?${params}`, { method: 'POST', body: '{}', headers: authHeaders() });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'mkdir failed');
      setCreatingFolder(false);
      setFolderName('');
      loadFiles(selectedServerId, currentPath);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  const breadcrumbs = currentPath.split('/').filter(Boolean);

  return (
    <div className="w-full max-w-none py-8">
      <div className="mb-6 flex justify-between items-center">
        <h1 className="text-2xl lg:text-3xl font-bold text-gray-900">{t('title')}</h1>
        <div className="flex items-center gap-2">
          <select
            className="border rounded px-3 py-2 text-sm"
            value={selectedServerId ?? ''}
            onChange={(e) => handleServerChange(Number(e.target.value))}
          >
            <option value="">{t('selectServer')}</option>
            {servers.map((s: Server) => (
              <option key={s.id} value={s.id}>{s.identifier} - {s.session_name}</option>
            ))}
          </select>
          {selectedServerId && (
            <Button variant="outline" size="sm" onClick={() => loadFiles(selectedServerId, currentPath)}>
              <RefreshCw className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">{error}</div>
      )}

      {!selectedServerId && (
        <div className="text-center py-20 text-gray-400">{t('selectServerHint')}</div>
      )}

      {selectedServerId && (
        <>
          {/* Breadcrumbs + actions */}
          <div className="flex items-center justify-between mb-4">
            <nav className="flex items-center gap-1 text-sm">
              <button onClick={() => setCurrentPath('/')} className="hover:text-blue-600 flex items-center gap-1">
                <Home className="h-4 w-4" /> {t('root')}
              </button>
              {breadcrumbs.map((part, i) => (
                <span key={i} className="flex items-center gap-1">
                  <ChevronRight className="h-3 w-3 text-gray-400" />
                  <button onClick={() => navigateBreadcrumb(i + 1)} className="hover:text-blue-600">
                    {part}
                  </button>
                </span>
              ))}
            </nav>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={() => fileInputRef.current?.click()} disabled={uploading}>
                <Upload className="h-4 w-4 mr-1" /> {t('upload')}
              </Button>
              <input
                ref={fileInputRef}
                type="file"
                multiple
                className="hidden"
                onChange={(e) => e.target.files && handleUpload(e.target.files)}
              />
              <Button variant="outline" size="sm" onClick={() => setCreatingFolder(true)}>
                <FolderPlus className="h-4 w-4 mr-1" /> {t('newFolder')}
              </Button>
            </div>
          </div>

          {/* Drag & drop zone */}
          <div
            ref={dropRef}
            className={`border-2 border-dashed rounded-lg min-h-[300px] ${dragOver ? 'border-blue-500 bg-blue-50' : 'border-gray-300'} ${uploading ? 'opacity-50' : ''}`}
            onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
            onDragLeave={() => setDragOver(false)}
            onDrop={handleDrop}
          >
            {loading ? (
              <div className="flex items-center justify-center py-20">
                <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
              </div>
            ) : files.length === 0 ? (
              <div className="text-center py-20 text-gray-400">
                <Upload className="h-12 w-12 mx-auto mb-2 opacity-30" />
                <p>{t('empty')}</p>
                <p className="text-sm mt-1">{t('dragDropHint')}</p>
              </div>
            ) : (
              <div className="divide-y">
                {/* Parent directory link */}
                {currentPath !== '/' && (
                  <div
                    className="flex items-center px-4 py-3 hover:bg-gray-50 cursor-pointer"
                    onClick={() => {
                      const parts = currentPath.split('/').filter(Boolean);
                      setCurrentPath('/' + parts.slice(0, -1).join('/'));
                    }}
                  >
                    <Folder className="h-5 w-5 text-gray-400 mr-3" />
                    <span className="text-gray-600">..</span>
                  </div>
                )}

                {/* File list */}
                {files.map((entry) => (
                  <div
                    key={entry.name}
                    className="flex items-center px-4 py-3 hover:bg-gray-50 group"
                  >
                    {renaming === entry.name ? (
                      <div className="flex-1 flex items-center gap-2">
                        {entry.is_dir ? <Folder className="h-5 w-5 text-yellow-500" /> : <File className="h-5 w-5 text-gray-400" />}
                        <input
                          className="border rounded px-2 py-1 text-sm flex-1 max-w-xs"
                          value={renameValue}
                          onChange={(e) => setRenameValue(e.target.value)}
                          onKeyDown={(e) => e.key === 'Enter' && handleRenameConfirm()}
                          autoFocus
                        />
                        <Button size="sm" onClick={handleRenameConfirm}>{tCommon('save')}</Button>
                        <Button size="sm" variant="outline" onClick={() => setRenaming(null)}>{tCommon('cancel')}</Button>
                      </div>
                    ) : (
                      <>
                        <div
                          className="flex-1 flex items-center cursor-pointer"
                          onClick={() => entry.is_dir && navigateToDir(entry.name)}
                        >
                          {entry.is_dir ? (
                            <Folder className="h-5 w-5 text-yellow-500 mr-3 flex-shrink-0" />
                          ) : (
                            <File className="h-5 w-5 text-gray-400 mr-3 flex-shrink-0" />
                          )}
                          <span className={`text-sm ${entry.is_dir ? 'font-medium' : ''}`}>{entry.name}</span>
                        </div>
                        <span className="text-xs text-gray-400 w-20 text-right">{entry.is_dir ? '-' : formatSize(entry.size)}</span>
                        <span className="text-xs text-gray-400 w-40 text-right hidden md:block">{formatTime(entry.mod_time)}</span>
                        <div className="flex items-center gap-1 ml-2 opacity-0 group-hover:opacity-100">
                          {!entry.is_dir && (
                            <a
                              href={`/api/plugins/download?server_id=${selectedServerId}&path=${currentPath === '/' ? '/' + entry.name : currentPath + '/' + entry.name}`}
                              className="p-1 hover:bg-gray-200 rounded"
                              title={tCommon('download')}
                            >
                              <Download className="h-4 w-4 text-gray-500" />
                            </a>
                          )}
                          {!entry.is_dir && /\.zip$/i.test(entry.name) && (
                            <button className="p-1 hover:bg-gray-200 rounded" onClick={() => handleUnzip(entry)} title={t('extract')}>
                              <FileArchive className="h-4 w-4 text-orange-500" />
                            </button>
                          )}
                          {!entry.is_dir && isEditableFile(entry.name) && (
                            <button className="p-1 hover:bg-gray-200 rounded" onClick={() => handleEditFile(entry)} title={t('edit')}>
                              <Edit3 className="h-4 w-4 text-blue-500" />
                            </button>
                          )}
                          {entry.is_dir && (
                            <a
                              href={`/api/plugins/zip-download?server_id=${selectedServerId}&path=${currentPath === '/' ? '/' + entry.name : currentPath + '/' + entry.name}`}
                              className="p-1 hover:bg-gray-200 rounded"
                              title={t('downloadZip')}
                            >
                              <FileDown className="h-4 w-4 text-green-600" />
                            </a>
                          )}
                          <button className="p-1 hover:bg-gray-200 rounded" onClick={() => handleRenameStart(entry)} title={tCommon('rename')}>
                            <Pencil className="h-4 w-4 text-gray-500" />
                          </button>
                          <button className="p-1 hover:bg-gray-200 rounded" onClick={() => handleDelete(entry)} title={tCommon('delete')}>
                            <Trash2 className="h-4 w-4 text-red-500" />
                          </button>
                        </div>
                      </>
                    )}
                  </div>
                ))}
              </div>
            )}

            {uploading && (
              <div className="flex items-center justify-center py-4 border-t bg-blue-50">
                <Loader2 className="h-5 w-5 animate-spin text-blue-500 mr-2" />
                <span className="text-sm text-blue-600">{t('uploading')}</span>
              </div>
            )}
          </div>

          {/* Create folder dialog */}
          {creatingFolder && (
            <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setCreatingFolder(false)}>
              <div className="bg-white rounded-lg p-6 shadow-xl w-96" onClick={(e) => e.stopPropagation()}>
                <h3 className="text-lg font-semibold mb-4">{t('newFolder')}</h3>
                <input
                  className="border rounded w-full px-3 py-2 mb-4"
                  placeholder={t('folderNamePlaceholder')}
                  value={folderName}
                  onChange={(e) => setFolderName(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleCreateFolder()}
                  autoFocus
                />
                <div className="flex justify-end gap-2">
                  <Button variant="outline" onClick={() => setCreatingFolder(false)}>{tCommon('cancel')}</Button>
                  <Button onClick={handleCreateFolder}>{tCommon('create')}</Button>
                </div>
              </div>
            </div>
          )}

          {/* Editor modal */}
          {editingFile && (
            <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setEditingFile(null)}>
              <div className="bg-white rounded-lg shadow-xl w-[800px] max-w-[95vw] max-h-[90vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
                <div className="flex items-center justify-between px-6 py-4 border-b">
                  <h3 className="text-lg font-semibold flex items-center gap-2">
                    <File className="h-5 w-5 text-blue-500" />
                    {t('editing')}: {editingFile.name}
                    {/\.json$/i.test(editingFile.name) && <span className="text-xs bg-blue-100 text-blue-700 px-2 py-0.5 rounded">JSON</span>}
                    {/\.ini$/i.test(editingFile.name) && <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded">INI</span>}
                  </h3>
                  <button onClick={() => setEditingFile(null)} className="p-1 hover:bg-gray-200 rounded">
                    <X className="h-5 w-5" />
                  </button>
                </div>
                {editorError && (
                  <div className="mx-6 mt-3 bg-red-50 border border-red-200 text-red-700 px-4 py-2 rounded text-sm">{editorError}</div>
                )}
                <div className="flex-1 overflow-auto p-6">
                  {editorLoading ? (
                    <div className="flex items-center justify-center py-20">
                      <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
                    </div>
                  ) : (
                    <textarea
                      className="w-full h-[50vh] font-mono text-sm border rounded p-3 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                      value={editorContent}
                      onChange={(e) => setEditorContent(e.target.value)}
                      spellCheck={false}
                    />
                  )}
                </div>
                <div className="flex justify-end gap-2 px-6 py-4 border-t">
                  <Button variant="outline" onClick={() => setEditingFile(null)}>{tCommon('cancel')}</Button>
                  <Button onClick={handleSaveFile} disabled={editorLoading}>
                    <Save className="h-4 w-4 mr-1" />
                    {tCommon('save')}
                  </Button>
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
