import { useUpgradePromptStore, type UpgradeReason } from './upgradePromptStore';

// Plain <a> on purpose — this modal renders as a sibling of RouterProvider
// in App.tsx, so it lives outside the router context. <Link> needs that
// context and crashes the whole app with "Cannot destructure property
// 'basename' of useContext(...) as it is null". /billing is a full page
// nav anyway; the modal closes via hide() in the click handler.

// UpgradePromptModal is a global modal that appears whenever the backend
// returns 402 (a Pro/BYO feature was attempted on free). Mounted once in
// App.tsx so every feature can trigger it via the store.

const COPY: Record<UpgradeReason, { title: string; body: string }> = {
  cloud_sync: {
    title: 'Cloud sync is a Pro feature',
    body: 'Keep your projects available on every device — or bring your own AI keys and save on managed generations.',
  },
  iac_export: {
    title: "You've used your free IaC export",
    body: 'Free includes one lifetime Terraform or CloudFormation export. Upgrade to Pro for unlimited — or BYO with your own AI keys.',
  },
  generic: {
    title: 'This feature requires a paid plan',
    body: 'Upgrade to Pro for managed AI, or BYO to use your own OpenAI / Anthropic / Gemini key.',
  },
};

export function UpgradePromptModal() {
  const reason = useUpgradePromptStore((s) => s.reason);
  const hide = useUpgradePromptStore((s) => s.hide);

  if (!reason) return null;
  const { title, body } = COPY[reason];

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="upgrade-title"
      className="fixed inset-0 z-50 flex items-center justify-center bg-charcoal/40 backdrop-blur-sm px-4"
      onClick={hide}
    >
      <div
        className="w-full max-w-md rounded-xl bg-paper-raised border border-stone-subtle shadow-paper-lg p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="upgrade-title" className="text-lg font-semibold text-charcoal">
          {title}
        </h2>
        <p className="mt-2 text-petit text-charcoal-muted leading-relaxed">{body}</p>

        <div className="mt-6 flex flex-col gap-2">
          <a
            href="/billing"
            onClick={hide}
            className="block w-full rounded-lg bg-charcoal px-4 py-2.5 text-center text-sm font-medium text-paper hover:bg-charcoal-soft transition-colors"
          >
            View plans
          </a>
          <button
            type="button"
            onClick={hide}
            className="block w-full rounded-lg border border-stone-default px-4 py-2.5 text-center text-sm font-medium text-charcoal-soft hover:border-stone-strong transition-colors"
          >
            Not now
          </button>
        </div>
      </div>
    </div>
  );
}
