import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

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
        'Content-Disposition': `attachment; filename="${fileName}"`,
        'Content-Type': 'application/octet-stream',
      },
    });
  } catch {
    return NextResponse.json({ error: 'download failed' }, { status: 500 });
  }
}
