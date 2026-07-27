import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

// Builds a Content-Disposition header that cannot be broken out of: control
// characters, quotes and backslashes are stripped, non-ASCII is carried by the
// RFC 5987 filename* form.
function contentDisposition(rawName: string) {
  const cleaned = rawName.replace(/[\u0000-\u001f\u007f"\\]/g, '').trim() || 'download';
  const ascii = cleaned.replace(/[^\u0020-\u007e]/g, '_') || 'download';
  const encoded = encodeURIComponent(cleaned).replace(
    /['()*]/g,
    (c) => `%${c.charCodeAt(0).toString(16).toUpperCase()}`
  );
  return `attachment; filename="${ascii}"; filename*=UTF-8''${encoded}`;
}

export async function GET(request: Request) {
  try {
    const h = await headers();
    const auth = h.get('authorization');
    if (!auth) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const url = new URL(request.url);
    const apiBase = process.env.NEXT_PUBLIC_API_BASE;
    const response = await axios.get(`${apiBase}/plugins/download?${url.searchParams.toString()}`, {
      headers: { Authorization: auth },
      responseType: 'stream',
    });

    const fileName = url.searchParams.get('path')?.split('/').pop() || 'download';
    return new NextResponse(response.data, {
      headers: {
        'Content-Disposition': contentDisposition(fileName),
        'Content-Type': 'application/octet-stream',
      },
    });
  } catch {
    return NextResponse.json({ error: 'download failed' }, { status: 500 });
  }
}
