import { useCallback, useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import * as mcpApi from '@/services/api/mcpTokens';
import type { MCPToken } from '@/services/api/mcpTokens';

// /settings/mcp panel — mint, list, revoke Personal Access Tokens for
// the MCP server (docs/mcp.md § Phase 1). Cleartext is shown exactly
// once via the CreatedTokenModal; we never store or re-fetch it.
//
// Design principles:
//   1. The 'you won't see this again' warning is the LOUDEST thing in
//      the modal, next to the copy button. Losing a token = revoke +
//      mint again, no recovery path exists (by design).
//   2. Claude Desktop config snippet is right there in the modal — the
//      user should be able to copy the whole JSON block, not just the
//      raw token, so setup is one paste.
//   3. Revoke is destructive; confirm inline with a two-step click
//      instead of a system dialog.

export function MCPTokensSection() {
  const [tokens, setTokens] = useState<MCPToken[] | null>(null);
  const [minting, setMinting] = useState(false);
  const [name, setName] = useState('');
  const [freshToken, setFreshToken] = useState<MCPToken | null>(null);
  const [confirmRevoke, setConfirmRevoke] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await mcpApi.listTokens();
      setTokens(res.tokens);
    } catch (err) {
      console.error('list mcp tokens failed', err);
      toast.error('Could not load MCP tokens');
      setTokens([]);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const handleMint = useCallback(async () => {
    const trimmed = name.trim();
    if (!trimmed) {
      toast.error('Name the token so you remember what it does');
      return;
    }
    setMinting(true);
    try {
      const res = await mcpApi.createToken(trimmed);
      setFreshToken(res.token);
      setName('');
      await load();
    } catch (err) {
      toast.error((err as Error).message || 'Failed to mint token');
    } finally {
      setMinting(false);
    }
  }, [name, load]);

  const handleRevoke = useCallback(
    async (id: string) => {
      try {
        await mcpApi.revokeToken(id);
        toast.success('Token revoked');
        setConfirmRevoke(null);
        await load();
      } catch (err) {
        toast.error((err as Error).message || 'Failed to revoke');
      }
    },
    [load],
  );

  const liveTokens = (tokens ?? []).filter((t) => !t.revoked_at);
  const revokedTokens = (tokens ?? []).filter((t) => t.revoked_at);

  return (
    <section className="rounded-2xl border border-c2-border bg-c2-surface/60 p-5">
      <header className="mb-4">
        <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          MCP · Model Context Protocol
        </p>
        <h2 className="mt-1 font-display text-lg font-semibold tracking-tight text-c2-text">
          API tokens
        </h2>
        <p className="mt-1 text-petit text-c2-text-muted">
          Personal access tokens for the Clyra MCP server. Paste one into
          Claude Desktop, Cursor, or any MCP-capable client to generate
          diagrams from a chat window.
        </p>
      </header>

      <div className="mb-5 flex gap-2">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Token label — e.g. 'claude-desktop'"
          maxLength={64}
          disabled={minting}
          onKeyDown={(e) => {
            if (e.key === 'Enter') void handleMint();
          }}
          className="flex-1 rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2 text-sm text-c2-text placeholder:text-c2-text-faint outline-none transition-colors focus:border-c2-primary/60 focus:bg-c2-surface disabled:opacity-50"
        />
        <button
          type="button"
          onClick={handleMint}
          disabled={minting || !name.trim()}
          className="shrink-0 rounded-lg bg-c2-primary px-4 py-2 text-sm font-medium text-c2-bg shadow-c2-glow-soft transition-all hover:bg-c2-primary-hover hover:shadow-c2-glow disabled:cursor-not-allowed disabled:opacity-50"
        >
          {minting ? 'Generating…' : 'Generate token'}
        </button>
      </div>

      {tokens === null ? (
        <p className="text-petit text-c2-text-faint">Loading…</p>
      ) : liveTokens.length === 0 ? (
        <p className="text-petit text-c2-text-faint">
          No active tokens yet. Generate one to get started.
        </p>
      ) : (
        <ul className="space-y-2">
          {liveTokens.map((t) => (
            <TokenRow
              key={t.id}
              token={t}
              confirmingRevoke={confirmRevoke === t.id}
              onConfirmRevoke={() => setConfirmRevoke(t.id)}
              onCancelRevoke={() => setConfirmRevoke(null)}
              onRevoke={() => void handleRevoke(t.id)}
            />
          ))}
        </ul>
      )}

      {revokedTokens.length > 0 && (
        <details className="mt-4">
          <summary className="cursor-pointer font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint hover:text-c2-text-muted">
            {revokedTokens.length} revoked
          </summary>
          <ul className="mt-2 space-y-1">
            {revokedTokens.map((t) => (
              <li
                key={t.id}
                className="flex items-center justify-between rounded-lg border border-c2-hairline bg-c2-bg/20 px-3 py-1.5 text-tiny text-c2-text-faint"
              >
                <span className="truncate line-through" title={t.name}>{t.name}</span>
                <span className="font-mono text-[10px] uppercase tracking-[0.14em]">
                  revoked · {formatDate(t.revoked_at)}
                </span>
              </li>
            ))}
          </ul>
        </details>
      )}

      {freshToken && (
        <CreatedTokenModal
          token={freshToken}
          onClose={() => setFreshToken(null)}
        />
      )}
    </section>
  );
}

function TokenRow({
  token,
  confirmingRevoke,
  onConfirmRevoke,
  onCancelRevoke,
  onRevoke,
}: {
  token: MCPToken;
  confirmingRevoke: boolean;
  onConfirmRevoke: () => void;
  onCancelRevoke: () => void;
  onRevoke: () => void;
}) {
  const lastUsed = token.last_used
    ? `Used ${formatRelative(token.last_used)}`
    : 'Never used';
  return (
    <li className="flex items-center gap-3 rounded-lg border border-c2-hairline bg-c2-bg/30 px-3 py-2.5">
      <div className="min-w-0 flex-1">
        <p className="truncate text-tiny font-medium text-c2-text" title={token.name}>
          {token.name}
        </p>
        <p className="mt-0.5 font-mono text-[10px] uppercase tracking-[0.14em] text-c2-text-faint">
          Created {formatRelative(token.created_at)} · {lastUsed}
        </p>
      </div>
      {confirmingRevoke ? (
        <div className="flex shrink-0 items-center gap-1">
          <button
            type="button"
            onClick={onRevoke}
            className="rounded-md border border-c2-danger/40 bg-c2-danger/10 px-2 py-1 text-micro font-medium text-c2-danger transition-colors hover:bg-c2-danger/20"
          >
            Confirm revoke
          </button>
          <button
            type="button"
            onClick={onCancelRevoke}
            className="rounded-md px-2 py-1 text-micro text-c2-text-muted transition-colors hover:text-c2-text"
          >
            Cancel
          </button>
        </div>
      ) : (
        <button
          type="button"
          onClick={onConfirmRevoke}
          className="shrink-0 rounded-md px-2 py-1 text-micro font-medium text-c2-text-muted transition-colors hover:bg-c2-danger/10 hover:text-c2-danger"
        >
          Revoke
        </button>
      )}
    </li>
  );
}

function CreatedTokenModal({
  token,
  onClose,
}: {
  token: MCPToken;
  onClose: () => void;
}) {
  const [copiedToken, setCopiedToken] = useState(false);
  const [copiedConfig, setCopiedConfig] = useState(false);

  const configSnippet = JSON.stringify(
    {
      mcpServers: {
        clyra: {
          url: 'https://your-clyra-instance.example.com/mcp',
          headers: {
            Authorization: `Bearer ${token.clear}`,
          },
        },
      },
    },
    null,
    2,
  );

  const copy = async (text: string, which: 'token' | 'config') => {
    try {
      await navigator.clipboard.writeText(text);
      if (which === 'token') {
        setCopiedToken(true);
        setTimeout(() => setCopiedToken(false), 1500);
      } else {
        setCopiedConfig(true);
        setTimeout(() => setCopiedConfig(false), 1500);
      }
    } catch {
      toast.error('Clipboard unavailable — select + copy manually');
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-c2-bg/70 backdrop-blur-sm px-4"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-labelledby="mcp-token-title"
    >
      <div
        className="w-full max-w-lg rounded-2xl border border-c2-border bg-c2-elevated p-6 shadow-c2-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-start gap-3">
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-c2-accent/20 text-c2-accent">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M20 6L9 17l-5-5" />
            </svg>
          </div>
          <div className="min-w-0">
            <h3 id="mcp-token-title" className="font-display text-base font-semibold text-c2-text">
              Token created
            </h3>
            <p className="mt-0.5 text-petit text-c2-text-muted">
              Copy it now. For security, the full token is only shown this once.
              If you lose it, revoke + generate a new one.
            </p>
          </div>
        </div>

        <div className="mb-4">
          <label className="mb-1.5 block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            Token
          </label>
          <div className="flex items-center gap-2">
            <code className="flex-1 truncate rounded-lg border border-c2-hairline bg-c2-bg/60 px-3 py-2 font-mono text-xs text-c2-text">
              {token.clear}
            </code>
            <button
              type="button"
              onClick={() => void copy(token.clear ?? '', 'token')}
              className="shrink-0 rounded-lg bg-c2-primary px-3 py-2 text-xs font-medium text-c2-bg transition-all hover:bg-c2-primary-hover"
            >
              {copiedToken ? 'Copied' : 'Copy'}
            </button>
          </div>
        </div>

        <div className="mb-4">
          <label className="mb-1.5 flex items-center justify-between font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            <span>Claude Desktop config</span>
            <button
              type="button"
              onClick={() => void copy(configSnippet, 'config')}
              className="rounded-md px-2 py-0.5 text-c2-primary transition-colors hover:bg-c2-primary/10"
            >
              {copiedConfig ? 'Copied' : 'Copy JSON'}
            </button>
          </label>
          <pre className="overflow-x-auto rounded-lg border border-c2-hairline bg-c2-bg/60 p-3 font-mono text-[11px] leading-relaxed text-c2-text-muted">
{configSnippet}
          </pre>
          <p className="mt-1.5 text-micro text-c2-text-faint">
            Paste into <code className="rounded bg-c2-bg/60 px-1 py-0.5 font-mono">~/Library/Application Support/Claude/claude_desktop_config.json</code> and restart Claude Desktop.
          </p>
        </div>

        <div className="flex justify-end">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-c2-border bg-c2-bg/40 px-4 py-2 text-sm font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface"
          >
            Done — I&rsquo;ve saved it
          </button>
        </div>
      </div>
    </div>
  );
}

function formatDate(iso: string | null | undefined): string {
  if (!iso) return '';
  const d = new Date(iso);
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

function formatRelative(iso: string | null | undefined): string {
  if (!iso) return '';
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 86400 * 30) return `${Math.floor(diff / 86400)}d ago`;
  return formatDate(iso);
}
