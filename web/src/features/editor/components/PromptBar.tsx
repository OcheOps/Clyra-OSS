import { type FormEvent, useState, useRef, useCallback, useEffect } from 'react';
import type { ProviderName, ArchitectureMode } from '@/types/diagram';
import { generateStream } from '@/services/api/diagram';
import { useEditorStore } from '../store';
import { revealDiagram } from '../lib/revealDiagram';
import { usePromptSessionStore } from '../promptSessionStore';
import { useBYOKeys, type BYOProvider } from '../hooks/useBYOKeys';
import toast from 'react-hot-toast';
import { maybeShowUpgradeNudge } from '@/features/billing/upgradeNudge';
import { useAuthStore } from '@/features/auth/store';
import { useUpgradeModalStore } from '@/features/billing/upgradeModalStore';
import { useUsageStore } from '@/features/billing/hooks/useDailyUsage';

const MODES: { value: ArchitectureMode; label: string; icon: string }[] = [
  { value: 'solution_architecture', label: 'Solution', icon: '🧩' },
  { value: 'network_architecture', label: 'Network', icon: '🔗' },
  { value: 'enterprise_architecture', label: 'Enterprise', icon: '🏢' },
];

type ProviderOption = 'managed' | 'openai' | 'anthropic' | 'gemini';

// Managed Bedrock is disabled in open-source builds — the hosted managed
// path is gone; bring your own OpenAI/Anthropic/Gemini key instead.
const PROVIDER_OPTIONS: { value: ProviderOption; label: string }[] = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'gemini', label: 'Gemini' },
];

function resolveProvider(opt: ProviderOption): ProviderName {
  return opt === 'managed' ? 'managed_bedrock' : opt;
}

