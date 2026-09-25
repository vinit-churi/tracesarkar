# Demo script — v0.1, 26 September 2026

Fifteen minutes of setup, then a five-minute walkthrough. Written to be followed verbatim when
you are short of sleep.

---

## 1. Before anyone is watching

Two terminals. Terminal one, from the repository root:

```sh
set -a && . ./.env && set +a
go run ./cmd/api serve --addr :8080
```

Wait for `listening`. Confirm it:

```sh
curl -s http://localhost:8080/healthz
```

Terminal two:

```sh
cd app
flutter run -d chrome \
  --dart-define=API_BASE=http://localhost:8080 \
  --dart-define=DEMO_POSITION=19.2094,72.8348
```

**Why `DEMO_POSITION`:** a laptop in a meeting room often cannot get a GPS fix, and with no position
there is nothing to send. The build-time fixture keeps the demo moving. The screen labels it
"Fixture position" in amber and says it is worthless as evidence — which is the honest thing to say,
and is itself worth pointing at.

On your own machine, with location permission granted, drop that flag and the real fix appears with
its accuracy in metres.

Have a photograph of a pothole in `~/Downloads` before you start.

### If something is wrong

| Symptom | Cause | Fix |
|---|---|---|
| `AUTH_SECRET is not set` | The env file was not exported | `set -a && . ./.env && set +a` — the API reads these from the environment, not the file |
| Browser console shows a CORS failure | The origin is not allowed | Add it to `ALLOWED_ORIGINS` in `.env` and restart |
| Position never arrives, no fixture | Chrome or macOS is refusing location | Rebuild with `--dart-define=DEMO_POSITION=19.2094,72.8348` |
| Everything is broken | — | Fall back to `curl`; §4 stands on its own |

---

## 2. The five minutes

**Create an account.** "Create an account", a real address, a password of at least ten characters.
There are no composition rules: they push people toward predictable substitutions without adding
entropy.

**Sign out and back in with a wrong password.** The answer is *email or password is incorrect* —
identical to what an unregistered address gets. On a platform where people report against local
interests, the list of who is registered here is worth something, so it is not leaked through
differential error messages.

**Take a photograph, send it.** Note what the button does *not* say. Sending records the capture;
the line beneath says nothing is filed with any authority. That is hard rule 1 of the project, and
there is a test that fails if a future screen breaks it.

**Show the report id that comes back.** Then, in a third terminal, show it is real:

```sh
curl -s http://localhost:8080/v1/auth/me -H "Authorization: Bearer $TOKEN"
```

---

## 3. The two things worth saying out loud

**"The capture is durable before the API replies."** The `202` promises the capture is safe. If the
photograph were still only in memory, that would be a lie — so the image goes to object storage
*before* the response, and a failure to store it returns a `5xx` the client will retry rather than a
cheerful `202`. Losing a citizen's report is the one failure this system cannot have.

**"Retrying is safe, and that is where the bug was."** Post the same capture twice with the same
`Idempotency-Key` and you get the same report id and `created: false`. That already worked. What did
*not* work: the media row was inserted twice, so one photograph became two rows pointing at the same
object. It surfaced in an end-to-end run against the real database, was reproduced as a failing test
first, and the fix makes the content digest the identity (migration `0007`).

That is a better story than a clean demo: it shows the end-to-end test earning its keep.

---

## 4. If the app will not run

The backend alone demonstrates the whole path:

```sh
EMAIL="demo+$(date +%s)@example.org"

curl -s -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"correct-horse-battery\",\"name\":\"Demo\"}"

TOKEN=<the token from that response>

curl -s -X POST http://localhost:8080/v1/reports \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: demo-1" \
  -F 'meta={"location":{"lat":19.2094,"lon":72.8348,"accuracy_m":6.2},"captured_at":"2026-09-26T09:00:00+05:30"}' \
  -F 'role=close' -F 'media=@pothole.jpg;type=image/jpeg'
```

Run the second command twice: the same report id, `created: false`.

---

## 5. Questions you should expect

**"What actually works today?"** Collection, storage, and capture. Twice a day a Cloud Run job in
Mumbai archives BMC's road-works data — 4,673 work records so far, every change kept. The API takes
a capture and stores it durably. The client signs in and sends one. What does *not* exist yet:
classification, jurisdiction resolution, and the contract join. Say that plainly; the honest boundary
is more convincing than a vague claim.

**"Why Go and hand-written SQL?"** The queries are spatial and hand-tuned — PostGIS containment
against ward boundaries, buffered joins against road geometry. An ORM fights that, and the one thing
that must never be wrong is which authority a point falls in. [ADR 0005](../04-adr/0005-no-orm.md).

**"Why doesn't it file the complaint for you?"** Because a legal instrument submitted by a machine
is a liability with nobody's name on it. Everything is generated as a draft that a person reviews
and submits. [ADR 0007](../04-adr/0007-never-auto-file.md) and
[ADR 0014](../04-adr/0014-assisted-filing-not-automated-submission.md).

**"How do you know the contractor data is right?"** You do not assert it — you cite it. Every fact
about a named party carries a source reference and a retrieval timestamp, and an unsourced field is
not rendered at all. The archived copy of the source document is the evidentiary record, not the
parsed row.

**"Isn't this legally risky?"** The platform states facts and joins — this contract covers this
location, its defect liability period is active — and never concludes that anyone did wrong. Every
surface starts private and only goes public deliberately.

**"What was hard?"** Two honest answers. Accuracy was silently rounded from 6.2 m to 6 m by
`NULLIF($n, 0)` coercing a numeric to an integer — and accuracy is exactly what decides whether a
capture can be attributed confidently; a live test caught it. And BMC's portal geo-restricts by
geography rather than by network type, verified from three vantage points, which is why collection
runs from Mumbai rather than from a CI runner.
