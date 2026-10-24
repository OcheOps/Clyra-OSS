import { api } from './client';
import type { Diagram } from '@/types/diagram';

// Mirrors internal/iac/reverse/handler.go IngestResponse.
//
// Diagram is the same shape useEditorStore.loadDiagram() takes, so a
// successful ingest can flow straight into the editor without any extra
// transformation on the client.
export interface IngestResponse {
  diagram: Diagram;
  // UX-level notes — "1 module rendered as group", "3 unsupported HCL
  // types rendered as placeholders". One per concern, all safe to render.
  warnings: string[];
  // Distinct HCL types we mapped to generic.service placeholder nodes,
  // sorted. Surface as a chip list so the user knows which resources were
  // approximated and can spot the gap at a glance.
  unsupported: string[];
}

// ingestTerraform sends raw HCL to the backend reverse pipeline. Backend
// caps the body at 50KB; we don't pre-check here because the user might
// paste something slightly over and a 413 is cheaper than rejecting them
// twice (once optimistically on the client, once on the wire).
export function ingestTerraform(source: string): Promise<IngestResponse> {
  return api<IngestResponse>('/iac/ingest', {
    method: 'POST',
    body: { source },
  });
}
