import { NextResponse } from 'next/server';
import axios from 'axios';
import { cookies } from 'next/headers';

const getApiBase = () => process.env.NEXT_PUBLIC_API_BASE;

export async function GET() {
  const cookieStore = await cookies();
  const token = cookieStore.get('auth-token');

  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const config = {
    headers: { Authorization: `Bearer ${token.value}`, 'Content-Type': 'application/json' },
  };

  try {
    const url = `${getApiBase()}/images/check-updates`;
    const response = await axios.get(url, config);
    return NextResponse.json(response.data);
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { error?: string }, status?: number } };
    return NextResponse.json({
      error: axiosError.response?.data?.error || 'Failed to check updates'
    }, { status: axiosError.response?.status || 500 });
  }
}
