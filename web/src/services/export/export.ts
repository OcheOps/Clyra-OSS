import { toPng, toSvg } from 'html-to-image';
import { jsPDF } from 'jspdf';
import type { Diagram } from '@/types/diagram';

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

function downloadDataURL(dataUrl: string, filename: string) {
  const a = document.createElement('a');
  a.href = dataUrl;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
}

function sanitizeFilename(name: string): string {
  return name.replace(/[^a-zA-Z0-9_-]/g, '_').slice(0, 64) || 'diagram';
}

function getFlowContainer(): HTMLElement | null {
  return document.querySelector('.react-flow') as HTMLElement | null;
}

function getViewportEl(): HTMLElement | null {
  return document.querySelector('.react-flow__viewport') as HTMLElement | null;
}

function filterExportNode(node: HTMLElement): boolean {
  if (!node.classList) return true;
  const exclude = [
    'react-flow__minimap',
    'react-flow__controls',
    'react-flow__attribution',
    'react-flow__panel',
  ];
  return !exclude.some((cls) => node.classList.contains(cls));
}

// ─── Export bounds ───────────────────────────────────────────────────
// The React Flow *viewport* element is what we actually want to
// capture — it contains every node at its true flow-coordinate
// position, unlike the outer `.react-flow` which is just the window.
// We walk the nodes, compute the union of their flow-space rectangles,
// and return the transform that, applied to the viewport, brings the
// top-left of that bounding box to (padding, padding) at zoom=1.
//
// Zoom=1 is intentional: it gives us native-resolution output
// independent of how the user currently has the canvas zoomed in
// the UI. No more "zoomed out = unreadable / zoomed in = cropped".

const NODE_SELECTOR = '.react-flow__viewport .react-flow__node';
const EXPORT_PADDING = 40; // px in flow coordinate space

