import { NextRequest, NextResponse } from 'next/server';
import { defaultLocale } from './lib/locale-server';

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  
  // Check for language switch request
  const searchParams = request.nextUrl.searchParams;
  const localeParam = searchParams.get('locale');

  if (localeParam && ['en', 'zh'].includes(localeParam)) {
    // Set new language cookie
    const response = NextResponse.redirect(new URL(request.nextUrl.pathname, request.url));
    response.cookies.set('NEXT_LOCALE', localeParam, {
      maxAge: 365 * 24 * 60 * 60, // 1 year
      path: '/',
      sameSite: 'lax'
    });
    return response;
  }

  // Ensure language cookie exists
  const currentLocale = request.cookies.get('NEXT_LOCALE')?.value;
  if (!currentLocale || !['en', 'zh'].includes(currentLocale)) {
    const response = NextResponse.next();
    response.cookies.set('NEXT_LOCALE', defaultLocale, {
      maxAge: 365 * 24 * 60 * 60, // 1 year
      path: '/',
      sameSite: 'lax'
    });
    return response;
  }

  // Auth check
  const token = request.cookies.get('auth-token')?.value;
  const isAuthPage = pathname.startsWith('/login') || pathname.startsWith('/init');
  const isProtectedPage = pathname.startsWith('/home') || pathname.startsWith('/servers') || pathname.startsWith('/plugins');
  const isRootPage = pathname === '/';

  // Redirect to login when accessing protected pages without a token
  if (isProtectedPage && !token) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  // Redirect to home when accessing root with a token
  if (isRootPage && token) {
    return NextResponse.redirect(new URL('/home', request.url));
  }

  // Redirect to login when accessing root without a token
  if (isRootPage && !token) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  // Redirect to home when accessing auth pages while logged in
  if (isAuthPage && token) {
    return NextResponse.redirect(new URL('/home', request.url));
  }

  return NextResponse.next();
}

export const config = {
  // Match all paths except static files and API routes
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ]
};
