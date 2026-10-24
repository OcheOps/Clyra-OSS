import { useState } from 'react';
import toast from 'react-hot-toast';
import { useBYOKeys, type BYOProvider } from '@/features/editor/hooks/useBYOKeys';

interface ProviderMeta {
  id: BYOProvider;
  label: string;
  hint: string;
  placeholder: string;
  docsUrl: string;
}

const PROVIDERS: ProviderMeta[] = [
  {
    id: 'openai',
    label: 'OpenAI',
    hint: 'Starts with sk-…',
    placeholder: 'sk-...',
    docsUrl: 'https://platform.openai.com/api-keys',
  },
  {
    id: 'anthropic',
    label: 'Anthropic',
    hint: 'Starts with sk-ant-…',
    placeholder: 'sk-ant-...',
    docsUrl: 'https://console.anthropic.com/settings/keys',
  },
  {
    id: 'gemini',
    label: 'Google Gemini',
    hint: 'Get a key from AI Studio.',
    placeholder: 'AIza...',
    docsUrl: 'https://aistudio.google.com/app/apikey',
  },
];

/**
 * BYOKeysSection — v2 restyle.
 *
 * Shown on Settings only when the user is actually on the BYO plan.
 * Three provider rows, each with: provider label + "Get a key ↗"
 * docs link, masked input with show/hide eye toggle, Save + Remove
 * buttons. Storage logic lives in useBYOKeys — keys stay in
 * localStorage, never transmitted to our servers.
 */
export function BYOKeysSection() {
  const { getKey, setKey, clearKey, hasKey } = useBYOKeys();

  return (
    <section className="overflow-hidden rounded-2xl border border-c2-border bg-c2-surface/60 p-6 transition-colors hover:border-c2-border-strong">
      <div className="flex items-center gap-2">
        <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
        <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
          AI keys
        </span>
      </div>
      <h2 className="mt-3 font-display text-base font-semibold tracking-[-0.005em] text-c2-text">
        Your AI keys
      </h2>
      <p className="mt-2 text-petit text-c2-text-muted">
        Stored in your browser only. We never see or transmit them to our servers.
      </p>

      <div className="mt-6 space-y-5">
        {PROVIDERS.map((p) => (
          <KeyRow
            key={p.id}
            meta={p}
            value={getKey(p.id)}
            present={hasKey(p.id)}
            onSave={(v) => {
              setKey(p.id, v);
              toast.success(`${p.label} key saved`);
            }}
            onClear={() => {
              clearKey(p.id);
              toast.success(`${p.label} key removed`);
            }}
          />
        ))}
      </div>
    </section>
  );
}

function KeyRow({
  meta,
  value,
  present,
  onSave,
  onClear,
}: {
  meta: ProviderMeta;
  value: string;
  present: boolean;
  onSave: (v: string) => void;
  onClear: () => void;
}) {
  const [draft, setDraft] = useState(value);
  const [show, setShow] = useState(false);
  const dirty = draft.trim() !== value;

  return (
    <div className="space-y-2">
      <div className="flex items-baseline justify-between">
        <label className="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
          <span>{meta.label}</span>
          {present && (
            <span className="inline-flex items-center gap-1 rounded-full border border-c2-accent/30 bg-c2-accent-soft px-1.5 py-0.5 text-[8px] font-semibold text-c2-accent">
              <span className="h-1 w-1 rounded-full bg-c2-accent" />
              Saved
            </span>
          )}
        </label>
        <a
          href={meta.docsUrl}
          target="_blank"
          rel="noreferrer"
          className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-primary transition-colors hover:text-c2-info"
        >
          Get a key ↗
        </a>
      </div>

      <div className="flex gap-2">
        <div className="relative flex-1">
          <input
            type={show ? 'text' : 'password'}
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            placeholder={meta.placeholder}
            autoComplete="off"
            spellCheck={false}
            className="w-full rounded-lg border border-c2-border bg-c2-bg/40 px-3 py-2 pr-9 font-mono text-[13px] text-c2-text placeholder:text-c2-text-faint focus:border-c2-primary/60 focus:bg-c2-surface focus:outline-none focus:ring-2 focus:ring-c2-primary-ring/40 transition-colors"
          />
          <button
            type="button"
            onClick={() => setShow((s) => !s)}
            className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-c2-text-faint transition-colors hover:text-c2-text"
            aria-label={show ? 'Hide key' : 'Show key'}
          >
            {show ? <EyeOff /> : <Eye />}
          </button>
        </div>

        <button
          type="button"
          disabled={!dirty}
          onClick={() => onSave(draft)}
          className="rounded-lg bg-c2-primary px-4 py-2 text-tiny font-medium text-c2-bg transition-all hover:bg-c2-primary-hover disabled:pointer-events-none disabled:opacity-30"
        >
          Save
        </button>
        {present && (
          <button
            type="button"
            onClick={() => {
              onClear();
              setDraft('');
            }}
            className="rounded-lg border border-c2-border bg-c2-surface/40 px-3 py-2 text-tiny font-medium text-c2-text-muted transition-colors hover:border-c2-danger/40 hover:bg-c2-danger/10 hover:text-c2-danger"
          >
            Remove
          </button>
        )}
      </div>

      <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-c2-text-faint">
        {meta.hint}
      </p>
    </div>
  );
}

function Eye() {
  return (
    <svg className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
      <path d="M10 4C5 4 1.73 7.11 1 10c.73 2.89 4 6 9 6s8.27-3.11 9-6c-.73-2.89-4-6-9-6zm0 10a4 4 0 110-8 4 4 0 010 8zm0-6a2 2 0 100 4 2 2 0 000-4z" />
    </svg>
  );
}
function EyeOff() {
  return (
    <svg className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
      <path d="M3.707 2.293a1 1 0 10-1.414 1.414l14 14a1 1 0 001.414-1.414l-14-14zM10 14c-5 0-8.27-3.11-9-4 .36-1.41 1.5-3.02 3.18-4.22L5.6 7.2A4 4 0 0010 14zm3.38-1.4L8.4 7.6A4 4 0 0114 12c0 .21-.01.4-.04.6.65-.58 1.18-1.24 1.55-1.9-.7-1.2-2.11-2.7-3.91-3.6l-1.03-1.03A9.72 9.72 0 0110 6c5 0 8.27 3.11 9 4-.3 1.13-1.05 2.4-2.27 3.5l-3.35-.9z" />
    </svg>
  );
}
