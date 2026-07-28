import { NextRequest, NextResponse } from 'next/server';
import { locales, defaultLocale, type Locale } from './lib/locale';

const LOCALE_COOKIE_OPTIONS = {
  maxAge: 365 * 24 * 60 * 60, // 1 year
  path: '/',
  sameSite: 'lax'
} as const;

// Pages rendered by app/(protected)/. The route group name is stripped from the
// URL and the filesystem is not readable from the edge runtime, so this list
// cannot be derived at runtime - it MUST be kept in sync with app/(protected)/.
const protectedPaths = ['/audit-logs', '/home', '/metrics', '/mods', '/players', '/plugins', '/servers'];

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Check for language switch request
  const searchParams = request.nextUrl.searchParams;
  const localeParam = searchParams.get('locale');

  if (localeParam && locales.includes(localeParam as Locale)) {
    // Set new language cookie
    const response = NextResponse.redirect(new URL(request.nextUrl.pathname, request.url));
    response.cookies.set('NEXT_LOCALE', localeParam, LOCALE_COOKIE_OPTIONS);
    return response;
  }

  // Ensure language cookie exists. This must not return early: the cookie is
  // attached to whichever response the auth check below produces, otherwise a
  // request without NEXT_LOCALE (fresh browser, curl, cleared cookies) would
  // skip the auth check entirely.
  const currentLocale = request.cookies.get('NEXT_LOCALE')?.value;
  const needsLocaleCookie = !currentLocale || !locales.includes(currentLocale as Locale);
  const withLocale = (response: NextResponse) => {
    if (needsLocaleCookie) {
      response.cookies.set('NEXT_LOCALE', defaultLocale, LOCALE_COOKIE_OPTIONS);
    }
    return response;
  };

  // Auth check
  const token = request.cookies.get('auth-token')?.value;
  const isAuthPage = pathname.startsWith('/login') || pathname.startsWith('/init');
  const isProtectedPage = protectedPaths.some((path) => pathname === path || pathname.startsWith(`${path}/`));
  const isRootPage = pathname === '/';

  // Redirect to login when accessing protected pages without a token
  if (isProtectedPage && !token) {
    return withLocale(NextResponse.redirect(new URL('/login', request.url)));
  }

  // Redirect to home when accessing root with a token
  if (isRootPage && token) {
    return withLocale(NextResponse.redirect(new URL('/home', request.url)));
  }

  // Redirect to login when accessing root without a token
  if (isRootPage && !token) {
    return withLocale(NextResponse.redirect(new URL('/login', request.url)));
  }

  // Redirect to home when accessing auth pages while logged in
  if (isAuthPage && token) {
    return withLocale(NextResponse.redirect(new URL('/home', request.url)));
  }

  return withLocale(NextResponse.next());
}

export const config = {
  // Match all paths except static files and API routes
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ]
};
