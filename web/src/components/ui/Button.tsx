import { type ButtonHTMLAttributes, forwardRef } from 'react';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
type Size = 'sm' | 'md' | 'lg';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
}

const variantStyles: Record<Variant, string> = {
  primary:
    'bg-brand-600 text-white hover:bg-brand-500 active:bg-brand-700 shadow-clyra-sm ' +
    '[body.paper_&]:bg-charcoal [body.paper_&]:text-paper [body.paper_&]:hover:bg-charcoal-soft ' +
    '[body.paper_&]:active:bg-charcoal [body.paper_&]:shadow-none',
  secondary:
    'bg-clyra-raised text-ink-primary border border-line-default hover:bg-clyra-overlay hover:border-line-strong ' +
    '[body.paper_&]:bg-paper-raised [body.paper_&]:text-charcoal ' +
    '[body.paper_&]:border-stone-default [body.paper_&]:hover:border-stone-strong [body.paper_&]:hover:bg-paper-soft',
  ghost:
    'text-ink-secondary hover:text-ink-primary hover:bg-clyra-raised ' +
    '[body.paper_&]:text-charcoal-muted [body.paper_&]:hover:text-charcoal [body.paper_&]:hover:bg-paper-soft',
  danger:
    'border border-danger/20 text-danger hover:bg-danger-muted',
};

const sizeStyles: Record<Size, string> = {
  sm: 'px-2.5 py-1 text-petit',
  md: 'px-3.5 py-1.5 text-sm',
  lg: 'px-5 py-2 text-sm',
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'md', loading, className = '', children, disabled, ...props }, ref) => {
    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={`inline-flex items-center justify-center gap-2 rounded-clyra font-medium transition-all duration-150 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-brand-500/50 focus-visible:ring-offset-1 focus-visible:ring-offset-clyra-bg disabled:pointer-events-none disabled:opacity-40 ${variantStyles[variant]} ${sizeStyles[size]} ${className}`}
        {...props}
      >
        {loading && (
          <svg className="h-3.5 w-3.5 animate-spin" viewBox="0 0 24 24" fill="none">
            <circle className="opacity-20" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
            <path className="opacity-80" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
        )}
        {children}
      </button>
    );
  },
);

Button.displayName = 'Button';
