import { useCallback, useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import type { IaCCloud, IaCFormat, IaCGenerateResponse } from '@/types/iac';
import type { ProviderName } from '@/types/diagram';
import { useIaC } from '../hooks/useIaC';
import { useBYOKeys, type BYOProvider } from '../hooks/useBYOKeys';
import { useAuthStore } from '@/features/auth/store';
import { useUpgradeModalStore } from '@/features/billing/upgradeModalStore';

interface IaCPanelProps {
  open: boolean;
  format: IaCFormat;
  onClose: () => void;
  projectName: string;
}

type ProviderOption = 'managed' | 'openai' | 'anthropic' | 'gemini';

// Managed Bedrock is disabled in open-source builds — bring your own
// OpenAI/Anthropic/Gemini key instead.
const PROVIDER_OPTIONS: { value: ProviderOption; label: string }[] = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'gemini', label: 'Gemini' },
];

function resolveProvider(opt: ProviderOption): ProviderName {
  return opt === 'managed' ? 'managed_bedrock' : opt;
}

// At launch we ship the AWS structured renderer only — GCP and Azure still
// use a legacy prompt-only path that produces lower-confidence output.
// Hiding them avoids selling a feature whose quality we can't stand behind
// yet. They come back here once their structured renderers ship.
const CLOUDS: { value: IaCCloud; label: string; comingSoon?: boolean }[] = [
  { value: 'aws', label: 'AWS' },
  { value: 'gcp', label: 'GCP', comingSoon: true },
  { value: 'azure', label: 'Azure', comingSoon: true },
];

const EXTENSIONS: Record<IaCFormat, string> = {
  terraform: 'tf',
  cloudformation: 'yaml',
};

const LABELS: Record<IaCFormat, string> = {
  terraform: 'Terraform',
  cloudformation: 'CloudFormation',
};

