import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

const getApiBase = () => process.env.NEXT_PUBLIC_API_BASE;

async function auth() {
  const headersList = await headers();
  return headersList.get('authorization');
}

function fail(error: unknown) {
  const e = error as { response?: { data?: { error?: string }, status?: number } };
  return NextResponse.json(
    { error: e.response?.data?.error || 'Request failed' },
    { status: e.response?.status || 500 },
  );
}

export async function PUT(
  request: Request,
  { params }: { params: Promise<{ id: string; workshopId: string }> },
) {
  const authorization = await auth();
  if (!authorization) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });

  const { id, workshopId } = await params;
  try {
    const body = await request.json();
    const res = await axios.put(
      `${getApiBase()}/servers/${id}/mods/${encodeURIComponent(workshopId)}/enabled`,
      body,
      { headers: { Authorization: authorization, 'Content-Type': 'application/json' } },
    );
    return NextResponse.json(res.data);
  } catch (error: unknown) {
    return fail(error);
  }
}
