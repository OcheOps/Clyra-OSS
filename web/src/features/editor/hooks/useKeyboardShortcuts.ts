import { useEffect } from 'react';
import { useEditorStore } from '../store';

/**
 * Registers global keyboard shortcuts for the editor.
 * - Ctrl+Z: undo
 * - Ctrl+Shift+Z / Ctrl+Y: redo
 * - Ctrl+D: duplicate selected node
 * - Escape: exit blast-radius view (no-op if not active)
 */
export function useKeyboardShortcuts() {
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      const mod = e.metaKey || e.ctrlKey;
      const target = e.target as HTMLElement;

      // Don't intercept when typing in inputs
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
        return;
      }

      // Undo: Ctrl+Z
      if (mod && e.key === 'z' && !e.shiftKey) {
        e.preventDefault();
        useEditorStore.getState().undo();
        return;
      }

      // Redo: Ctrl+Shift+Z or Ctrl+Y
      if (mod && ((e.key === 'z' && e.shiftKey) || e.key === 'y')) {
        e.preventDefault();
        useEditorStore.getState().redo();
        return;
      }

      // Exit blast-radius view: Escape. Cheap no-op when not active.
      // Mod-less so it can't collide with browser shortcuts.
      if (!mod && e.key === 'Escape') {
        if (useEditorStore.getState().blastRadiusFocus !== null) {
          e.preventDefault();
          useEditorStore.getState().setBlastRadiusFocus(null);
          return;
        }
      }

      // Duplicate: Ctrl+D
      if (mod && e.key === 'd') {
        e.preventDefault();
        const { selectedNodeId } = useEditorStore.getState();
        if (selectedNodeId) {
          useEditorStore.getState().duplicateNode(selectedNodeId);
        }
        return;
      }
    }

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);
}
