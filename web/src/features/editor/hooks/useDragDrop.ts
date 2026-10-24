import { useCallback } from 'react';
import { useReactFlow } from '@xyflow/react';
import { getIconDef } from '@/lib/icons';
import { useEditorStore } from '../store';

let nodeCounter = 0;

export function useDragDrop() {
  const { screenToFlowPosition } = useReactFlow();
  const addNode = useEditorStore((s) => s.addNode);

  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      const nodeType = e.dataTransfer.getData('application/diagram-node-type');
      if (!nodeType) return;

      const position = screenToFlowPosition({
        x: e.clientX,
        y: e.clientY,
      });

      const icon = getIconDef(nodeType);
      nodeCounter++;

      addNode({
        id: `node-${Date.now()}-${nodeCounter}`,
        type: 'cloud',
        position,
        data: {
          label: icon.label,
          nodeType,
          metadata: {},
        },
      });
    },
    [screenToFlowPosition, addNode],
  );

  return { onDragOver, onDrop };
}
