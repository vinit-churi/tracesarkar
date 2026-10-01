# Releasing the app

How a build reaches a phone, and the two things that will break it.

---

## The short version

```
git tag v0.1.1 && git push origin v0.1.1
```

That runs [`.github/workflows/android.yml`](../../.github/workflows/android.yml), which builds a
signed APK, uploads it to the public R2 downloads bucket, writes `android/latest.json`, and cuts a
GitHub release. The page at **<https://tracesarkar-app.infoyantra.workers.dev/download/>** reads
that JSON, so it starts offering the new build within a minute. Nothing else needs redeploying.

The web client is separate and is deployed by hand:

```
make app-deploy
```

---

## The two things that break it

### 1. Signing with the wrong key

The APK is signed with `app/android/tracesarkar-release.jks`. Its SHA-1 is registered with Google
as an Android OAuth client, and Android refuses to install an update signed by a different key.

So a debug-signed build is worse than no build: Google sign-in fails, and anyone who already has
the app cannot install over it. The workflow therefore **fails rather than publishes** if the
keystore secret is missing, and verifies the certificate subject after building.

**The keystore is not in the repository and cannot be regenerated.** Losing it means a new package
identity and every existing install stranded. Back it up somewhere off this machine.

| Fact | Value |
|---|---|
| Package | `org.tracesarkar.app` |
| Key alias | `tracesarkar` |
| SHA-1 | `30:A3:36:B2:D5:03:17:66:4A:90:9C:80:22:36:47:4E:7D:CD:BB:D2` |
| SHA-256 | `F8:2B:75:31:4B:60:18:25:00:79:62:24:4C:64:1E:9E:E4:ED:6B:F1:BF:67:8F:07:5A:81:00:C9:DC:0E:7F:15` |

### 2. Building without `API_BASE`

The server address is compiled in through `--dart-define`. A build without it points at
`localhost:8080` and fails silently on a phone — no error, just nothing working. `make app-apk`
and the workflow both refuse to run when it is unset.

---

## What has to exist in GitHub

**Secrets** (Settings → Secrets and variables → Actions → Secrets):

| Name | What |
|---|---|
| `ANDROID_KEYSTORE_BASE64` | `base64 -i app/android/tracesarkar-release.jks` |
| `ANDROID_KEYSTORE_PASSWORD` | From `.env` |
| `ANDROID_KEY_ALIAS` | `tracesarkar` |
| `R2_DOWNLOADS_ACCESS_KEY` | R2 API token, scoped to `tracesarkar-downloads`, object read+write |
| `R2_DOWNLOADS_SECRET_KEY` | Its secret |
| `R2_ENDPOINT` | `https://dba7a9b5ed108516b714250de2a0dd0f.r2.cloudflarestorage.com` |

**Variables** (same page → Variables) — these are not secret and are easier to read in logs:

| Name | What |
|---|---|
| `API_BASE` | The Cloud Run URL |
| `GOOGLE_CLIENT_ID` | The OAuth web client id |

Scope the R2 token to `tracesarkar-downloads` only. The archive bucket holds citizen photographs
and a release workflow has no business being able to read it.

---

## The two buckets, and why they are two

| Bucket | Public | Holds |
|---|---|---|
| `tracesarkar` | **No** | Raw source artefacts and citizen photographs — the evidentiary archive |
| `tracesarkar-downloads` | Yes | APKs and `latest.json` |

They are separate because public access on R2 is per bucket, not per prefix. Putting the APK in
the archive bucket would mean making the archive public, and the archive contains photographs
taken by people who did not agree to publish them. See hard rules 4 and 5 in
[`CLAUDE.md`](../../CLAUDE.md).

CORS for the downloads bucket lives in
[`infra/r2/downloads-cors.json`](../../infra/r2/downloads-cors.json) and is applied with:

```
cd app && npx wrangler r2 bucket cors set tracesarkar-downloads \
  --file ../infra/r2/downloads-cors.json
```

Without it the download page cannot read `latest.json` and says no build is published.

---

## Why `latest.json` is written last

The page reads `latest.json` to find the APK. If the JSON were written first, every visitor
between the two uploads would get a link that 404s. Written last, the worst case is that the page
offers the previous build for a few seconds longer.

---

## Checking a build is what you think it is

```
apksigner verify --print-certs tracesarkar-0.1.0.apk   # must say CN=TraceSarkar
shasum -a 256 tracesarkar-0.1.0.apk                    # must match latest.json
```
