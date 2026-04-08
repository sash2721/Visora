const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:3000';

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = localStorage.getItem('token');
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string>),
  };

  if (token) headers['Authorization'] = `Bearer ${token}`;
  // Don't set Content-Type for FormData (browser sets boundary automatically)
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  let res: Response;
  try {
    res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  } catch {
    throw new Error('Unable to connect. Please check your internet and try again.');
  }

  if (!res.ok) {
    // Don't expose internal error details to the user
    const status = res.status;
    if (status === 401) throw new Error('Session expired. Please log in again.');
    if (status === 403) throw new Error('You don\'t have permission to do that.');
    if (status === 404) throw new Error('The requested resource was not found.');
    if (status >= 500) throw new Error('Service temporarily unavailable. Please try again later.');
    throw new Error('Something went wrong. Please try again.');
  }

  const data = await res.json();
  return data as T;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, {
      method: 'POST',
      body: body instanceof FormData ? body : JSON.stringify(body),
    }),
};
