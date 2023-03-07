HTTP Destination
================
This repository contains documentation and samples for building publish
destinations using the "generic" HTTP destination built into Kapost.

Contact your customer success manager to learn more.

Getting Started
---------------
The HTTP destination that is built into Kapost allows building custom publish
destinations by the means of hitting a user built custom HTTP (REST) endpoint.

This is not a 100% traditional endpoint and in a sense it is more similar to
an endpoint that would receive and handle an incoming _webhook event_.

### Configuration
When setting up an HTTP destination in the Kapost App Center, one has the ability
to configure a number of options.

#### Authentication
There are four options when it comes to authentication.

| Authentication   | Description |
| ---------------- | ----------- |
| None             | No authentication |
| Basic Auth       | Basic authentication with username and password |
| Bearer (API Key) | API Key based authentication |
| Payload (JSON)   | User defined custom JSON authentication payload |

The Basic Auth and Bearer (API Key) authentication types should be self
explanatory.

As for the Payload (JSON), well like its name suggests, it allows you to define
a custom "JSON" payload with your own key / value pairs which is then sent over
with every request.

#### Signature
The signature options allow you to configure the signature type as well as a
signature secret to be used when "computing" the signature.

| Signature | Description  |
| --------- | ------------ |
| None | No signature |
| SHA1 | Compute SHA1 signature |
| SHA256 | Compute SHA256 signature |
| Secret | Secret to compute the SHA1 or SHA256 |

#### Endpoint
The endpoint lets you define a canonical HTTP endpoint that the HTTP destination
will interact with and make requests to.

| Endpoint | Description |
| -------- | ----------- |
| URL | Canonical HTTP endpoint URL |

The provided [samples][2] assume that you configured the HTTP destination in the
Kapost App Center with Bearer (API Key) authentication and a SHA256 signature.

### Actions
At this point in time, any HTTP endpoint defined in the App Center is expected to
implement three distinct "actions" for a successful connection and integration.

All actions come in the form of a "POST" request to the defined HTTP endpoint.

In addition, regardless of the authentication and signature configuration, all
requests made to the defined HTTP endpoint will have the following mandatory
request headers.

| Name | Value |
| ---- | ----- |
| Content-Type | application/json |
| X-Kapost-Platform | HTTP |
| X-Kapost-Action | auth, publish or republish |

When "Basic Auth" has been configured as the preferred authentication method
then there will be an additional header.

| Name | Value |
| ---- | ----- |
| Authorization | Basic base64\_encoded\_username\_password |

If you are an unsure of what this means in practice, please click [here][3] in
order to learn more.

When "Bearer (API Key)" has been configured as the preferred authentication
method then there will be an additional header.

| Name | Value |
| ---- | ----- |
| Authorization | Bearer api\_key |

If you are an unsure of what this means in practice, please click [here][4] in
order to learn more.

When "SHA1" or "SHA256" has been configured as the preferred signature method,
then there will be an additional header.

| Name | Value |
| ---- | ----- |
| X-Kapost-Signature | sha1 or sha256 signature (i.e sha256=xyz) |

#### Responses
All responses returned by the defined HTTP endpoint must set the `Content-Type`
header to `application/json`.

##### Errors
In the event of any errors during any of the "actions", the HTTP status code
must be set and a well formed JSON response must be returned.

For instance, in the case of an authentication error, the HTTP status code
must be set to 401, and the returned response must look like the one presented
below.

```json
{
  "error": "Authentication error. Invalid credentials."
}
```

All error messages will be shown to users, therefore try to make them concise
and to the point, without cryptic error codes or magic numbers.

If Kapost fails to "extract" the error from the response, it will simply show
"Unexpected error" to the users.

#### Signature
When "SHA1" or "SHA256" has been configured as the preferred signature method,
then you must compute the signature and check it against the value of the
`X-Kapost-Signature` header.

```ruby
computed_signature = [
  "sha256",
  OpenSSL::HMAC.hexdigest(
    OpenSSL::Digest.new("sha256"),
    ENV.fetch("SIGNATURE_SECRET"),
    [
      request.headers["X-Kapost-Platform"],
      request.headers["X-Kapost-Action"],
      request.body.read
    ].join
  )
].join("=")

unless Rack::Utils.secure_compare?(headers["X-Kapost-Signature"], computed_signature)
  raise_error(403, "Invalid signature")
end
```

In order to avoid "timing" based attacks, one should never compare the computed
signatures with the good old regular equality operator provided by the language of
choice, but rather use a "secure compare" method provided by the _crypto library_.

Please consult the provided samples for concrete implementations of signature
verification by clicking [here][2].

#### Authentication
The "auth" action is called when the connection is made in the Kapost App
Center and it is used to verify the credentials as well as discover the
supported capabilities of the endpoint.

##### Request Headers
| Name | Value |
| ---- | ----- |
| Content-Type | application/json |
| X-Kapost-Platform | HTTP |
| X-Kapost-Action | auth |

