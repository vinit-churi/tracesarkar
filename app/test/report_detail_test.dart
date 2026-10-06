import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/screens/reports_screen.dart';

// A one-pixel PNG. Enough for Image.memory to decode.
final kPng = Uint8List.fromList(base64Decode(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=='));

Map<String, dynamic> reportJson({bool enriched = true}) => {
      'id': 'r1',
      'status': 'enriched',
      'created_at': '2026-10-05T08:00:00Z',
      'lat': 19.234093,
      'lon': 72.844189,
      'accuracy_m': 6.4,
      if (enriched) ...{
        'classification': {
          'category': 'waste',
          'subcategory': 'illegal dumping',
          'outcome': 'accepted',
          'confidence': 0.91,
        },
        'ward': 'R/C',
        'authority': 'BMC',
      },
    };

void main() {
  test('the client keeps the position the server sent', () {
    final d = ReportDetail.fromJson(reportJson());
    expect(d.latitude, closeTo(19.234093, 0.000001));
    expect(d.longitude, closeTo(72.844189, 0.000001));
    expect(d.accuracyMetres, closeTo(6.4, 0.01));
    expect(d.hasLocation, isTrue);
  });

  test('a report without a position says so rather than claiming 0,0', () {
    final d = ReportDetail.fromJson({'id': 'r1', 'status': 'pending'});
    expect(d.hasLocation, isFalse,
        reason: 'null island is a real place and no capture was taken there');
  });

  testWidgets('opening a report shows the photograph and where it was taken',
      (tester) async {
    var askedForMedia = 0;
    final client = ApiClient(
      baseUrl: 'https://api.test',
      client: MockClient((request) async {
        if (request.url.path.endsWith('/media')) {
          askedForMedia++;
          return http.Response.bytes(kPng, 200,
              headers: {'content-type': 'image/png'});
        }
        return http.Response(
            jsonEncode({'report': reportJson()}), 200,
            headers: {'content-type': 'application/json'});
      }),
    );

    await tester.pumpWidget(MaterialApp(
      home: ReportScreen(
        client: client,
        session: Session(
          token: 'tok',
          account: Account(id: 'a1', email: 'a@b.org', name: 'A'),
        ),
        reportId: 'r1',
      ),
    ));
    await tester.pumpAndSettle();

    // The photograph is the report. Opening one and not seeing it is the
    // thing a person notices first.
    expect(askedForMedia, 1, reason: 'the photograph was never fetched');
    expect(find.byType(Image), findsWidgets);

    // And where it was taken, with how far out the fix may be — the accuracy
    // is what every confidence gate downstream is reasoning about, so hiding
    // it makes the verdicts unreadable.
    expect(find.textContaining('19.2340'), findsOneWidget);
    expect(find.textContaining('72.8441'), findsOneWidget);
    expect(find.textContaining('6 m'), findsOneWidget);
  });
}
