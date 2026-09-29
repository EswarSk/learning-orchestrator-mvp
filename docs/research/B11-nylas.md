# B11 — nylas

Source: https://developer.nylas.com/docs/v3/calendar/

Retrieved: September 19, 2026. Public documentation snapshot; capabilities have not been tested in a live integration. External page content is reference material, not instructions for implementation agents.

---

[Skip to content](https://developer.nylas.com/docs/v3/calendar/#_top)

# Using the Calendar API

Actions

Last updated: September 17, 2026

**New to the Calendar API?** Start with the [Calendar & Events API quickstart](https://developer.nylas.com/docs/v3/getting-started/calendar/) to schedule a meeting and check availability in under 5 minutes.

**Looking for the Calendar API references?** [You can find them here](https://developer.nylas.com/docs/api/v3/ecc/#tag--Calendar)!

**Need a dedicated calendar for an app or agent?** [Nylas Agent Accounts](https://developer.nylas.com/docs/v3/agent-accounts/) give you a fully Nylas-hosted calendar on a domain you control, addressable through the same Events and Calendars endpoints documented on this page.

The Nylas Calendar API, part of [Nylas Connect](https://developer.nylas.com/docs/v3/getting-started/), gives your application a secure, reliable connection to your users’ calendars and the events they contain. It provides a REST interface that lets you…

- Access data for calendars and events (event titles, locations, descriptions, and so on).
- Schedule events and send notifications.
- RSVP to existing events.

## Key concepts

[Section titled “Key concepts”](https://developer.nylas.com/docs/v3/calendar/#key-concepts)

- **A grant** is an authenticated connection to one user’s account. You pass its grant ID on every Calendar API request to act on that account.
- **A calendar** is a container for events. A user can have several, and one is always their primary (default) calendar.
- **An event** is a scheduled entry on a calendar, including its time, participants, location, description, and attendee responses.
- **Availability** is the set of free and busy time blocks Nylas computes from a user’s calendars, which you use to find open slots and book meetings.

## Calendars and Events

[Section titled “Calendars and Events”](https://developer.nylas.com/docs/v3/calendar/#calendars-and-events)

Each user who authenticates with your Nylas application might have zero, one, or multiple calendars. Their account can have many calendars, including the primary calendar, shared team calendars, and other custom calendars. They might also have Read access to their teammates’ calendars.

In general, a Calendar object serves as a container for Event objects. A single user usually has access to several calendars, but one is always considered their “primary” or default calendar. Calendars can be shared amongst users, and users might have full Read/Write access to a calendar (owner or organizer), Read-only access (subscriber), or they might be able to see only the free/busy information.

By default, when a user authenticates their account with your application, Nylas generates notifications for their primary calendar only. You can [modify settings in the Nylas Dashboard](https://developer.nylas.com/docs/v3/calendar/#receive-notifications-for-all-calendars) to receive notifications for all of your users’ calendars.

Event objects are containers for information about scheduled events. This includes a list of the people involved, details about the time, meeting location (in-person address or virtual conferencing details), and a description. They can also include attachments and information about who has confirmed their attendance.

There are several types of calendars that you might encounter when working with the Nylas Calendar APIs:

- [Provider calendars](https://developer.nylas.com/docs/v3/calendar/#provider-calendars)
- [Agent Account calendars](https://developer.nylas.com/docs/v3/calendar/#agent-account-calendars)

### Provider calendars

[Section titled “Provider calendars”](https://developer.nylas.com/docs/v3/calendar/#provider-calendars)

These are the most common type of calendar. There are several types of provider calendars: personal, group, and resource. You can view, add, modify, and delete events on calendars where `read_only` is `false`. Nylas makes changes directly on the provider.

#### Microsoft shared calendars

[Section titled “Microsoft shared calendars”](https://developer.nylas.com/docs/v3/calendar/#microsoft-shared-calendars)

You can use the full set of Calendar APIs on Microsoft shared calendars only if they are correctly configured, and used with the correct permissions.

- Set the `Calendars.ReadWrite.Shared` API permission in the [Azure application](https://developer.nylas.com/docs/provider-guides/microsoft/create-azure-app/).
- Make sure that the calendar owner specifically gives individual users access to the shared calendar. The owner can grant `Can view all details` or `Can edit` permissions, depending on your use cases.
If the shared calendar owner grants “organizational access” to a shared calendar, meaning default-on access granted to the entire domain (sometimes called “tenant-based sharing”), then the owner must explicitly grant the `Can edit` permission to users who need access using Nylas. The `Can view all details` permission does not allow Nylas access to the shared calendar.

Shared calendars come back in the same list as the user’s own calendars. To tell them apart, check [the `owner_email` field](https://developer.nylas.com/docs/v3/calendar/#the-owner_email-field), which names the account that owns each calendar.

### Agent Account calendars

[Section titled “Agent Account calendars”](https://developer.nylas.com/docs/v3/calendar/#agent-account-calendars)

An [Agent Account](https://developer.nylas.com/docs/v3/agent-accounts/calendars/) is a Nylas-hosted mailbox and calendar on a domain you control. Use one when you need a calendar for something that has no account on Google, Microsoft, or another provider: a meeting room, a piece of equipment, an external contractor, or an AI agent. You create it with one API call and no OAuth flow, and it never expires.

Because the calendar sits on a real mailbox, it behaves like a person’s calendar to everyone else. It sends invitations when you create an event with `notify_participants=true`, receives invitations that anyone sends to its address, and answers them through the [Send RSVP endpoint](https://developer.nylas.com/docs/reference/api/events/send-rsvp/). Updates and cancellations reach every attendee’s calendar. All of the Calendar and Events endpoints on this page work on an Agent Account grant, and so does [Scheduler](https://developer.nylas.com/docs/v3/scheduler/agent-accounts/).

### Limited calendar support for IMAP providers

[Section titled “Limited calendar support for IMAP providers”](https://developer.nylas.com/docs/v3/calendar/#limited-calendar-support-for-imap-providers)

IMAP is a protocol for _receiving_ messages only. “IMAP” service providers use the [IMAP protocol](https://en.wikipedia.org/wiki/Internet_Message_Access_Protocol) to receive messages, and the [SMTP protocol](https://en.wikipedia.org/wiki/Simple_Mail_Transfer_Protocol) to send them. Neither of these protocols includes calendar functionality.

## Calendar availability

[Section titled “Calendar availability”](https://developer.nylas.com/docs/v3/calendar/#calendar-availability)

The Nylas APIs offer multiple ways to get information about users’ availability, so you can book meetings or run other time-based operations.

### The Free/Busy endpoint

[Section titled “The Free/Busy endpoint”](https://developer.nylas.com/docs/v3/calendar/#the-freebusy-endpoint)

The simplest way to get availability information is using the [Free/Busy endpoint](https://developer.nylas.com/docs/reference/api/calendar/post-calendars-free-busy/). When you make a Free/Busy request, Nylas queries the provider, which returns simple information about blocks of time when the user is booked and _not_ available. For more information, see [Check free/busy information](https://developer.nylas.com/docs/v3/calendar/check-free-busy/).

If you’re querying a user within your organization, this information is usually public within the org. If you’re querying for a user outside of your organization, the calendar’s owner must have published or made public the free/busy information.

![A calendar UI displaying the simple busy data returned by the Free/Busy endpoint.](https://developer.nylas.com/_images/calendar/busy-data.png)

### The Availability endpoint

[Section titled “The Availability endpoint”](https://developer.nylas.com/docs/v3/calendar/#the-availability-endpoint)

You can use the [Availability endpoint](https://developer.nylas.com/docs/reference/api/calendar/post-availability/) to find space in a user’s calendar for a meeting with specific and detailed criteria. Use the endpoint to find a time that works, which you can then feed to the Events API to create an actual booking. For more information, see [Check calendar availability](https://developer.nylas.com/docs/v3/calendar/calendar-availability/).

## Before you begin

[Section titled “Before you begin”](https://developer.nylas.com/docs/v3/calendar/#before-you-begin)

To follow along with the samples on this page, you first need to [sign up for a Nylas developer account](https://dashboard-v3.nylas.com/register?utm_source=docs&utm_medium=devrel-surfaces&utm_campaign=&utm_content=using-apis), which gets you a free Nylas application and API key.

For a guided introduction, you can follow the [Getting started guide](https://developer.nylas.com/docs/v3/getting-started/) to set up a Nylas account and Sandbox application. When you have those, you can connect an account from a calendar provider (such as Google, Microsoft, or iCloud) and use your API key with the sample API calls on this page to access that account’s data.

## Get available calendars

[Section titled “Get available calendars”](https://developer.nylas.com/docs/v3/calendar/#get-available-calendars)

Most users subscribe to multiple calendars. To view a list of all calendars they have access to, make a [Get all Calendars request](https://developer.nylas.com/docs/reference/api/calendar/get-all-calendars/).

**To follow along with this guide, copy the ID of a calendar that you’re willing to make modifications to**. You can always create a calendar for this purpose if you don’t want to test features out on any of your existing calendars.

To make examples easier to read, this guide uses a query parameter to limit the number of results to five. See the [pagination references](https://developer.nylas.com/docs/v3/calendar/#pagination) to learn more about controlling the number of objects that Nylas returns.

The following example shows how to get a user’s available calendars, and the type of JSON response you can expect.

RequestResponse (JSON)Node.js SDKPython SDKRuby SDKJava SDKKotlin SDK

```
curl -X GET "https://api.us.nylas.com/v3/grants/<NYLAS_GRANT_ID>/calendars?limit=5" \

  -H "accept: application/json"\

  -H "authorization: Bearer <NYLAS_API_KEY>"
```

```
{

  "request_id" : "123-456-789",

  "data" : [\
\
    {\
\
      "grant_id": "<NYLAS_GRANT_ID>"\
\
      "description": "Board game nights!",\
\
      "id": "zd08j9stfph95u449vti",\
\
      "is_primary": false,\
\
      "name": "Game nights",\
\
      "object": "calendar",\
\
      "read_only": true,\
\
      "timezone": "America/Toronto",\
\
      "hex_color": "#000000",\
\
      "is_owned_by_user": true,\
\
      "owner_email": "nyla@example.com"\
\
    },\
\
    {\
\
    "grant_id": "<NYLAS_GRANT_ID>"\
\
      "description": null,\
\
      "id": "joujhadwh59pz9rvfjfw",\
\
      "is_primary": false,\
\
      "name": "jane.doe@example.com",\
\
      "object": "calendar",\
\
      "read_only": true,\
\
      "timezone": "America/Toronto",\
\
      "hex_color": "#000000",\
\
      "is_owned_by_user": false,\
\
      "owner_email": "jane.doe@example.com"\
\
    }\
\
  ]

}
```

```
import Nylas from "nylas";

const nylas = new Nylas({

  apiKey: "<NYLAS_API_KEY>",

  apiUri: "<NYLAS_API_URI>",

});

async function fetchFiveAvailableCalendars() {

  try {

    const calendars = await nylas.calendars.list({

      identifier: "<NYLAS_GRANT_ID>",

      queryParams: {

        limit: 5,

      },

    });

    console.log("Available Calendars:", calendars);

  } catch (error) {

    console.error("Error fetching calendars:", error);

  }

}

fetchFiveAvailableCalendars();
```

```
from nylas import Client

nylas = Client(

    "<NYLAS_API_KEY>",

    "<NYLAS_API_URI>"

)

grant_id = "<NYLAS_GRANT_ID>"

calendars = nylas.calendars.list(grant_id)

print(calendars)
```

```
require 'nylas'

nylas = Nylas::Client.new(api_key: "<NYLAS_API_KEY>")

calendars, _request_ids = nylas.calendars.list(identifier: "<NYLAS_GRANT_ID>", query_params: {limit: 5})

calendars.each {|calendar|

    puts calendar

}
```

```
import com.nylas.NylasClient;

import com.nylas.models.*;

import java.util.List;

public class read_calendars {

  public static void main(String[] args) throws NylasSdkTimeoutError, NylasApiError {

    NylasClient nylas = new NylasClient.Builder("<NYLAS_API_KEY>").build();

    ListCalendersQueryParams listCalendersQueryParams = new ListCalendersQueryParams.

        Builder().

        limit(5).

        build();

    List<Calendar> calendars = nylas.calendars().list(dotenv.get("CALENDAR_ID"), listCalendersQueryParams).getData();

    for (Calendar calendar : calendars) {

      System.out.println(calendar);

    }

  }

}
```

```
import com.nylas.NylasClient

import com.nylas.models.*

import com.nylas.resources.Calendars

fun main(args: Array<String>) {

  val nylas: NylasClient = NylasClient(apiKey = "<NYLAS_API_KEY>")

  val calendarQueryParams: ListCalendersQueryParams = ListCalendersQueryParams(limit = 5)

  val calendars: List<Calendar> = nylas.calendars().list("<CALENDAR_ID>", calendarQueryParams).data

  for(calendar in calendars) {

    println(calendar)

  }

}
```

The `is_primary` field is included for Google and EWS calendars. If it’s `true`, it’s the account’s primary calendar. If `false`, it’s either a secondary or child calendar. For other providers, the primary calendar usually uses the account’s email address as the name, otherwise it’s called **Calendar**.

### The owner\_email field

[Section titled “The owner\_email field”](https://developer.nylas.com/docs/v3/calendar/#the-owner_email-field)

Microsoft calendars include an `owner_email` field that holds the email address of the account the calendar belongs to. It’s most useful when a Microsoft user has access to shared calendars: `is_owned_by_user` tells you whether the calendar is theirs, and `owner_email` tells you whose it is when it isn’t.

Nylas returns `owner_email` for Microsoft grants only. Google, iCloud, EWS, and virtual calendars omit the field entirely, so read it defensively rather than assuming it’s present on every Calendar object. The value is read-only, and you can’t set or change it with a `PUT` request.

## Get a calendar

[Section titled “Get a calendar”](https://developer.nylas.com/docs/v3/calendar/#get-a-calendar)

You can return information about a single calendar by making a [Get Calendar request](https://developer.nylas.com/docs/reference/api/calendar/get-calendars-id/) that includes the appropriate calendar `id`.

RequestResponse (JSON)Node.js SDKPython SDKRuby SDKJava SDKKotlin SDK

```
curl --compressed --request GET \

  --url 'https://api.us.nylas.com/v3/grants/<NYLAS_GRANT_ID>/calendars/<CALENDAR_ID>' \

  --header 'Accept: application/json' \

  --header 'Authorization: Bearer <NYLAS_API_KEY>' \

  --header 'Content-Type: application/json'
```

```
{

  "request_id": "123-456-789",

  "data": {

    "grant_id": "<NYLAS_GRANT_ID>",

    "description": "Board game nights!",

    "id": "zd08j9stfph95u449vti",

    "is_primary": false,

    "name": "Game nights",

    "object": "calendar",

    "read_only": false,

    "timezone": "UTC",

    "hex_color": "#a47ae2",

    "hex_foreground_color": "#000000",

    "is_owned_by_user": true,

    "owner_email": "nyla@example.com"

  }

}
```

```
import Nylas from "nylas";

const nylas = new Nylas({

  apiKey: "<NYLAS_API_KEY>",

  apiUri: "<NYLAS_API_URI>",

});

async function fetchCalendar() {

  try {

    const calendar = await nylas.calendars.find({

      identifier: "<NYLAS_GRANT_ID>",

      calendarId: "<CALENDAR_ID>",

    });

    console.log("Calendar:", calendar);

  } catch (error) {

    console.error("Error fetching calendars:", error);

  }

}

fetchCalendar();
```

```
from nylas import Client

nylas = Client(

    "<NYLAS_API_KEY>",

    "<NYLAS_API_URI>"

)

grant_id = "<NYLAS_GRANT_ID>"

calendar = nylas.calendars.find(

    grant_id,

    "<CALENDAR_ID>"

)

print(calendar)
```

```
require 'nylas'

nylas = Nylas::Client.new(api_key: "<NYLAS_API_KEY>")

calendar, _request_ids = nylas.calendars.find(

    identifier: "<NYLAS_GRANT_ID>",

    calendar_id: "<CALENDAR_ID>"

)

puts calendar
```

```
import com.nylas.NylasClient;

import com.nylas.models.*;

public class GetCalendar {

  public static void main(String[] args) throws NylasSdkTimeoutError, NylasApiError {

  NylasClient nylas = new NylasClient.Builder("<NYLAS_API_KEY>").build();

  Response<Calendar> calendar = nylas.calendars().find("<NYLAS_GRANT_ID>", "<CALENDAR_ID>");

  System.out.println("Id: " + calendar.getData().getId() +

      " | Name: " + calendar.getData().getName() +

      " | Description: " + calendar.getData().getDescription());

  }

}
```

```
import com.nylas.NylasClient

import com.nylas.models.*

fun main(args: Array<String>) {

  val nylas: NylasClient = NylasClient(apiKey = "<NYLAS_API_KEY>")

  val calendar: Response<Calendar> = nylas.calendars().find("<NYLAS_GRANT_ID>", "<CALENDAR_ID>")

  println("Id: " + calendar.data.id +

      " | Name: " + calendar.data.name +

      " | Description: " + calendar.data.description)

}
```

## Create a calendar

[Section titled “Create a calendar”](https://developer.nylas.com/docs/v3/calendar/#create-a-calendar)

Because a calendar is a container for Event objects, it doesn’t have a lot of parameters or options available. In fact, when you create a calendar for a Microsoft account, you can specify the name as a string only. For other providers, you can also add a description, a location, an [IANA-formatted timezone](https://en.wikipedia.org/wiki/Tz_database), and optional metadata.

The following example shows how to use a [Create Calendar request](https://developer.nylas.com/docs/reference/api/calendar/create-calendar/) and the type of JSON response you can expect.

RequestResponse (JSON)Node.js SDKPython SDKRuby SDKJava SDKKotlin SDK

```
curl --compressed --request POST \

  --url 'https://api.us.nylas.com/v3/grants/<NYLAS_GRANT_ID>/calendars' \

  --header 'Accept: application/json' \

  --header 'Authorization: Bearer <NYLAS_API_KEY>' \

  --header 'Content-Type: application/json' \

  --data '{

    "name": "My New Calendar",

    "description": "Description of my new calendar",

    "location": "Location description",

    "timezone": "America/Los_Angeles"

  }'
```

```
{

  "request_id": "1",

  "data": {

    "grant_id": "1",

    "description": "<string>",

    "id": "2",

    "is_primary": false,

    "location": "<string>",

    "metadata": {

      "your-key": "<string>"

    },

    "name": "<string>",

    "object": "calendar",

    "owner_email": "nyla@example.com",

    "read_only": false,

    "timezone": "UTC",

    "is_owned_by_user": false

  }

}
```

```
import Nylas from "nylas";

const nylas = new Nylas({

  apiKey: "<NYLAS_API_KEY>",

  apiUri: "<NYLAS_API_URI>",

});

async function createCalendar() {

  try {

    const calendar = await nylas.calendars.create({

      identifier: "<NYLAS_GRANT_ID>",

      requestBody: {

        name: "Nylas DevRel",

        description: "Nylas Developer Relations",

      },

    });

    console.log("Calendar:", calendar);

  } catch (error) {

    console.error("Error to create calendar:", error);

  }

}

createCalendar();
```

```
from nylas import Client

nylas = Client(

    "<NYLAS_API_KEY>",

    "<NYLAS_API_URI>"

)

grant_id = "<NYLAS_GRANT_ID>"

calendar = nylas.calendars.create(

    grant_id,

    request_body={

      "name": 'Nylas DevRel',

      "description": 'Nylas Developer Relations'

    }

)

print(calendar)
```

```
require 'nylas'

nylas = Nylas::Client.new(api_key: "<NYLAS_API_KEY>")

query_params = {

  calendar_id: "<CALENDAR_ID>"

}

request_body = {

  "name": "My New Calendar",

  "description": "Description of my new calendar",

  "location": "Location description",

  "timezone": "America/Toronto",

  "metadata": { "key1":"This is my metadata" }

}

calendar, _request_ids = nylas.calendars.create(

    identifier: "<NYLAS_GRANT_ID>",

    request_body: request_body)

puts calendar
```

```
import com.nylas.NylasClient;

import com.nylas.models.*;

import java.util.Map;

public class CreateCalendar {

  public static void main(String[] args) throws NylasSdkTimeoutError, NylasApiError {

    NylasClient nylas = new NylasClient.Builder("<NYLAS_API_KEY>").build();

    CreateCalendarRequest requestBody = new CreateCalendarRequest.Builder("My New Calendar")

        .description("Description of my new calendar")

        .location("Location description")

        .timezone("America/Toronto")

        .metadata(Map.of("key1", "This is my metadata"))

        .build();

    Response<Calendar> calendar = nylas.calendars().

        create("<NYLAS_GRANT_ID>", requestBody);

    System.out.println(calendar.getData());

  }

}
```

```
import com.nylas.NylasClient

import com.nylas.models.*

fun main(args: Array<String>) {

  val nylas: NylasClient = NylasClient(apiKey = "<NYLAS_API_KEY>")

  val requestBody = CreateCalendarRequest(

      "My New Calendar",

      "Description of my new calendar",

      "Location description",

      "America/Toronto",

      mapOf("key1" to "This is my metadata")

  )

  val calendar: Response<Calendar> = nylas.calendars().

      create("<NYLAS_GRANT_ID>", requestBody)

  print(calendar.data)

}
```

## Update a calendar

[Section titled “Update a calendar”](https://developer.nylas.com/docs/v3/calendar/#update-a-calendar)

You can update a calendar to change any of the available fields, except if the user’s provider is Microsoft. For Microsoft calendars, you can update only the name string. To update any other information, you must delete the calendar and re-create it.

To update a calendar, make an [Update Calendar request](https://developer.nylas.com/docs/reference/api/calendar/put-calendars-id/) that includes the calendar’s `id`.

RequestResponse (JSON)Node.js SDKPython SDKRuby SDKJava SDKKotlin SDK

```
curl --compressed --request PUT \

  --url 'https://api.us.nylas.com/v3/grants/<NYLAS_GRANT_ID>/calendars/<CALENDAR_ID>' \

  --header 'Accept: application/json' \

  --header 'Authorization: Bearer <NYLAS_API_KEY>' \

  --header 'Content-Type: application/json' \

  --data '{

    "name": "My New Calendar",

    "description": "Description of my new calendar",

    "location": "Location description",

    "timezone": "America/Los_Angeles"

  }'
```

```
{

  "request_id": "1",

  "data": {

    "grant_id": "1",

    "description": "<string>",

    "id": "2",

    "is_primary": false,

    "location": "<string>",

    "metadata": {

      "your-key": "<string>"

    },

    "name": "<string>",

    "object": "calendar",

    "owner_email": "nyla@example.com",

    "read_only": false,

    "timezone": "UTC",

    "is_owned_by_user": false

  }

}
```

```
import Nylas from "nylas";

const nylas = new Nylas({

  apiKey: "<NYLAS_API_KEY>",

  apiUri: "<NYLAS_API_URI>",

});

async function updateCalendar() {

  try {

    const calendar = await nylas.calendars.update({

      identifier: "<NYLAS_GRANT_ID>",

      calendarId: "<CALENDAR_ID>",

      requestBody: {

        name: "Nylas DevRel Calendar",

        description: "Nylas Developer Relations",

      },

    });

    console.log("Updated Calendar:", calendar);

  } catch (error) {

    console.error("Error to update calendar:", error);

  }

}

updateCalendar();
```

```
from nylas import Client

nylas = Client(

    "<NYLAS_API_KEY>",

    "<NYLAS_API_URI>"

)

grant_id = "<NYLAS_GRANT_ID>"

calendar = nylas.calendars.update(

    grant_id,

    calendar_id="<CALENDAR_ID>",

    request_body={

      "name": 'Nylas DevRel Calendar',

      "description": 'Nylas Developer Relations'

    }

)

print(calendar)
```

```
require 'nylas'

nylas = Nylas::Client.new(api_key: "<NYLAS_API_KEY>")

request_body = {

  "name": "\"New Test Calendar (changed)\"",

  "description": "\"this calendar has been updated!\"",

}

calendar, _request_ids = nylas.calendars.update(

    identifier: "<NYLAS_GRANT_ID>",

    calendar_id: "<CALENDAR_ID",

    request_body: request_body)

puts calendar
```

```
import com.nylas.NylasClient;

import com.nylas.models.*;

public class UpdateCalendar {

  public static void main(String[] args) throws NylasSdkTimeoutError, NylasApiError {

    NylasClient nylas = new NylasClient.Builder("<NYLAS_API_KEY>").build();

    UpdateCalendarRequest requestBody = new UpdateCalendarRequest.Builder().

        name("My New Calendar").

        description("Description of my new calendar").

        location("Location description").

        timezone("America/Los_Angeles").

        build();

    Response<Calendar> calendar = nylas.calendars().update(

        "<CALENDAR_ID>",

        "<CALENDAR_ID>",

        requestBody);

    System.out.println(calendar.getData());

  }

}
```

```
import com.nylas.NylasClient

import com.nylas.models.*

import com.nylas.resources.Calendars

fun main(args: Array<String>) {

  val nylas: NylasClient = NylasClient(apiKey = "<NYLAS_API_KEY>")

  val requestBody = UpdateCalendarRequest.Builder().

      name("\"New Test Calendar (changed)\"").

      description("\"this calendar has been updated!\"").

      location("Location description").

      timezone("America/Los_Angeles").

      build()

  val calendar: Response<Calendar> = nylas.calendars().update(

      "<CALENDAR_ID>",

      "<CALENDAR_ID>",

      requestBody)

  print(calendar.data)

}
```

## Receive notifications for all calendars

[Section titled “Receive notifications for all calendars”](https://developer.nylas.com/docs/v3/calendar/#receive-notifications-for-all-calendars)

**If you choose to receive webhook notifications for all of your users’ calendars, Nylas might generate a large number of webhooks**. Prepare for this possibility by scaling your webhook processing _before_ you change your subscriptions. For more information, see [Best practices for webhooks](https://developer.nylas.com/docs/dev-guide/best-practices/webhook-best-practices/).

By default, when a user authenticates their account with your application, Nylas generates notifications for their primary calendar only. You can change this in your organization settings to receive notifications for all of your users’ calendars. This is an organization-level setting that applies to every application in your organization.

1. Log in to the [Nylas Dashboard](https://dashboard-v3.nylas.com/login?utm_source=docs&utm_content=using-calendar-api).
2. Expand the **account** menu at the top-right of the page, then click the **gear** symbol next to your organization’s name to open your organization settings.
3. Select **Customizations** in the left navigation.
![The Nylas Dashboard organization Customizations page, on the Notifications tab, showing the "Subscribe to primary calendars only" setting.](https://developer.nylas.com/_images/calendar/v3-primary-calendar-webhook-settings.png)
4. Deselect the **Subscribe to primary calendars only** option and **save** your changes.

After you update this setting, you might need to wait a few minutes for your change to take effect.

**Because iCloud doesn’t support marking calendars as primary, Nylas sends `calendar.updated` webhook notifications for changes to all calendars on an iCloud account**. This happens even if you have “Subscribe to primary calendars only” enabled.

## Calendar limitations

[Section titled “Calendar limitations”](https://developer.nylas.com/docs/v3/calendar/#calendar-limitations)

- When creating a calendar on Microsoft, you can define only its name.
  - Similarly, you can update only the name of a Microsoft calendar.

## More resources

[Section titled “More resources”](https://developer.nylas.com/docs/v3/calendar/#more-resources)

The [Nylas Samples Github repository](https://github.com/nylas-samples#calendar-and-events-api-samples) includes working example applications in several languages that use the Nylas Calendar API. The following examples are interesting places to start exploring:

- [Map the relationship between events and messages](https://github.com/nylas-samples/map_events_and_emails).
- [Manage team schedules](https://github.com/nylas-samples/team_schedules).
- [Optimize user schedules](https://github.com/nylas-samples/optimize_user_schedules).

Key concepts

Menu