##### Request Payload
```json
{
  "authentication": {
    "instance": {
      "id": "5c536647480a0e68ee000001",
      "subdomain": "instance"
    },
    "destination": {
      "id": "5f9a8f6b480a0e23f5b0d456",
      "platform": "HTTP"
    }
  }
}
```

If Payload (JSON) was configured as the preferred authentication method and the
defined payload looks like the one shown below.

```json
{
  "client_id": "xxx",
  "client_secret": "yyy"
}
```

Then the payload would take the following form.

```json
{
  "authentication": {
    "client_id": "xxx",
    "client_secret": "yyy",
    "instance": {
      "id": "5c536647480a0e68ee000001",
      "subdomain": "instance"
    },
    "destination": {
      "id": "5f9a8f6b480a0e23f5b0d456",
      "platform": "HTTP"
    }
  }
}
```

##### Response
The "auth" action must return a "hash" of supported capabilities by the
endpoint.

```json
{
  "capabilities": {
    "html": true
  }
}
```

| Capability | Description |
| ---------- | ----------- |
| any\_file | Available as a primary destination for Content Types with an Any File body type |
| document | Available as a primary destination for Content Types with a Document body type |
| html | Available as a primary destination for Content Types with an HTML body type |
| social\_media | Available as a primary destination for Content Types with a Social Media body type |
| social\_media\_external\_schedule | Enables support for external scheduling for Content Types with a Social Media body type. Learn more below. |
| locations | Enables support for locations. Learn more below. |
| draft | Enables support for publishing as draft. Learn more below. |
| tracking\_url | Enables support for auto-generated Bit.ly tracking URL. Learn more below. |

If `social_media_external_schedule` is enabled then it is assumed that the endpoint implements
external scheduling for content types with a social media body type.

If `draft` is enabled then users will have the ability to publish/republish as draft
as well as publish/republish as live.

If `tracking_url` is enabled then Kapost will auto-generate a Bit.ly tracking URL
using the returned `published_url`. If the `published url` changes between
republishes, then the `tracking url` will be regenerated.

If `locations` are enabled as a supported capability, then the response is
expected to also contain an `array` of locations. These locations are then
available for the user to pick from during the publishing process. This is ideal
for building an endpoint that allows to user to publish to let's say several
blogs or in conjunction with social media, allowing them to publish
to several social media destinations based on the social media channel of the
content they are working on.

```json
{
  "capabilities": {
    "html": true,
    "locations": true
  },
  "locations": [
    {
      "id": "uniqueid",
      "name": "My Location"
    },
    {
      "id": "uniqueid2",
      "name": "My Second Location"
    }
  ]
}
```

Each location must have at least an `id` and `name`. They can also be "grouped"
by returning a `type` attribute for each.

In addition, when used in conjunction with the social media body type, they must
have a `channel` attribute, indicating which social media channel they support.

| Channel |
| ------- |
| Facebook |
| Instagram |
| LinkedIn |
| Pinterest |
| Twitter |
| YouTube |

The response can also return a "hash" of user defined `metadata`. This `metadata`
is then included in the `authentication` hash of the payloads of other actions.

```json
{
  "authentication": {
    "metadata": {
      "uniqueid": "dc42"
    }
  }
}
```

#### Publish
The publish action is called when a user publishes a piece of content that
hasn't been previously published.

##### Request Headers
| Name | Value |
| ---- | ----- |
| Content-Type | application/json |
| X-Kapost-Platform | HTTP |
| X-Kapost-Action | publish |

##### Request Payload
```json
{
  "action": {
    "draft": true,
    "scheduled_date": "2023-03-07 18:11:00 UTC"
  },
  "metadata": {
  },
  "content": {
    "id": "5fa13129480a0e3ca3db3a4f",
    "title": "test",
    "body": "<p>test</p>",
    "progress_stage":"planned",
    "updated_at": "2020-11-03T10:30:52Z",
    "created_at": "2020-11-03T10:30:01Z",
    "file": {
      "url": "https://asseturl/asset.pdf",
      "file_name": "asset.pdf"
    },
    "attachments": [
      {
        "url": "https://asseturl/attachment.pdf",
        "file_name": "attachment.pdf"
      }
    ],
    "media": [
      {
        "url": "https://asseturl/image.png",
        "file_name": "image.png"
      },
      {
        "url": "https://asseturl/video.mp4",
        "file_name": "video.mp4"
      }
    ],
    "custom_fields": {
      "hello_world_select": [
        "43",
        "World"
      ]
    },
    "type": {
      "id": "5f7c6139480a0ec80dc16d7e",
      "name": "Test Content Type",
      "field_name": "test_content_type"
    },
    "idea": {
      "id": "5c536610480a0c68ee000000"
    },
    "initiatives": [
    {
      "id": "5d536610480a0e68ee000000",
      "title": "My Initiative"
    }
    ],
    "author": {
      "id": "5c536610480a0e68ee000000",
      "name": "user",
      "email": "user@example.com"
    },
    "creator": {
      "id": "5c536610480a0e68ee000000",
      "name": "user",
      "email": "user@example.com"
    },
    "last_updated_by": {
      "id": "5c536610480a0e68ee000000",
      "name": "user",
      "email": "user@example.com"
    }
  },
  "authentication": {
    "metadata": {
    },
    "instance": {
      "id": "5c536647480a0e68ee000001",
      "subdomain": "instance"
    },
    "destination": {
      "id": "5f7c43e5480a0e75849d8bf0",
      "platform": "HTTP"
    }
  }
}
```

