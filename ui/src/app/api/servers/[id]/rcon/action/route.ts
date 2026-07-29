import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

const getApiBase = () => process.env.NEXT_PUBLIC_API_BASE;

function fail(error: unknown) {
  const e = error as { response?: { data?: { error?: string; message?: string; detail?: string }, status?: number } };
  const d = e.response?.data;
  return NextResponse.json(
    { error: d?.error || d?.detail || d?.message || 'Request failed' },
    { status: e.response?.status || 500 },
  );
}

export async function POST(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const authorization = (await headers()).get('authorization');
  if (!authorization) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });

  const { id } = await params;
  try {
    const body = await request.json();
    const res = await axios.post(`${getApiBase()}/servers/${id}/rcon/action`, body, {
      headers: { Authorization: authorization, 'Content-Type': 'application/json' },
    });
    return NextResponse.json(res.data);
  } catch (error: unknown) {
    return fail(error);
  }
}
