import { lazy, Suspense, memo, type ComponentType, type LazyExoticComponent } from 'react';
import { type LucideProps } from 'lucide-react';
import { lucideIconImports } from '@/lib/generated/lucideIconImports';

interface IconRendererProps extends Omit<LucideProps, 'ref'> {
  name: string;
}

type LucideComponent = ComponentType<Omit<LucideProps, 'ref'>>;
const lazyIcons: Record<string, LazyExoticComponent<LucideComponent>> = {};
for (const [name, loader] of Object.entries(lucideIconImports)) {
  lazyIcons[name] = lazy(loader as () => Promise<{ default: LucideComponent }>);
}

// Dispatch:
//   - "cdn:AWS/svg/Compute/Lambda.svg" → vendored same-origin SVG
//   - "local:aws/compute/EC2.svg"       → explicit vendored SVG
//   - "iconify:logos:kubernetes"       → vendored same-origin brand SVG
//   - "lucide:server" or "server"      → bundled lucide outline icon
// Vendor + brand icons keep native colors; lucide tints to the
// family color set in the node chip via inherited stroke.
function IconRendererInner({ name, ...props }: IconRendererProps) {
  if (name.startsWith('cdn:') || name.startsWith('local:')) {
    const path = name.startsWith('cdn:')
      ? vendoredPath(name.slice('cdn:'.length))
      : `/icons/${name.slice('local:'.length).replace(/^\/+/, '')}`;
    const size = props.size ?? 16;
    return (
      <img
        src={path}
        width={size}
        height={size}
        alt=""
        loading="lazy"
        decoding="async"
        // crossOrigin='anonymous' is safe on same-origin too and
        // primes html-to-image's serialization path.
        crossOrigin="anonymous"
        style={{ display: 'block', width: size, height: size }}
      />
    );
  }

  if (name.startsWith('iconify:')) {
    const id = name.slice('iconify:'.length);
    const size = props.size ?? 16;
    return (
      <img
        src={`/icons/brands/${id.replace(':', '/')}.svg`}
        width={size}
        height={size}
        alt=""
        loading="lazy"
        decoding="async"
        crossOrigin="anonymous"
        style={{ display: 'block' }}
      />
    );
  }

	const iconName = name.startsWith('lucide:') ? name.slice('lucide:'.length) : name;
	const LucideIcon = lazyIcons[iconName] ?? lazyIcons.circle!;

  return (
    <Suspense fallback={<div style={{ width: props.size ?? 16, height: props.size ?? 16 }} />}>
      <LucideIcon {...props} />
    </Suspense>
  );
}

function vendoredPath(cdnPath: string): string {
  const parts = cdnPath.split('/');
  const provider = (parts.shift() ?? '').toLowerCase();
  if (parts[0]?.toLowerCase() === 'svg') parts.shift();
  if (provider === 'aws' && parts.length > 1) {
    parts[0] = parts[0]!.toLowerCase();
  }
  return `/icons/${provider}/${parts.join('/')}`;
}

export const IconRenderer = memo(IconRendererInner);
