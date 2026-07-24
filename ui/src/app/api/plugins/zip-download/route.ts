import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

export async function GET(request: Request) {
  try {
    const h = await headers();
    const auth = h.get('authorization');
    if (!auth) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    const url = new URL(request.url);
    const res = await axios.get(
      `${process.env.NEXT_PUBLIC_API_BASE}/plugins/zip-download?${url.searchParams.toString()}`,
      { headers: { Authorization: auth }, responseType: 'arraybuffer' }
    );
    const cd = res.headers['content-disposition'] || 'attachment; filename="plugins.zip"';
    return new NextResponse(res.data, {
      headers: {
        'Content-Type': 'application/zip',
        'Content-Disposition': cd,
      },
    });
  } catch (e: unknown) {
    const ae = e as { response?: { data?: string, status?: number } };
    return NextResponse.json(
      { error: ae?.response?.data || 'zip download failed' },
      { status: ae?.response?.status || 500 }
    );
  }
}
