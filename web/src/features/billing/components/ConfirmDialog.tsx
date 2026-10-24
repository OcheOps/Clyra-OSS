import { Button } from '@/components/ui/Button';

interface ConfirmDialogProps {
  open: boolean;
  title: string;
  body: string;
  confirmLabel?: string;
  cancelLabel?: string;
  confirmVariant?: 'primary' | 'danger';
  loading?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}

export function ConfirmDialog({
  open,
  title,
  body,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  confirmVariant = 'primary',
  loading,
  onConfirm,
  onClose,
}: ConfirmDialogProps) {
  if (!open) return null;
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm px-4"
      onClick={onClose}
    >
      <div
        className="relative w-full max-w-sm rounded-clyra-lg border border-line-default bg-clyra-bg shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-5">
          <h2 className="text-sm font-semibold text-ink-primary">{title}</h2>
          <p className="mt-2 text-petit text-ink-secondary">{body}</p>
        </div>
        <div className="flex justify-end gap-2 border-t border-line-subtle bg-clyra-raised/30 p-4">
          <Button variant="secondary" size="sm" onClick={onClose} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button
            size="sm"
            variant={confirmVariant === 'danger' ? 'danger' : 'primary'}
            loading={loading}
            onClick={onConfirm}
          >
            {confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  );
}
