import { useEffect, useRef, useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import { useEditorStore } from '../store';
import { useAuthStore } from '@/features/auth/store';
import * as projectsDB from '@/services/persistence/projects';
import { loadCloudProject } from '@/services/api/cloudsync';

const AUTOSAVE_INTERVAL = 3_000; // 3 seconds

interface AutosaveState {
  projectId: string | null;
  projectName: string;
  saving: boolean;
  lastSaved: number | null;
}

export function useAutosave(initialProjectId?: string) {
  const userId = useAuthStore((s) => s.user?.id) ?? '';

  // Pick up pending metadata from a cloud-project open. ProjectsPage stashes
  // the id+name in the store before navigating; consume + clear here so the
  // title bar starts on the correct project name instead of "Untitled".
  const pendingId = useEditorStore.getState().pendingProjectId;
  const pendingName = useEditorStore.getState().pendingProjectName;

  const [state, setState] = useState<AutosaveState>({
    projectId: initialProjectId ?? pendingId ?? null,
    projectName: pendingName ?? 'Untitled Diagram',
    saving: false,
    lastSaved: null,
  });

  useEffect(() => {
    if (pendingId || pendingName) {
      useEditorStore.getState().setPendingProject(null, null);
    }
    // Run-once on mount — the values were already snapshotted into local state.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const dirtyRef = useRef(false);

  // Mark dirty on any store change
  useEffect(() => {
    return useEditorStore.subscribe(() => {
      dirtyRef.current = true;
    });
  }, []);

  // Load project on mount if ID provided
  useEffect(() => {
    if (!initialProjectId || !userId) return;
    void projectsDB.getProject(initialProjectId, userId).then(async (project) => {
      if (project) {
        useEditorStore.getState().loadDiagram(project.diagram);
        setState((s) => ({
          ...s,
          projectId: project.id,
          projectName: project.name,
          lastSaved: project.updatedAt,
        }));
        return;
      }

      // Direct links from MCP point at cloud projects. Cache the cloud payload
      // under the same ID so normal local autosave can continue immediately and
      // a later Save-to-cloud updates the original project instead of forking it.
      try {
        const response = await loadCloudProject(initialProjectId);
        const now = Date.now();
        const cached = {
          id: response.project.id,
          name: response.project.title,
          diagram: response.diagram,
          createdAt: new Date(response.project.created_at).getTime() || now,
          updatedAt: new Date(response.project.updated_at).getTime() || now,
        };
        await projectsDB.saveProject(cached, userId);
        useEditorStore.getState().loadDiagram(response.diagram);
        dirtyRef.current = false;
        setState((s) => ({
          ...s,
          projectId: response.project.id,
          projectName: response.project.title,
          lastSaved: cached.updatedAt,
        }));
      } catch (err) {
        // 401 is already handled by the api client (session-expired
        // redirect to /login). Toasting on top of that just adds noise
        // during the redirect. Other statuses get the user-facing toast.
        const status = (err as { status?: number })?.status;
        if (status === 401) return;
        toast.error("Couldn't load this project");
      }
    });
  }, [initialProjectId, userId]);

  // Toast dedupe: don't spam the user on every failed 3s tick.
  const errorToastAtRef = useRef<number>(0);

  const save = useCallback(async () => {
    if (!dirtyRef.current || !userId) return;

    const diagram = useEditorStore.getState().getDiagram();
    const isEmpty = diagram.nodes.length === 0 && diagram.edges.length === 0;

    // Audit fix #4: a cleared existing project MUST persist its empty
    // state, otherwise deleting every node then reloading brings the
    // diagram back from the last non-empty save. Only skip the write
    // when there's no project yet AND nothing to write (fresh mount).
    if (isEmpty && !state.projectId) {
      dirtyRef.current = false;
      return;
    }

    setState((s) => ({ ...s, saving: true }));

    try {
      if (state.projectId) {
        const existing = await projectsDB.getProject(state.projectId, userId);
        if (existing) {
          await projectsDB.saveProject({ ...existing, diagram, name: state.projectName }, userId);
        } else {
          // Project row vanished (user deleted it in another tab, DB
          // wipe, etc). Fall through to create-new so their in-memory
          // work isn't lost on the next reload.
          const project = await projectsDB.createProject(state.projectName, diagram, userId);
          setState((s) => ({ ...s, projectId: project.id }));
        }
      } else {
        const project = await projectsDB.createProject(state.projectName, diagram, userId);
        setState((s) => ({ ...s, projectId: project.id }));
      }
      // Audit fix #5: clear dirty ONLY after the write commits, so a
      // transient IndexedDB failure leaves dirty=true and the next tick
      // retries instead of silently abandoning the version.
      dirtyRef.current = false;
      setState((s) => ({ ...s, saving: false, lastSaved: Date.now() }));
    } catch (err) {
      setState((s) => ({ ...s, saving: false }));
      // Surface autosave failures so the user knows their edits aren't
      // safe yet. Rate-limit to at most once every 15s to avoid a
      // toast storm if IndexedDB is durably broken.
      const now = Date.now();
      if (now - errorToastAtRef.current > 15_000) {
        errorToastAtRef.current = now;
        toast.error("Couldn't save your changes — we'll keep trying.");
      }
      console.error('autosave failed', err);
    }
  }, [state.projectId, state.projectName, userId]);

  // Autosave interval
  useEffect(() => {
    timerRef.current = setInterval(save, AUTOSAVE_INTERVAL);
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [save]);

  // Save on unmount
  useEffect(() => {
    return () => { save(); };
  }, [save]);

  const setProjectName = useCallback((name: string) => {
    setState((s) => ({ ...s, projectName: name }));
    dirtyRef.current = true;
  }, []);

  const saveNow = useCallback(() => {
    dirtyRef.current = true;
    return save();
  }, [save]);

  return {
    projectId: state.projectId,
    projectName: state.projectName,
    saving: state.saving,
    lastSaved: state.lastSaved,
    setProjectName,
    saveNow,
  };
}

export function formatLastSaved(ts: number | null): string {
  if (!ts) return '';
  const diff = Math.floor((Date.now() - ts) / 1000);
  if (diff < 5) return 'just now';
  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  return `${Math.floor(diff / 3600)}h ago`;
}
