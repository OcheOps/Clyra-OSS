import { type FormEvent, useCallback, useEffect, useRef, useState } from 'react';
import toast from 'react-hot-toast';
import { generateStream } from '@/services/api/diagram';
import { maybeShowUpgradeNudge } from '@/features/billing/upgradeNudge';
import { useUsageStore } from '@/features/billing/hooks/useDailyUsage';
import { useBYOKeys } from '../hooks/useBYOKeys';
import { revealDiagram } from '../lib/revealDiagram';
import { usePromptSessionStore } from '../promptSessionStore';

// Empty-canvas prompt card. Shown only when the editor has no nodes and
// no project loaded. Gives a first-time user something to do instead of
// staring at a blank background. Calls the same streaming endpoint as
// the topbar PromptBar — no provider picker, no modes, no settings.
//
// Open-source build: managed Bedrock is disabled, so this card routes
// through the user's stored BYO key (OpenAI first, then Anthropic, then
// Gemini). If no key is set, submitting points them at Settings.

const SEED_PROMPTS = [
  'A three-tier web app on AWS — ALB, ECS Fargate, RDS Postgres, Redis, S3',
  'A serverless API with DynamoDB and Lambda',
  'A RAG assistant with S3, Bedrock, and OpenSearch',
];

const INITIAL_PROMPT_KEY = 'clyra-initial-prompt';

interface Props {
  onProjectRename?: (name: string) => void;
  /** Opens the reverse-from-IaC dialog. Optional so the component
   * still works in surfaces (tests, storybook) without the plumbing. */
  onOpenReverse?: () => void;
}

