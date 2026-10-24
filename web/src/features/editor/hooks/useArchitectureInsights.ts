import { useMemo } from 'react';
import { useEditorStore } from '../store';
import { summarize, type CostReport } from '@/lib/costCatalog';
import { validate, type ValidationReport } from '@/lib/architectureRules';

export interface ArchitectureInsights {
  cost: CostReport;
  validation: ValidationReport;
}

// Computes cost + validation on every render of the current diagram.
// Memoized against node/edge references so recomputation is cheap — both
// functions are synchronous and run on the client.
export function useArchitectureInsights(): ArchitectureInsights {
  const nodes = useEditorStore((s) => s.nodes);
  const edges = useEditorStore((s) => s.edges);

  return useMemo(() => {
    const input = nodes
      .filter((n) => n.type !== 'group')
      .map((n) => {
        const d = n.data as { nodeType?: string; label?: string } | undefined;
        return {
          id: n.id,
          type: d?.nodeType ?? '',
          label: d?.label ?? n.id,
        };
      });
    return {
      cost: summarize(input),
      validation: validate(nodes, edges),
    };
  }, [nodes, edges]);
}
