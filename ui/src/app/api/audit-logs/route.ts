import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

const getApiBase = () => process.env.NEXT_PUBLIC_API_BASE;

export async function GET(request: Request) {
  const headersList = await headers();
  const authorization = headersList.get('authorization');

  if (!authorization) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const config = {
    headers: { Authorization: authorization, 'Content-Type': 'application/json' },
  };

  try {
    const url = new URL(request.url);
    const queryString = url.searchParams.toString();
    const response = await axios.get(`${getApiBase()}/audit-logs?${queryString}`, config);
    return NextResponse.json(response.data);
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { error?: string }, status?: number } };
    return NextResponse.json({
      error: axiosError.response?.data?.error || 'Failed to fetch audit logs',
    }, { status: axiosError.response?.status || 500 });
  }
}
