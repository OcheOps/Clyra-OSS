import { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import toast from 'react-hot-toast';
import { APIError } from '@/types/api';
import {
  type ProjectVersion,
  listVersions,
  restoreVersion,
} from '@/services/api/history';
import { useUpgradePromptStore } from '@/features/billing/upgradePromptStore';

// HistoryPanel shows the last ~10 cloud-save snapshots for the active
// project and lets the user restore one. Pro/BYO-only — the backend
// returns 402 for free users and we bounce that into the global upgrade
// prompt.

interface Props {
  projectID: string | null;
  open: boolean;
  onClose: () => void;
  onRestored: (diagram: unknown) => void;
}

function timeAgo(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

export function HistoryPanel({ projectID, open, onClose, onRestored }: Props) {
  const [versions, setVersions] = useState<ProjectVersion[]>([]);
  const [loading, setLoading] = useState(false);
  const [restoringID, setRestoringID] = useState<string | null>(null);

  useEffect(() => {
    if (!open || !projectID) return;
    setLoading(true);
    listVersions(projectID)
      .then((v) => setVersions(v ?? []))
      .catch((err) => {
        if (err instanceof APIError && err.status === 402) {
          useUpgradePromptStore.getState().show('cloud_sync');
          onClose();
          return;
        }
        toast.error('Could not load history');
      })
      .finally(() => setLoading(false));
  }, [open, projectID, onClose]);

  const onRestore = async (v: ProjectVersion) => {
    if (!projectID) return;
    if (!confirm(`Restore the save from ${timeAgo(v.created_at)}? The current state is saved first so you can undo.`)) return;
    setRestoringID(v.id);
    try {
      const res = await restoreVersion(projectID, v.id);
      onRestored(res.diagram);
      toast.success('Restored');
      onClose();
    } catch (err) {
      if (err instanceof APIError && err.status === 402) {
        useUpgradePromptStore.getState().show('cloud_sync');
      } else {
        toast.error('Restore failed');
      }
    } finally {
      setRestoringID(null);
    }
  };

  if (!open) return null;

  // Portal to document.body so React Flow's transformed ancestors can't
  // re-anchor our `fixed` positioning and clip the modal.
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
            Version history
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
          The last 10 saves of this project. Older snapshots roll off.
        </p>

        <div className="mt-4 max-h-80 overflow-y-auto">
          {!projectID ? (
            <p className="py-6 text-center text-petit text-c2-text-muted">
              Save this project to the cloud first.
            </p>
          ) : loading ? (
            <p className="py-6 text-center text-sm text-c2-text-muted">Loading…</p>
          ) : versions.length === 0 ? (
            <p className="py-6 text-center text-petit text-c2-text-muted">
              No prior saves yet.
            </p>
          ) : (
            <ul className="space-y-1">
              {versions.map((v) => (
                <li
                  key={v.id}
                  className="flex items-center justify-between rounded-md border border-c2-hairline px-3 py-2"
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm text-c2-text">
                      {v.title}
                    </p>
                    <p className="mt-0.5 text-micro text-c2-text-muted">
                      {timeAgo(v.created_at)} · {(v.byte_size / 1024).toFixed(1)} KB
                    </p>
                  </div>
                  <button
                    type="button"
                    disabled={restoringID === v.id}
                    onClick={() => onRestore(v)}
                    className="shrink-0 rounded-md border border-c2-border px-2.5 py-1 text-tiny text-c2-text-muted hover:border-c2-border-strong hover:text-c2-text disabled:opacity-40"
                  >
                    {restoringID === v.id ? 'Restoring…' : 'Restore'}
                  </button>
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
