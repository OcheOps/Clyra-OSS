import { useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import { saveCloudProject } from '@/services/api/cloudsync';
import type { CloudProject } from '@/types/cloudsync';
import { APIError } from '@/types/api';
import { useUpgradePromptStore } from '@/features/billing/upgradePromptStore';
import { useEditorStore } from '../store';

interface CloudSaveState {
  saving: boolean;
  lastSyncedAt: number | null;
  project: CloudProject | null;
}

/**
 * useCloudSave wraps the manual "Save to cloud" action from the editor.
 * Auto-sync is deliberately NOT here — this is invoked by a button press only,
 * matching the product spec: manual for everyone, optional interval is PRO-only
 * and will be added when there's demand.
 */
export function useCloudSave(projectIdRef: { current: string | null }) {
  const [state, setState] = useState<CloudSaveState>({
    saving: false,
    lastSyncedAt: null,
    project: null,
  });
  const getDiagram = useEditorStore((s) => s.getDiagram);

  const save = useCallback(
    async (title: string) => {
      const diagram = getDiagram();
      if (diagram.nodes.length === 0) {
        toast.error('Add at least one node before saving to the cloud');
        return null;
      }

      setState((s) => ({ ...s, saving: true }));
      try {
        // Strip locally-generated draft IDs (e.g. "proj-<ts>-<rand>" from the
        // IndexedDB autosave layer) — those only exist client-side and the
        // backend would 404 looking them up. Server IDs are 24-char hex.
        const currentID = projectIdRef.current;
        const cloudID = currentID && !currentID.startsWith('proj-') ? currentID : undefined;
        const project = await saveCloudProject({
          id: cloudID,
          title,
          diagram,
        });
        projectIdRef.current = project.id;
        setState({ saving: false, lastSyncedAt: Date.now(), project });
        toast.success('Saved to your account');
        return project;
      } catch (err) {
        if (err instanceof APIError && err.status === 402) {
          useUpgradePromptStore.getState().show('cloud_sync');
        } else {
          const msg = (err as Error).message ?? 'Cloud save failed';
          toast.error(msg);
        }
        setState((s) => ({ ...s, saving: false }));
        return null;
      }
    },
    [getDiagram, projectIdRef],
  );

  return { ...state, save };
}
