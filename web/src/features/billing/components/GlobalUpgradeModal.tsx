import { useUpgradeModalStore } from '../upgradeModalStore';
import { UpgradeModal } from './UpgradeModal';

// Mount-once wrapper so any feature can call `useUpgradeModalStore.openModal()`
// (or `useUpgradeModalStore.getState().openModal()` outside React) without
// threading props or hoisting state. Put this in AppLayout.
export function GlobalUpgradeModal() {
  const { open, defaultPlan, headline, subline, closeModal } = useUpgradeModalStore();
  return (
    <UpgradeModal
      open={open}
      onClose={closeModal}
      defaultPlan={defaultPlan}
      headline={headline}
      subline={subline}
    />
  );
}
