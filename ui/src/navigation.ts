// Use standard Next.js navigation, no longer depends on next-intl path navigation
export { redirect } from 'next/navigation';
export { usePathname, useRouter } from 'next/navigation';

// Re-export Next.js Link component
import NextLink from 'next/link';
export const Link = NextLink;

export const locales = ['en', 'zh'] as const;