function deriveProjectName(prompt: string): string {
  const cleaned = prompt.replace(/["'`]/g, '').replace(/\s+/g, ' ').trim();
  if (cleaned.length <= 40) return cleaned;
  const cut = cleaned.slice(0, 40);
  const lastSpace = cut.lastIndexOf(' ');
  return (lastSpace > 15 ? cut.slice(0, lastSpace) : cut) + '...';
}

const LOADING_PHASES = [
  'Analyzing your architecture...',
  'Identifying components...',
  'Mapping relationships...',
  'Generating diagram layout...',
  'Almost there...',
];

const MAX_FOLLOWUPS = 2;

interface PromptBarProps {
  onProjectRename?: (name: string) => void;
}

export function PromptBar({ onProjectRename }: PromptBarProps) {
  const [prompt, setPrompt] = useState('');
  const [followup, setFollowup] = useState('');
  const mode = useEditorStore((s) => s.mode);
  const setMode = useEditorStore((s) => s.setMode);
  const [providerOpt, setProviderOpt] = useState<ProviderOption>('openai');
  const user = useAuthStore((s) => s.user);
  const openUpgrade = useUpgradeModalStore((s) => s.openModal);
  // BYO-provider selection gated: only users on the BYO plan can actually
  // point the generator at their own keys. Everyone else sees the tab but
  // clicking it opens the upgrade modal instead of flipping state.
  const canUseBYO = (user?.effective_plan ?? user?.plan) === 'byo';
  const { getKey, setKey } = useBYOKeys();
  const [showKey, setShowKey] = useState(false);
  const [loading, setLoading] = useState(false);
  const [loadingPhase, setLoadingPhase] = useState(0);
  const [showSettings, setShowSettings] = useState(false);
  // Shared with EmptyCanvasPrompt so a generation from either surface
  // transitions the user into follow-up mode in the topbar.
  const lastPrompt = usePromptSessionStore((s) => s.lastPrompt);
  const followupCount = usePromptSessionStore((s) => s.followupCount);
  const setLastPrompt = usePromptSessionStore((s) => s.setLastPrompt);
  const incrementFollowup = usePromptSessionStore((s) => s.incrementFollowup);
  const resetSession = usePromptSessionStore((s) => s.resetSession);
  const abortRef = useRef<AbortController | null>(null);
  const phaseTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const loadDiagram = useEditorStore((s) => s.loadDiagram);
  const getDiagram = useEditorStore((s) => s.getDiagram);
  const chunkCountRef = useRef(0);
  const followupInputRef = useRef<HTMLInputElement>(null);

  const isByo = providerOpt !== 'managed';
  const apiKey = isByo ? getKey(providerOpt as BYOProvider) : '';
  const hasGenerated = lastPrompt !== '' && !loading;
  const canFollowup = hasGenerated && followupCount < MAX_FOLLOWUPS;

  useEffect(() => {
    if (loading) {
      setLoadingPhase(0);
      chunkCountRef.current = 0;
      phaseTimerRef.current = setInterval(() => {
        setLoadingPhase((p) => Math.min(p + 1, LOADING_PHASES.length - 1));
      }, 3000);
    } else {
      if (phaseTimerRef.current) {
        clearInterval(phaseTimerRef.current);
        phaseTimerRef.current = null;
      }
    }
    return () => {
      if (phaseTimerRef.current) clearInterval(phaseTimerRef.current);
    };
  }, [loading]);

  // Focus followup input after generation
  useEffect(() => {
    if (canFollowup) {
      followupInputRef.current?.focus();
    }
  }, [canFollowup]);

  const handleChunk = useCallback(() => {
    chunkCountRef.current++;
    if (chunkCountRef.current === 5) setLoadingPhase((p) => Math.max(p, 2));
    else if (chunkCountRef.current === 15) setLoadingPhase((p) => Math.max(p, 3));
  }, []);

  const doGenerate = useCallback(
    async (userPrompt: string, isFollowup: boolean) => {
      // Prevent duplicate requests
      if (loading) return;

      if (isByo && !apiKey.trim()) {
        toast.error('Enter your API key in settings');
        setShowSettings(true);
        return;
      }

      // Cancel any in-flight request
      if (abortRef.current) {
        abortRef.current.abort();
        abortRef.current = null;
      }

      let fullPrompt = userPrompt;
      if (isFollowup) {
        const current = getDiagram();
        const nodeNames = current.nodes.map((n) => `${n.label} (${n.type})`).join(', ');
        fullPrompt = `I have an existing diagram with these components: ${nodeNames}.\n\nOriginal request: ${lastPrompt}\n\nPlease modify the diagram: ${userPrompt}`;
      }

      setLoading(true);
      abortRef.current = new AbortController();

      try {
        const diagram = await generateStream(
          {
            prompt: fullPrompt,
            provider: resolveProvider(providerOpt),
            mode,
            ...(isByo ? { api_key: apiKey } : {}),
          },
          handleChunk,
          abortRef.current.signal,
        );

        // First generation gets the staggered reveal for the "painting in"
        // effect. Follow-ups would be jarring to stagger — user already
        // sees a diagram, just swap it atomically.
        if (isFollowup) loadDiagram(diagram);
        else revealDiagram(diagram);

        if (isFollowup) {
          incrementFollowup();
          setFollowup('');
          toast.success('Diagram updated');
        } else {
          setLastPrompt(userPrompt);
          setFollowup('');
          toast.success(`Generated ${diagram.nodes.length} nodes, ${diagram.edges.length} edges`);
          if (onProjectRename) onProjectRename(deriveProjectName(userPrompt));
        }
        // Keep the topbar usage pill accurate right after a successful gen,
        // so free users see the counter tick toward the cap in real time.
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
    [providerOpt, mode, isByo, apiKey, handleChunk, loadDiagram, getDiagram, lastPrompt, onProjectRename, loading, incrementFollowup, setLastPrompt],
  );

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      const trimmed = prompt.trim();
      if (!trimmed) { toast.error('Enter a prompt'); return; }
      doGenerate(trimmed, false);
    },
    [prompt, doGenerate],
  );

  const handleFollowupSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      const trimmed = followup.trim();
      if (!trimmed) return;
      doGenerate(trimmed, true);
    },
    [followup, doGenerate],
  );

  const handleCancel = useCallback(() => {
    abortRef.current?.abort();
  }, []);

  const handleNewDiagram = useCallback(() => {
    resetSession();
    setFollowup('');
    setPrompt('');
  }, [resetSession]);

  return (
    <div className="flex flex-col gap-1.5">
      {/* Main prompt row. `min-w-0` on the inner input wrapper lets the
          input shrink below its intrinsic width so sibling buttons
          (Generate / Stop) never overlap the gear icon. */}
      <form onSubmit={handleSubmit} className="flex min-w-0 items-center gap-2">
        {/* Mode selector — compact native select so it only takes the
            width of its longest label instead of three inline tabs. */}
        <div className="relative shrink-0">
          <select
            value={mode}
            onChange={(e) => setMode(e.target.value as ArchitectureMode)}
            disabled={loading}
            title="Architecture mode"
            className="cursor-pointer appearance-none rounded-full border border-c2-border bg-c2-bg/40 py-1.5 pl-2.5 pr-7 text-micro font-medium text-c2-primary outline-none transition-colors focus:border-c2-primary/60 focus:bg-c2-surface disabled:opacity-40"
          >
            {MODES.map((m) => (
              <option key={m.value} value={m.value} className="bg-c2-elevated text-c2-text">
                {m.icon} {m.label}
              </option>
            ))}
          </select>
          <svg
            className="pointer-events-none absolute right-2 top-1/2 h-3 w-3 -translate-y-1/2 text-c2-text-faint"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden
          >
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </div>

        {/* Input or loading */}
        <div className="relative min-w-0 flex-1">
          {loading ? (
            <div className="flex items-center gap-2.5 rounded-full border border-c2-primary/30 bg-c2-primary-soft px-3 py-1.5">
              <div className="flex gap-1">
                <span
                  className="h-1 w-1 animate-bounce rounded-full bg-c2-primary"
                  style={{ animationDelay: '0ms' }}
                />
                <span
                  className="h-1 w-1 animate-bounce rounded-full bg-c2-primary"
                  style={{ animationDelay: '150ms' }}
                />
                <span
                  className="h-1 w-1 animate-bounce rounded-full bg-c2-primary"
                  style={{ animationDelay: '300ms' }}
                />
              </div>
              <span className="truncate text-tiny font-medium text-c2-primary">
                {LOADING_PHASES[loadingPhase]}
              </span>
            </div>
          ) : (
            <>
              <input
                type="text"
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                placeholder={
                  mode === 'network_architecture'
                    ? 'Describe your network topology…'
                    : mode === 'enterprise_architecture'
                      ? 'Describe your business systems…'
                      : 'Describe your architecture…'
                }
                className="w-full rounded-full border border-c2-border bg-c2-bg/40 px-3 py-1.5 pr-8 text-xs text-c2-text placeholder:text-c2-text-faint outline-none transition-colors focus:border-c2-primary/60 focus:bg-c2-surface focus:ring-2 focus:ring-c2-primary-ring/30"
              />
              <button
                type="button"
                onClick={() => setShowSettings((s) => !s)}
                className={`absolute right-2 top-1/2 -translate-y-1/2 transition-colors ${
                  showSettings ? 'text-c2-primary' : 'text-c2-text-faint hover:text-c2-text-muted'
                }`}
                title="Provider settings"
              >
                <svg
                  className="h-3.5 w-3.5"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
              </button>
            </>
          )}
        </div>

        {loading ? (
          <button
            type="button"
            onClick={handleCancel}
            className="rounded-full border border-c2-border bg-c2-bg/40 px-3 py-1 text-tiny font-medium text-c2-text-muted transition-colors hover:border-c2-border-strong hover:text-c2-text"
          >
            Stop
          </button>
        ) : (
          <button
            type="submit"
            className="rounded-full bg-c2-primary px-4 py-1 text-tiny font-medium text-c2-bg shadow-c2-glow-soft transition-all hover:bg-c2-primary-hover hover:shadow-c2-glow"
          >
            Generate
          </button>
        )}
      </form>

      {/* Follow-up bar — appears after generation, max 2 */}
      {canFollowup && (
        <form
          onSubmit={handleFollowupSubmit}
          className="flex items-center gap-2 pl-[1px]"
        >
          <div className="flex shrink-0 items-center gap-1 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
            <svg
              className="h-3 w-3"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <polyline points="9 14 4 9 9 4" />
              <path d="M20 20v-7a4 4 0 0 0-4-4H4" />
            </svg>
            <span>{MAX_FOLLOWUPS - followupCount} left</span>
          </div>
          <input
            ref={followupInputRef}
            type="text"
            value={followup}
            onChange={(e) => setFollowup(e.target.value)}
            placeholder='Tweak it…  e.g. "Add a Redis cache between API and DB"'
            disabled={loading}
            className="flex-1 rounded-full border border-c2-border bg-c2-bg/40 px-3 py-1 text-tiny text-c2-text placeholder:text-c2-text-faint outline-none transition-colors focus:border-c2-primary/60 focus:bg-c2-surface"
          />
          <button
            type="submit"
            disabled={loading || !followup.trim()}
            className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-primary transition-colors hover:text-c2-info disabled:cursor-default disabled:text-c2-text-faint"
          >
            Refine
          </button>
          <button
            type="button"
            onClick={handleNewDiagram}
            className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint transition-colors hover:text-c2-text"
          >
            New
          </button>
        </form>
      )}

      {/* Settings panel */}
      {showSettings && !loading && (
        <div className="space-y-3 rounded-xl border border-c2-border bg-c2-elevated/80 p-3 backdrop-blur-sm">
          <div className="space-y-2">
            <label className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              AI Provider
            </label>
            <div className="flex rounded-full border border-c2-border bg-c2-bg/40 p-1">
              {PROVIDER_OPTIONS.map((p) => {
                const locked = p.value !== 'managed' && !canUseBYO;
                return (
                  <button
                    key={p.value}
                    type="button"
                    onClick={() => {
                      if (locked) {
                        openUpgrade({ defaultPlan: 'byo' });
                        return;
                      }
                      setProviderOpt(p.value);
                    }}
                    className={`flex-1 rounded-full px-2 py-1 text-micro font-medium transition-all ${
                      providerOpt === p.value
                        ? 'bg-c2-primary text-c2-bg'
                        : locked
                          ? 'text-c2-text-faint hover:text-c2-text-muted'
                          : 'text-c2-text-muted hover:text-c2-text'
                    }`}
                    title={locked ? 'Upgrade to BYO to use your own keys' : p.label}
                  >
                    {p.label}
                    {locked && <span className="ml-1 opacity-70">🔒</span>}
                  </button>
                );
              })}
            </div>
          </div>

          {isByo ? (
            <div className="space-y-2">
              <label className="block font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                API Key
              </label>
              <div className="relative">
                <input
                  type={showKey ? 'text' : 'password'}
                  value={apiKey}
                  onChange={(e) => setKey(providerOpt as BYOProvider, e.target.value)}
                  placeholder={`${
                    providerOpt === 'openai'
                      ? 'OpenAI'
                      : providerOpt === 'anthropic'
                        ? 'Anthropic'
                        : 'Gemini'
                  } API key`}
                  className="w-full rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-1.5 pr-14 font-mono text-tiny text-c2-text placeholder:text-c2-text-faint outline-none transition-colors focus:border-c2-primary/60 focus:bg-c2-surface focus:ring-2 focus:ring-c2-primary-ring/30"
                />
                <button
                  type="button"
                  tabIndex={-1}
                  onClick={() => setShowKey((s) => !s)}
                  className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded-md px-1.5 py-0.5 font-mono text-[9px] font-medium uppercase tracking-[0.18em] text-c2-text-muted transition-colors hover:bg-c2-bg/60 hover:text-c2-text"
                >
                  {showKey ? 'Hide' : 'Show'}
                </button>
              </div>
              <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
                Stored locally · never transmitted except in the generation request
              </p>
            </div>
          ) : (
            <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
              Built-in AI · Amazon Bedrock
            </p>
          )}
        </div>
      )}
    </div>
  );
}
