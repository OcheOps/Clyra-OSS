import type { NodeTypes } from '@xyflow/react';
import { CloudNode } from './components/CloudNode';
import { GroupNode } from './components/GroupNode';

export const nodeTypes: NodeTypes = {
  cloud: CloudNode,
  group: GroupNode,
};
