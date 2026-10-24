import { type InputHTMLAttributes, forwardRef } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

// Input renders differently on paper surfaces (marketing, auth) vs the
// dark app. The `[body.paper_&]:` selectors kick in when the body has
// the .paper class from usePaperTheme().
export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, className = '', id, ...props }, ref) => {
    const inputId = id ?? label?.toLowerCase().replace(/\s+/g, '-');

    const base =
      'w-full rounded-lg px-3 py-2 text-sm transition-colors duration-150 ' +
      // Dark defaults (signed-in app).
      'border border-line-default bg-clyra-bg text-ink-primary placeholder:text-ink-muted ' +
      'focus:border-brand-500/50 focus:outline-none focus:ring-1 focus:ring-brand-500/30 ' +
      // Paper overrides (marketing / auth).
      '[body.paper_&]:border-stone-default [body.paper_&]:bg-paper-raised ' +
      '[body.paper_&]:text-charcoal [body.paper_&]:placeholder:text-charcoal-faint ' +
      '[body.paper_&]:focus:border-teal-500/60 [body.paper_&]:focus:ring-teal-500/20 ' +
      'disabled:opacity-40';

    const errorClasses = error
      ? 'border-danger/40 focus:border-danger/60 focus:ring-danger/30 ' +
        '[body.paper_&]:border-danger/60 [body.paper_&]:focus:ring-danger/20'
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
        <input
          ref={ref}
          id={inputId}
          className={`${base} ${errorClasses} ${className}`}
          {...props}
        />
        {error && <p className="text-petit text-danger">{error}</p>}
      </div>
    );
  },
);

Input.displayName = 'Input';
