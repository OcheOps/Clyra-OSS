# Clyra MCP server

Clyra exposes an MCP server at `https://your-clyra-instance.example.com/mcp`. It uses the official Go MCP SDK and the Streamable HTTP transport. Protocol versions and JSON-RPC framing are negotiated by the SDK.

## Connect a client

Create a personal access token in **Settings → MCP tokens**. The clear token is shown once and has this form:

```text
clyra_pat_<64 lowercase hex characters>
```

Configure an MCP client with the endpoint and bearer header:

```json
{
  "mcpServers": {
    "clyra": {
      "url": "https://your-clyra-instance.example.com/mcp",
      "headers": {
        "Authorization": "Bearer clyra_pat_…"
      }
    }
  }
}
```

Discovery metadata is available from `GET /.well-known/mcp`. A human-readable tool catalog is available from `GET /mcp/tools`.

Tokens are bound to one user, stored only as SHA-256 hashes, and can be revoked from Settings. A user can have at most 20 active tokens. MCP tools require an active Pro or BYO plan.

## Tools

### `generate_diagram`

Generates a diagram with Clyra-managed AI and saves it to the caller's cloud project library.

```ts
input: {
  prompt: string // 20–2,000 characters
  mode?: 'solution_architecture' | 'network_architecture' | 'enterprise_architecture'
  title?: string // up to 140 characters
}
```

### `reverse_from_iac`

Parses Terraform HCL locally, turns it into an editable diagram, and saves it to the caller's cloud project library. It accepts 20–50,000 characters and an optional title of up to 140 characters.

Both tools return structured content with this shape:

```ts
{
  ok: true
  url: string
  diagram_id: string
  summary: string
  node_count: number
  edge_count: number
  message: string
}
```

The returned `/editor/<diagram_id>` link loads the cloud project directly. Projects created through either tool are tagged `source: "mcp"` and shown with a **Created via MCP** label in the project library.

## Quotas and errors

`generate_diagram` consumes one managed-AI quota unit. The current limit is 20 calls per UTC day per user, stored in Redis with an in-memory fallback when Redis is not configured. BYO accounts use managed inference through MCP today, so they have the same limit. `reverse_from_iac` performs local parsing and does not consume that quota.

Tool-level failures use MCP `CallToolResult` responses with `isError: true`. This lets clients receive plan, validation, quota, and execution errors without treating them as transport failures. Invalid bearer tokens are rejected by HTTP authentication. Request bodies are capped at 1 MiB by the MCP transport; the Terraform field has a tighter 50,000-byte limit.

Every tool invocation records an `mcp.tool_called` audit event with the actor, tool, outcome, duration, plan, and failure category. Cross-origin protection and stateful session expiry are handled by the SDK.

## Current scope

The server currently supports personal access tokens, `generate_diagram`, and Terraform `reverse_from_iac`. OAuth client registration, CloudFormation import, preview-image URLs, cost-estimation tools, and blast-radius tools are future work and are not advertised by the server.
