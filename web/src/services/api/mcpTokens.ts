import { api } from './client';

// MCP Personal Access Tokens — client for /api/v1/settings/mcp/tokens.
// See docs/mcp.md § Auth for the token model and internal/apitoken/*
// for the server contract.

export interface MCPToken {
  id: string;
  name: string;
  source: 'pat' | 'oauth';
  last_used?: string | null;
  created_at: string;
  revoked_at?: string | null;
  /** Cleartext is populated only on the creation response — shown once,
   * never returned again by subsequent list calls. */
  clear?: string;
}

export function listTokens(): Promise<{ tokens: MCPToken[] }> {
  return api<{ tokens: MCPToken[] }>('/settings/mcp/tokens', { method: 'GET' });
}

export function createToken(name: string): Promise<{ token: MCPToken }> {
  return api<{ token: MCPToken }>('/settings/mcp/tokens', {
    method: 'POST',
    body: { name },
  });
}

export function revokeToken(id: string): Promise<void> {
  return api<void>(`/settings/mcp/tokens/${id}`, { method: 'DELETE' });
}
