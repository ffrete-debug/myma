"use client";

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Alert, AlertDescription } from '@/components/ui/alert';
import axios from 'axios';
import Cookies from 'js-cookie';

interface AffectedServer {
  id: number;
  identifier: string;
  status: string;
}

interface ImageUpdateConfirmModalProps {
  imageName: string;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

export function ImageUpdateConfirmModal({
  imageName,
  isOpen,
  onClose,
  onConfirm,
}: ImageUpdateConfirmModalProps) {
  const t = useTranslations('servers.dockerImages');
  const [affectedServers, setAffectedServers] = useState<AffectedServer[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchAffected = async () => {
    try {
      const token = Cookies.get('auth-token');
      const response = await axios.get<AffectedServer[]>(
        `/api/images/affected`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          params: { image_name: imageName },
        }
      );
      setAffectedServers(response.data);
    } catch {
      // Silently ignore — affected servers list is optional context
    }
  };

  const handleOpen = (open: boolean) => {
    if (open) {
      fetchAffected();
    } else {
      setError('');
      setAffectedServers([]);
    }
  };

  const handleConfirm = async () => {
    setLoading(true);
    setError('');
    try {
      const token = Cookies.get('auth-token');
      const response = await axios.post(
        '/api/images/update',
        { image_name: imageName },
        { headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' } }
      );
      if (response.status >= 400) throw new Error('Update failed');
      onConfirm();
    } catch {
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleOpen}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('updateConfirm')}</DialogTitle>
          <DialogDescription>
            {t('imageInfo')}: <strong>{imageName}</strong>
          </DialogDescription>
        </DialogHeader>

        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {affectedServers.length > 0 && (
          <div className="space-y-2 py-2">
            <h4 className="text-sm font-medium text-foreground">
              {t('affectedServers')}
            </h4>
            <ul className="space-y-1">
              {affectedServers.map((server) => (
                <li key={server.id} className="flex items-center gap-2 text-sm">
                  <span className="font-mono">{server.identifier}</span>
                  <span className="text-muted-foreground text-xs">({server.status})</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        <Alert>
          <AlertDescription>
            <p className="font-medium">{t('updateWarning')}</p>
            <ul className="mt-1 list-disc list-inside text-sm text-muted-foreground space-y-1">
              <li>{t('warningDownloadTime')}</li>
              <li>{t('warningContainerRecreate')}</li>
              <li>{t('warningDataSafety')}</li>
            </ul>
          </AlertDescription>
        </Alert>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {t('cancel') || 'Cancel'}
          </Button>
          <Button onClick={handleConfirm} disabled={loading}>
            {loading && <span className="mr-2 inline-block h-3 w-3 animate-spin rounded-full border-2 border-current border-t-transparent" />}
            {t('confirmUpdate')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
