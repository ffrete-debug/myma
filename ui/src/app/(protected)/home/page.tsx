"use client";

import { isAxiosError } from 'axios';
import { useTranslations } from 'next-intl';
import { Card, CardContent } from '@/components/ui/card';
import { Link } from '@/navigation';
import { useAuthUser } from '@/stores/auth';
import useServersStore, { useImageStatus, serversActions } from '@/stores/servers';
import { ImageStatus } from '@/components/docker/ImageStatus';
import { ImageUpdateConfirmModal } from '@/components/docker/ImageUpdateConfirmModal';
import { useEffect, useRef, useCallback, useState } from 'react';
import { Server, Monitor, Activity, Users } from 'lucide-react';
import Cookies from 'js-cookie';

export default function HomePage() {
    const t = useTranslations('home');
    const profile = useAuthUser();
    const imageStatus = useImageStatus();
    const { getImageStatus } = serversActions;
    const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);
    const [updateModalOpen, setUpdateModalOpen] = useState(false);
    const [selectedImage, setSelectedImage] = useState('');

    const refreshImageStatus = useCallback(async () => {
        try { await getImageStatus(); }
        catch (error) {
          // A 401 means the session is dead; the shared interceptor clears it and
          // moves the user on. Logging it here is what filled the console on a
          // fresh install.
          if (!isAxiosError(error) || error.response?.status !== 401) {
            console.error('Failed to get image status:', error);
          }
        }
    }, [getImageStatus]);

    const handleManualDownload = async () => {
        if (!imageStatus?.images) return;
        const notReadyImage = Object.entries(imageStatus.images).find(([, status]) => !status.ready);
        if (!notReadyImage) { console.log('All images are ready'); return; }
        await handleDownloadImage(notReadyImage[0]);
    };

    const handleDownloadImage = async (imageName: string) => {
        try {
            const token = Cookies.get('auth-token');
            const response = await fetch('/api/images/pull', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
                body: JSON.stringify({ image_name: imageName })
            });
            if (!response.ok) throw new Error('Failed to download image');
            startPolling();
        } catch (error) { console.error('Failed to download image:', error); }
    };

    const handleUpdateImage = (imageName: string) => {
        setSelectedImage(imageName);
        setUpdateModalOpen(true);
    };

    const handleConfirmUpdate = async () => {
        try {
            const token = Cookies.get('auth-token');
            const response = await fetch('/api/images/update', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
                body: JSON.stringify({ image_name: selectedImage })
            });
            if (!response.ok) throw new Error('Failed to update image');
            startPolling();
        } catch (error) {
            console.error('Failed to update image:', error);
        }
        setUpdateModalOpen(false);
    };

    const handleCheckUpdates = async () => {
        try {
            const token = Cookies.get('auth-token');
            const response = await fetch('/api/images/check-updates', {
                headers: { 'Authorization': `Bearer ${token}` },
            });
            if (!response.ok) throw new Error('Failed to check updates');
            await refreshImageStatus();
        } catch (error) { console.error('Failed to check updates:', error); }
    };

    const stopPolling = useCallback(() => {
        if (pollingIntervalRef.current) { clearInterval(pollingIntervalRef.current); pollingIntervalRef.current = null; }
    }, []);

    const startPolling = useCallback(() => {
        stopPolling();
        pollingIntervalRef.current = setInterval(async () => {
            await refreshImageStatus();
            // Read the freshly written value straight from the store: the
            // `imageStatus` binding from this render is frozen in the closure
            // and would never reflect the refresh above.
            if (!useServersStore.getState().imageStatus?.any_pulling) stopPolling();
        }, 2000);
    }, [refreshImageStatus, stopPolling]);

    useEffect(() => {
        refreshImageStatus();
        return () => stopPolling();
    }, [refreshImageStatus, stopPolling]);

    const quickLinks = [
        { href: '/servers', icon: Server, label: t('serverManagement'), desc: t('serverManagementDesc'), color: 'blue' },
        { href: '/players', icon: Users, label: t('playerManagement'), desc: t('playerManagementDesc'), color: 'blue' },
        { href: '/servers', icon: Activity, label: t('logMonitoring'), desc: t('logMonitoringDesc'), color: 'green' },
    ];

    return (
        <div className="h-[calc(100vh-4rem)] flex flex-col gap-4 py-4">
            {/* Top row: welcome + image status */}
            <div className="flex gap-4 flex-1 min-h-0">
                {/* Welcome card */}
                <Card className="flex-[2] border-0 shadow-sm">
                    <CardContent className="p-6 h-full flex flex-col justify-center">
                        <div className="flex items-center gap-4">
                            <div className="w-14 h-14 bg-gradient-to-br from-primary to-primary/70 rounded-2xl flex items-center justify-center shadow-lg">
                                <Monitor className="w-7 h-7 text-primary-foreground" />
                            </div>
                            <div>
                                <h1 className="text-2xl font-bold text-foreground">{t('title')}</h1>
                                <p className="text-sm text-muted-foreground">{t('subtitle')}</p>
                            </div>
                        </div>
                        <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
                            <div className="bg-muted/50 rounded-lg p-3">
                                <span className="text-muted-foreground">{t('username')}</span>
                                <p className="font-semibold text-foreground">{profile?.username}</p>
                            </div>
                            <div className="bg-muted/50 rounded-lg p-3">
                                <span className="text-muted-foreground">{t('userID')}</span>
                                <p className="font-semibold text-foreground">#{profile?.id}</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* Image status card */}
                <Card className="flex-[3] border-0 shadow-sm">
                    <CardContent className="p-6 h-full flex flex-col">
                        <h2 className="section-title mb-3">{t('imageManagement')}</h2>
                        <div className="flex-1 min-h-0 overflow-auto">
                            {imageStatus ? (
                                <ImageStatus
                                    imageStatus={imageStatus}
                                    onRefresh={refreshImageStatus}
                                    onManualDownload={handleManualDownload}
                                    onCheckUpdates={handleCheckUpdates}
                                    onDownloadImage={handleDownloadImage}
                                    onUpdateImage={handleUpdateImage}
                                />
                            ) : (
                                <div className="flex items-center justify-center h-full">
                                    <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
                                    <p className="ml-2 text-sm text-muted-foreground">Loading...</p>
                                </div>
                            )}
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Quick links grid */}
            <div className="grid grid-cols-3 gap-4">
                {quickLinks.map(({ href, icon: Icon, label, desc, color }) => (
                    <Link key={label} href={href}>
                        <Card className="border-0 shadow-sm transition-all duration-200 hover:shadow-md hover:-translate-y-0.5 cursor-pointer">
                            <CardContent className="p-4 flex items-center gap-3">
                                <div className={`w-10 h-10 rounded-xl flex items-center justify-center bg-${color}-100 dark:bg-${color}-950/30`}>
                                    <Icon className={`w-5 h-5 text-${color}-600`} />
                                </div>
                                <div className="min-w-0">
                                    <p className="text-sm font-semibold text-foreground truncate">{label}</p>
                                    <p className="text-xs text-muted-foreground truncate">{desc}</p>
                                </div>
                            </CardContent>
                        </Card>
                    </Link>
                ))}
            </div>
            <ImageUpdateConfirmModal
                imageName={selectedImage}
                isOpen={updateModalOpen}
                onClose={() => setUpdateModalOpen(false)}
                onConfirm={handleConfirmUpdate}
            />
        </div>
    );
}
