import { NextResponse } from 'next/server';
import axios from 'axios';
import { headers } from 'next/headers';

const getApiBase = () => process.env.NEXT_PUBLIC_API_BASE;

export async function POST(request: Request) {
  const headersList = await headers();
  const authorization = headersList.get('authorization');

  if (!authorization) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const body = await request.json();
    const { image_name } = body;

    if (!image_name) {
      return NextResponse.json({ error: 'Image name cannot be empty' }, { status: 400 });
    }

    const config = {
      headers: { Authorization: authorization, 'Content-Type': 'application/json' },
    };

    const url = `${getApiBase()}/images/pull`;
    const response = await axios.post(url, { image_name }, config);
    return NextResponse.json(response.data);
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { error?: string }, status?: number } };
    return NextResponse.json({
      error: axiosError.response?.data?.error || 'Failed to pull image'
    }, { status: axiosError.response?.status || 500 });
  }
}
