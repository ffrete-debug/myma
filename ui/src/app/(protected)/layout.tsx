"use client";

import { useEffect } from 'react';
import { useTranslations } from 'next-intl';
import { Link, usePathname, useRouter } from '@/navigation';
import { useIsAuthenticated, useAuthActions, useAuthIsInitialized } from '@/stores/auth';
import { LogOut } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function ProtectedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const isAuthenticated = useIsAuthenticated();
  const isInitialized = useAuthIsInitialized();
  const { initFromStorage, logout } = useAuthActions();
  const t = useTranslations('navigation');
  const pathname = usePathname();

  useEffect(() => {
    if (!isInitialized) initFromStorage();
  }, [initFromStorage, isInitialized]);

  useEffect(() => {
    if (isInitialized && !isAuthenticated) router.replace('/login');
  }, [isAuthenticated, isInitialized, router]);

  if (!isInitialized) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Initializing...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Verifying...</p>
        </div>
      </div>
    );
  }

  const navLink = (href: string, label: string, active: boolean) => (
    <Link
      href={href}
      className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
        active ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
      }`}
    >
      {label}
    </Link>
  );

  return (
    <div className="min-h-screen bg-background">
      <header className="bg-card border-b border-border sticky top-0 z-50">
        <div className="w-full px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-14">
            <div className="flex items-center gap-6">
              <h1 className="text-lg font-semibold text-foreground">ARK Server Manager</h1>
              <nav className="flex items-center gap-1">
                {navLink('/home', t('home'), pathname === '/home')}
                {navLink('/servers', t('servers'), pathname.startsWith('/servers'))}
                {navLink('/plugins', t('plugins'), pathname.startsWith('/plugins'))}
              </nav>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="ghost" size="sm" className="h-8 text-muted-foreground hover:text-destructive" onClick={logout}>
                <LogOut className="h-4 w-4 mr-1" /> Log out
              </Button>
            </div>
          </div>
        </div>
      </header>
      <main className="w-full py-4 px-4 sm:px-6 lg:px-8">{children}</main>
    </div>
  );
}
