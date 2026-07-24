import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

const getApiBase = () => process.env.NEXT_PUBLIC_API_BASE;

export async function GET(request: Request) {
  try {
    const h = await headers();
    const auth = h.get('authorization');
    if (!auth) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    const url = new URL(request.url);
    const res = await axios.get(`${getApiBase()}/plugins?${url.searchParams.toString()}`, {
      headers: { Authorization: auth },
    });
    return NextResponse.json(res.data);
  } catch (e: unknown) {
    const ae = e as { response?: { data?: { error?: string }, status?: number } };
    return NextResponse.json(
      { error: ae?.response?.data?.error || ' ' },
      { status: ae?.response?.status || 500 }
    );
  }
}

export async function POST(request: Request) {
  try {
    const h = await headers();
    const auth = h.get('authorization');
    if (!auth) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    const url = new URL(request.url);
    const action = url.searchParams.get('action') || 'upload';
    const serverId = url.searchParams.get('server_id');

    if (action === 'upload') {
      const fd = await request.formData();
      const destPath = url.searchParams.get('path') || '/';
      const res = await axios.post(
        `${getApiBase()}/plugins/upload?server_id=${serverId}&path=${destPath}`,
        fd,
        { headers: { Authorization: auth } }
      );
      return NextResponse.json(res.data);
    }

    const body = await request.json();
    const ep = action === 'mkdir' ? 'mkdir' : action === 'rename' ? 'rename' : 'upload';
    const params = new URLSearchParams({ server_id: serverId || '' });
    for (const k of ['path', 'old_path', 'new_path']) {
      if (url.searchParams.has(k)) params.set(k, url.searchParams.get(k)!);
    }
    const res = await axios.post(`${getApiBase()}/plugins/${ep}?${params}`, body, {
      headers: { Authorization: auth, 'Content-Type': 'application/json' },
    });
    return NextResponse.json(res.data);
  } catch (e: unknown) {
    const ae = e as { response?: { data?: { error?: string }, status?: number } };
    return NextResponse.json(
      { error: ae?.response?.data?.error || ' ' },
      { status: ae?.response?.status || 500 }
    );
  }
}

export async function DELETE(request: Request) {
  try {
    const h = await headers();
    const auth = h.get('authorization');
    if (!auth) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    const url = new URL(request.url);
    const res = await axios.delete(`${getApiBase()}/plugins/delete?${url.searchParams.toString()}`, {
      headers: { Authorization: auth },
    });
    return NextResponse.json(res.data);
  } catch (e: unknown) {
    const ae = e as { response?: { data?: { error?: string }, status?: number } };
    return NextResponse.json(
      { error: ae?.response?.data?.error || ' ' },
      { status: ae?.response?.status || 500 }
    );
  }
}
