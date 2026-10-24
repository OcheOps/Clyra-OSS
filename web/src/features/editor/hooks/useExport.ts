import { useCallback } from 'react';
import toast from 'react-hot-toast';
import { useEditorStore } from '../store';
import { exportJSON, exportSVG, exportPNG, exportPDF } from '@/services/export/export';

export function useExport(projectName: string) {
  const getDiagram = useEditorStore((s) => s.getDiagram);

  const handleExportJSON = useCallback(() => {
    const diagram = getDiagram();
    if (diagram.nodes.length === 0) {
      toast.error('Nothing to export');
      return;
    }
    exportJSON(diagram, projectName);
    toast.success('Exported as JSON');
  }, [getDiagram, projectName]);

  const handleExportSVG = useCallback(async () => {
    const ok = await exportSVG(projectName);
    if (ok) {
      toast.success('Exported as SVG');
    } else {
      toast.error('SVG export failed');
    }
  }, [projectName]);

  const handleExportPNG = useCallback(async () => {
    const ok = await exportPNG(projectName);
    if (ok) {
      toast.success('Exported as PNG');
    } else {
      toast.error('PNG export failed');
    }
  }, [projectName]);

  const handleExportPDF = useCallback(async () => {
    const ok = await exportPDF(projectName);
    if (ok) {
      toast.success('Exported as PDF');
    } else {
      toast.error('PDF export failed');
    }
  }, [projectName]);

  return { handleExportJSON, handleExportSVG, handleExportPNG, handleExportPDF };
}
