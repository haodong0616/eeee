import { NextRequest, NextResponse } from 'next/server';

// 后端API地址
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

export async function GET(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) {
  const params = await context.params;
  const path = params.path.join('/');
  const searchParams = request.nextUrl.searchParams.toString();
  const url = `${BACKEND_URL}/api/${path}${searchParams ? `?${searchParams}` : ''}`;

  console.log('🔵 GET API Proxy:', url);

  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
        'Content-Type': 'application/json',
      },
    });

    const data = await response.json();
    console.log('✅ GET Response:', response.status);
    return NextResponse.json(data, { status: response.status });
  } catch (error: any) {
    console.error('❌ GET API Proxy Error:', error.message);
    return NextResponse.json(
      { error: 'Failed to fetch from backend' },
      { status: 500 }
    );
  }
}

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) {
  const params = await context.params;
  const path = params.path.join('/');
  const url = `${BACKEND_URL}/api/${path}`;
  const body = await request.text();

  console.log('🟢 POST API Proxy:', url);
  console.log('📦 Body:', body.slice(0, 200));

  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
        'Content-Type': 'application/json',
      },
      body,
    });

    const data = await response.json();
    console.log('✅ POST Response:', response.status);
    return NextResponse.json(data, { status: response.status });
  } catch (error: any) {
    console.error('❌ POST API Proxy Error:', error.message);
    return NextResponse.json(
      { error: 'Failed to fetch from backend' },
      { status: 500 }
    );
  }
}

export async function PUT(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) {
  const params = await context.params;
  const path = params.path.join('/');
  const url = `${BACKEND_URL}/api/${path}`;
  const body = await request.text();

  console.log('🟡 PUT API Proxy:', url);

  try {
    const response = await fetch(url, {
      method: 'PUT',
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
        'Content-Type': 'application/json',
      },
      body,
    });

    const data = await response.json();
    console.log('✅ PUT Response:', response.status);
    return NextResponse.json(data, { status: response.status });
  } catch (error: any) {
    console.error('❌ PUT API Proxy Error:', error.message);
    return NextResponse.json(
      { error: 'Failed to fetch from backend' },
      { status: 500 }
    );
  }
}

export async function DELETE(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) {
  const params = await context.params;
  const path = params.path.join('/');
  const url = `${BACKEND_URL}/api/${path}`;

  console.log('🔴 DELETE API Proxy:', url);

  try {
    const response = await fetch(url, {
      method: 'DELETE',
      headers: {
        'Authorization': request.headers.get('Authorization') || '',
        'Content-Type': 'application/json',
      },
    });

    const data = await response.json();
    console.log('✅ DELETE Response:', response.status);
    return NextResponse.json(data, { status: response.status });
  } catch (error: any) {
    console.error('❌ DELETE API Proxy Error:', error.message);
    return NextResponse.json(
      { error: 'Failed to fetch from backend' },
      { status: 500 }
    );
  }
}