function parseTranslate(style: string): { x: number; y: number } {
  // React Flow sets node transforms as "translate(Xpx, Ypx)".
  const m = style.match(/translate\(\s*(-?\d*\.?\d+)px\s*,\s*(-?\d*\.?\d+)px/);
  if (!m) return { x: 0, y: 0 };
  return { x: parseFloat(m[1]!), y: parseFloat(m[2]!) };
}

interface Bounds {
  minX: number;
  minY: number;
  width: number;
  height: number;
}

function computeNodeBounds(): Bounds | null {
  const nodes = document.querySelectorAll<HTMLElement>(NODE_SELECTOR);
  if (nodes.length === 0) return null;
  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
  nodes.forEach((el) => {
    const { x, y } = parseTranslate(el.style.transform || '');
    const w = el.offsetWidth;
    const h = el.offsetHeight;
    if (x < minX) minX = x;
    if (y < minY) minY = y;
    if (x + w > maxX) maxX = x + w;
    if (y + h > maxY) maxY = y + h;
  });
  if (!Number.isFinite(minX)) return null;
  return {
    minX,
    minY,
    width: maxX - minX,
    height: maxY - minY,
  };
}

// applyExportViewport overrides the viewport's transform so the full
// diagram is visible at scale=1, and returns a restore function.
function applyExportViewport(viewport: HTMLElement, bounds: Bounds): () => void {
  const prevTransform = viewport.style.transform;
  const prevTransition = viewport.style.transition;
  viewport.style.transition = 'none';
  viewport.style.transform = `translate(${-bounds.minX + EXPORT_PADDING}px, ${-bounds.minY + EXPORT_PADDING}px) scale(1)`;
  return () => {
    viewport.style.transform = prevTransform;
    viewport.style.transition = prevTransition;
  };
}

/**
 * Before capture: inline all computed styles on SVG edge paths so they
 * survive serialization. html-to-image loses CSS-based SVG styling.
 */
function inlineEdgeStyles(container: HTMLElement): () => void {
  const originals: { el: SVGElement; attrs: Record<string, string> }[] = [];

  // Inline styles on edge paths
  container.querySelectorAll('.react-flow__edge path').forEach((path) => {
    const el = path as SVGElement;
    const computed = window.getComputedStyle(el);
    const saved: Record<string, string> = {};

    const props = ['stroke', 'stroke-width', 'stroke-dasharray', 'stroke-opacity', 'fill', 'opacity'];
    for (const prop of props) {
      saved[prop] = el.getAttribute(prop) ?? '';
      const val = computed.getPropertyValue(prop);
      if (val) el.setAttribute(prop, val);
    }
    originals.push({ el, attrs: saved });
  });

  // Inline styles on edge labels
  container.querySelectorAll('.react-flow__edge text, .react-flow__edge textPath').forEach((text) => {
    const el = text as SVGElement;
    const computed = window.getComputedStyle(el);
    const saved: Record<string, string> = {};

    const props = ['fill', 'font-size', 'font-family', 'font-weight'];
    for (const prop of props) {
      saved[prop] = el.getAttribute(prop) ?? '';
      const val = computed.getPropertyValue(prop);
      if (val) el.setAttribute(prop, val);
    }
    originals.push({ el, attrs: saved });
  });

  // Inline label bg rect styles
  container.querySelectorAll('.react-flow__edge rect').forEach((rect) => {
    const el = rect as SVGElement;
    const computed = window.getComputedStyle(el);
    const saved: Record<string, string> = {};

    const props = ['fill', 'fill-opacity', 'rx', 'ry'];
    for (const prop of props) {
      saved[prop] = el.getAttribute(prop) ?? '';
      const val = computed.getPropertyValue(prop);
      if (val) el.setAttribute(prop, val);
    }
    originals.push({ el, attrs: saved });
  });

  // Return restore function
  return () => {
    for (const { el, attrs } of originals) {
      for (const [prop, val] of Object.entries(attrs)) {
        if (val) el.setAttribute(prop, val);
        else el.removeAttribute(prop);
      }
    }
  };
}

/**
 * Hide UI chrome before capture, return restore function.
 */
function hideChrome(container: HTMLElement): () => void {
  const hidden: { el: HTMLElement; prev: string }[] = [];

  ['.react-flow__minimap', '.react-flow__controls', '.react-flow__attribution', '.react-flow__panel']
    .forEach((sel) => {
      const el = container.querySelector(sel) as HTMLElement | null;
      if (el) {
        hidden.push({ el, prev: el.style.display });
        el.style.display = 'none';
      }
    });

  // Hide handle dots
  container.querySelectorAll('.react-flow__handle').forEach((h) => {
    const el = h as HTMLElement;
    hidden.push({ el, prev: el.style.visibility });
    el.style.visibility = 'hidden';
  });

  return () => {
    hidden.forEach(({ el, prev }) => {
      if (el.style.display === 'none') el.style.display = prev;
      else el.style.visibility = prev;
    });
  };
}

// --- JSON export ---

export function exportJSON(diagram: Diagram, projectName: string) {
  const json = JSON.stringify(diagram, null, 2);
  const blob = new Blob([json], { type: 'application/json' });
  downloadBlob(blob, `${sanitizeFilename(projectName)}.json`);
}

// --- SVG export ---

export async function exportSVG(projectName: string): Promise<boolean> {
  const outer = getFlowContainer();
  const viewport = getViewportEl();
  if (!outer || !viewport) return false;
  const bounds = computeNodeBounds();
  if (!bounds) return false;

  const restoreChrome = hideChrome(outer);
  const restoreEdges = inlineEdgeStyles(outer);
  const restoreViewport = applyExportViewport(viewport, bounds);

  const width = Math.ceil(bounds.width + EXPORT_PADDING * 2);
  const height = Math.ceil(bounds.height + EXPORT_PADDING * 2);
  const bg = document.body.classList.contains('paper') ? '#fbfaf7' : '#0f1117';

  try {
    const dataUrl = await toSvg(viewport, {
      filter: filterExportNode,
      backgroundColor: bg,
      width,
      height,
      style: {
        width: `${width}px`,
        height: `${height}px`,
      },
    });
    downloadDataURL(dataUrl, `${sanitizeFilename(projectName)}.svg`);
    return true;
  } catch (e) {
    console.error('SVG export failed:', e);
    return false;
  } finally {
    restoreViewport();
    restoreEdges();
    restoreChrome();
  }
}

// --- PDF export ---
// Capture the full viewport at native resolution then wrap it in a
// single-page PDF sized to the diagram's natural aspect ratio.
export async function exportPDF(projectName: string): Promise<boolean> {
  const outer = getFlowContainer();
  const viewport = getViewportEl();
  if (!outer || !viewport) return false;
  const bounds = computeNodeBounds();
  if (!bounds) return false;

  const restoreChrome = hideChrome(outer);
  const restoreEdges = inlineEdgeStyles(outer);
  const restoreViewport = applyExportViewport(viewport, bounds);

  const width = Math.ceil(bounds.width + EXPORT_PADDING * 2);
  const height = Math.ceil(bounds.height + EXPORT_PADDING * 2);

  try {
    const dataUrl = await toPng(viewport, {
      pixelRatio: 2,
      filter: filterExportNode,
      backgroundColor: '#ffffff',
      quality: 1,
      width,
      height,
      style: {
        width: `${width}px`,
        height: `${height}px`,
      },
    });
    // Probe image dimensions so the PDF page matches the aspect ratio.
    const img = new Image();
    await new Promise<void>((resolve, reject) => {
      img.onload = () => resolve();
      img.onerror = () => reject(new Error('image load'));
      img.src = dataUrl;
    });
    // Convert pixels → millimeters at 96 DPI (1 px = 0.2645833 mm) with an
    // 8 mm margin so the diagram isn't bleeding to the edge.
    const pxToMm = 0.2645833;
    const margin = 8;
    const widthMm = img.width * pxToMm + margin * 2;
    const heightMm = img.height * pxToMm + margin * 2;
    const orientation = widthMm >= heightMm ? 'l' : 'p';
    const pdf = new jsPDF({ unit: 'mm', format: [widthMm, heightMm], orientation });
    pdf.addImage(dataUrl, 'PNG', margin, margin, widthMm - margin * 2, heightMm - margin * 2, undefined, 'FAST');
    pdf.save(`${sanitizeFilename(projectName)}.pdf`);
    return true;
  } catch (e) {
    console.error('PDF export failed:', e);
    return false;
  } finally {
    restoreViewport();
    restoreEdges();
    restoreChrome();
  }
}

// --- PNG export ---

export async function exportPNG(projectName: string): Promise<boolean> {
  const outer = getFlowContainer();
  const viewport = getViewportEl();
  if (!outer || !viewport) return false;
  const bounds = computeNodeBounds();
  if (!bounds) return false;

  const restoreChrome = hideChrome(outer);
  const restoreEdges = inlineEdgeStyles(outer);
  const restoreViewport = applyExportViewport(viewport, bounds);

  const width = Math.ceil(bounds.width + EXPORT_PADDING * 2);
  const height = Math.ceil(bounds.height + EXPORT_PADDING * 2);
  const bg = document.body.classList.contains('paper') ? '#fbfaf7' : '#0f1117';

  try {
    const dataUrl = await toPng(viewport, {
      pixelRatio: 3,
      filter: filterExportNode,
      backgroundColor: bg,
      quality: 1,
      width,
      height,
      style: {
        width: `${width}px`,
        height: `${height}px`,
      },
    });
    downloadDataURL(dataUrl, `${sanitizeFilename(projectName)}.png`);
    return true;
  } catch (e) {
    console.error('PNG export failed:', e);
    return false;
  } finally {
    restoreViewport();
    restoreEdges();
    restoreChrome();
  }
}