export function EmptyCanvasPrompt({ onProjectRename, onOpenReverse }: Props) {
  const [prompt, setPrompt] = useState('');
  const [loading, setLoading] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const abortRef = useRef<AbortController | null>(null);
  const setLastPrompt = usePromptSessionStore((s) => s.setLastPrompt);
  const dismissed = usePromptSessionStore((s) => s.emptyCardDismissed);
  const dismissCard = usePromptSessionStore((s) => s.dismissEmptyCard);
  const { getKey } = useBYOKeys();

  // Accept a template-stashed initial prompt and auto-submit it. Writes
  // from the template cards land here via sessionStorage — delete on
  // read so a refresh doesn't re-fire the generation.
  useEffect(() => {
    const stashed = sessionStorage.getItem(INITIAL_PROMPT_KEY);
    if (stashed) {
      sessionStorage.removeItem(INITIAL_PROMPT_KEY);
      setPrompt(stashed);
      // queue submit on next tick so state lands first
      setTimeout(() => submit(stashed), 0);
    } else {
      textareaRef.current?.focus();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const submit = useCallback(
    async (raw: string) => {
      const trimmed = raw.trim();
      if (!trimmed) {
        toast.error('Describe what you want to build');
        return;
      }
      if (loading) return;

      if (abortRef.current) {
        abortRef.current.abort();
        abortRef.current = null;
      }
      abortRef.current = new AbortController();
      setLoading(true);

      // Pick the first BYO key the user has set. Managed Bedrock is no
      // longer registered in the OSS build, so we always need a user-
      // supplied key here. If nothing is configured, send them to Settings.
      const providerOrder = [
        { name: 'openai' as const, label: 'OpenAI' },
        { name: 'anthropic' as const, label: 'Anthropic' },
        { name: 'gemini' as const, label: 'Gemini' },
      ];
      const chosen = providerOrder
        .map((p) => ({ ...p, key: getKey(p.name) }))
        .find((p) => p.key && p.key.length > 0);
      if (!chosen) {
        toast.error(
          'Add an OpenAI, Anthropic, or Gemini API key in Settings → BYO keys to generate.',
        );
        setLoading(false);
        return;
      }

      try {
        const diagram = await generateStream(
          {
            prompt: trimmed,
            provider: chosen.name,
            api_key: chosen.key,
            mode: 'solution_architecture',
          },
          () => {
            /* chunks visualized by PromptBar's phase indicator once the
               canvas has content; here we just wait for the final result. */
          },
          abortRef.current.signal,
        );
        revealDiagram(diagram);
        // Hand the prompt to the shared session so the topbar PromptBar
        // enters follow-up mode (2 refines remaining) on next render.
        setLastPrompt(trimmed);
        if (onProjectRename) {
          const name = trimmed.length <= 40 ? trimmed : trimmed.slice(0, 40) + '…';
          onProjectRename(name);
        }
        toast.success(`Generated ${diagram.nodes.length} nodes, ${diagram.edges.length} edges`);
        void useUsageStore.getState().refresh();
      } catch (err) {
        if ((err as Error).name !== 'AbortError') {
          if (!maybeShowUpgradeNudge(err)) {
            toast.error((err as Error).message || 'Generation failed');
          }
        }
      } finally {
        setLoading(false);
        abortRef.current = null;
      }
    },
    [loading, onProjectRename, setLastPrompt, getKey],
  );

  if (dismissed) return null;

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    void submit(prompt);
  };

  const pickSeed = (seed: string) => {
    setPrompt(seed);
    void submit(seed);
  };

  return (
    <div className="absolute inset-0 flex items-center justify-center px-4 pointer-events-none">
      <div className="relative w-full max-w-xl rounded-2xl border border-c2-hairline bg-c2-elevated/80 p-6 shadow-c2-lg backdrop-blur-md pointer-events-auto">
        <button
          type="button"
          onClick={dismissCard}
          title="Close — draw manually from the sidebar or use the top prompt bar"
          aria-label="Close"
          className="absolute right-3 top-3 flex h-7 w-7 items-center justify-center rounded-full text-c2-text-muted hover:text-c2-text hover:bg-c2-bg/60/60 transition-colors"
        >
          <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <path d="M6 6l12 12M18 6L6 18" />
          </svg>
        </button>
        <p className="text-micro font-medium uppercase tracking-[0.14em] text-c2-primary">
          New diagram
        </p>
        <h2 className="mt-2 text-xl font-semibold tracking-tight-ish text-c2-text">
          What are you building?
        </h2>
        <p className="mt-1.5 text-petit text-c2-text-muted">
          Describe your system in plain English. Clyra drafts the diagram —
          you refine it.
        </p>

        <form onSubmit={onSubmit} className="mt-5 space-y-3">
          <textarea
            ref={textareaRef}
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="A three-tier web app on AWS with Postgres and Redis…"
            rows={3}
            disabled={loading}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                e.preventDefault();
                void submit(prompt);
              }
            }}
            className="w-full resize-none rounded-lg border border-c2-hairline bg-c2-bg/70 px-3 py-2.5 text-sm text-c2-text placeholder:text-c2-text-muted/50 outline-none focus:border-c2-primary/50 focus:bg-c2-bg disabled:opacity-50"
          />

          <div className="flex items-center justify-between gap-3">
            <p className="text-micro text-c2-text-faint">
              ⌘/Ctrl + Enter to generate
            </p>
            <button
              type="submit"
              disabled={loading || !prompt.trim()}
              className="rounded-lg bg-c2-primary px-4 py-1.5 text-tiny font-medium text-white hover:bg-c2-primary/90 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              {loading ? 'Generating…' : 'Generate'}
            </button>
          </div>
        </form>

        <div className="mt-5 pt-5 border-t border-c2-hairline">
          <p className="text-micro text-c2-text-faint mb-2">Or try one of these</p>
          <div className="flex flex-wrap gap-2">
            {SEED_PROMPTS.map((seed) => (
              <button
                key={seed}
                type="button"
                disabled={loading}
                onClick={() => pickSeed(seed)}
                className="rounded-full border border-c2-hairline bg-c2-bg/50 px-3 py-1.5 text-micro text-c2-text-muted hover:text-c2-text hover:border-c2-border-strong disabled:opacity-40 disabled:cursor-not-allowed transition-colors text-left"
              >
                {truncate(seed, 50)}
              </button>
            ))}
          </div>
        </div>

        {/* Second CTA — importing existing infra is a much softer ask
            than 'design me a new system'. Anyone with a running AWS
            stack has a .tf file to paste; that's the lowest-friction
            way for a stranger to try Clyra and see value in 5 seconds. */}
        {onOpenReverse && (
          <div className="mt-5 pt-5 border-t border-c2-hairline">
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0">
                <p className="text-tiny font-medium text-c2-text">
                  Have existing Terraform?
                </p>
                <p className="mt-0.5 text-micro text-c2-text-muted">
                  Paste it in and we&rsquo;ll map it to a diagram in seconds.
                </p>
              </div>
              <button
                type="button"
                onClick={onOpenReverse}
                disabled={loading}
                className="shrink-0 rounded-lg border border-c2-border bg-c2-bg/60 px-3 py-1.5 text-tiny font-medium text-c2-text transition-colors hover:border-c2-border-strong hover:bg-c2-surface disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Import Terraform
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function truncate(s: string, n: number): string {
  return s.length <= n ? s : s.slice(0, n - 1) + '…';
}
