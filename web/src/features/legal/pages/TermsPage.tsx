import { Link } from 'react-router-dom';

// Clyra is open source. Self-hosters write their own terms for their
// deployed instance. This page is a placeholder that redirects visitors
// to the project README; replace it with the real policy for your fork.
export function TermsPage() {
  return (
    <div className="min-h-screen flex items-center justify-center p-8 text-center">
      <div className="max-w-xl">
        <h1 className="text-2xl font-semibold mb-4">Terms of Service</h1>
        <p className="opacity-80 mb-4">
          This Clyra instance is self-hosted. The operator sets their own
          terms — ask whoever runs this deployment for the applicable
          policy.
        </p>
        <Link to="/" className="underline">← Back to editor</Link>
      </div>
    </div>
  );
}
