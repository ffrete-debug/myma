import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

export async function POST(request: Request) {
  try {
    const h = await headers();
    const auth = h.get('authorization');
    if (!auth) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    const url = new URL(request.url);
    const body = await request.json();
    const res = await axios.post(
      `${process.env.NEXT_PUBLIC_API_BASE}/plugins/write?${url.searchParams.toString()}`,
      body,
      { headers: { Authorization: auth, 'Content-Type': 'application/json' } }
    );
    return NextResponse.json(res.data);
  } catch (e: unknown) {
    const ae = e as { response?: { data?: { error?: string }, status?: number } };
    return NextResponse.json(
      { error: ae?.response?.data?.error || 'write failed' },
      { status: ae?.response?.status || 500 }
    );
  }
}