export function IaCPanel({ open, format, onClose, projectName }: IaCPanelProps) {
  const { loading, result, error, generate, reset } = useIaC();
  const { getKey, setKey } = useBYOKeys();
  // Mirror PromptBar's BYO gate. Free / Pro users see the BYO buttons but
  // they're locked — clicking opens the upgrade modal pre-selected to BYO,
  // never flips local state to a provider the backend would reject.
  const user = useAuthStore((s) => s.user);
  // effective_plan respects expiry — a lapsed BYO shouldn't see the
  // BYO provider picker (backend enforceBYOKey would 403 the call).
  const canUseBYO = (user?.effective_plan ?? user?.plan) === 'byo';
  const openUpgrade = useUpgradeModalStore((s) => s.openModal);
  const [cloud, setCloud] = useState<IaCCloud>('aws');
  const [providerOpt, setProviderOpt] = useState<ProviderOption>('openai');
  const [explain, setExplain] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [showKey, setShowKey] = useState(false);
  const [currentFormat, setCurrentFormat] = useState<IaCFormat>(format);

  const isByo = providerOpt !== 'managed';
  const apiKey = isByo ? getKey(providerOpt as BYOProvider) : '';

  useEffect(() => {
    setCurrentFormat(format);
  }, [format]);

  // CloudFormation is AWS-only — snap cloud back if the user switches to it from a non-AWS selection.
  useEffect(() => {
    if (currentFormat === 'cloudformation' && cloud !== 'aws') {
      setCloud('aws');
    }
  }, [currentFormat, cloud]);

  // Reset result when the panel closes, so reopening starts fresh-looking.
  useEffect(() => {
    if (!open) reset();
  }, [open, reset]);

  const handleGenerate = useCallback(() => {
    if (isByo && !apiKey.trim()) {
      toast.error('Enter your API key in settings');
      setShowSettings(true);
      return;
    }
    generate({
      format: currentFormat,
      cloud,
      provider: resolveProvider(providerOpt),
      apiKey: isByo ? apiKey : undefined,
      explain,
    });
  }, [isByo, apiKey, generate, currentFormat, cloud, providerOpt, explain]);

  const handleCopy = useCallback(() => {
    if (!result?.code) return;
    navigator.clipboard.writeText(result.code);
    toast.success('Copied to clipboard');
  }, [result]);

  const handleDownload = useCallback(() => {
    if (!result?.code) return;
    const ext = EXTENSIONS[result.format];
    const safeName = projectName.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '') || 'diagram';
    const filename = `${safeName}.${ext}`;
    const blob = new Blob([result.code], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, [result, projectName]);

  if (!open) return null;

  return (
    <div className="fixed right-0 top-14 bottom-0 w-[460px] border-l border-c2-border bg-c2-surface/95/95 backdrop-blur-md shadow-c2-lg z-30 flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-c2-border px-4 py-3">
        <div>
          <h2 className="text-sm font-semibold text-c2-text">Generate Infrastructure</h2>
          <p className="mt-0.5 text-micro text-c2-text-faint">
            Starter scaffold — review before apply
          </p>
        </div>
        <button
          onClick={onClose}
          className="rounded-lg p-1 text-c2-text-muted hover:bg-c2-bg/60 hover:text-c2-text transition-colors"
          aria-label="Close"
        >
          <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>

      {/* Controls */}
      <div className="space-y-3 border-b border-c2-hairline px-4 py-3">
        {/* Format tabs */}
        <div className="flex rounded-lg border border-c2-border bg-c2-bg/50">
          {(Object.keys(LABELS) as IaCFormat[]).map((f) => (
            <button
              key={f}
              type="button"
              onClick={() => setCurrentFormat(f)}
              className={`flex-1 px-2.5 py-1.5 text-tiny font-medium transition-all first:rounded-l-clyra last:rounded-r-clyra ${
                currentFormat === f
                  ? 'bg-c2-primary-soft text-c2-primary'
                  : 'text-c2-text-muted hover:text-c2-text'
              }`}
            >
              {LABELS[f]}
            </button>
          ))}
        </div>

        {/* Cloud + explain */}
        <div className="flex items-center gap-2">
          <div className="flex flex-1 rounded-lg border border-c2-border bg-c2-bg/50">
            {CLOUDS.map((c) => {
              const cfOnlyAws = currentFormat === 'cloudformation' && c.value !== 'aws';
              const disabled = cfOnlyAws || !!c.comingSoon;
              const tip = c.comingSoon
                ? `${c.label} support is coming soon`
                : cfOnlyAws
                  ? 'CloudFormation is AWS-only'
                  : undefined;
              return (
                <button
                  key={c.value}
                  type="button"
                  onClick={() => !disabled && setCloud(c.value)}
                  disabled={disabled}
                  className={`flex-1 px-2 py-1 text-micro font-medium transition-all first:rounded-l-clyra last:rounded-r-clyra disabled:cursor-not-allowed disabled:opacity-30 ${
                    cloud === c.value && !disabled
                      ? 'bg-c2-primary-soft text-c2-primary'
                      : 'text-c2-text-muted/70 hover:text-c2-text'
                  }`}
                  title={tip}
                >
                  {c.label}
                  {c.comingSoon && <span className="ml-1 text-[9px] uppercase tracking-wider text-c2-text-faint">soon</span>}
                </button>
              );
            })}
          </div>

          <label className="flex shrink-0 cursor-pointer items-center gap-1.5 text-micro text-c2-text-muted">
            <input
              type="checkbox"
              checked={explain}
              onChange={(e) => setExplain(e.target.checked)}
              className="h-3 w-3 accent-c2-primary"
            />
            Explain
          </label>

          <button
            type="button"
            onClick={() => setShowSettings((s) => !s)}
            className={`shrink-0 rounded-lg p-1 transition-colors ${
              showSettings ? 'text-c2-primary' : 'text-c2-text-muted hover:text-c2-text-muted'
            }`}
            title="Provider settings"
          >
            <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="3" />
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
            </svg>
          </button>
        </div>

        <div className="rounded-lg border border-c2-border bg-c2-bg/50 px-3 py-2 text-micro text-c2-text-muted">
          AWS Terraform &amp; CloudFormation pass <span className="text-c2-text">terraform validate</span> and{' '}
          <span className="text-c2-text">cfn-lint</span> on the structured renderer. GCP &amp; Azure are coming soon.
        </div>

        {showSettings && (
          <div className="space-y-2 rounded-lg border border-c2-border bg-c2-elevated/70 p-2.5">
            <div className="space-y-1">
              <label className="block text-micro font-medium uppercase tracking-wider text-c2-text-muted/70">
                AI Provider
              </label>
              <div className="flex rounded-lg border border-c2-border bg-c2-bg/50">
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
                      className={`flex-1 px-2 py-1 text-micro font-medium transition-all first:rounded-l-clyra last:rounded-r-clyra ${
                        providerOpt === p.value && !locked
                          ? 'bg-c2-primary-soft text-c2-primary'
                          : locked
                            ? 'text-c2-text-faint hover:text-c2-text-muted'
                            : 'text-c2-text-muted/60 hover:text-c2-text'
                      }`}
                      title={locked ? 'Upgrade to BYO to use your own keys' : p.label}
                    >
                      {p.label}
                      {locked && <span className="ml-1 text-c2-text-faint">🔒</span>}
                    </button>
                  );
                })}
              </div>
            </div>
            {isByo && (
              <div className="space-y-1">
                <label className="block text-micro font-medium uppercase tracking-wider text-c2-text-muted/70">
                  API Key
                </label>
                <div className="relative">
                  <input
                    type={showKey ? 'text' : 'password'}
                    value={apiKey}
                    onChange={(e) => setKey(providerOpt as BYOProvider, e.target.value)}
                    placeholder={`${providerOpt === 'openai' ? 'OpenAI' : providerOpt === 'anthropic' ? 'Anthropic' : 'Gemini'} API key`}
                    className="w-full rounded-lg border border-c2-border bg-c2-bg/50 px-2.5 py-1.5 pr-12 text-tiny text-c2-text placeholder:text-c2-text-muted/40 outline-none focus:border-c2-primary/40"
                  />
                  <button
                    type="button"
                    tabIndex={-1}
                    onClick={() => setShowKey((s) => !s)}
                    className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded-lg px-1.5 py-0.5 text-micro font-medium text-c2-text-muted hover:text-c2-text hover:bg-c2-bg/60 transition-colors"
                  >
                    {showKey ? 'Hide' : 'Show'}
                  </button>
                </div>
                <p className="text-micro text-c2-text-faint">
                  Stored in your browser only.
                </p>
              </div>
            )}
          </div>
        )}

        <button
          type="button"
          onClick={handleGenerate}
          disabled={loading}
          className="w-full rounded-lg bg-c2-primary-soft px-3 py-1.5 text-xs font-medium text-c2-primary transition-all hover:bg-c2-primary/25 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading
            ? `Generating ${LABELS[currentFormat]}...`
            : result
              ? `Regenerate ${LABELS[currentFormat]}`
              : `Generate ${LABELS[currentFormat]}`}
        </button>
      </div>

      {/* Result area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {error && (
          <div className="mx-4 mt-3 rounded-lg border border-danger/40 bg-danger/10 px-3 py-2 text-tiny text-danger">
            {error}
          </div>
        )}

        {!result && !loading && !error && (
          <div className="flex flex-1 items-center justify-center px-6">
            <div className="space-y-2 text-center">
              <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-2xl bg-c2-primary-soft text-c2-primary">
                <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M16 18l6-6-6-6" />
                  <path d="M8 6l-6 6 6 6" />
                </svg>
              </div>
              <p className="text-tiny text-c2-text-muted">
                Generates starter {LABELS[currentFormat]} code from your diagram.
              </p>
              <p className="text-micro text-c2-text-faint">
                Max 25 resources. Not production-ready.
              </p>
            </div>
          </div>
        )}

        {loading && (
          <div className="flex flex-1 items-center justify-center">
            <div className="flex items-center gap-2 text-tiny text-c2-text-muted">
              <span className="h-1 w-1 rounded-full bg-c2-primary animate-bounce" style={{ animationDelay: '0ms' }} />
              <span className="h-1 w-1 rounded-full bg-c2-primary animate-bounce" style={{ animationDelay: '150ms' }} />
              <span className="h-1 w-1 rounded-full bg-c2-primary animate-bounce" style={{ animationDelay: '300ms' }} />
              <span className="ml-2">Synthesizing infrastructure...</span>
            </div>
          </div>
        )}

        {result && !loading && (
          <ResultView
            result={result}
            onCopy={handleCopy}
            onDownload={handleDownload}
          />
        )}
      </div>
    </div>
  );
}

interface ResultViewProps {
  result: IaCGenerateResponse;
  onCopy: () => void;
  onDownload: () => void;
}

function ResultView({ result, onCopy, onDownload }: ResultViewProps) {
  return (
    <>
      {/* Action bar */}
      <div className="flex items-center gap-2 border-b border-c2-hairline px-4 py-2">
        <span className="text-micro text-c2-text-faint">
          {result.code.split('\n').length} lines · {result.language}
        </span>
        <div className="ml-auto flex items-center gap-1">
          <button
            onClick={onCopy}
            className="rounded-lg border border-c2-border bg-c2-bg/50 px-2 py-1 text-tiny text-c2-text-muted hover:text-c2-text hover:border-c2-border-strong transition-colors"
          >
            Copy
          </button>
          <button
            onClick={onDownload}
            className="rounded-lg border border-c2-border bg-c2-bg/50 px-2 py-1 text-tiny text-c2-text-muted hover:text-c2-text hover:border-c2-border-strong transition-colors"
          >
            Download
          </button>
        </div>
      </div>

      {/* Warnings */}
      {result.warnings && result.warnings.length > 0 && (
        <div className="mx-4 mt-2 rounded-lg border border-warning/40 bg-warning/10 px-3 py-1.5 text-micro text-warning">
          {result.warnings.map((w, i) => (
            <div key={i}>• {w}</div>
          ))}
        </div>
      )}

      {/* Explanation */}
      {result.explanation && (
        <div className="mx-4 mt-2 rounded-lg border border-c2-hairline bg-c2-bg/60 p-2.5 text-micro text-c2-text-muted whitespace-pre-wrap">
          {result.explanation}
        </div>
      )}

      {/* Code */}
      <div className="flex-1 overflow-auto px-4 py-3">
        <pre className="whitespace-pre text-tiny leading-relaxed text-c2-text font-mono">
          <code>{result.code}</code>
        </pre>
      </div>
    </>
  );
}
