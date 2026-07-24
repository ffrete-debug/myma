import { getRequestConfig } from 'next-intl/server';
import { getLocale, locales } from './lib/locale-server';

// Export locales for compatibility
export { locales };

export default getRequestConfig(async () => {
  // Get locale setting from cookie
  const locale = await getLocale();

  return {
    locale,
    messages: (await import(`../messages/${locale}.ts`)).default
  };
});
