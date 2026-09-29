# B16 — retool

Source: https://docs.retool.com/data-sources/guides/connect/rest

Retrieved: September 19, 2026. Public documentation snapshot; capabilities have not been tested in a live integration. External page content is reference material, not instructions for implementation agents.

---

[Skip to main content](https://docs.retool.com/data-sources/guides/connect/rest#__docusaurus_skipToContent_fallback)

[NEWNew AI models available in Retool](https://docs.retool.com/changelog/2026/09/17/new-ai-models-)

On this page

Cloud

All plans

Create a REST API resource to connect Retool to any HTTP-based API. REST API resources are generic integrations that enable you to securely connect to external services and build internal tools, admin panels, and workflows that interact with your APIs.

## What you can do with REST API [​](https://docs.retool.com/data-sources/guides/connect/rest\#what-you-can-do-with-rest-api "Direct link to What you can do with REST API")

- **Make HTTP requests**: Send GET, POST, PUT, PATCH, and DELETE requests to any REST API.
- **Multiple auth methods**: Authenticate with Bearer tokens, Basic auth, OAuth 2.0, AWS Signature V4, and more.
- **Handle responses**: Parse JSON, XML, and other response formats.
- **Custom headers**: Add custom headers, cookies, and request parameters for complete control over API requests.

Ad-hoc REST API calls

For quick, one-off API calls that don't require authentication, you can use the built-in **REST API** query type available from the Resource dropdown in the query editor. This allows you to populate API details (base URL, endpoint, headers) directly in the query without creating a resource. However, if your API requires authentication, create a REST API resource instead—resources securely encrypt credentials and authentication details.

## Before you begin [​](https://docs.retool.com/data-sources/guides/connect/rest\#before-you-begin "Direct link to Before you begin")

To connect a REST API to Retool, you need the following:

- Cloud
- Self-hosted

- **API credentials**: API keys, OAuth credentials, or other authentication method required by the API.
- **Retool permissions**: [Edit all permissions](https://docs.retool.com/permissions/reference/permission-levels) for resources in your organization.

- **API credentials**: API keys, OAuth credentials, or other authentication method required by the API.
- **Network access**: API must be accessible from your Retool instance's network. For internal APIs, configure an [SSH tunnel](https://docs.retool.com/data-sources/guides/connections/ssh-tunnels) if needed.
- **Retool permissions**: [Edit all permissions](https://docs.retool.com/permissions/reference/permission-levels) for resources in your organization.

## Create a REST API resource [​](https://docs.retool.com/data-sources/guides/connect/rest\#create-a-rest-api-resource "Direct link to Create a REST API resource")

Follow these steps to create a REST API resource in Retool.

### 1\.  Create a new resource [​](https://docs.retool.com/data-sources/guides/connect/rest\#1-create-a-new-resource "Direct link to 1. Create a new resource")

In your Retool organization, navigate to [Resources](https://docs.retool.com/data-sources/tutorials/create-resource) in the main navigation and click **Create new** → **Resource**. Search for `REST API` and click the REST API tile to begin configuration.

![](<Base64-Image-Removed>)
REST API resource selection.

### 2\.  Configure connection settings [​](https://docs.retool.com/data-sources/guides/connect/rest\#2-configure-connection-settings "Direct link to 2. Configure connection settings")

Configure the following connection settings for your REST API resource.

![](https://docs.retool.com/assets/images/config-connection-top-c0a5c17009a830e49d5b4c13ece84dfe.png)
REST API connection settings.

#### Resource name and description [​](https://docs.retool.com/data-sources/guides/connect/rest\#resource-name-and-description "Direct link to Resource name and description")

Specify a name for the resource that indicates which API it connects to. Include a description that can provide more context to users and [Assist](https://docs.retool.com/apps/guides/assist/) about how to use the resource.

| Example Name | Example Description |
| --- | --- |
| OpenWeatherMap API | A REST API resource for the OpenWeatherMap weather data service. |
| Internal customer API | A REST API resource for our internal customer management API. |

#### Query mode [​](https://docs.retool.com/data-sources/guides/connect/rest\#query-mode "Direct link to Query mode")

Choose between two query modes for your REST API resource:

- Use an API spec
- Manual queries

AvailabilityAll plansPublic beta

**Spec URL**

Import an OpenAPI or Swagger specification to enable guided query creation with endpoint autocompletion and improved [Assist](https://docs.retool.com/apps/guides/assist/) integration.

Examples

```text
https://api.example.com/openapi.json

https://api.example.com/swagger.yaml

https://petstore.swagger.io/v2/swagger.json
```

After entering the URL, click **Load spec** to fetch and validate the specification. Retool will:

- Parse the specification and extract available endpoints.
- Display base URLs from the spec for selection, or auto-populate the base URL if there is only one available.
- Enable guided query creation with endpoint selection and parameter autocompletion.

Best practice

- **Autocomplete**: Get endpoint and parameter suggestions when writing queries.
- **Validation**: Retool validates requests against the specification schema.
- **Better Assist**: [Assist](https://docs.retool.com/apps/guides/assist/) can more effectively generate queries based on the specification.
- **Documentation**: Endpoint descriptions from the specification appear in the query editor.

When working with resources, users can also manually create endpoints or parameters that aren't included in the specification.

**Server URL**

If your spec defines multiple server URLs, select the appropriate server for your environment. You can also enter a custom URL if needed.

**Base URL**

Manually configure the base URL without importing a specification. Use this mode if your API doesn't provide an OpenAPI/Swagger spec or you prefer manual configuration. Query paths are appended to this base URL.

Examples

```text
https://api.openweathermap.org/data/2.5

https://api.example.com/v2

https://internal-api.company.com/api
```

#### URL parameters [​](https://docs.retool.com/data-sources/guides/connect/rest\#url-parameters "Direct link to URL parameters")

Global URL parameters applied to all requests. Add parameters as key-value pairs.

#### Headers [​](https://docs.retool.com/data-sources/guides/connect/rest\#headers "Direct link to Headers")

Global headers applied to all requests. Add headers as key-value pairs.

### 3\.  Configure authentication [​](https://docs.retool.com/data-sources/guides/connect/rest\#3-configure-authentication "Direct link to 3. Configure authentication")

REST API supports multiple [authentication methods](https://docs.retool.com/data-sources/concepts/authentication) based on your API requirements.

| Authentication method | Use cases |
| --- | --- |
| **Auth0 Client Credentials** | Auth0 for identity and access management. Retool handles OAuth flow and token management automatically. |
| **AWS Signature V4** | AWS-hosted APIs requiring signed requests (API Gateway). Retool signs each request using AWS Signature Version 4. |
| **Azure Identity** | Azure-hosted APIs that use Azure Identity credentials. |
| **Basic authentication** | Username and password credentials sent with each request. Retool encodes credentials as base64 in the `Authorization` header. Common for simple APIs. |
| **Bearer token** | Static API token authentication. Retool sends the token as `Authorization: Bearer {token}`. |
| **Custom** | Custom authentication logic not covered by standard methods, such as multi-step flows. Add custom headers and authentication workflow using JavaScript. Test connection disabled for custom auth. |
| **Digest authentication** | Digest authentication (RFC 2617). Uses MD5 hashing for improved security over Basic auth. |
| **Google Service Account** | Google service account authentication for Google Cloud services. Upload JSON key file for automatic JWT generation and token refresh. |
| **None** | Public endpoints or custom authentication via headers. Add authentication tokens as custom headers. |
| **OAuth 1.0** | Legacy OAuth 1.0 authentication (older services). Retool signs requests using OAuth 1.0a signature method. |
| **OAuth 2.0** | User authentication via OAuth 2.0. Use Authorization Code flow for user credentials or Client Credentials for server-to-server authentication with automatic token refresh. |
| **Session-based** Deprecated | Session cookies for authentication. Configure login endpoint and Retool maintains session cookies across requests. |

Select an authentication method from the **Authentication** dropdown and provide the required credentials. For sensitive values like API keys, tokens, and secrets, use [configuration variables](https://docs.retool.com/org-users/guides/configuration/config-vars) or Retool secrets rather than hardcoding them.

![](https://docs.retool.com/assets/images/config-authentication-02f230ff5229570f372e3fa958007b64.webp)
REST API authentication settings.

### 4\.  Test the connection [​](https://docs.retool.com/data-sources/guides/connect/rest\#4-test-the-connection "Direct link to 4. Test the connection")

Click **Test connection** to verify Retool can connect to your REST API. A successful test confirms the base URL is accessible and authentication credentials are valid.

If the connection test fails, verify:

- Base URL is correct and accessible from Retool
- Authentication credentials are valid and not expired
- API accepts requests from Retool's IP addresses (for Cloud organizations)
- Custom headers are formatted correctly

### 5\.  Save the resource [​](https://docs.retool.com/data-sources/guides/connect/rest\#5-save-the-resource "Direct link to 5. Save the resource")

Click **Create resource** to save your REST API resource. The resource is now available to use in apps, workflows, and agent tools across your Retool organization.

## Query REST API data [​](https://docs.retool.com/data-sources/guides/connect/rest\#query-rest-api-data "Direct link to Query REST API data")

Once you've created a REST API resource, you can query it in apps, workflows, and agent tools.

### Create a query [​](https://docs.retool.com/data-sources/guides/connect/rest\#create-a-query "Direct link to Create a query")

You can create a REST API query in a Retool app using [Assist](https://docs.retool.com/apps/guides/assist/) to generate queries with natural language, or manually using code.

- Assist
- Code

Use [Assist](https://docs.retool.com/apps/guides/assist/) to generate queries from natural language prompts. Assist can create REST API queries to retrieve, send, and manipulate data from your REST API resource.

To create a query with Assist:

1. In the Retool app [IDE](https://docs.retool.com/apps/concepts/ide), click the **Assist** button at the bottom of the left toolbar to open the Assist panel (if not already visible).
2. Write a prompt describing the API request you want to make, referencing your resource using `@`.
3. Press **Enter** to submit the prompt.
4. Select your REST API resource when prompted.
5. Review the generated query and click **Run query** to add it to your app.

Example prompt

```text
get current weather data from the weather endpoint using @OpenWeatherMap API
```

![](https://docs.retool.com/assets/images/assist-query-preview-232df47596ec4304b0f2cd1b902d72de.webp)
REST API query with Assist.

To manually create a REST API query in a Retool app:

1. In the Retool app [IDE](https://docs.retool.com/apps/concepts/ide), open the **Code** tab, then click **+** in the page or global scope.
2. Select **Resource query**.
3. Choose your REST API resource.
4. Configure the HTTP method, URL path, headers, and request body.

### Query configuration fields [​](https://docs.retool.com/data-sources/guides/connect/rest\#query-configuration-fields "Direct link to Query configuration fields")

Each REST API query has the following configuration fields:

- [HTTP method](https://docs.retool.com/data-sources/guides/connect/rest#http-method)
- [URL path](https://docs.retool.com/data-sources/guides/connect/rest#url-path)
- [Headers](https://docs.retool.com/data-sources/guides/connect/rest#headers)
- [Body](https://docs.retool.com/data-sources/guides/connect/rest#body)
- [Body type](https://docs.retool.com/data-sources/guides/connect/rest#body-type)

#### HTTP method [​](https://docs.retool.com/data-sources/guides/connect/rest\#http-method "Direct link to HTTP method")

The HTTP method for the request.

| Method | Use case |
| --- | --- |
| **GET** | Retrieve data from the API. Most common read operation. |
| **POST** | Create new objects or submit data to the API. |
| **PUT** | Replace an existing object completely. |
| **PATCH** | Partially update an existing object. |
| **DELETE** | Remove an object from the API. |

#### URL path [​](https://docs.retool.com/data-sources/guides/connect/rest\#url-path "Direct link to URL path")

The path appended to the resource's base URL. Include dynamic values using embedded expressions.

URL path examples

```text
// Static path

/users

// Dynamic with ID

/users/{{ table1.selectedRow.data.id }}

// With query parameters

/search?q={{ searchInput.value }}&limit=20

// Multiple path segments

/customers/{{ customerId.value }}/orders
```

#### URL parameters [​](https://docs.retool.com/data-sources/guides/connect/rest\#url-parameters-1 "Direct link to URL parameters")

Query parameters appended to the URL. Add parameters as key-value pairs to build the query string.

#### Headers [​](https://docs.retool.com/data-sources/guides/connect/rest\#headers-1 "Direct link to Headers")

Request-specific headers that override or extend the resource's global headers.

#### Body [​](https://docs.retool.com/data-sources/guides/connect/rest\#body "Direct link to Body")

The request body for POST, PUT, and PATCH requests. Use embedded expressions to include dynamic data.

Body example

```javascript
{

  "name": {{ nameInput.value }},

  "email": {{ emailInput.value }},

  "role": {{ roleSelect.value }},

  "metadata": {

    "created_by": {{ current_user.email }},

    "created_at": {{ new Date().toISOString() }}

  }

}
```

#### Body type [​](https://docs.retool.com/data-sources/guides/connect/rest\#body-type "Direct link to Body type")

The format of the request body.

| Body type | Description | Content-Type header |
| --- | --- | --- |
| **JSON** | JavaScript object serialized to JSON. | `application/json` |
| **Raw** | Plain text or custom format. | Custom or none |
| **None** | No request body. | None |

### OpenAPI deepObject body encoding [​](https://docs.retool.com/data-sources/guides/connect/rest\#openapi-deepobject-body-encoding "Direct link to OpenAPI deepObject body encoding")

REST API resources imported from an OpenAPI 3.0 specification can include request body fields that use the `deepObject` encoding style. When an OpenAPI operation declares `encoding: { style: "deepObject" }` on an `application/x-www-form-urlencoded` request body, Retool serializes nested object properties as bracket-notation keys at the root of the body. This matches the [OpenAPI 3.0 `deepObject` style](https://spec.openapis.org/oas/v3.0.3#style-examples).

For example, a body field at `shipping.address.line1` is serialized as:

```text
shipping[address][line1]=...
```

The `deepObject` encoding only applies to application/x-www-form-urlencoded request bodies on REST resources imported from OpenAPI specs. It does not change the behavior of query parameters or JSON request bodies, and it does not affect REST resources created manually.

## Connect to a SOAP API [​](https://docs.retool.com/data-sources/guides/connect/rest\#connect-to-a-soap-api "Direct link to Connect to a SOAP API")

You can also use a REST API resource to connect to a SOAP API. To perform SOAP requests:

- Use the **POST** HTTP method.
- Use the **Raw** [body type](https://docs.retool.com/data-sources/guides/connect/rest#body-type).
- If you use WSDL, provide the path to the WSDL in the base URL or URL path.

Set the correct headers for SOAP requests

SOAP requests must also include the following header.

| Key | Value |
| --- | --- |
| `Content-Type` | `text/xml` |

If you use a SOAP 1.1 service, also include a `SOAPAction` header with an appropriate value. You may need additional headers depending on your SOAP API's requirements; refer to your SOAP API specification for details. The name of the SOAP method is usually provided in the SOAP body but may need to be provided in a header instead.

## Common use cases [​](https://docs.retool.com/data-sources/guides/connect/rest\#common-use-cases "Direct link to Common use cases")

The following examples demonstrate typical REST API operations in Retool apps using a sample REST API.

Use the mock API generator below to generate a fully functional temporary REST API with sample data that you can use.

1. Select the sample data set.
2. Specify the number of **Items** that the data set should include.
3. Click **Generate API** to generate the API with sample data.
4. Click **Copy URL** to copy the API URL to your clipboard.

[Mock API generator](https://docsdemos.retool.com/p/mock-api-generator?data=customers)

Caution

The mock API is for testing purposes only and does not require authentication. Do not attempt to use it for production purposes or with any type of sensitive information.

fetch data from an API

First, create a REST API resource with the base URL from the mock API generator, for example: `https://retoolapi.dev/SPq6yI`.

Next, create a query to fetch customer data:

| Field | Value |
| --- | --- |
| **HTTP method** | GET |
| **URL path** | `/customers` |

Then, add a [Table](https://docs.retool.com/apps/guides/data/table/) component to the app and set its **Data** property to `{{ getCustomersQuery.data }}`.

filter data with query parameters

Query parameters allow you to filter and refine API responses.

Create a query with query parameters:

| Field | Value |
| --- | --- |
| **HTTP method** | GET |
| **URL path** | `/customers` |
| **Query parameters** | `[["company", {{ companySelect.value }}], ["_limit", "50"]]` |

The query sends a request to `/customers?company=Acme&_limit=50` and returns filtered results.

fetch a single record by ID

Retrieve a specific customer record using their ID.

Create a query with a dynamic ID in the path:

| Field | Value |
| --- | --- |
| **HTTP method** | GET |
| **URL path** | `/customers/{{ table1.selectedRow.data.id }}` |

The query fetches the selected customer's details and you can display them in a form or detail view.

create a new record via POST

First, add a [Form](https://docs.retool.com/apps/reference/components/form) component (`form1`) with input fields for the customer data you want to collect.

Next, create a POST query:

| Field | Value |
| --- | --- |
| **HTTP method** | POST |
| **URL path** | `/customers` |
| **Body type** | JSON |

**Body:**

```javascript
{

  "name": {{ form1.data.name }},

  "email": {{ form1.data.email }},

  "company": {{ form1.data.company }}

}
```

Then, add an event handler to the form's **Submit** event that runs the POST query and displays a success notification.

update a record via PATCH

First, add a [Form](https://docs.retool.com/apps/reference/components/form) component with fields pre-populated from selected row data.

Next, create a PATCH query:

| Field | Value |
| --- | --- |
| **HTTP method** | PATCH |
| **URL path** | `/customers/{{ table1.selectedRow.data.id }}` |
| **Body type** | JSON |

**Body:**

```javascript
{

  "name": {{ form1.data.name }},

  "email": {{ form1.data.email }},

  "company": {{ form1.data.company }}

}
```

Then, add an event handler that runs the PATCH query when the form is submitted and refreshes the data query.

delete a record with confirmation

First, add a [Button](https://docs.retool.com/apps/reference/components/button) component in your table's action column.

Next, create a DELETE query:

| Field | Value |
| --- | --- |
| **HTTP method** | DELETE |
| **URL path** | `/customers/{{ table1.selectedRow.data.id }}` |

Then, add an event handler to the button's **Click** event:

1. Action: **Show confirmation modal**
2. If confirmed, trigger the DELETE query
3. Then refresh the data query

handle paginated responses

For APIs that return paginated data, use query parameters to implement pagination.

Create a query with pagination parameters:

| Field | Value |
| --- | --- |
| **HTTP method** | GET |
| **URL path** | `/customers` |
| **Query parameters** | `[["_page", {{ table1.pagination.pageIndex + 1 }}], ["_limit", "10"]]` |

The query sends a request to `/customers?_page=2&_limit=10` and returns paginated results. Connect the query to a table component and enable server-side pagination.

## Best practices [​](https://docs.retool.com/data-sources/guides/connect/rest\#best-practices "Direct link to Best practices")

Follow these best practices to optimize performance, maintain security, and ensure data integrity when working with REST APIs.

### Performance [​](https://docs.retool.com/data-sources/guides/connect/rest\#performance "Direct link to Performance")

- **Cache responses**: For data that doesn't change frequently, enable [query caching](https://docs.retool.com/queries/concepts/caching) to reduce API calls and improve response times.
- **Implement pagination**: For endpoints that return large datasets, use query parameters to implement pagination and reduce payload size.
- **Batch requests**: When available, use batch API endpoints to combine multiple operations into a single request.
- **Minimize payload size**: Request only the fields you need using query parameters or API-specific field selection features.
- **Set appropriate timeouts**: Configure query timeouts based on expected API response times to prevent hung requests.

### Security [​](https://docs.retool.com/data-sources/guides/connect/rest\#security "Direct link to Security")

- **Use configuration variables**: Store API keys and tokens in [configuration variables](https://docs.retool.com/org-users/guides/configuration/config-vars) or Retool secrets rather than hardcoding them.
- **Use HTTPS only**: Always connect to APIs over HTTPS to encrypt data in transit and protect authentication credentials.
- **Rotate credentials regularly**: Follow your API provider's recommendations for credential rotation and key management.
- **Validate SSL certificates**: Keep [SSL certificate](https://docs.retool.com/data-sources/guides/connections/ssl) verification enabled unless absolutely necessary for development environments.
- **Use resource environments**: Organizations on an Enterprise plan can configure multiple [resource environments](https://docs.retool.com/org-users/guides/configuration/environments) to maintain separate configurations for production, staging, and development.
- **Apply least privilege**: Use API keys with minimal required permissions. Create separate keys for different environments.

### Data integrity [​](https://docs.retool.com/data-sources/guides/connect/rest\#data-integrity "Direct link to Data integrity")

- **Validate user input**: Sanitize and validate all user input before including it in API requests to prevent injection attacks.
- **Handle errors gracefully**: Configure error notifications and fallback behavior for failed API calls to improve user experience.
- **Use idempotency keys**: For APIs that support idempotency keys, include them in POST/PATCH requests to prevent duplicate operations.
- **Verify responses**: Check response status codes and validate response data structure before using it in your app.
- **Implement retry logic**: For transient failures, use Retool's automatic retry settings or implement custom retry logic with exponential backoff.
- **Log API interactions**: Enable query logging to track API requests and responses for debugging and auditing.

- [What you can do with REST API](https://docs.retool.com/data-sources/guides/connect/rest#what-you-can-do-with-rest-api "What you can do with REST API")
- [Before you begin](https://docs.retool.com/data-sources/guides/connect/rest#before-you-begin "Before you begin")
- [Create a REST API resource](https://docs.retool.com/data-sources/guides/connect/rest#create-a-rest-api-resource "Create a REST API resource")
  - [1\. Create a new resource](https://docs.retool.com/data-sources/guides/connect/rest#1-create-a-new-resource "1. Create a new resource")
  - [2\. Configure connection settings](https://docs.retool.com/data-sources/guides/connect/rest#2-configure-connection-settings "2. Configure connection settings")
  - [3\. Configure authentication](https://docs.retool.com/data-sources/guides/connect/rest#3-configure-authentication "3. Configure authentication")
  - [4\. Test the connection](https://docs.retool.com/data-sources/guides/connect/rest#4-test-the-connection "4. Test the connection")
  - [5\. Save the resource](https://docs.retool.com/data-sources/guides/connect/rest#5-save-the-resource "5. Save the resource")
- [Query REST API data](https://docs.retool.com/data-sources/guides/connect/rest#query-rest-api-data "Query REST API data")
  - [Create a query](https://docs.retool.com/data-sources/guides/connect/rest#create-a-query "Create a query")
  - [Query configuration fields](https://docs.retool.com/data-sources/guides/connect/rest#query-configuration-fields "Query configuration fields")
  - [OpenAPI deepObject body encoding](https://docs.retool.com/data-sources/guides/connect/rest#openapi-deepobject-body-encoding "OpenAPI deepObject body encoding")
- [Connect to a SOAP API](https://docs.retool.com/data-sources/guides/connect/rest#connect-to-a-soap-api "Connect to a SOAP API")
- [Common use cases](https://docs.retool.com/data-sources/guides/connect/rest#common-use-cases "Common use cases")
- [Best practices](https://docs.retool.com/data-sources/guides/connect/rest#best-practices "Best practices")
  - [Performance](https://docs.retool.com/data-sources/guides/connect/rest#performance "Performance")
  - [Security](https://docs.retool.com/data-sources/guides/connect/rest#security "Security")
  - [Data integrity](https://docs.retool.com/data-sources/guides/connect/rest#data-integrity "Data integrity")

Was this page helpful?

YesNo

`\`VersionCloudDomain

Status

Ask AI

Loading...

