import { useState, useCallback, useMemo, useEffect } from 'react';
import { getServicesByPrefix, type IconDef } from '@/lib/icons';
import { IconRenderer } from '@/components/ui/IconRenderer';
import { useEditorStore } from '../store';
import type { ArchitectureMode } from '@/types/diagram';

const allServices = getServicesByPrefix();

const FAMILY_LABELS: Record<string, string> = {
  aws: 'AWS',
  gcp: 'Google Cloud',
  azure: 'Azure',
  k8s: 'Kubernetes',
  oci: 'Oracle Cloud',
  cloudflare: 'Cloudflare',
  solution: 'Solution',
  network: 'Network',
  enterprise: 'Enterprise',
  data: 'Data & Analytics',
  edge: 'Edge & CDN',
  iot: 'IoT & Protocols',
  generic: 'Generic',
};

// Which families are promoted (expanded + listed first) for each mode
const MODE_PROMOTED_FAMILIES: Record<ArchitectureMode, string[]> = {
  solution_architecture: ['solution', 'aws', 'gcp', 'azure', 'k8s', 'data', 'oci', 'cloudflare', 'generic', 'network', 'edge', 'iot', 'enterprise'],
  network_architecture: ['network', 'aws', 'azure', 'gcp', 'cloudflare', 'k8s', 'edge', 'oci', 'solution', 'enterprise', 'iot', 'data', 'generic'],
  enterprise_architecture: ['enterprise', 'solution', 'aws', 'azure', 'gcp', 'data', 'k8s', 'oci', 'cloudflare', 'network', 'edge', 'iot', 'generic'],
};

function getDefaultExpanded(mode: ArchitectureMode): Record<string, boolean> {
  const promoted = MODE_PROMOTED_FAMILIES[mode];
  const result: Record<string, boolean> = {};
  for (const family of Object.keys(FAMILY_LABELS)) {
    // Expand the first family (the mode-primary one) and generic
    result[family] = family === promoted[0] || family === 'generic';
  }
  return result;
}

interface PaletteItemProps {
  type: string;
  icon: IconDef;
}

function PaletteItem({ type, icon }: PaletteItemProps) {
  const onDragStart = useCallback(
    (e: React.DragEvent) => {
      e.dataTransfer.setData('application/diagram-node-type', type);
      e.dataTransfer.effectAllowed = 'move';
    },
    [type],
  );

  return (
    <div
      draggable
      onDragStart={onDragStart}
      className="flex cursor-grab items-center gap-2 rounded-md px-2 py-1.5 text-xs transition-colors hover:bg-c2-bg/40 active:cursor-grabbing"
      title={type}
    >
      <span
        className="flex h-6 w-6 shrink-0 items-center justify-center rounded"
        style={{ backgroundColor: `${icon.color}15`, color: icon.color }}
      >
        <IconRenderer name={icon.icon} size={14} strokeWidth={1.5} />
      </span>
      <span className="truncate text-c2-text">{icon.label}</span>
    </div>
  );
}

export function NodePalette() {
  const mode = useEditorStore((s) => s.mode);
  const [search, setSearch] = useState('');
  const [expanded, setExpanded] = useState<Record<string, boolean>>(() => getDefaultExpanded(mode));

  // When mode changes, reset expansion to promote the relevant families
  useEffect(() => {
    setExpanded(getDefaultExpanded(mode));
  }, [mode]);

  const toggle = useCallback((prefix: string) => {
    setExpanded((prev) => ({ ...prev, [prefix]: !prev[prefix] }));
  }, []);

  const lowerSearch = search.toLowerCase();

  // Order families by mode relevance
  const orderedFamilies = useMemo(() => {
    return MODE_PROMOTED_FAMILIES[mode];
  }, [mode]);

  return (
    <div className="flex h-full flex-col">
      <div className="border-b border-c2-hairline p-3">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search components…"
          className="w-full rounded-lg border border-c2-border bg-c2-bg/40 px-2.5 py-1.5 text-xs text-c2-text placeholder:text-c2-text-faint transition-colors focus:border-c2-primary/60 focus:bg-c2-surface focus:outline-none focus:ring-2 focus:ring-c2-primary-ring/40"
        />
      </div>

      <div className="flex-1 overflow-y-auto px-2 py-3">
        {orderedFamilies.map((prefix) => {
          const items = allServices[prefix];
          if (!items) return null;

          const filtered = lowerSearch
            ? items.filter(
                (item) =>
                  item.icon.label.toLowerCase().includes(lowerSearch) ||
                  item.type.toLowerCase().includes(lowerSearch),
              )
            : items;

          if (filtered.length === 0) return null;

          const isOpen = search ? true : expanded[prefix];

          return (
            <div key={prefix} className="mb-1">
              <button
                onClick={() => toggle(prefix)}
                className="flex w-full items-center gap-1.5 rounded px-1 py-1.5 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text"
              >
                <svg
                  className={`h-3 w-3 transition-transform ${isOpen ? 'rotate-90' : ''}`}
                  viewBox="0 0 24 24"
                  fill="currentColor"
                >
                  <path d="M9 6l6 6-6 6z" />
                </svg>
                {FAMILY_LABELS[prefix] ?? prefix}
                <span className="ml-auto tabular-nums text-c2-text-faint">
                  {filtered.length}
                </span>
              </button>

              {isOpen && (
                <div className="ml-1">
                  {filtered.map((item) => (
                    <PaletteItem key={item.type} type={item.type} icon={item.icon} />
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
