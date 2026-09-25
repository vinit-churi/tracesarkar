# TraceSarkar client

One Flutter codebase for Android and the web. At v0.1 it does two things: it signs a person in, and
it sends a capture — a photograph plus the position it was taken at — to `POST /v1/reports`.

It deliberately does **not** file anything. Sending records the capture; any complaint, RTI or
tribunal application is a draft a person reviews and submits themselves
([hard rule 1](../CLAUDE.md#hard-rules)).

---

## Layout

| Path | What it is |
|---|---|
| `lib/api.dart` | The HTTP client. One type per endpoint the app uses; nothing else talks to the network |
| `lib/session_store.dart` | The signed-in session, kept across launches |
| `lib/theme.dart` | The tokens from [`11-screen-spec.md`](../docs/02-product/11-screen-spec.md) §2 |
| `lib/screens/sign_in_screen.dart` | Sign in / create account |
| `lib/screens/capture_screen.dart` | Position, photograph, send |
| `test/` | 17 tests — client, session, and the sign-in flow end to end |

## Running it

The client needs a backend. Start one from the repository root:

```sh
set -a && . ./.env && set +a      # API_TOKEN and AUTH_SECRET must be set
go run ./cmd/api serve --addr :8080
```

Then:

```sh
cd app
flutter run -d chrome --dart-define=API_BASE=http://localhost:8080
flutter run -d <android-device> --dart-define=API_BASE=http://<your-lan-ip>:8080
```

`API_BASE` defaults to `http://localhost:8080`, which is only right for the web build on the same
machine. An Android device cannot reach your laptop's `localhost`; use the LAN address.

The origin you serve the web client from must appear in the backend's `ALLOWED_ORIGINS`, or the
browser will refuse the call before it leaves the page.

## Tests

```sh
flutter test        # 17 tests, no network, no backend needed
flutter analyze
```

The tests inject a fake `http.Client`, so they exercise the real request construction — headers,
multipart shape, idempotency key — without a server. Two of them encode hard rules rather than
mechanics: one asserts the capture screen never offers to file anything, and one asserts signing out
actually removes the token rather than merely navigating away.

## Building for the web

```sh
flutter build web --release --dart-define=API_BASE=https://<your-api-host>
```

Output is `build/web`. For Cloudflare Pages, that directory is the publish directory; there is no
build command to run on Cloudflare's side if you upload the built output.

Geolocation requires a secure context. `https://` and `http://localhost` work; a plain-HTTP LAN
address does not, and the position will simply never arrive.

### A position for machines that cannot get one

A laptop in a meeting room often cannot get a fix, and with no position there is nothing to send:

```sh
flutter run -d chrome --dart-define=DEMO_POSITION=19.2094,72.8348
```

The build then never asks the device. The screen labels the position "Fixture position" in amber and
says it is worthless as evidence, because a position that was typed in must never be mistaken for
one that was measured. Leave the flag out and the app uses the device, as it does in the field.

## What is not here yet

- **Google sign-in.** The backend endpoint `POST /v1/auth/google` is implemented and tested, and
  `ApiClient.signInWithGoogle` is written against it. The button is not wired, because the web flow
  needs `google_sign_in_web`'s rendered button rather than a plain `signIn()` call, and that was not
  worth risking on the night before a demo. Wiring it needs a Web OAuth client ID from the GCP
  console in both `GOOGLE_CLIENT_ID` (backend) and the web client.
- **History.** The app shows what it sent during the current session only; `GET /v1/reports/{id}`
  and a list endpoint do not exist yet.
- **Offline queue.** The field kit at `internal/api/fieldkit/` queues captures in IndexedDB and
  retries them when the connection returns. The Flutter client does not yet; a failed send reports
  the failure and keeps the photograph on screen so it can be retried by hand.
- **Redaction.** Faces and number plates are not yet redacted, so this build writes no public
  derivative at all ([hard rule 5](../CLAUDE.md#hard-rules)).
