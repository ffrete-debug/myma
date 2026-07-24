import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

const getApiBase = () => process.env.NEXT_PUBLIC_API_BASE;

export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const headersList = await headers();
  const authorization = headersList.get('authorization');

  if (!authorization) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const config = {
    headers: { Authorization: authorization },
  };

  try {
    const url = new URL(`${getApiBase()}/servers/${id}/logs`);
    url.searchParams.set('tail', new URL(request.url).searchParams.get('tail') || '200');
    const response = await axios.get(url.toString(), config);
    return NextResponse.json(response.data);
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { error?: string }, status?: number } };
    return NextResponse.json({
      error: axiosError.response?.data?.error || 'Failed to fetch logs'
    }, { status: axiosError.response?.status || 500 });
  }
}
