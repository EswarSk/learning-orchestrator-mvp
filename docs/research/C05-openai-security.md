# OpenAI connector security and privacy

Source: https://developers.openai.com/plugins/guides/security-privacy
Retrieved: September 18, 2026 (Pacific time)
Evidence type: primary public documentation or vendor-published description; not independently tested.

For the complete documentation index, see [llms.txt](https://developers.openai.com/llms.txt). Markdown versions of documentation pages are available by appending
`.md` to the page URL.

## Search the ChatGPT docs

Search docs

### Suggested

toolscomponentsstateauth

Primary navigation

Search docs

### Suggested

toolscomponentsstateauth

Plugins  Workspace Agents  Commerce  Ads

PluginsWorkspace AgentsCommerceAdsDocsPlugins

- [Home](https://developers.openai.com/plugins)
- [Quickstart](https://developers.openai.com/plugins/quickstart)

### Core concepts

- [Plugin architecture](https://developers.openai.com/plugins/concepts/plugins)
- [Skills](https://developers.openai.com/plugins/concepts/skills)
- [MCP server](https://developers.openai.com/plugins/concepts/mcp-server)

### Plan

- [Brainstorm use cases](https://developers.openai.com/plugins/plan/use-case)
- [Define tools](https://developers.openai.com/plugins/plan/tools)

### Build

- [Build an MCP server](https://developers.openai.com/plugins/build/mcp-server)
- [Add UI to your MCP server (optional)](https://developers.openai.com/plugins/build/chatgpt-ui)
- [Authenticate users](https://developers.openai.com/plugins/build/auth)
- [Build skills](https://developers.openai.com/plugins/build/skills)
- [Package your plugin](https://developers.openai.com/plugins/build/plugins)
- [Examples](https://developers.openai.com/plugins/build/examples)

### Test and publish

- [Connect and test your plugin](https://developers.openai.com/plugins/deploy/connect-chatgpt)
- [Submit and publish](https://developers.openai.com/plugins/deploy/submission)
- [Submission error reference](https://developers.openai.com/plugins/deploy/submission-errors)

### Conversion specs

- [Restaurant reservation spec](https://developers.openai.com/plugins/guides/restaurant-reservation-conversion-spec)
- [Get Quote spec](https://developers.openai.com/plugins/guides/local-services-request-quote-conversion-spec)
- [Product checkout spec](https://developers.openai.com/plugins/guides/product-checkout-conversion-spec)

### Guides

- [UI guidelines](https://developers.openai.com/plugins/concepts/ui-guidelines)
- [Optimize Metadata](https://developers.openai.com/plugins/guides/optimize-metadata)
- [Submit a Claude Code plugin](https://developers.openai.com/plugins/guides/submit-claude-plugin)
- [Security & Privacy](https://developers.openai.com/plugins/guides/security-privacy)
- [Troubleshooting](https://developers.openai.com/plugins/deploy/troubleshooting)

### Resources

- [Changelog](https://developers.openai.com/plugins/changelog)
- [Plugin guidelines](https://developers.openai.com/plugins/app-guidelines)
- [MCP server review requirements](https://developers.openai.com/plugins/deploy/app-review)
- [Plugin UI reference](https://developers.openai.com/plugins/reference)
- [Checkout API reference](https://developers.openai.com/plugins/build/monetization)

[API Dashboard](https://platform.openai.com/login)

[Try ChatGPT](https://chatgpt.com/)

Copy Page

## Principles

Plugin tools can access user data, third-party APIs, and write actions. Treat
every MCP server and UI component as production software:

- **Least privilege:** Only request the scopes, storage access, and network permissions you need.
- **Explicit user consent:** Make sure users understand when they are linking
accounts or granting write access. Use the host’s confirmation prompts for
destructive actions.
- **Defense in depth:** Assume prompt injection and malicious inputs will reach your server. Check every input and keep audit logs.

## Data handling

- **Structured content:** Include only the data required for the current prompt. Avoid embedding secrets or tokens in component props.
- **Storage:** Decide how long you keep user data and publish a retention policy. Respect deletion requests.
- **Logging:** Redact PII before writing to logs. Store correlation IDs for debugging but avoid storing raw prompt text unless necessary.

## Prompt injection and write actions

Developer mode enables full MCP access, including write tools. Mitigate risk by:

- Reviewing tool descriptions regularly to discourage misuse (“Do not use to delete records”).
- Validating all inputs server-side even if the model provided them.
- Requiring human confirmation for irreversible operations.

Share your best prompts for testing injections with your QA team so they can probe weak spots early.

## Network access

Widgets run inside an isolated iframe with a strict Content Security Policy.
They cannot access privileged browser APIs such as `window.alert`,
`window.prompt`, `window.confirm`, or `navigator.clipboard`. The CSP controls
standard `fetch` requests. Nested frames are unavailable by default; enable
specific origins in resource CSP metadata such as
`_meta.ui.csp.frameDomains`. Plugins can embed pages from their MCP server’s
own registrable domain, including existing editors and admin interfaces. See
the [iframe policy](https://developers.openai.com/plugins/app-guidelines#iframes-and-embedded-pages) for
domain ownership, required justifications, and review requirements.

The widget CSP restricts which iframe destinations can load. An embedded
page uses its own CSP; the widget `connectDomains` and `resourceDomains`
allowlists do not restrict network requests made inside that page. Keep iframe
origins specific and include the embedded experience in your security review.

Server-side code has no network restrictions beyond what your hosting environment enforces. Follow normal best practices for outbound calls (TLS verification, retries, timeouts).

## Authentication & authorization

- Use OAuth 2.1 authorization-code flows when integrating external accounts.
Prefer Client ID Metadata Documents (CIMD) when your authorization server
supports CIMD and the plugin builder chooses it. Use `none` for public-client
token exchange or `private_key_jwt` when your authorization server requires
client authentication. Support DCR when the plugin builder chooses it or CIMD
is not available.
- Verify and enforce scopes on every tool call. Return a `401` response for expired or malformed tokens.
- For built-in identity, avoid storing long-lived secrets; use the provided auth context instead.

## Operational readiness

- Run security reviews before launch, especially if you handle regulated data.
- Monitor for anomalous traffic patterns and set up alerts for repeated errors or failed auth attempts.
- Keep third-party dependencies, libraries, and build tooling patched to mitigate supply chain risks.

Security and privacy are foundational to user trust. Bake them into your planning, implementation, and deployment workflows rather than treating them as an afterthought.

Ask AI

Loading docs agent...

