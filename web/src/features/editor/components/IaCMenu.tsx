import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import type { IaCFormat } from '@/types/iac';
import { useEditorStore } from '../store';
import { useAuthStore } from '@/features/auth/store';
import { useUpgradeModalStore } from '@/features/billing/upgradeModalStore';

interface IaCMenuProps {
  onChoose: (format: IaCFormat) => void;
}

export function IaCMenu({ onChoose }: IaCMenuProps) {
  const [open, setOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState({ top: 0, right: 0 });
  const mode = useEditorStore((s) => s.mode);
  const nodes = useEditorStore((s) => s.nodes);
  const user = useAuthStore((s) => s.user);
  const openUpgrade = useUpgradeModalStore((s) => s.openModal);

  // IaC export is a Pro/BYO feature — creates a concrete upgrade reason.
  // Free users still see the button so they know the feature exists; clicking
  // routes them to the upgrade modal instead of a hidden no-op.
  // Use effective_plan so a lapsed Pro/BYO doesn't see the menu as
  // unlocked — the backend would 402 anyway, but a paywalled control
  // that looks unlocked is a worse experience than seeing the padlock.
  const effective = user?.effective_plan ?? user?.plan;
  const isPaid = effective === 'pro' || effective === 'byo';
  const disabled = mode !== 'solution_architecture' || nodes.length === 0;
  const disabledReason =
    mode !== 'solution_architecture'
      ? 'Available in Solution mode only'
      : nodes.length === 0
        ? 'Add nodes first'
        : '';

  useEffect(() => {
    if (!open || !buttonRef.current) return;
    const rect = buttonRef.current.getBoundingClientRect();
    setPos({
      top: rect.bottom + 4,
      right: window.innerWidth - rect.right,
    });
  }, [open]);

  useEffect(() => {
    if (!open) return;
    function handleClick(e: MouseEvent) {
      const target = e.target as Node;
      if (
        buttonRef.current && !buttonRef.current.contains(target) &&
        menuRef.current && !menuRef.current.contains(target)
      ) {
        setOpen(false);
      }
    }
    const handleKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false); };
    document.addEventListener('mousedown', handleClick);
    document.addEventListener('keydown', handleKey);
    return () => {
      document.removeEventListener('mousedown', handleClick);
      document.removeEventListener('keydown', handleKey);
    };
  }, [open]);

  const choose = (format: IaCFormat) => {
    setOpen(false);
    onChoose(format);
  };

  const handleClick = () => {
    if (disabled) return;
    if (!isPaid) {
      openUpgrade({ defaultPlan: 'pro' });
      return;
    }
    setOpen((s) => !s);
  };

  return (
    <>
      <button
        ref={buttonRef}
        type="button"
        onClick={handleClick}
        disabled={disabled}
        title={
          disabled
            ? disabledReason
            : isPaid
              ? 'Generate infrastructure code'
              : 'Upgrade to Pro or BYO to export Terraform / CloudFormation'
        }
        className="shrink-0 rounded-full border border-c2-primary/40 bg-c2-primary-soft px-3 py-1 text-tiny font-medium text-c2-primary transition-colors hover:bg-c2-primary/25 disabled:cursor-not-allowed disabled:border-c2-border disabled:bg-transparent disabled:text-c2-text-faint disabled:opacity-50"
      >
        Generate IaC
        {!isPaid && (
          // Small 'Pro' chip instead of a padlock — same information
          // (this needs an upgrade), no security-alarm connotation.
          <span className="ml-1.5 rounded-full border border-c2-primary/50 bg-c2-primary/15 px-1.5 py-0.5 font-mono text-[9px] font-semibold uppercase tracking-[0.14em] text-c2-primary">
            Pro
          </span>
        )}
      </button>

      {open &&
        createPortal(
          <div
            ref={menuRef}
            className="fixed z-50 w-56 rounded-xl border border-c2-border bg-c2-elevated/95 py-1.5 shadow-c2-lg backdrop-blur-md"
            style={{ top: pos.top, right: pos.right }}
          >
            <IaCMenuItem
              label="Terraform"
              description="HCL · AWS / GCP / Azure"
              onClick={() => choose('terraform')}
            />
            <IaCMenuItem
              label="CloudFormation"
              description="YAML · AWS"
              onClick={() => choose('cloudformation')}
            />
          </div>,
          document.body,
        )}
    </>
  );
}

function IaCMenuItem({
  label,
  description,
  onClick,
}: {
  label: string;
  description: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center justify-between px-3 py-2 text-left transition-colors hover:bg-c2-primary-soft"
    >
      <span className="text-xs font-medium text-c2-text">{label}</span>
      <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        {description}
      </span>
    </button>
  );
}
