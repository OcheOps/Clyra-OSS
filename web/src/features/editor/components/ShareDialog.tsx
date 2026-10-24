import { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import toast from 'react-hot-toast';
import { APIError } from '@/types/api';
import { createShare, listShares, revokeShare, shareURL, type ShareToken } from '@/services/api/share';
import { useUpgradePromptStore } from '@/features/billing/upgradePromptStore';

interface Props {
  projectID: string | null;
  open: boolean;
  onClose: () => void;
}

function timeAgo(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

// ShareDialog: list + create + revoke public read-only links for a
// project. Pro/BYO only — 402s bounce to the global upgrade prompt.
export function ShareDialog({ projectID, open, onClose }: Props) {
  const [shares, setShares] = useState<ShareToken[]>([]);
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    if (!open || !projectID) return;
    setLoading(true);
    listShares(projectID)
      .then((s) => setShares(s ?? []))
      .catch((err) => {
        if (err instanceof APIError && err.status === 402) {
          useUpgradePromptStore.getState().show('cloud_sync');
          onClose();
          return;
        }
        toast.error('Could not load shares');
      })
      .finally(() => setLoading(false));
  }, [open, projectID, onClose]);

  const onCreate = async () => {
    if (!projectID) return;
    setCreating(true);
    try {
      const s = await createShare(projectID);
      setShares((prev) => [s, ...prev]);
      navigator.clipboard.writeText(shareURL(s.token));
      toast.success('Link created & copied');
    } catch (err) {
      if (err instanceof APIError && err.status === 402) {
        useUpgradePromptStore.getState().show('cloud_sync');
      } else {
        toast.error('Share failed');
      }
    } finally {
      setCreating(false);
    }
  };

  const onRevoke = async (token: string) => {
    if (!projectID) return;
    if (!confirm('Revoke this link? Anyone with it will stop being able to view.')) return;
    try {
      await revokeShare(projectID, token);
      setShares((prev) => prev.filter((s) => s.token !== token));
      toast.success('Link revoked');
    } catch {
      toast.error('Revoke failed');
    }
  };

  const onCopy = (token: string) => {
    navigator.clipboard.writeText(shareURL(token));
    toast.success('Copied');
  };

  if (!open) return null;

  return createPortal(
    <div
      role="dialog"
      aria-modal="true"
      className="fixed inset-0 z-[60] flex items-center justify-center bg-c2-bg/70 backdrop-blur-sm px-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-xl border border-c2-border bg-c2-surface/95 p-6 shadow-c2-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold text-c2-text">
            Share a read-only link
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="text-tiny text-c2-text-muted hover:text-c2-text"
          >
            Close
          </button>
        </div>
        <p className="mt-1 text-petit text-c2-text-muted">
          Anyone with the link can view this diagram — no account needed. Revoke anytime.
        </p>

        <div className="mt-5">
          <button
            type="button"
            onClick={onCreate}
            disabled={!projectID || creating}
            className="w-full rounded-lg bg-c2-primary px-4 py-2 text-sm font-medium text-white hover:bg-c2-primary disabled:opacity-40"
          >
            {creating ? 'Creating…' : !projectID ? 'Save to cloud first' : 'Create new link'}
          </button>
        </div>

        <div className="mt-6">
          <p className="text-tiny font-semibold uppercase tracking-[0.14em] text-c2-text-muted">
            Active links
          </p>
          {loading ? (
            <p className="mt-3 text-sm text-c2-text-muted">Loading…</p>
          ) : shares.length === 0 ? (
            <p className="mt-3 text-petit text-c2-text-muted">No active links yet.</p>
          ) : (
            <ul className="mt-3 space-y-2">
              {shares.map((s) => (
                <li
                  key={s.token}
                  className="flex items-center justify-between gap-2 rounded-md border border-c2-hairline px-3 py-2"
                >
                  <div className="min-w-0">
                    <code className="block truncate font-mono text-tiny text-c2-text-muted">
                      /share/{s.token}
                    </code>
                    <p className="mt-0.5 text-micro text-c2-text-muted">
                      Created {timeAgo(s.created_at)}
                    </p>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <button
                      type="button"
                      onClick={() => onCopy(s.token)}
                      className="text-tiny text-c2-text-muted hover:text-c2-text"
                    >
                      Copy
                    </button>
                    <button
                      type="button"
                      onClick={() => onRevoke(s.token)}
                      className="text-tiny text-danger/70 hover:text-danger"
                    >
                      Revoke
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>,
    document.body,
  );
}
