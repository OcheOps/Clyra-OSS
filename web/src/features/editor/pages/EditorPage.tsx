import { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { ReactFlowProvider } from '@xyflow/react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Sidebar } from '@/components/layout/Sidebar';
import { Canvas } from '../components/Canvas';
import { NodePalette } from '../components/NodePalette';
import { NodeProperties } from '../components/NodeProperties';
import { EditorStatusBar } from '../components/EditorStatusBar';
import { FloatingPromptDock } from '../components/FloatingPromptDock';
import { usePromptSessionStore } from '../promptSessionStore';
import { IaCMenu } from '../components/IaCMenu';
import { IaCPanel } from '../components/IaCPanel';
import { OnboardingModal } from '../components/OnboardingModal';
import { EmptyCanvasPrompt } from '../components/EmptyCanvasPrompt';
import { TweakCoach } from '../components/TweakCoach';
import { EditorOverflowMenu } from '../components/EditorOverflowMenu';
import { SavedIndicator } from '../components/SavedIndicator';
import { ReverseFromIaCDialog } from '../components/ReverseFromIaCDialog';
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts';
import { useAutosave } from '../hooks/useAutosave';
import { useEditorStore } from '../store';
import { useAuthStore } from '@/features/auth/store';
import { BlastRadiusProvider } from '../lib/blastRadiusContext';
import type { IaCFormat } from '@/types/iac';

function onboardingKey(userId: string | undefined): string | null {
  return userId ? `clyra-onboarded-${userId}` : null;
}

function EditorInner({ projectId }: { projectId?: string }) {
  useKeyboardShortcuts();
  const {
    projectId: autosaveProjectId,
    projectName,
    setProjectName,
    saving,
    lastSaved,
  } = useAutosave(projectId);
  // Owned here so both CloudSaveButton (writes it on save) and the
  // overflow menu (reads it to gate History/Share) share one source.
  //
  // Audit fix #6: initialize from useAutosave's resolved id (route
  // param OR pendingProjectId from a cloud-project open). Previously we
  // read only the route param, so opening a cloud project via
  // ProjectsPage (which navigates to /editor and stashes the id in
  // pending state) left projectIdRef = null. The next CloudSave then
  // sent "new" instead of the opened id and forked a duplicate.
  const projectIdRef = useRef<string | null>(projectId ?? autosaveProjectId ?? null);
  // Keep the ref in sync as useAutosave resolves the id (async load,
  // or a fresh-create's server-assigned id landing later).
  useEffect(() => {
    projectIdRef.current = autosaveProjectId ?? null;
  }, [autosaveProjectId]);
  const [iacOpen, setIacOpen] = useState(false);
  const [iacFormat, setIacFormat] = useState<IaCFormat>('terraform');

  // Landing-page "Import your Terraform" deep-link. Route arrives as
  // /editor?reverse=1 — auto-open the reverse-ingest dialog once, then
  // strip the query param so a page refresh doesn't reopen it.
  const [searchParams, setSearchParams] = useSearchParams();
  const [reverseOpen, setReverseOpen] = useState(false);
  useEffect(() => {
    if (searchParams.get('reverse') === '1') {
      setReverseOpen(true);
      const next = new URLSearchParams(searchParams);
      next.delete('reverse');
      setSearchParams(next, { replace: true });
    }
    // Run-once on mount — subsequent state changes are user-driven.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  const [onboardingOpen, setOnboardingOpen] = useState(false);
  const nodes = useEditorStore((s) => s.nodes);
  const addNode = useEditorStore((s) => s.addNode);
  const user = useAuthStore((s) => s.user);

  const handleIaCChoose = useCallback((format: IaCFormat) => {
    setIacFormat(format);
    setIacOpen(true);
  }, []);

  const dismissOnboarding = useCallback(() => {
    const key = onboardingKey(user?.id);
    if (key) localStorage.setItem(key, '1');
    setOnboardingOpen(false);
  }, [user?.id]);

  // Onboarding modal is superseded by the empty-canvas prompt card, which
  // gives a new user an immediate thing to do. Keep the modal code around
  // but don't auto-open it — `EmptyCanvasPrompt` is the first-run surface.


  // No clearCanvas on mount — the editor used to wipe the store whenever
  // there was no :projectId in the URL, which silently destroyed diagrams
  // pre-loaded by ProjectsPage (cloud open) and TemplateGallery. Now the
  // "fresh canvas" intent lives at the call sites: any button that means
  // "new diagram" calls useEditorStore.getState().clearCanvas() before
  // navigating here. Callers that pre-populate the store just navigate.

  const projectNameRef = useRef(projectName);
  projectNameRef.current = projectName;

  const handleProjectRename = useCallback(
    (name: string) => {
      if (projectNameRef.current === 'Untitled Diagram') {
        setProjectName(name);
      }
    },
    [setProjectName],
  );

  const handleAddGroup = useCallback(() => {
    const groupId = `group-${Date.now()}`;
    addNode({
      id: groupId,
      type: 'group',
      position: { x: 100, y: 100 },
      data: { label: 'New Group', nodeType: 'generic.container' },
      style: { width: 500, height: 350 },
    });
  }, [addNode]);

  const isEmpty = nodes.length === 0;
  // When the user dismisses EmptyCanvasPrompt without generating,
  // isEmpty stays true — but the card is gone, so without this flag
  // the FloatingPromptDock would never mount and the user would be
  // stranded with no way to type a prompt. Once dismissed OR once a
  // diagram exists, the dock takes over as the fresh-prompt affordance.
  const emptyCardDismissed = usePromptSessionStore((s) => s.emptyCardDismissed);
  const showDock = !isEmpty || emptyCardDismissed;

  return (
    <AppLayout
      topbarChildren={
        <>
          {/* Project name + saved status — one calm cluster, no manual
              Save competing with the primary Generate action. Save-to-
              cloud lives inside SavedIndicator as an icon button. */}
          <input
            value={projectName}
            onChange={(e) => setProjectName(e.target.value)}
            className="w-40 shrink-0 truncate rounded-lg border border-transparent bg-transparent px-2 py-0.5 text-sm font-medium text-c2-text outline-none hover:border-c2-border focus:border-c2-primary/50 transition-colors"
            title="Project name"
          />
          <SavedIndicator
            projectName={projectName}
            localSavedAt={lastSaved}
            saving={saving}
            projectIdRef={projectIdRef}
          />

          {/* Spacer — pushes the right cluster to the right. Prompt
              bar has moved to the floating dock at the bottom of the
              canvas; the topbar only carries context + secondary
              actions now. */}
          <div className="flex-1" />

          <IaCMenu onChoose={handleIaCChoose} />

          <EditorOverflowMenu
            projectIdRef={projectIdRef}
            projectName={projectName}
            onAddGroup={handleAddGroup}
          />
        </>
      }
    >
      <Sidebar side="left">
        <NodePalette />
      </Sidebar>

      <div className="flex flex-1 flex-col overflow-hidden relative">
        <Canvas />
        {isEmpty && (
          <EmptyCanvasPrompt
            onProjectRename={handleProjectRename}
            onOpenReverse={() => setReverseOpen(true)}
          />
        )}
        {/* Cursor-style floating prompt at bottom-center. Mounts once
            EITHER the canvas has nodes OR the empty-card was dismissed
            — otherwise a user who X'd the empty card would be stranded
            with no prompt affordance. EmptyCanvasPrompt owns the
            fresh-prompt surface while it's still visible. */}
        {showDock && <FloatingPromptDock onProjectRename={handleProjectRename} />}
        <TweakCoach hasContent={!isEmpty} />
        <EditorStatusBar />
      </div>

      <Sidebar side="right" width="w-64">
        <NodeProperties />
      </Sidebar>

      <IaCPanel
        open={iacOpen}
        format={iacFormat}
        onClose={() => setIacOpen(false)}
        projectName={projectName}
      />

      <OnboardingModal open={onboardingOpen} onDismiss={dismissOnboarding} />

      {/* Deep-link modal from landing "Import your Terraform" CTA.
          EditorOverflowMenu owns its own instance for the manual open,
          so both can coexist without conflict. */}
      <ReverseFromIaCDialog open={reverseOpen} onClose={() => setReverseOpen(false)} />
    </AppLayout>
  );
}

export function EditorPage() {
  const { projectId } = useParams<{ projectId: string }>();

  return (
    <ReactFlowProvider key={projectId ?? 'new'}>
      <BlastRadiusProvider>
        <EditorInner projectId={projectId} />
      </BlastRadiusProvider>
    </ReactFlowProvider>
  );
}
