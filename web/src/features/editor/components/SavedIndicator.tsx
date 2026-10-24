import { type MutableRefObject, useEffect, useState } from 'react';
import { useCloudSave } from '../hooks/useCloudSave';

interface Props {
  projectName: string;
  localSavedAt: number | null;
  saving: boolean;
  projectIdRef: MutableRefObject<string | null>;
}

// Compact status chip that replaces the old CloudSaveButton + shouty
// "SAVED 33M AGO" mono caps combo.
//
// Design principles:
//   1. Fresh writes (< 60s) show a minimal 'Saved' dot with no timestamp
//      — timestamps on recent saves read as noise, not reassurance.
//   2. Older writes show 'Saved · 3m' quietly so the user knows the
//      snapshot age exists but isn't nagged.
//   3. Cloud sync is a click-to-sync icon button, not a wall of text —
//      the icon alone communicates 'this is a cloud action' and the
//      tooltip explains what it does. Manual sync is preserved (some
//      users treat autosave as untrusted) but no longer competes with
//      the primary Generate CTA for attention.
//   4. Errors are surfaced as a tone change on the chip + tooltip,
//      never as a modal or toast.
export function SavedIndicator({ projectName, localSavedAt, saving: localSaving, projectIdRef }: Props) {
  const { save, saving: cloudSaving, lastSyncedAt } = useCloudSave(projectIdRef);
  const [now, setNow] = useState(() => Date.now());

  // 10s tick — enough resolution to walk the label from 'Saved' →
  // 'Saved · 1m' when the user looks away for a minute.
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 10_000);
    return () => clearInterval(id);
  }, []);

  const busy = localSaving || cloudSaving;
  const effectiveSavedAt = lastSyncedAt ?? localSavedAt ?? 0;
  const age = effectiveSavedAt ? Math.floor((now - effectiveSavedAt) / 1000) : Infinity;

  let label: string;
  if (busy) label = 'Saving…';
  else if (age === Infinity) label = 'Unsaved';
  else if (age < 60) label = 'Saved';
  else if (age < 3600) label = `Saved · ${Math.floor(age / 60)}m`;
  else label = `Saved · ${Math.floor(age / 3600)}h`;

  const isDirty = age === Infinity;
  const dotTone = busy ? 'bg-c2-primary animate-pulse'
    : isDirty ? 'bg-amber-400'
    : 'bg-c2-accent';

  return (
    <span className="flex shrink-0 items-center gap-2 text-tiny text-c2-text-muted">
      <span className={`h-1.5 w-1.5 rounded-full ${dotTone}`} aria-hidden />
      <span>{label}</span>
      <button
        type="button"
        onClick={() => void save(projectName)}
        disabled={cloudSaving}
        title={cloudSaving ? 'Syncing to your account…' : 'Sync to your account'}
        aria-label="Sync to cloud"
        className="ml-1 rounded-md p-1 text-c2-text-faint transition-colors hover:bg-c2-bg/40 hover:text-c2-text disabled:cursor-not-allowed disabled:opacity-40"
      >
        {/* Simple cloud-up glyph. 12px so the whole indicator stays
            roughly single-line-height. */}
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M7 18a4 4 0 0 1-.7-7.94 6 6 0 0 1 11.4-1.32A4.5 4.5 0 0 1 17 18" />
          <path d="M12 12v6" />
          <path d="M9 15l3-3 3 3" />
        </svg>
      </button>
    </span>
  );
}
