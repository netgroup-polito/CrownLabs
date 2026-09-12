// The instance gate redirects while the session is invalid and its cookies are
// HttpOnly, so asking the server is the only way to know. `redirect: 'manual'`
// keeps the request same-origin, so the check is never blocked by CORS.
export async function hasValidInstanceSession(
  envUrl: string,
): Promise<boolean> {
  try {
    const res = await fetch(envUrl, {
      credentials: 'include',
      redirect: 'manual',
      cache: 'no-store',
    });
    return res.type !== 'opaqueredirect';
  } catch {
    return false;
  }
}

// Only a real navigation can complete the OIDC redirect chain, since fetch is
// blocked by CORS at the identity provider, and only fetch can observe its
// outcome: the opened window performs the handshake while this page polls the
// gate until it lets the request through. Resolves with the window once the
// session is valid, or null if it was blocked or closed by the user.
export function refreshInstanceSession(envUrl: string): Promise<Window | null> {
  const sessionWindow = window.open(envUrl, '_blank');
  if (!sessionWindow) return Promise.resolve(null);

  return new Promise(resolve => {
    const poll = async () => {
      if (sessionWindow.closed) return resolve(null);
      if (await hasValidInstanceSession(envUrl)) return resolve(sessionWindow);
      window.setTimeout(poll, 500);
    };
    poll();
  });
}
