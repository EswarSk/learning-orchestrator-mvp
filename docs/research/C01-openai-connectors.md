# ChatGPT connection and plugin controls

Source: https://learn.chatgpt.com/docs/enterprise/apps-and-connectors
Retrieved: September 18, 2026 (Pacific time)
Evidence type: primary public documentation or vendor-published description; not independently tested.

For the complete documentation index, see [llms.txt](https://learn.chatgpt.com/llms.txt). Markdown versions of documentation pages are available by appending
`.md` to the page URL.

## Search the docs

Search docs

### Suggested

responses createreasoning\_effortrealtimeprompt caching

Primary navigation

Search docs

### Suggested

responses createreasoning\_effortrealtimeprompt caching

Overview  Features  Configuration  Developers  Security  Administration  Use Cases  Resources

OverviewFeaturesConfigurationDevelopersSecurityAdministrationUse CasesResourcesDocsAdministration

- [Overview](https://learn.chatgpt.com/docs/administration)

### Getting started

- [Admin rollout guide](https://learn.chatgpt.com/docs/enterprise/admin-setup)

### ChatGPT Work

- [ChatGPT Work Overview](https://learn.chatgpt.com/docs/enterprise/chatgpt-work-overview)
- [ChatGPT Work cloud security](https://learn.chatgpt.com/docs/enterprise/chatgpt-work-cloud-security)
- [ChatGPT Work local security](https://learn.chatgpt.com/docs/enterprise/chatgpt-work-local-security)
- [ChatGPT Work admin FAQ](https://learn.chatgpt.com/docs/enterprise/work-admin-faq)
- [ChatGPT Work: usage and cost](https://learn.chatgpt.com/docs/enterprise/chatgpt-work-usage-and-cost)

### Identity and authentication

- [Authentication overview](https://learn.chatgpt.com/docs/auth)
- [Personal Access Tokens](https://learn.chatgpt.com/docs/enterprise/access-tokens)
- [Service accounts](https://learn.chatgpt.com/docs/enterprise/service-accounts)

### Workspace access, policy, and models

- [Groups and provisioning](https://learn.chatgpt.com/docs/enterprise/groups-and-provisioning)
- [User lifecycle management](https://learn.chatgpt.com/docs/enterprise/user-lifecycle)
- [Roles and workspace permissions](https://learn.chatgpt.com/docs/enterprise/roles-and-workspace-permissions)
- [GPTs and Sharing](https://learn.chatgpt.com/docs/enterprise/gpts-and-sharing)
- [Managed configuration](https://learn.chatgpt.com/docs/enterprise/managed-configuration)
- [Prisma AIRS](https://learn.chatgpt.com/docs/enterprise/prisma-airs)
- [HIPAA configuration](https://learn.chatgpt.com/docs/hipaa-configuration)
- [Workspace model availability](https://learn.chatgpt.com/docs/enterprise/workspace-model-availability)

### Plugin and connector controls

- [Plugin controls](https://learn.chatgpt.com/docs/enterprise/apps-and-connectors)
- [Plugin management](https://learn.chatgpt.com/docs/enterprise/plugin-management)
- [Skill controls](https://learn.chatgpt.com/docs/enterprise/skills)

### Usage, governance, and compliance

- [Governance](https://learn.chatgpt.com/docs/enterprise/governance)
- [Admin plugin](https://learn.chatgpt.com/docs/enterprise/admin-plugin)
- [Workspace analytics](https://learn.chatgpt.com/docs/enterprise/workspace-analytics)
- [Usage Insights](https://learn.chatgpt.com/docs/enterprise/usage-insights)
- [Analytics API](https://learn.chatgpt.com/docs/enterprise/analytics-api)
- [Compliance API and audit events](https://learn.chatgpt.com/docs/enterprise/compliance-api)

### Deployment and model providers

- [Manage app updates](https://learn.chatgpt.com/docs/enterprise/manage-app-updates)
- [Windows app deployment](https://learn.chatgpt.com/docs/enterprise/windows-deployment)
- [Remote connections](https://learn.chatgpt.com/docs/remote-connections)
- [Amazon Bedrock](https://learn.chatgpt.com/docs/amazon-bedrock)

[API Dashboard](https://platform.openai.com/login)

[Try ChatGPT](https://chatgpt.com/)

Copy Page

Plugins package reusable workflows and can include skills and MCP servers that connect
to other tools. ChatGPT and Codex use the same public plugin directory on
supported surfaces, while admins decide which plugins are available in their workspace.
Learn more about [plugins](https://learn.chatgpt.com/docs/plugins),
[skills](https://learn.chatgpt.com/docs/skills-and-plugins), and
[connected services](https://help.openai.com/en/articles/11487775).

In this guide, **app** and **MCP server** refer to the same connected
integration and are interchangeable terms. We use **MCP server** in the prose,
but preserve **app** in UI labels such as **Workspace apps** and
**App permissions**, and in CSV column names.

A member can use an MCP server’s capabilities only when the plugin and MCP server are
available to their role and they have access to the connected service.

Plugins work in Chat and Work across ChatGPT on the web, desktop, and mobile,
in Codex in the ChatGPT desktop app, and through the Codex CLI plugin browser.
They aren’t available in the IDE extension.

To see how these controls fit with workspace roles and permissions, see
[Roles and workspace permissions](https://learn.chatgpt.com/docs/enterprise/roles-and-workspace-permissions).

## Understand the capability chain

A plugin can span these control layers:

| Layer | What it determines | Where to manage it |
| --- | --- | --- |
| Availability | Whether the plugin bundle is available to the user | [Workspace settings](https://chatgpt.com/admin/settings) for supported web and desktop surfaces; the CLI plugin browser for CLI |
| Included skills | Which reusable instructions the installed plugin contributes | The plugin package and [Skill controls](https://learn.chatgpt.com/docs/enterprise/skills) |
| MCP server access | Whether users can use an MCP server’s capabilities | [Workspace apps](https://chatgpt.com/admin/ca) and [Permissions & roles](https://chatgpt.com/admin/settings) |
| Actions and permissions | Which actions users can run and when ChatGPT asks before using its tools | The connection’s **Action control** and **App permissions** in [Workspace apps](https://chatgpt.com/admin/ca) |
| Service authorization | Which external data and actions the authenticated identity can access | The connected service and its identity provider |
| Runtime permissions | What an agent can do after it receives data or a tool | The runtime, sandbox, and approval controls for the active surface |

Use these layers as a two-step rollout: first make the right plugins available,
then configure the capabilities and permissions each workflow needs.

## Step 1: Enable plugin availability

For supported web and desktop surfaces, workspace plugin controls determine
which roles can use or install a plugin. The Codex CLI uses its own plugin
browser for installation. See
[Build plugins](https://developers.openai.com/plugins/build/plugins) for
packaging and distribution.

To import workspace plugins from GitHub and keep them up to date, see
[Plugin management](https://learn.chatgpt.com/docs/enterprise/plugin-management).

### Export the public catalog for review

Eligible ChatGPT Enterprise workspace owners and admins can download a CSV of
the public plugins available to their workspace. Use the export to review
plugin, MCP server, and skill metadata before changing plugin availability.

1. Open [Admin > Plugins](https://chatgpt.com/admin/plugins).
2. Select **Public**.
3. Select the download icon ( **Export CSV**) in the page header.

The download uses the filename `public-plugins-security-review.csv` and includes:

- Plugin metadata: `Plugin Name`, `Plugin Description`, `Date Added (UTC)`,
`OpenAI Verified`, `Developer Name`, and `Version`.
- MCP server metadata: `App Name(s)` and `App Description(s)`.
- Chat skill metadata: `Skill Name(s)` and `Skill Description(s)`.

When a plugin includes more than one MCP server or skill, semicolons separate the
corresponding values. The export uses a public-catalog snapshot that can be up
to 48 hours old,
includes only public plugins visible to the current workspace, and does not
include plugins created for that workspace. It isn’t available in FedRAMP
workspaces.

## Step 2: Manage capabilities

Making an MCP server or plugin available in ChatGPT doesn’t grant access to
files, records, or actions in the connected service. Before troubleshooting or
expanding access, check the member’s workspace role and approved action
settings. Then confirm the authenticated account or shared connection has the
expected permissions in the connected service.

Plugins in ChatGPT and Codex can include MCP server connections that search, retrieve, sync,
or act on external systems. Plugin availability and the access and actions
granted to each connection are separate controls.

Manage MCP server capabilities from
[Workspace apps](https://chatgpt.com/admin/ca) and
[Permissions & roles](https://chatgpt.com/admin/settings). Available controls
let admins:

- Enable MCP server connections and assign access by workspace role.
- For connections that support **Action control**, allow read-only actions or an
approved custom set, including how the workspace handles newly added actions.
- Set **App permissions** that determine when ChatGPT asks before using a connection.
- Keep access within the scopes and permissions granted by each connected
service and authenticated user.

For current availability and procedures, see
[Admin controls, security, and compliance in apps](https://help.openai.com/en/articles/11509118).

## Choose a focused initial set

Start with plugins that support a clear business need. Decide whether to make
each plugin available to everyone, limit it to a role or pilot group, or require
further review.

For each connected service, record the business owner, permitted data, approved
read or write actions, authentication method, and a support or removal contact.

Before enabling write actions or publishing a new connected capability, verify
its role scope and test with an account that has only the intended permissions
in the connected service.

For a broad rollout, begin with categories teams use every day, such as email,
calendar, and file or document systems. Use the
[Plugins Directory](https://chatgpt.com/apps) to confirm current availability
and capabilities across supported ChatGPT and Codex surfaces.

Whatever the initial set, start with read actions. Before enabling write
actions, identify the plugin owner, review MCP server scopes and service
permissions, confirm data access, and document external effects and a recovery
path.

## Understand data flow and security

When ChatGPT uses an MCP server included with a plugin, it sends a request
to the connected service and returns data or action results allowed by the
authenticated user’s permissions in that service.

ChatGPT handles data from connected services in two ways:

- **Non-synced:** ChatGPT processes data from Chat and deep research transiently
and doesn’t index it.
- **Synced:** ChatGPT indexes selected connected content in advance. You can see
whether a connection supports sync on its plugin page.

The mode changes how ChatGPT indexes connected content; it doesn’t replace
normal chat-retention controls. ChatGPT conversations that use these connections remain
available through the Compliance API.

OpenAI’s connected-service guidance documents encryption in transit and at rest, per-user
authorization, role and action controls, restricted network access for
conversations that use these connections, and no model training on information accessed
through these connections for Business, Enterprise, and Edu customers. When a request reaches
a connected service, that service’s scopes, retention, data residency, and other
policies also apply.

See [security and compliance for connected services](https://help.openai.com/en/articles/11509118)
and [connections with sync](https://help.openai.com/en/articles/10847137) for current
data-handling details. For locally configured MCP servers in the ChatGPT desktop
app, Codex CLI, or IDE extension, see
[Codex MCP configuration](https://learn.chatgpt.com/docs/extend/mcp).

## Use current procedures and references

- [Admin controls, security, and compliance in apps](https://help.openai.com/en/articles/11509118)
- [Apps in ChatGPT](https://help.openai.com/en/articles/11487775)
- [Apps with sync](https://help.openai.com/en/articles/10847137)
- [Manage workspace settings](https://help.openai.com/en/articles/8411955)
- [Plugins](https://learn.chatgpt.com/docs/plugins)
- [Skills and plugins](https://learn.chatgpt.com/docs/skills-and-plugins)
- [Build plugins](https://developers.openai.com/plugins/build/plugins)
- [Admin rollout guide](https://learn.chatgpt.com/docs/enterprise/admin-setup)

Ask AI

Loading docs agent...

