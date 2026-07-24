"use client";

import { useEffect } from 'react';
import { useRouter } from '@/navigation';
import { useIsAuthenticated, useAuthActions, useAuthIsInitialized } from '@/stores/auth';

export default function HomePage() {
  const router = useRouter();
  const isAuthenticated = useIsAuthenticated();
  const isInitialized = useAuthIsInitialized();
  const { initFromStorage } = useAuthActions();

  useEffect(() => {
    if (!isInitialized) {
      initFromStorage();
    }
  }, [initFromStorage, isInitialized]);

  useEffect(() => {
    // Only redirect after initialization is complete
    if (isInitialized) {
      if (isAuthenticated) {
        router.replace('/home');
      } else {
        router.replace('/login');
      }
    }
  }, [isAuthenticated, isInitialized, router]);

  // Show loading state
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto mb-4"></div>
        <p>{isInitialized ? 'Redirecting...' : 'Loading...'}</p>
      </div>
    </div>
  );
}