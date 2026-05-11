export async function requestFaucet(address: string): Promise<{ status?: string; hash?: string; amount?: number; error?: string; retryAfter?: number; retryAfterH?: string }> {
  const res = await fetch('/api/v1/faucet', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ address }),
  });
  return res.json();
}
