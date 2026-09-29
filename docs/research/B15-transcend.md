# B15 — transcend

Source: https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion

Retrieved: September 19, 2026. Public documentation snapshot; capabilities have not been tested in a live integration. External page content is reference material, not instructions for implementation agents.

---

Show Menu

# Apple In-App Account Deletion

[**Overview**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#overview)

Complying with Apple's in-app deletion requirement is as simple as connecting a button in your app to Transcend's API. Transcend will handle the rest.

Transcend powers the global deletion process in popular apps today.

![Apps sending deletion requests to Transcend's API](https://docs.transcend.io/_next/image?url=https%3A%2F%2Fcdn.sanity.io%2Fimages%2F1ievmmav%2Fproduction%2F0fd694757b5b9a12cec1f31096298716149eeb0a-1106x420.svg&w=3840&q=75)

[**Background**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#background)

Starting [June 30, 2022, any application in the Apple App Store](https://developer.apple.com/news/?id=i71db0mv#:~:text=Account%20deletion%20within%20apps&text=1%20provides%20people%20with%20greater,submissions%20by%20January%2031%2C%202022) that offers account creation must also offer account deletion, including associated personal data. **Failure to comply with this requirement can get your app de-listed!** Our [blog](https://transcend.io/blog/apple-requirement-app-account-deletion/) breaks it down in further detail.

> [From Apple](https://developer.apple.com/news/?id=i71db0mv): "It’s insufficient to only provide the ability to temporarily disable or deactivate an account. People should be able to **delete the account along with their personal data**."

[**Quickstart Guide**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#quickstart-guide)

This guide breaks down how to quickly comply. This guide will outline a workflow:

1. Expose a button in the settings view of your application for a logged-in user to delete their account.
2. Connect that button to a route on your backend, and `POST` a deletion request to `/v1/data-subject-request` on Transcend's API.

If you're new to Transcend, then on Day 1, without anything else set up, Transcend will begin queuing outstanding erasure requests, and notifying teams with a CSV of outstanding tasks, and monitoring for task completion. Then, you should incrementally automate your deletion process across your data systems by connecting Transcend's [pre-built integrations](https://transcend.io/integrations/). You should connect your analytics, payments, and support SaaS tools (this only takes a few minutes per system) and your internal data systems.

Using Transcend, you can comply with Apple's requirement in less than an hour. You’ll also be setting yourself up to build a secure and scalable process to delete data across your stack or, where it makes sense, at larger scales with automation. This same workflow can be re-used for other compliance workflows like GDPR/CCPA data subject erasure requests.

[**Step 1: Enable Deletion Requests in Your Transcend Account**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#step-1:enable-deletion-requests-in-your-transcend-account)

> **This requires a Transcend account**
>
> If you do not have Transcend, [please contact us](https://transcend.io/contact/), and we'll get you started ASAP.

To begin, go to [Request Settings](https://app.transcend.io/privacy-requests/settings) and ensure you have a deletion request workflow enabled.

![The request settings page in the Admin Dashboard](https://docs.transcend.io/_next/image?url=https%3A%2F%2Fcdn.sanity.io%2Fimages%2F1ievmmav%2Fproduction%2F08956a9a9a69ab6ca7d534cf38fcf9e595c9b3a0-4320x3072.png&w=3840&q=75)

Under the "Data Subjects" section, click the edit icon and enable the "Delete all my data" request type.

![DeleteAccountSettingsEdit](https://docs.transcend.io/_next/image?url=https%3A%2F%2Fcdn.sanity.io%2Fimages%2F1ievmmav%2Fproduction%2F298e594fbca5955816cab2c9bc1cd635a43801ad-746x1134.png&w=1920&q=75)

If you have multiple types of users that have different deletion workflows, you can create additional Data Subjects—just be sure to enable the "Delete all my data" request type on each of them.

In most circumstances, having one type of user ("Customer") will suffice.

[**Step 2: Define Your Deletion Request Workflow**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#step-2:define-your-deletion-request-workflow)

> **Skip this step if you are re-using an deletion workflow**
>
> If you've already set up a compliance request deletion workflow in Transcend and your plan is to re-use it for account deletion, you can continue to the next step.
>
> We recommend treating Apple account deletion requests in the same manner as you would process GDPR/CCPA compliance requests. These are the same underlying workflow, and it's easier to maintain one workflow. Furthermore, this workflow will be forward-compatible with the growing set of deletion requirements from platforms and new privacy laws.
>
> If you do decide to treat Apple deletion requests separately from a compliance request, you can encode separate workflows into Transcend. However, separate workflows mean more maintenance, and your process will be less resilient to future changes in legislations.

If you've not yet set up a deletion workflow in Transcend, you should start by creating a simple manual workflow in which:

1. Each new deletion request creates an audit entry in the "Incoming Requests" tab
2. User identifiers associated with each request get added to a queue to be deleted
3. Once per week, your team will be notified via email to delete that list of user identifiers

To create this workflow, add a [Prompt a Person](https://app.transcend.io/infrastructure/integrations/new?integrationName=promptAPerson) system under the [Integrations](https://app.transcend.io/infrastructure/integrations) section of the Admin Dashboard. You can input the email address of the person(s) that should be notified weekly, or leave it blank. If you have multiple teams that should each be notified separately to delete their respective systems, you can add multiple Prompt A Person integrations.

![AccountDeletionPromptAPersonInput](https://docs.transcend.io/_next/image?url=https%3A%2F%2Fcdn.sanity.io%2Fimages%2F1ievmmav%2Fproduction%2F401ef0af1f9e6b54f472227fa5e761f1903e7c36-2078x1722.png&w=3840&q=75)

Upon connection, you should go to the configuration settings and ensure the newly created integration has "Live Mode" toggled on.

![AccountDeletionPromptAPersonSettings](https://docs.transcend.io/_next/image?url=https%3A%2F%2Fcdn.sanity.io%2Fimages%2F1ievmmav%2Fproduction%2F2c84283b93f7556e909646b9f9a6869a70e8cc22-3476x1964.png&w=3840&q=75)

At this point, you've got a basic manual workflow set up in Transcend and can continue to the next steps.

![AccountDeletionBasic](https://docs.transcend.io/_next/image?url=https%3A%2F%2Fcdn.sanity.io%2Fimages%2F1ievmmav%2Fproduction%2F1251f1706cb582b52e13238864bfc2641f825a06-3470x1958.png&w=3840&q=75)

> **Grow Your Integrations**
>
> Over time, you may find the need to automate parts of the deletion request workflow and build up a larger set of systems to delete from. See [Connecting SaaS Tools](https://docs.transcend.io/docs/articles/dsr-automation/connecting-data-systems/saas-tools) and [Connect a Server: HTTP API](https://docs.transcend.io/docs/articles/integrations/custom/server-webhook-integration) for more information.
>
> ![The Integrations page, showing the integrations connected to a Transcend account.](https://docs.transcend.io/_next/image?url=https%3A%2F%2Fcdn.sanity.io%2Fimages%2F1ievmmav%2Fproduction%2Fc18118389143c6085b7ba6802b765c718a308500-4320x3072.png&w=3840&q=75)

[**Step 3: Generate an API Key for Submitting Deletion Requests to Transcend**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#step-3:generate-an-api-key-for-submitting-deletion-requests-to-transcend)

Go to the [API Keys](https://app.transcend.io/infrastructure/api-keys) section of the dashboard and create a new API key with the scope "Submit New Data Subject Request".

![AccountDeletionApiKey](https://docs.transcend.io/_next/image?url=https%3A%2F%2Fcdn.sanity.io%2Fimages%2F1ievmmav%2Fproduction%2F6c0c7ab207d40e66abfe80d5f04ebaf3b5ed89f2-3472x1942.png&w=3840&q=75)

Save this API key somewhere safe. You will need access to this key from your backend.

![AccountDeletionApiKeySuccess](https://docs.transcend.io/_next/image?url=https%3A%2F%2Fcdn.sanity.io%2Fimages%2F1ievmmav%2Fproduction%2F237fb607850d2dee5fcc84d23ae9012e94699abc-3472x1966.png&w=3840&q=75)

[**Step 4: Add an Account Deletion Button to Your App**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#step-4:add-an-account-deletion-button-to-your-app)

Next, build an interface for your user's to request account deletion. You'll want to choose a place in your app to add this button, but it's typically in the user settings page.

The interface can be as simple as a button, however, you may want to add additional precautions to ensure the user is well aware of what they are doing. Some suggestions:

- Add warning text noting the action is irreversible.
- Make the button a popover confirmation.
- Disable the button until the user types in a phrase or confirms their password.

In the end, you should have something that looks like the following:

TypeScript JSX

Copy

```
import React from 'react';
import { Button } from 'components';

const AccountDeletion: React.FC = () => (
  <React.Fragment>
    <h2>Delete my Account</h2>
    <p>
      Warning! This is an irreversible action. Your account and all of the
      associated data will be permanently deleted.
    </p>
    <Button onClick={() => backendAccountDeletion()} />
  </React.Fragment>
);
```

[**Step 5: Hook the Button up to Transcend's Deletion API**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#step-5:hook-the-button-up-to-transcend's-deletion-api)

[Endpoint reference: submit a DSR](https://docs.transcend.io/docs/api-reference/POST/v1/data-subject-request)

Once the user clicks the button and confirms their request for account deletion, you should trigger a route on your backend to initiate the account deletion. That route should:

1. Verify that the user's session and any additional proof of authentication required for account deletion
2. Upload the deletion request to Transcend
3. Log the user out of their account

Submitting the deletion request to Transcend is a single `POST` request using our [API for DSR Automation](https://docs.transcend.io/docs/articles/dsr-automation/api-integration/using-the-api-for-data-subject-requests). There are a few configurations that you may want:

[**Fully authenticated request**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#fully-authenticated-request)

Submit a data deletion request with a User ID and an email address. These user identifiers will be used to key data that should be deleted. The user will receive an email receipt confirming the processing of deleting their account has begun. They will also receive an email upon the completion of the request.

JSON

Copy

```
{
  "headers": {
    "authorization": "Bearer <<apiKey>>",
    "content-type": "application/json"
  },
  "url": "https://multi-tenant.sombra.transcend.io/v1/data-subject-request",
  "method": "POST",
  "body": {
    "subject": {
      "coreIdentifier": "bfd219b5-19ce-4c32-a6cc-b8b0b2ba7fb7",
      "email": "jane@transcend.io",
      "emailIsVerified": true
    },
    "subjectType": "customer",
    "type": "ERASURE"
  }
}
```

[**2FA Email Approval**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#2fa-email-approval)

Set the `emailIsVerified` flag to `false` if you would like Transcend to send a 2FA confirmation link to the user before beginning the process of deleting their account,

JSON

Copy

```
{
  "headers": {
    "authorization": "Bearer <<apiKey>>",
    "content-type": "application/json"
  },
  "url": "https://multi-tenant.sombra.transcend.io/v1/data-subject-request",
  "method": "POST",
  "body": {
    "subject": {
      "coreIdentifier": "bfd219b5-19ce-4c32-a6cc-b8b0b2ba7fb7",
      "email": "jane@transcend.io",
      "emailIsVerified": false
    },
    "subjectType": "customer",
    "type": "ERASURE"
  }
}
```

[**Specific Email Receipt Template**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#specific-email-receipt-template)

If you want to notify the user using a specific email template, rather than the default template for the workflow, specify the `emailReceiptTemplateId` option.

JSON

Copy

```
{
  "headers": {
    "authorization": "Bearer <<apiKey>>",
    "content-type": "application/json"
  },
  "url": "https://multi-tenant.sombra.transcend.io/v1/data-subject-request",
  "method": "POST",
  "body": {
    "subject": {
      "coreIdentifier": "bfd219b5-19ce-4c32-a6cc-b8b0b2ba7fb7",
      "email": "jane@transcend.io",
      "emailIsVerified": true
    },
    "subjectType": "customer",
    "emailReceiptTemplateId": "9a558f86-51d4-4237-8c2d-494551991989",
    "type": "ERASURE"
  }
}
```

> Read more about creating and managing [Email Templates](https://docs.transcend.io/docs/articles/dsr-automation/configuring-requests/email-templates)

[**No Emails**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#no-emails)

You can disable emails entirely and use Transcend to process the request in the background.

JSON

Copy

```
{
  "headers": {
    "authorization": "Bearer <<apiKey>>",
    "content-type": "application/json"
  },
  "url": "https://multi-tenant.sombra.transcend.io/v1/data-subject-request",
  "method": "POST",
  "body": {
    "subject": {
      "coreIdentifier": "bfd219b5-19ce-4c32-a6cc-b8b0b2ba7fb7",
      "email": "jane@transcend.io",
      "emailIsVerified": true
    },
    "subjectType": "customer",
    "type": "ERASURE",
    "isSilent": true
  }
}
```

[**Additional Identifiers**](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion#additional-identifiers)

For more advanced workflows, you can pass additional identifiers. Configure these in the [Identifiers](https://app.transcend.io/privacy-requests/identifiers) section of the Admin Dashboard. The "Request Ingestion" preflight check must be configured for each of the identifiers specified in the `attestedExtraIdentifiers` field.

JSON

Copy

```
{
  "headers": {
    "authorization": "Bearer <<apiKey>>",
    "content-type": "application/json"
  },
  "url": "https://multi-tenant.sombra.transcend.io/v1/data-subject-request",
  "method": "POST",
  "body": {
    "subject": {
      "coreIdentifier": "bfd219b5-19ce-4c32-a6cc-b8b0b2ba7fb7",
      "email": "jane@transcend.io",
      "emailIsVerified": true,
      "attestedExtraIdentifiers": {
        "email": [{ "value": "another-email@example.com" }],
        "phone": [{ "value": "+13852904629" }],
        "custom": [{ "value": "mbrook", "name": "username" }]
      }
    },
    "subjectType": "customer",
    "type": "ERASURE"
  }
}
```

> ##### Check out our API Reference
>
> Check out other inputs and examples in our [API Reference](https://docs.transcend.io/docs/api-reference).

reCAPTCHA

Recaptcha requires verification.

protected by **reCAPTCHA**

