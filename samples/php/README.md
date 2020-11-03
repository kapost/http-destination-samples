PHP Sample
==========
This directory contains a sample HTTP endpoint that the Kapost HTTP Destination
can connect to and use for publishing.

Getting Started
---------------
In order to start this sample, you'll need PHP 7.4.11 or later.

```bash
$ API_KEY=myapi_key SIGNATURE_SECRET=long_signature_secret php -S localhost:3342 -t .
```

This assumes that you configured the HTTP App in the Kapost App Center
with Bearer (API Key) authentication and a SHA256 signature.

| Name | Value |
| ---- | ----- |
| Authentication | Bearer (API Key) |
| API Key | myapi\_key |
| Signature | SHA256 |
| Signature Secret | long\_signature\_secret |
| Endpoint | Ngrok URL forwarding to port 3342 on localhost |

Do not forget to use a tool like [ngrok][1] to expose your local machine so that
Kapost can connect to it.

In a production environment, you would use Apache or Nginx and not expose the
_built-in PHP development server_ directly to the internet.

[1]: https://ngrok.com/
