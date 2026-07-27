import { NextResponse } from 'next/server';
import { headers } from 'next/headers';

const getApiBase = () => process.env.NEXT_PUBLIC_API_BASE;

// Builds a Content-Disposition header that cannot be broken out of: control
// characters, quotes and backslashes are stripped, non-ASCII is carried by the
// RFC 5987 filename* form.
function contentDisposition(rawName: string) {
  const cleaned = rawName.replace(/[\u0000-\u001f\u007f"\\]/g, '').trim() || 'backup';
  const ascii = cleaned.replace(/[^\u0020-\u007e]/g, '_') || 'backup';
  const encoded = encodeURIComponent(cleaned).replace(
    /['()*]/g,
    (c) => `%${c.charCodeAt(0).toString(16).toUpperCase()}`
  );
  return `attachment; filename="${ascii}"; filename*=UTF-8''${encoded}`;
}

export async function GET(_request: Request, { params }: { params: Promise<{ filename: string }> }) {
  const { filename } = await params;
  const headersList = await headers();
  const authorization = headersList.get('authorization');

  if (!authorization) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const url = `${getApiBase()}/backups/download/${encodeURIComponent(filename)}`;
    const response = await fetch(url, {
      headers: { Authorization: authorization },
      cache: 'no-store',
    });

    if (!response.ok) {
      let message = 'Failed to download backup';
      try {
        const data: unknown = await response.json();
        if (data && typeof data === 'object' && 'error' in data) {
          const { error } = data as { error?: unknown };
          if (typeof error === 'string' && error) message = error;
        }
      } catch {
        // backend did not return JSON, keep the generic message
      }
      return NextResponse.json({ error: message }, { status: response.status });
    }

    // Stream the archive straight through instead of buffering it in memory.
    return new NextResponse(response.body, {
      status: response.status,
      headers: {
        'Content-Type': response.headers.get('content-type') || 'application/octet-stream',
        'Content-Disposition':
          response.headers.get('content-disposition') || contentDisposition(filename),
      },
    });
  } catch {
    return NextResponse.json({ error: 'Failed to download backup' }, { status: 500 });
  }
}
