import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/screens/reports_screen.dart';

// A one-pixel PNG, so the photograph resolves and the spinner stops; an
// indeterminate progress indicator never lets pumpAndSettle finish.
final kPng = Uint8List.fromList(base64Decode(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=='));

Map<String, dynamic> withNextStep({bool available = true}) => {
      'id': 'r1',
      'status': 'enriched',
      'created_at': '2026-10-05T08:00:00Z',
      'lat': 19.234093,
      'lon': 72.844189,
      'classification': {'category': 'waste', 'subcategory': 'illegal dumping'},
      'ward': 'R/C',
      'next_step': {
        'available': available,
        if (available) ...{
          'text': 'To: Assistant Engineer (SWM) R/Central\n\nThere is illegal dumping…',
          'channels': [
            {'name': 'MyBMC MARG', 'how': 'Gives a reference number.'},
            {'name': '1916', 'how': 'Ask for the complaint number.'},
          ],
          'deadline_hours': 24,
          'deadline_citation': 'R/Central SWM RTI handbook §4(1)(b)(iii)',
        },
        'what_happens_next': available
            ? 'Send this yourself. Once BMC has been told, it has 24 hours.'
            : 'TraceSarkar does not yet have the department responsible.',
      },
    };

Future<void> pump(WidgetTester tester, Map<String, dynamic> json) async {
  await tester.pumpWidget(MaterialApp(
    home: ReportScreen(
      client: ApiClient(
        baseUrl: 'https://api.test',
        client: MockClient((request) async {
          if (request.url.path.endsWith('/media')) {
            return http.Response.bytes(kPng, 200,
                headers: {'content-type': 'image/png'});
          }
          return http.Response(jsonEncode({'report': json}), 200,
              headers: {'content-type': 'application/json'});
        }),
      ),
      session: Session(
        token: 'tok',
        account: Account(id: 'a1', email: 'a@b.org', name: 'A'),
      ),
      reportId: 'r1',
    ),
  ));
  // Pumped rather than settled: the screen shows an indeterminate progress
  // indicator while the photograph loads, and pumpAndSettle never returns
  // while one is on screen.
  for (var i = 0; i < 6; i++) {
    await tester.pump(const Duration(milliseconds: 50));
  }
}

void main() {
  testWidgets('a report says what to do, not just what it is', (tester) async {
    await pump(tester, withNextStep());

    await tester.dragUntilVisible(
      find.textContaining('What to do'),
      find.byType(Scrollable).first,
      const Offset(0, -200),
    );
    await tester.pump();
    expect(find.textContaining('What to do'), findsOneWidget);
    expect(find.textContaining('MyBMC MARG'), findsOneWidget);
    expect(find.textContaining('24 hours'), findsOneWidget);
    // The citation travels with the deadline (hard rule 2).
    expect(find.textContaining('SWM RTI handbook'), findsOneWidget);
    // The text is copied, never sent (hard rule 1).
    expect(find.textContaining('Copy'), findsOneWidget);
    for (final forbidden in ['File complaint', 'Send to BMC', 'Submit']) {
      expect(find.text(forbidden), findsNothing);
    }
  });

  testWidgets('with no department it says so rather than inventing a desk',
      (tester) async {
    await pump(tester, withNextStep(available: false));

    await tester.dragUntilVisible(
      find.textContaining('does not yet have the department'),
      find.byType(Scrollable).first,
      const Offset(0, -200),
    );
    await tester.pump();
    expect(find.textContaining('does not yet have the department'), findsOneWidget);
    expect(find.textContaining('Copy'), findsNothing);
  });
}
