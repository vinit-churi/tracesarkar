import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/outbox.dart';
import 'package:tracesarkar_app/screens/capture_screen.dart';

/// In memory, because a widget test cannot touch the disk — pumpAndSettle runs
/// in a fake-async zone where a real file operation never completes. What the
/// real outbox does with the disk is covered in outbox_test.dart.
class FakeOutbox implements CaptureStore {
  final List<Capture> held = [];
  int flushes = 0;
  bool finished = false;

  @override
  Future<void> add(Capture capture) async => held.add(capture);

  @override
  Future<int> count() async => held.length;

  @override
  Future<FlushResult> flush(Future<String> Function(Capture) send) async {
    flushes++;
    finished = false;
    var sent = 0, failed = 0;
    for (final c in List.of(held)) {
      try {
        await send(c);
        held.remove(c);
        sent++;
      } catch (_) {
        failed++;
      }
    }
    finished = true;
    return FlushResult(sent: sent, failed: failed);
  }
}

Capture aCapture() => Capture(
      bytes: Uint8List.fromList(List.filled(8, 1)),
      filename: 'a.jpg',
      latitude: 19.23,
      longitude: 72.84,
      accuracyMetres: 6,
      capturedAt: DateTime.utc(2026, 10, 4),
    );

Future<void> pumpCapture(
  WidgetTester tester, {
  required http.Client client,
  required CaptureStore outbox,
}) async {
  await tester.pumpWidget(MaterialApp(
    home: CaptureScreen(
      client: ApiClient(baseUrl: 'https://api.test', client: client),
      session: Session(
        token: 'tok',
        account: Account(id: 'a1', email: 'a@b.org', name: 'A'),
      ),
      outbox: outbox,
      onSignOut: () async {},
    ),
  ));
  await tester.pumpAndSettle();
}

void main() {
  setUp(() => SharedPreferences.setMockInitialValues({}));

  testWidgets('captures still waiting are shown, not hidden', (tester) async {
    // Someone who has walked for three hours has to be able to see that
    // nothing is stuck. A queue that drains silently is indistinguishable from
    // one that has lost the photographs.
    final outbox = FakeOutbox()..held.add(aCapture());

    await pumpCapture(
      tester,
      client: MockClient((_) async => http.Response('{}', 503)),
      outbox: outbox,
    );
    await tester.pumpAndSettle();

    expect(outbox.held, hasLength(1), reason: 'a failed send stays queued');

    // The screen is a ListView, so anything below the fold has not been built
    // and cannot be found until it is scrolled to. It sits under the send
    // button, which is where someone reaching for it will read it.
    await tester.scrollUntilVisible(
      find.textContaining('waiting to send'),
      200,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.textContaining('waiting to send'), findsOneWidget);
    expect(find.text('Send now'), findsOneWidget);
  });

  testWidgets('opening the screen sends what is waiting', (tester) async {
    final outbox = FakeOutbox()..held.add(aCapture());
    var posted = 0;

    await pumpCapture(
      tester,
      client: MockClient((_) async {
        posted++;
        return http.Response(jsonEncode({'report_id': 'r1'}), 202);
      }),
      outbox: outbox,
    );
    await tester.pumpAndSettle();

    expect(outbox.flushes, 1, reason: 'the queue should drain on open');
    expect(posted, 1);
    expect(outbox.held, isEmpty);
    // Nothing queued, so nothing to say about it — even after scrolling.
    await tester.drag(find.byType(Scrollable).first, const Offset(0, -600));
    await tester.pumpAndSettle();
    expect(find.textContaining('waiting to send'), findsNothing);
  });
}