If the content that is being published has custom fields, then these custom
fields will be included in the payload keyed by their user defined custom field
name.

Custom field mappings can be used to customize this field name as well as map
the values of dropdown (select) and multi-select custom fields on a per
destination basis.

If locations are supported, then there will be a `location` hash in the payload
that will look like the one presented below. This is the `location` the user has picked
during the publishing process.

```json
{
  "location": {
    "id": "locationid",
    "name": "My Location"
  }
}
```

If publishing as draft is supported, then the `action` hash of the payload will
also contain a `draft` attribute, indicating whether the user has published as draft
or not.

```json
{
  "action": {
    "draft": true
  }
}
```

If `social_media_external_schedule` is supported, then the `action` hash of the payload will
also contain a `scheduled_date` attribute if the user decided to do a scheduled publish.

The date is in UTC.

If the user decided not to schedule, then there will be no `scheduled_date` attribute in the payload.

```json
{
  "action": {
    "draft": true,
    "scheduled_date": "2023-03-07 18:11:00 UTC"
  }
}
```

##### Response
The response must contain at least an `external_id` and `published_url`
attribute.

Failing to return these mandatory attributes, will result in an `Unexpected error`.

```json
{
  "metadata": {
    "external_id": "abc33",
    "published_url": "https://domain.com/content"
  }
}
```

The `external_id` attribute should be returned as a `string`.

In addition, an `embed_code` attribute can also be returned. This `embed code`
is then displayed to the user together with the published url.

Any additional attributes included in the `metadata` will be persisted and
sent back in the request payload during the republish action.

#### Republish
The republish action is called when a user republishes a previously published
piece of content. It has the exact same payload and response as the `publish` action
with one crucial difference. The `external_id` and any other previously returned `metadata`
during the initial `publish` action are also included in the payload.

##### Request Headers
| Name | Value |
| ---- | ----- |
| Content-Type | application/json |
| X-Kapost-Platform | HTTP |
| X-Kapost-Action | republish |

##### Request Payload
The `external_id` has special significance in Kapost, as a result it is also
included in the `action` hash in addition to the `metadata`.

The full content payload has been omitted for brevity.

```json
{
  "action": {
    "external_id": "abc33"
  },
  "metadata": {
    "external_id": "abc33",
    "published_url": "https://domain.com/content"
  },
  "content": {
    "id": "5fa13129480a0e3ca3db3a4f",
  },
  "authentication": {
    "metadata": {
    },
    "instance": {
      "id": "5c536647480a0e68ee000001",
      "subdomain": "instance"
    },
    "destination": {
      "id": "5f7c43e5480a0e75849d8bf0",
      "platform": "HTTP"
    }
  }
}
```

##### Response
Just like in the case of the `publish` action, the response must contain at least
the `external_id` and `published_url` attributes.

Failing to return these mandatory attributes, will result in an `Unexpected error`.

```json
{
  "metadata": {
    "external_id": "abc33",
    "published_url": "https://domain.com/content"
  }
}
```

#### Unhandled or unsupported "future" actions
New "actions" might be added in the future, therefore in order to _future
proof_ your integration, you should handle any _unknown_ actions by simply
returning an error with the HTTP status code set to 405.

```json
{
  "error": "Action is not supported by this endpoint"
}
```

Samples
-------
Kapost provides sample _HTTP endpoints_ usable by the HTTP destination in three
languages, namely: Ruby, Go and PHP. Samples in other languages might be added
later.

It is also possible to use automation services like Microsoft Flow, Dell Boomi,
Tray and more by leveraging so called _request triggers_ which provide an
HTTP endpoint that the Kapost HTTP destination can hit. This way one doesn't
have to write a single line of code, thus significantly reducing the barrier
of entry for building publish integrations.

To view the samples click [here][2].

License
-------
For more information see [LICENSE][1].

[1]: LICENSE
[2]: samples
[3]: https://en.wikipedia.org/wiki/Basic_access_authentication
[4]: https://swagger.io/docs/specification/authentication/bearer-authentication/
[modeline]: # ( vim: set ts=2 sw=2 sts=2 expandtab )
