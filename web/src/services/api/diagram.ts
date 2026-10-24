import { api } from './client';
import type { Diagram, GenerateRequest, StreamChunk } from '@/types/diagram';

function readCsrfToken(): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(/(?:^|; )([A-Za-z0-9_-]+_csrf)=([^;]*)/);
  return match ? decodeURIComponent(match[2]!) : '';
}

export function generate(req: GenerateRequest): Promise<Diagram> {
  return api<Diagram>('/diagrams/generate', {
    method: 'POST',
    body: req,
  });
}

/**
 * Streams diagram generation via SSE. Calls onChunk for each partial delta,
 * and returns the final parsed Diagram when the "result" event arrives.
 */
export function generateStream(
  req: GenerateRequest,
  onChunk: (chunk: StreamChunk) => void,
  signal?: AbortSignal,
): Promise<Diagram> {
  return new Promise<Diagram>((resolve, reject) => {
    // Early bail if already aborted
    if (signal?.aborted) {
      reject(new DOMException('Aborted', 'AbortError'));
      return;
    }

    const csrf = readCsrfToken();
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    if (csrf) headers['X-CSRF-Token'] = csrf;

    fetch('/api/v1/diagrams/generate/stream', {
      method: 'POST',
      credentials: 'include',
      headers,
      body: JSON.stringify(req),
      signal,
    })
      .then((res) => {
        if (!res.ok) {
          return res.json().then(
            (data) => reject(new Error(data?.error ?? `Request failed (${res.status})`)),
            () => reject(new Error(`Request failed (${res.status})`)),
          );
        }

        const reader = res.body?.getReader();
        if (!reader) {
          reject(new Error('Streaming not supported'));
          return;
        }

        const decoder = new TextDecoder();
        let buffer = '';
        let settled = false;
        // currentEvent persists across reads — without this, a chunk boundary
        // between the `event: result` and `data: {...}` lines would orphan
        // the data line (function-scoped state would reset on each call),
        // and a 20KB result payload would never resolve. Symptom: "Stream
        // ended without result" despite the backend logging a successful
        // 200 with a full body.
        let currentEvent = '';

        // Listen for abort to cancel the reader
        const onAbort = () => {
          if (!settled) {
            settled = true;
            reader.cancel().catch(() => {});
            reject(new DOMException('Aborted', 'AbortError'));
          }
        };
        signal?.addEventListener('abort', onAbort);

        function processLines() {
          const lines = buffer.split('\n');
          buffer = lines.pop() ?? '';

          for (const line of lines) {
            if (settled) return;
            if (line.startsWith('event: ')) {
              currentEvent = line.slice(7).trim();
            } else if (line.startsWith('data: ')) {
              const data = line.slice(6);
              try {
                const parsed = JSON.parse(data);
                if (currentEvent === 'chunk') {
                  onChunk(parsed as StreamChunk);
                } else if (currentEvent === 'result') {
                  settled = true;
                  signal?.removeEventListener('abort', onAbort);
                  resolve(parsed as Diagram);
                } else if (currentEvent === 'error') {
                  settled = true;
                  signal?.removeEventListener('abort', onAbort);
                  reject(new Error(parsed.error ?? 'Stream error'));
                }
              } catch {
                // Skip malformed JSON lines
              }
              currentEvent = '';
            }
          }
        }

        function read(): Promise<void> {
          if (settled) return Promise.resolve();
          return reader!.read().then(({ done, value }) => {
            if (settled) return;
            if (done) {
              // Flush remaining buffer — add trailing newline so processLines parses it
              if (buffer.trim()) {
                buffer += '\n';
                processLines();
              }
              if (!settled) {
                settled = true;
                signal?.removeEventListener('abort', onAbort);
                reject(new Error('Stream ended without result'));
              }
              return;
            }
            buffer += decoder.decode(value, { stream: true });
            processLines();
            return read();
          });
        }

        read().catch((err) => {
          if (!settled) {
            settled = true;
            signal?.removeEventListener('abort', onAbort);
            reject(err);
          }
        });
      })
      .catch((err) => {
        reject(err);
      });
  });
}
