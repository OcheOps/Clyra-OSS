import { usePublicConfig } from '@/features/auth/hooks/usePublicConfig';

// OAuthButtons — v2 dark restyle.
//
// Renders the Google / GitHub pair above the email form on Login + Signup.
// Full-page navigation (not fetch) because the OAuth spec requires
// top-level redirects.
//
// The "or use email" divider is the centerpiece — a hairline with a
// floating mono label, matching the eyebrow voice from the landing
// page. Buttons themselves are dark surfaces with subtle borders that
// brighten on hover; no glow (that's reserved for primary CTAs only).
export function OAuthButtons({ label = 'Continue with' }: { label?: string }) {
  const { googleEnabled, githubEnabled, loading } = usePublicConfig();

  if (loading || (!googleEnabled && !githubEnabled)) return null;

  return (
    <div className="space-y-2.5">
      {googleEnabled && (
        <a
          href="/api/v1/auth/google/start"
          className="group flex w-full items-center justify-center gap-2.5 rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2.5 text-sm font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-elevated/40"
        >
          <GoogleIcon />
          <span>{label} Google</span>
        </a>
      )}
      {githubEnabled && (
        <a
          href="/api/v1/auth/github/start"
          className="group flex w-full items-center justify-center gap-2.5 rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2.5 text-sm font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-elevated/40"
        >
          <GitHubIcon />
          <span>{label} GitHub</span>
        </a>
      )}
      <div className="relative py-2">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-c2-hairline" />
        </div>
        <div className="relative flex justify-center">
          <span className="bg-c2-surface/80 px-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            or use email
          </span>
        </div>
      </div>
    </div>
  );
}

function GoogleIcon() {
  return (
    <svg className="h-4 w-4" viewBox="0 0 48 48" aria-hidden="true">
      <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z" />
      <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z" />
      <path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z" />
      <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z" />
    </svg>
  );
}

function GitHubIcon() {
  return (
    <svg className="h-4 w-4 text-c2-text" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.438 9.8 8.205 11.385.6.113.82-.26.82-.577 0-.285-.01-1.04-.015-2.04-3.338.725-4.042-1.61-4.042-1.61-.546-1.387-1.333-1.757-1.333-1.757-1.09-.744.083-.729.083-.729 1.205.085 1.84 1.237 1.84 1.237 1.07 1.835 2.807 1.305 3.492.998.108-.776.42-1.305.763-1.605-2.665-.304-5.466-1.332-5.466-5.93 0-1.31.469-2.38 1.236-3.22-.124-.303-.536-1.524.117-3.176 0 0 1.008-.322 3.3 1.23a11.5 11.5 0 013.003-.404c1.02.005 2.047.138 3.003.404 2.29-1.552 3.297-1.23 3.297-1.23.655 1.652.243 2.873.12 3.176.77.84 1.235 1.91 1.235 3.22 0 4.61-2.805 5.624-5.475 5.921.43.37.816 1.1.816 2.22 0 1.606-.015 2.898-.015 3.293 0 .32.216.695.825.577C20.565 21.795 24 17.295 24 12c0-6.63-5.37-12-12-12z" />
    </svg>
  );
}
