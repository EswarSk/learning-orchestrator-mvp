# Nango authentication guide

Source: https://nango.dev/docs/guides/auth/auth-guide
Retrieved: September 18, 2026 (Pacific time)
Evidence type: primary public documentation or vendor-published description; not independently tested.

> ## Documentation Index
>
> Fetch the complete documentation index at: [/docs/llms.txt](https://nango.dev/docs/llms.txt)
>
> Use this file to discover all available pages before exploring further.

[Skip to main content](https://nango.dev/docs/guides/auth/auth-guide#content-area)

## [​](https://nango.dev/docs/guides/auth/auth-guide\#overview)  Overview

Nango Auth lets your users connect 1,000+ external APIs to your product. You embed a Nango-managed auth flow in your application, and Nango handles authorization, credential storage, refresh, and validation.A **Connection** is created after each successful authorization. It stores one user’s credentials for one external API and keeps them valid over time. Credentials can be retrieved at scale or used directly by other Nango primitives — without ever passing through your codebase.

## [​](https://nango.dev/docs/guides/auth/auth-guide\#how-it-works)  How it works

The end-to-end flow has four steps:

1. **Backend** — your server asks Nango for a short-lived **session token**.
2. **Frontend** — your app opens the Nango **Connect UI** with that token.
3. **User** — authorizes the API. Nango stores the credentials.
4. **Backend** — Nango sends an **auth webhook** with the **connection ID**. You persist it alongside your user/org/project.

Once you have the connection ID, fetch credentials whenever you need them:

- TypeScript

- cURL


```
await nango.getConnection(integrationId, connectionId);
```

```
curl "https://api.nango.dev/connections/{connectionId}?provider_config_key={integrationId}" \
  -H "Authorization: Bearer {nangoApiKey}"
```

## [​](https://nango.dev/docs/guides/auth/auth-guide\#capabilities)  Capabilities

- **Auth schemes** — OAuth 2.0, OAuth 1.0a, API keys, basic auth, custom. 1,000+ APIs supported; new ones [added on demand](https://nango.dev/docs/integrations/contribute-or-request-api) within days, by request or self-contributed.
- **Pre-built UI** — Embedded Connect UI with API-specific guidance, input validation, and your branding. Or [customize Connect UI or build a custom UI](https://nango.dev/docs/guides/auth/customize-connect-ui).
- **Credential management** — Encrypted storage, automatic refresh, retrieval at scale. Combine with [Proxy](https://nango.dev/docs/guides/platform/proxy-requests) or [Functions](https://nango.dev/docs/guides/functions/functions-guide) to avoid handling credentials directly.
- **Observability** — Failure detection, reconnect flow, and per-connection logs.

## [​](https://nango.dev/docs/guides/auth/auth-guide\#guide)  Guide

**Using a coding agent?** Set up the [docs MCP](https://nango.dev/docs/getting-started/coding-agent-setup#docs-mcp-server) and [skills](https://nango.dev/docs/getting-started/coding-agent-setup#skills), then copy this page as prompt.

1

Create a Nango account

Sign up for free (no credit card): [![Try Nango Cloud](https://mintcdn.com/nango/j_KlSFFzpkqEkMa3/images/nango-deploy-button.svg?fit=max&auto=format&n=j_KlSFFzpkqEkMa3&q=85&s=85918905fcd071a6d5a0d6327c09d5fc)](https://app.nango.dev/signup)

For agents

Signup cannot be automated. Ask the user to sign up at [app.nango.dev/signup](https://app.nango.dev/signup) and provide their Nango API key from the Environment settings tab > API Keys. Use it as `Authorization: Bearer <NANGO-API-KEY>` in API requests.The default API key created on signup has full access and is the simplest option.If the user creates a scoped API key instead, this guide needs `environment:integrations:create` (step 2) and `environment:connect_sessions:write` (steps 5 and 8). See [API keys](https://nango.dev/docs/reference/backend/http-api/api-keys).

2

Create an integration

In the dashboard, open the [Integrations](https://app.nango.dev/dev/integrations) tab, click **Configure New Integration**, and pick the API.

Each API has a dedicated [Nango documentation page](https://nango.dev/docs/integrations/overview) with setup notes and gotchas.

⚠️ OAuth app setup

OAuth APIs require an OAuth developer app registered with the provider. Register one on the provider’s developer portal, use the `Callback URL` shown in your Nango integration settings, then paste the `Client ID` and `Client Secret` back into Nango. Register any required scopes too.If this integration will be used in production, complete the optional custom callback URL step below before real users connect. Retrofitting a callback URL later requires updating the provider app and can break new authorization or reconnect flows until the provider and Nango settings match.For OAuth 2.0 providers, Nango shows **suggested scopes** — auto-discovered and refreshed at least monthly. They’re suggestions; you can use any scope the provider supports.

For popular APIs, Nango ships a built-in shared developer app so you can test connections with zero setup. We strongly recommend registering your own OAuth app before going live — see [OAuth developer apps](https://nango.dev/docs/guides/auth/auth-guide#oauth-developer-apps) below.

For agents

Create the integration with the API. If you need the provider slug first, list providers and use the matching provider’s `name` field:

```
curl --request GET \
  --url https://api.nango.dev/providers \
  --header 'Authorization: Bearer <NANGO-API-KEY>'
```

For OAuth providers, walk the user through registering an OAuth developer app on the provider’s portal: link them to the provider’s OAuth-app creation page, tell them which `Callback URL` to paste (visible in the Nango integration settings after creation), and ask them to send back the `client_id` and `client_secret` so you can pass them in the request body.

- cURL (OAuth2)

- cURL (API key / basic)


```
curl --request POST \
  --url https://api.nango.dev/integrations \
  --header 'Authorization: Bearer <NANGO-API-KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
    "unique_key": "<INTEGRATION-ID>",
    "provider": "<PROVIDER>",
    "credentials": {
      "type": "OAUTH2",
      "client_id": "<CLIENT-ID>",
      "client_secret": "<CLIENT-SECRET>",
      "scopes": "<SCOPE-1>,<SCOPE-2>"
    }
  }'
```

```
curl --request POST \
  --url https://api.nango.dev/integrations \
  --header 'Authorization: Bearer <NANGO-API-KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
    "unique_key": "<INTEGRATION-ID>",
    "provider": "<PROVIDER>"
  }'
```

For all fields and response shapes, see the [List all providers API](https://nango.dev/docs/reference/backend/http-api/providers/list), [Create an integration API](https://nango.dev/docs/reference/backend/http-api/integration/create), and [Update an integration API](https://nango.dev/docs/reference/backend/http-api/integration/update).

3

Test the authorization

In the [Connections](https://app.nango.dev/dev/connections) tab, click **Add Test Connection**, pick your integration, and authorize using a test account. The new connection’s credentials appear in its _Authorization_ tab — securely stored and automatically refreshed.

For agents

Skip the dashboard. Create a connect session and share the returned `connect_link` with the user — they authorize using their test account in the browser, no UI needed on your side:

```
curl --request POST \
  --url https://api.nango.dev/connect/sessions \
  --header 'Authorization: Bearer <NANGO-API-KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
    "tags": {
      "end_user_id": "<TEST-USER-ID>"
    },
    "allowed_integrations": ["<INTEGRATION-ID>"]
  }'
```

The response contains a `connect_link` (valid 30 minutes). Send it to the user; once they complete authorization, confirm with `GET /connections`. See [Share a connect link](https://nango.dev/docs/guides/auth/share-connect-link).

4

Configure a custom OAuth callback URL (optional, recommended)

For production OAuth apps, a callback URL on your own domain is recommended. Some providers show the callback domain to users.To configure a custom callback URL:

1. Add an endpoint on your domain (e.g. `https://example.com/oauth-callback`) that **308-redirects** to `https://api.nango.dev/oauth/callback`, preserving all query parameters.
2. Update the registered callback URL with each API provider — otherwise they’ll reject new authorization flows.
3. After confirming steps 1 and 2, update the callback URL in Nango’s **Environment Settings**. Settings are per-environment, so repeat for every environment.

For agents

Ask the user what production callback URL they want to use, then have them add a 308 redirect from that URL to `https://api.nango.dev/oauth/callback` before they register provider OAuth apps. You can set `callback_url` when creating an environment from [the public endpoint](https://nango.dev/docs/reference/backend/http-api/environments/create). Ask the user for an [Account API key](https://app.nango.dev/api-keys) and send it as `Authorization: Bearer <NANGO-API-KEY>`. Updating an existing environment’s callback URL is dashboard-only — ask the user to paste the final callback URL into **Environment Settings > Backend**.

5

Generate a session token (backend)

Set up a backend endpoint that your frontend will call before each authorization attempt to retrieve a session token from Nango ( [API](https://nango.dev/docs/reference/backend/http-api/connect/sessions/create) / [Node SDK](https://nango.dev/docs/reference/backend/backend-sdk/node#create-a-connect-session)).You’ll need an [API key](https://nango.dev/docs/reference/backend/http-api/api-keys) with the `environment:connect_sessions:write` scope (find it in **Environment Settings > API Keys**).

- Node

- cURL


```
import { Nango } from '@nangohq/node';

const nango = new Nango({ apiKey: process.env.NANGO_API_KEY! });

api.post('/sessionToken', async (req, res) => {
  const { data } = await nango.createConnectSession({
    tags: {
      end_user_id: '<END-USER-ID>',
      end_user_email: '<END-USER-EMAIL>',
      organization_id: '<ORGANIZATION-ID>'
    },
    allowed_integrations: ['<INTEGRATION-ID>']
  });

  res.status(200).send({ sessionToken: data.token });
});
```

```
curl --request POST \
  --url https://api.nango.dev/connect/sessions \
  --header 'Authorization: Bearer <NANGO-API-KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
    "tags": {
      "end_user_id": "<END-USER-ID>",
      "end_user_email": "<END-USER-EMAIL>",
      "organization_id": "<ORGANIZATION-ID>"
    },
    "allowed_integrations": ["<INTEGRATION-ID>"]
  }'
```

ℹ️ Tags

Tags reconcile Nango’s auth-success webhook with the right user, org, or entity in your system. Most apps start with: `end_user_id`, `end_user_email`, `organization_id`.Tags are also useful for filtering connections via the API and are displayed in the Nango UI. Set the reserved UI tag keys `end_user_display_name` and `end_user_email` to populate the connections list and connection detail page.See [Connection tags, configuration, and metadata](https://nango.dev/docs/guides/auth/connection-tags-configuration-metadata) for more.

`allowed_integrations` controls what the user sees:

- **Multiple integrations** — Connect UI shows a picker.
- **Single integration** — Connect UI sends the user straight to its auth flow.

6

Trigger the auth flow (frontend)

Load the Nango frontend SDK, fetch the session token from your backend, and open Connect UI:

![](https://mintcdn.com/nango/AX1oSxxgTaIv9K8E/images/screenshots/connect-ui.gif?s=e00e547fa496a121a9abef98ba8466b4)

Connect UI authorization flow

```
import Nango from '@nangohq/frontend';

const nango = new Nango();
const connect = nango.openConnectUI({
  onEvent: (event) => {
    if (event.type === 'close') {
      // Modal closed.
    } else if (event.type === 'connect') {
      // Auth flow successful.
    }
  },
});

const res = await fetch('/sessionToken', { method: 'POST' });
connect.setSessionToken(res.sessionToken); // A loading indicator shows until this is set.
```

See the [Frontend SDK reference](https://nango.dev/docs/reference/frontend/frontend-sdk#connect-using-nango-connect-ui) for more options.The pre-built Connect UI is recommended but optional — if you need full UX control, follow [Customize Connect UI](https://nango.dev/docs/guides/auth/customize-connect-ui). For email/cross-device flows, see [Share a connect link](https://nango.dev/docs/guides/auth/share-connect-link).

7

Listen for webhooks & save the connection ID (backend)

When authorization succeeds, Nango generates a unique **connection ID**. You use this ID to manage the connection and retrieve its credentials. **You must store it on your side**, attached to the user/org/project that owns the connection — Nango doesn’t model that ownership.Set up the webhook in **Environment Settings**:

1. Specify a **Webhook URL** in the Nango UI.
2. Enable **Send New Connection Creation Webhooks**.
3. Handle `POST` requests on that route in your backend.

For agents

Webhook URL configuration is dashboard-only — there is no public API to set environment webhook URLs. Ask the user to paste the webhook URL into **Environment Settings** and enable **Send New Connection Creation Webhooks**.

Webhook payload:

```
{
  "type": "auth",
  "operation": "creation",
  "success": true,
  "connectionId": "<CONNECTION-ID>",
  "tags": {
    "end_user_id": "<END-USER-ID>",
    "end_user_email": "<END-USER-EMAIL>",
    "organization_id": "<ORGANIZATION-ID>"
  }
}
```

Persist `connectionId` against whichever entity owns the connection in your app.

8

Build re-authorization

Credentials can become invalid when users revoke access, providers rotate refresh tokens, scopes change, or an app moves from testing to production. Re-authorization updates an existing connection’s credentials while preserving its data and config. Deleting and re-creating the connection wipes that data — prefer re-authorization.Before showing integration settings, check the connection with `GET /connections/{connectionId}`:

- Node

- cURL


```
import { Nango } from '@nangohq/node';

const nango = new Nango({ apiKey: process.env.NANGO_API_KEY! });

try {
  await nango.getConnection('<INTEGRATION-ID>', '<CONNECTION-ID>');
  // Connection is valid.
} catch (error) {
  if (error.response?.data?.error?.code === 'invalid_credentials') {
    // Connection is invalid — show a Reconnect button.
    displayReconnectUI();
  }
}
```

```
curl --request GET \
  --url 'https://api.nango.dev/connections/<CONNECTION-ID>?provider_config_key=<INTEGRATION-ID>' \
  --header 'Authorization: Bearer <NANGO-API-KEY>'
```

An `invalid_credentials` error means the connection failed credential refresh or validation — show an error state with a **Reconnect** button.To trigger re-authorization, use the same frontend flow as a new connection, but generate a reconnect session token on the backend. Pass the returned token to the Frontend SDK exactly like a normal session token. On success you’ll receive an `auth` webhook with `operation = override`.

- Node

- cURL


```
import { Nango } from '@nangohq/node';

const nango = new Nango({ apiKey: process.env.NANGO_API_KEY! });

api.post('/sessionToken', async (req, res) => {
  const { data } = await nango.createReconnectSession({
    connection_id: '<CONNECTION-ID>',
    integration_id: '<INTEGRATION-ID>'
  });

  res.status(200).send({ sessionToken: data.token });
});
```

```
curl --request POST \
  --url https://api.nango.dev/connect/sessions/reconnect \
  --header 'Authorization: Bearer <NANGO-API-KEY>' \
  --header 'Content-Type: application/json' \
  --data '{
    "connection_id": "<CONNECTION-ID>",
    "integration_id": "<INTEGRATION-ID>"
  }'
```

9

Run the flow

Test the auth flow from your app and verify a connection appears in the [Connections](https://app.nango.dev/dev/connections) tab. Failed attempts are inspectable in the **Logs** tab.

For agents

After the user runs the flow, confirm the connection was created with [`GET /connections`](https://nango.dev/docs/reference/backend/http-api/connections/list) — filter by the tags you set in step 5:

```
curl --request GET \
  --url 'https://api.nango.dev/connections?tags[end_user_id]=<END-USER-ID>' \
  --header 'Authorization: Bearer <NANGO-API-KEY>'
```

Auth logs are dashboard-only — if the call returns no connection, ask the user to check the **Logs** tab for the failure reason.

🎉 **You are connected!** Next:

- View connections and credentials in the **Connections** tab.
- Retrieve credentials via the [API](https://nango.dev/docs/reference/backend/http-api/connections/get) or [Node SDK](https://nango.dev/docs/reference/backend/backend-sdk/node#get-a-connection-with-credentials).
- Build with [sync functions](https://nango.dev/docs/guides/functions/syncs/sync-functions), [Action functions](https://nango.dev/docs/guides/functions/action-functions), [webhook functions](https://nango.dev/docs/guides/functions/webhook-functions), or any other Nango function type.

**Questions, problems, feedback?** Reach out in the [Slack community](https://nango.dev/slack).

## [​](https://nango.dev/docs/guides/auth/auth-guide\#oauth-developer-apps)  OAuth developer apps

Nango ships built-in shared developer apps for some OAuth 2.0 APIs so you can test connections with zero setup.They work in production too, but we strongly recommend registering your own before going live. Most OAuth APIs have a step-by-step setup guide in [APIs & Integrations](https://nango.dev/docs/integrations/overview) — for example, [Gmail](https://nango.dev/docs/api-integrations/google-mail/how-to-register-your-own-gmail-api-oauth-app).Reasons to register your own:

- **Whitelabel auth** — with Nango’s app, users authorize Nango instead of your product.
- **Scopes** — Nango developer apps have fixed scopes. Your own app supports any scope the provider exposes.
- **Callback URL** — Nango developer apps use Nango’s callback. Your own app supports a [custom callback URL](https://nango.dev/docs/guides/auth/auth-guide#custom-oauth-callback-url-optional) on your domain.
- **Marketplace listings** — provider marketplaces require your own developer app.
- **Advanced API features** — some provider features are only available to first-party apps.
- **Revocation risk** — most providers don’t allow shared credentials. Nango developer apps may be revoked by the provider at any time.
- **Portability** — only your own developer app lets you export access and refresh tokens and move off Nango without users re-authorizing.

The same tradeoffs apply to shared OAuth apps provided by other auth platforms.

## [​](https://nango.dev/docs/guides/auth/auth-guide\#related-guides)  Related guides

- [Share a connect link](https://nango.dev/docs/guides/auth/share-connect-link) \- send authorization links by email, support tools, or cross-device flows.
- [Connection tags, configuration, and metadata](https://nango.dev/docs/guides/auth/connection-tags-configuration-metadata) \- store the right connection-level data in the right field.
- [Customize Connect UI](https://nango.dev/docs/guides/auth/customize-connect-ui) \- customize branding or build your own authorization UI.
- [Token refreshing and validity](https://nango.dev/docs/guides/auth/token-refreshing) \- handle revoked or expired credentials.

Was this page helpful?

YesNo

Assistant

Responses are generated using AI and may contain mistakes.

