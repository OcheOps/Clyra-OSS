import { type MutableRefObject, useEffect, useState } from 'react';
import { useCloudSave } from '../hooks/useCloudSave';

interface CloudSaveButtonProps {
  projectName: string;
  localSavedText: string;  // "Saved 3m ago" from the local autosave hook
  saving: boolean;         // local autosave in-flight
  projectIdRef: MutableRefObject<string | null>;
}

/**
 * CloudSaveButton shows the combined local + cloud status and triggers manual
 * "Save to account". History and Share actions have moved to the overflow
 * menu; this component now just owns Save.
 */
export function CloudSaveButton({ projectName, localSavedText, saving: localSaving, projectIdRef }: CloudSaveButtonProps) {
  const { save, saving: cloudSaving, lastSyncedAt } = useCloudSave(projectIdRef);
  const [cloudText, setCloudText] = useState('');

  useEffect(() => {
    const tick = () => {
      if (!lastSyncedAt) { setCloudText(''); return; }
      const diff = Math.floor((Date.now() - lastSyncedAt) / 1000);
      if (diff < 30) setCloudText('Synced to account');
      else if (diff < 60) setCloudText(`Synced ${diff}s ago`);
      else if (diff < 3600) setCloudText(`Synced ${Math.floor(diff / 60)}m ago`);
      else setCloudText(`Synced ${Math.floor(diff / 3600)}h ago`);
    };
    tick();
    const id = setInterval(tick, 10_000);
    return () => clearInterval(id);
  }, [lastSyncedAt]);

  const handleClick = () => {
    void save(projectName);
  };

  let statusLine = localSavedText;
  if (cloudText) statusLine = cloudText;
  if (cloudSaving) statusLine = 'Saving to account…';
  if (localSaving && !cloudSaving) statusLine = 'Saving locally…';

  return (
    <div className="flex shrink-0 items-center gap-2">
      {statusLine && (
        <span className="hidden font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint md:inline">
          {statusLine}
        </span>
      )}
      <button
        onClick={handleClick}
        disabled={cloudSaving}
        className="shrink-0 rounded-full border border-c2-border bg-c2-bg/40 px-3 py-1 text-tiny font-medium text-c2-text-muted transition-colors hover:border-c2-border-strong hover:text-c2-text disabled:cursor-not-allowed disabled:opacity-50"
        title="Save a snapshot of this diagram to your account"
      >
        {cloudSaving ? 'Saving…' : 'Save'}
      </button>
    </div>
  );
}
