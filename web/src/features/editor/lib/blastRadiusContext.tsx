import { createContext, useContext, useEffect, useMemo, type ReactNode } from 'react';
import { useEditorStore } from '../store';
import {
  computeBlastRadius,
  type AffectedNode,
  type CostImpact,
  type Mitigation,
} from './blastRadius';

interface BlastRadiusContextValue {
  focus: string | null;
  /** Ranked list of affected nodes (severity, distance, type). */
  affected: AffectedNode[];
  /** Set of ids for O(1) canvas / node membership checks. */
  affectedIds: Set<string>;
  /** Diagram-wide SPOFs (permanent badges, focus-independent). */
  spofs: Set<string>;
  /** Concrete mitigation suggestions for the current blast. */
  mitigations: Mitigation[];
  /** Rough downtime cost estimate. */
  cost: CostImpact;
}

const EMPTY_SET: Set<string> = new Set();
const EMPTY_COST: CostImpact = { usdPerHour: 0, breakdown: [] };

const Ctx = createContext<BlastRadiusContextValue>({
  focus: null,
  affected: [],
  affectedIds: EMPTY_SET,
  spofs: EMPTY_SET,
  mitigations: [],
  cost: EMPTY_COST,
});

/**
 * BlastRadiusProvider computes the enriched blast state once per
 * (nodes, edges, focus) change and broadcasts it via context. Every
 * consumer (canvas nodes, insights panel, mitigation checklist) reads
 * from one memoized traversal.
 */
export function BlastRadiusProvider({ children }: { children: ReactNode }) {
  const nodes = useEditorStore((s) => s.nodes);
  const edges = useEditorStore((s) => s.edges);
  const focus = useEditorStore((s) => s.blastRadiusFocus);
  const selectedNodeId = useEditorStore((s) => s.selectedNodeId);
  const setBlastRadiusFocus = useEditorStore((s) => s.setBlastRadiusFocus);

  // Follow-selection: when blast mode is on and the user clicks a
  // different node, follow the click. Preserves the "click around to
  // see each node's blast" flow without re-triggering the prompt.
  useEffect(() => {
    if (focus !== null && selectedNodeId !== null && selectedNodeId !== focus) {
      setBlastRadiusFocus(selectedNodeId);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedNodeId, setBlastRadiusFocus]);

  const value = useMemo<BlastRadiusContextValue>(() => {
    const r = computeBlastRadius(nodes, edges, focus ?? '');
    if (!focus) {
      // No focus: only SPOF badges need to render; keep everything else empty.
      return {
        focus,
        affected: [],
        affectedIds: EMPTY_SET,
        spofs: r.spofs,
        mitigations: [],
        cost: EMPTY_COST,
      };
    }
    return {
      focus,
      affected: r.affected,
      affectedIds: r.affectedIds,
      spofs: r.spofs,
      mitigations: r.mitigations,
      cost: r.cost,
    };
  }, [nodes, edges, focus]);

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useBlastRadius(): BlastRadiusContextValue {
  return useContext(Ctx);
}
