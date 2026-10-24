import { type ReactNode } from 'react';
import { Topbar } from './Topbar';
import { GlobalUpgradeModal } from '@/features/billing/components/GlobalUpgradeModal';

interface AppLayoutProps {
  topbarChildren?: ReactNode;
  children: ReactNode;
}

export function AppLayout({ topbarChildren, children }: AppLayoutProps) {
  return (
    // bg-c2-bg + font-brand bring the signed-in app onto the v2 substrate.
    // Topbar internals haven't migrated yet (that's the editor session) but
    // the slight delta between Topbar's bg-clyra-elevated and c2-bg is
    // close enough to hold as a transition state — and the moment any
    // page below opts into c2-* tokens, it inherits the right substrate
    // without any further per-page wiring.
    <div className="flex h-screen flex-col bg-c2-bg font-brand text-c2-text">
      <Topbar>{topbarChildren}</Topbar>
      <div className="flex flex-1 overflow-hidden">{children}</div>
      <GlobalUpgradeModal />
    </div>
  );
}
