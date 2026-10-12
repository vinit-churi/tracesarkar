import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/screens/reports_screen.dart';

final kPng = Uint8List.fromList(base64Decode(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=='));

Map<String, dynamic> report({int sightings = 1, bool lodged = false}) => {
      'id': 'r1',
      'status': 'enriched',
      'created_at': '2026-10-12T08:00:00Z',
      'lat': 19.234, 'lon': 72.844,
      'classification': {'category': 'road_defect', 'subcategory': 'pothole'},
      'ward': 'R/C',
      'issue': {
        'issue_id': 'i1',
        'status': 'open',
        'sightings': sightings,
        if (lodged) ...{
          'reference': 'MARG-2026-1012-0001',
          'notified_at': '2026-10-12T09:00:00Z',
        },
      },
    };

Future<void> pump(WidgetTester tester, Map<String, dynamic> json,
    {void Function(String)? onLodge}) async {
  await tester.pumpWidget(MaterialApp(
    home: ReportScreen(
      client: ApiClient(
        baseUrl: 'https://api.test',
        client: MockClient((request) async {
          if (request.url.path.endsWith('/media')) {
            return http.Response.bytes(kPng, 200,
                headers: {'content-type': 'image/png'});
          }
          if (request.url.path.endsWith('/lodged')) {
            onLodge?.call(request.body);
            return http.Response('{"lodged":true}', 200,
                headers: {'content-type': 'application/json'});
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
  for (var i = 0; i < 6; i++) {
    await tester.pump(const Duration(milliseconds: 50));
  }
}

Future<void> scrollTo(WidgetTester tester, Finder f) async {
  await tester.dragUntilVisible(
      f, find.byType(Scrollable).first, const Offset(0, -200));
  await tester.pump();
}

void main() {
  testWidgets('a second sighting is shown as corroboration, not a duplicate',
      (tester) async {
    await pump(tester, report(sightings: 3));
    await scrollTo(tester, find.textContaining('3 people'));
    expect(find.textContaining('3 people'), findsOneWidget);
  });

  testWidgets('one sighting does not claim company it does not have',
      (tester) async {
    await pump(tester, report(sightings: 1));
    // "1 people have reported this" is worse than saying nothing.
    expect(find.textContaining('1 people'), findsNothing);
  });

  testWidgets('before it is lodged, no clock is running and it says what would start one',
      (tester) async {
    await pump(tester, report());
    await scrollTo(tester, find.textContaining('Nothing is running yet'));

    expect(find.textContaining('Nothing is running yet'), findsOneWidget);
    // It must not imply a deadline is already ticking.
    expect(find.textContaining('left of'), findsNothing);
    expect(find.textContaining('overdue'), findsNothing);
    // And it must offer the thing that starts it.
    expect(find.textContaining('reference'), findsWidgets);
  });

  testWidgets('recording the reference sends it', (tester) async {
    String? sent;
    await pump(tester, report(), onLodge: (b) => sent = b);

    // Scrolled to the button first: the field is inside a lazy ListView and
    // is not built until it is on screen, so a finder for it throws before
    // the scroll rather than after.
    await scrollTo(tester, find.widgetWithText(FilledButton, 'I have lodged it'));
    await tester.enterText(
        find.widgetWithText(TextField, 'Complaint reference, once you have lodged it'),
        'MARG-2026-1012-0001');
    await tester.pump();
    await tester.tap(find.widgetWithText(FilledButton, 'I have lodged it'));
    for (var i = 0; i < 6; i++) {
      await tester.pump(const Duration(milliseconds: 50));
    }

    expect(sent, isNotNull);
    expect(sent, contains('MARG-2026-1012-0001'));
  });
}
