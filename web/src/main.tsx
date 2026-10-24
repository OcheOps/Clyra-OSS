import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import * as Sentry from '@sentry/react';
import App from './App';
import './index.css';

// Sentry — only initializes if DSN is set at build time. No DSN = silent skip.
const sentryDsn = import.meta.env.VITE_SENTRY_DSN;
if (sentryDsn) {
  Sentry.init({
    dsn: sentryDsn,
    environment: import.meta.env.MODE,
    tracesSampleRate: 0,
    replaysSessionSampleRate: 0,
    replaysOnErrorSampleRate: 0,
  });
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
      refetchOnWindowFocus: false,
    },
  },
});

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <Sentry.ErrorBoundary fallback={<div style={{ padding: 24, color: '#e2e8f0' }}>Something went wrong. Please refresh.</div>}>
        <App />
      </Sentry.ErrorBoundary>
      <Toaster
        position="bottom-right"
        toastOptions={{
          style: {
            background: '#1a1d27',
            color: '#e2e8f0',
            border: '1px solid #2a2d3a',
          },
        }}
      />
    </QueryClientProvider>
  </StrictMode>,
);
