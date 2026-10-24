import { useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  ReactFlow,
  ReactFlowProvider,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  type Node,
  type Edge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { nodeTypes } from '@/features/editor/nodeTypes';
import { edgeTypes } from '@/features/editor/edgeTypes';
import { loadSharedProject } from '@/services/api/share';
import { diagramToFlow } from '@/lib/normalize';
import type { Diagram } from '@/types/diagram';
import type { CloudProject } from '@/types/cloudsync';

// PublicSharePage renders a diagram from a public share token without
// requiring auth. Read-only: no edit, no drag, no add. The editor store
// is intentionally bypassed — this page holds state locally so a
// visitor's interactions can't leak into a signed-in session.

export function PublicSharePage() {
  const { token } = useParams<{ token: string }>();

  // The public share view is intentionally always dark: diagrams read
  // better against a dark canvas, and we want visitors to see the
  // "product" look regardless of their stored preference.
  useEffect(() => {
    const had = document.body.classList.contains('paper');
    document.body.classList.remove('paper');
    return () => {
      if (had) document.body.classList.add('paper');
    };
  }, []);
  const [diagram, setDiagram] = useState<Diagram | null>(null);
  const [project, setProject] = useState<CloudProject | null>(null);
  const [error, setError] = useState<'not_found' | 'unknown' | null>(null);
  const [loading, setLoading] = useState(true);

  // React Flow needs nodes with position: {x, y}. Our API returns
  // DiagramNodes (no positions — the editor auto-layouts on load), so
  // we run the same normalizer here.
  const flow = useMemo<{ nodes: Node[]; edges: Edge[] } | null>(() => {
    if (!diagram) return null;
    const f = diagramToFlow(diagram);
    return { nodes: f.nodes as Node[], edges: f.edges as Edge[] };
  }, [diagram]);

  useEffect(() => {
    if (!token) return;
    setLoading(true);
    loadSharedProject(token)
      .then((r) => {
        setDiagram(r.diagram);
        setProject(r.project);
      })
      .catch((e) => {
        setError(e?.status === 404 ? 'not_found' : 'unknown');
      })
      .finally(() => setLoading(false));
  }, [token]);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-clyra-bg text-ink-muted">
        Loading…
      </div>
    );
  }

  if (error === 'not_found') {
    return (
      <EmptyState
        title="This link isn't active."
        body="It may have been revoked, or never existed."
      />
    );
  }

  if (error || !diagram) {
    return (
      <EmptyState
        title="Something went wrong."
        body="Try again in a moment. If the issue persists, reach out to the person who shared this link."
      />
    );
  }

  return (
    <div className="flex h-screen flex-col bg-clyra-bg">
      <header className="flex items-center justify-between border-b border-line-subtle bg-clyra-base/80 px-4 py-2 backdrop-blur-md">
        <div className="flex items-center gap-3">
          <Link to="/" className="text-sm font-semibold tracking-tight-ish text-ink-primary">
            clyra<span className="text-brand-400">.</span>
          </Link>
          <span className="text-tiny text-ink-faint">·</span>
          <span className="truncate text-tiny text-ink-secondary max-w-[40ch]">
            {project?.title ?? 'Shared diagram'}
          </span>
          <span className="rounded-full border border-line-subtle px-2 py-0.5 text-micro text-ink-muted">
            Read-only
          </span>
        </div>
        <Link
          to="/signup"
          className="rounded-md bg-brand-600 px-3 py-1 text-tiny font-medium text-white hover:bg-brand-500 transition-colors"
        >
          Try Clyra free
        </Link>
      </header>

      <div className="flex-1 relative">
        <ReactFlowProvider>
          <ReactFlow
            nodes={flow?.nodes ?? []}
            edges={flow?.edges ?? []}
            nodeTypes={nodeTypes}
            edgeTypes={edgeTypes}
            fitView
            fitViewOptions={{ padding: 0.15 }}
            minZoom={0.1}
            maxZoom={2}
            nodesDraggable={false}
            nodesConnectable={false}
            elementsSelectable={false}
            panOnDrag
            zoomOnScroll
            className="bg-clyra-bg"
          >
            <Background variant={BackgroundVariant.Dots} gap={24} size={0.6} color="#1e2433" />
            <Controls showInteractive={false} />
            <MiniMap
              nodeColor={() => '#1eaaff'}
              maskColor="rgba(10, 13, 20, 0.85)"
              pannable
              zoomable
            />
          </ReactFlow>
        </ReactFlowProvider>
      </div>
    </div>
  );
}

function EmptyState({ title, body }: { title: string; body: string }) {
  return (
    <div className="flex h-screen items-center justify-center bg-clyra-bg px-4">
      <div className="text-center">
        <h1 className="text-xl font-semibold text-ink-primary">{title}</h1>
        <p className="mt-2 text-sm text-ink-muted">{body}</p>
        <Link
          to="/"
          className="mt-6 inline-block rounded-md border border-line-default px-4 py-2 text-tiny font-medium text-ink-secondary hover:text-ink-primary hover:border-line-strong transition-colors"
        >
          Go to Clyra
        </Link>
      </div>
    </div>
  );
}
