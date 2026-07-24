import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

export async function POST(request: Request) {
  try {
    const h = await headers();
    const auth = h.get('authorization');
    if (!auth) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    const url = new URL(request.url);
    const res = await axios.post(
      `${process.env.NEXT_PUBLIC_API_BASE}/plugins/unzip?${url.searchParams.toString()}`,
      {},
      { headers: { Authorization: auth } }
    );
    return NextResponse.json(res.data);
  } catch (e: unknown) {
    const ae = e as { response?: { data?: { error?: string }, status?: number } };
    return NextResponse.json(
      { error: ae?.response?.data?.error || 'unzip failed' },
      { status: ae?.response?.status || 500 }
    );
  }
}
