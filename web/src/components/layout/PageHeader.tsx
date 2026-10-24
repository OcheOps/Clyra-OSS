import { type ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';

interface PageHeaderProps {
  title: string;
  description?: string;
  /** Optional mono eyebrow that sits above the title. */
  eyebrow?: string;
  backTo?: string;
  actions?: ReactNode;
}

/**
 * PageHeader — v2 restyle.
 *
 * Used by signed-in subpages (Projects, Settings sub-routes, etc.) for
 * a consistent "back chevron + title + description + actions" header
 * rhythm. Optional `eyebrow` slot picks up the same mono small-caps
 * voice as every other v2 surface.
 */
export function PageHeader({
  title,
  description,
  eyebrow,
  backTo,
  actions,
}: PageHeaderProps) {
  const navigate = useNavigate();

  const handleBack = () => {
    if (backTo) {
      navigate(backTo);
    } else if (window.history.length > 1) {
      navigate(-1);
    } else {
      navigate('/dashboard');
    }
  };

  return (
    <header className="mb-10 flex items-start justify-between gap-4">
      <div className="flex items-start gap-3">
        <button
          type="button"
          onClick={handleBack}
          className="mt-1.5 flex h-7 w-7 items-center justify-center rounded-md border border-c2-border text-c2-text-muted transition-colors hover:border-c2-border-strong hover:bg-c2-surface hover:text-c2-text focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-c2-primary-ring"
          aria-label="Go back"
          title="Back"
        >
          <svg
            className="h-3.5 w-3.5"
            viewBox="0 0 20 20"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M12.5 4L6 10l6.5 6" />
          </svg>
        </button>
        <div>
          {eyebrow && (
            <div className="flex items-center gap-2">
              <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
              <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
                {eyebrow}
              </span>
            </div>
          )}
          <h1
            className={`font-display text-3xl font-semibold leading-[1.05] tracking-[-0.025em] text-c2-text md:text-[34px] ${
              eyebrow ? 'mt-3' : ''
            }`}
          >
            {title}
          </h1>
          {description && (
            <p className="mt-3 max-w-xl text-[14px] leading-relaxed text-c2-text-muted">
              {description}
            </p>
          )}
        </div>
      </div>
      {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
    </header>
  );
}
