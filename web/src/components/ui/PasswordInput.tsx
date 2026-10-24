import { type InputHTMLAttributes, forwardRef, useState } from 'react';

interface PasswordInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string;
  error?: string;
}

export const PasswordInput = forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ label, error, className = '', id, ...props }, ref) => {
    const [show, setShow] = useState(false);
    const inputId = id ?? label?.toLowerCase().replace(/\s+/g, '-');

    const base =
      'w-full rounded-lg px-3 py-2 pr-10 text-sm transition-colors duration-150 ' +
      'border border-line-default bg-clyra-bg text-ink-primary placeholder:text-ink-muted ' +
      'focus:border-brand-500/50 focus:outline-none focus:ring-1 focus:ring-brand-500/30 ' +
      '[body.paper_&]:border-stone-default [body.paper_&]:bg-paper-raised ' +
      '[body.paper_&]:text-charcoal [body.paper_&]:placeholder:text-charcoal-faint ' +
      '[body.paper_&]:focus:border-teal-500/60 [body.paper_&]:focus:ring-teal-500/20 ' +
      'disabled:opacity-40';

    const errorClasses = error
      ? 'border-danger/40 focus:border-danger/60 focus:ring-danger/30 ' +
        '[body.paper_&]:border-danger/60'
      : '';

    return (
      <div className="space-y-1.5">
        {label && (
          <label
            htmlFor={inputId}
            className="block text-petit font-medium text-ink-secondary [body.paper_&]:text-charcoal-soft"
          >
            {label}
          </label>
        )}
        <div className="relative">
          <input
            ref={ref}
            id={inputId}
            type={show ? 'text' : 'password'}
            className={`${base} ${errorClasses} ${className}`}
            {...props}
          />
          <button
            type="button"
            tabIndex={-1}
            onClick={() => setShow((s) => !s)}
            aria-label={show ? 'Hide password' : 'Show password'}
            title={show ? 'Hide password' : 'Show password'}
            className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-ink-muted hover:bg-clyra-overlay hover:text-ink-primary transition-colors [body.paper_&]:text-charcoal-faint [body.paper_&]:hover:bg-paper-soft [body.paper_&]:hover:text-charcoal"
          >
            {show ? <EyeOffIcon /> : <EyeIcon />}
          </button>
        </div>
        {error && <p className="text-petit text-danger">{error}</p>}
      </div>
    );
  },
);

PasswordInput.displayName = 'PasswordInput';

function EyeIcon() {
  return (
    <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

function EyeOffIcon() {
  return (
    <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24" />
      <line x1="1" y1="1" x2="23" y2="23" />
    </svg>
  );
}
