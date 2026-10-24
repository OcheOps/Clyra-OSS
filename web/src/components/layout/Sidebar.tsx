interface SidebarProps {
  side: 'left' | 'right';
  width?: string;
  children: React.ReactNode;
  className?: string;
}

export function Sidebar({ side, width = 'w-56', children, className = '' }: SidebarProps) {
  const borderClass = side === 'left' ? 'border-r' : 'border-l';

  return (
    <aside
      className={`hidden md:block ${width} shrink-0 overflow-y-auto ${borderClass} border-c2-hairline bg-c2-surface/40 ${className}`}
    >
      {children}
    </aside>
  );
}

interface SidebarSectionProps {
  title: string;
  children: React.ReactNode;
}

export function SidebarSection({ title, children }: SidebarSectionProps) {
  return (
    <div className="p-3">
      <div className="mb-3 flex items-center gap-2">
        <span className="inline-block h-1 w-1 rounded-full bg-c2-primary" />
        <h3 className="font-mono text-[10px] uppercase tracking-[0.22em] text-c2-text-faint">
          {title}
        </h3>
      </div>
      {children}
    </div>
  );
}
