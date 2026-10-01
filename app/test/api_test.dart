import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:tracesarkar_app/api.dart';

Capture sampleCapture() => Capture(
      bytes: Uint8List.fromList([1, 2, 3]),
      filename: 'close.jpg',
      latitude: 19.2094,
      longitude: 72.8348,
      accuracyMetres: 6.2,
      capturedAt: DateTime.utc(2026, 9, 25, 8, 14, 22),
    );

void main() {
  test('login sends the credentials and returns the session', () async {
    late http.Request seen;
    final client = ApiClient(
      baseUrl: 'https://api.test',
      client: MockClient((request) async {
        seen = request;
        return http.Response(
          jsonEncode({
            'token': 'tok-1',
            'account': {'id': 'acct-1', 'email': 'a@b.org', 'name': 'A'},
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      }),
    );

    final session = await client.login(email: 'a@b.org', password: 'secret');

    expect(seen.url.toString(), 'https://api.test/v1/auth/login');
    expect(jsonDecode(seen.body), {'email': 'a@b.org', 'password': 'secret'});
    expect(session.token, 'tok-1');
    expect(session.account.email, 'a@b.org');
  });

  test('a rejected sign-in surfaces the server message, not a status code',
      () async {
    final client = ApiClient(
      baseUrl: 'https://api.test',
      client: MockClient((_) async => http.Response(
            jsonEncode({
              'error': {'message': 'email or password is incorrect'}
            }),
            401,
          )),
    );

    expect(
      () => client.login(email: 'a@b.org', password: 'wrong'),
      throwsA(isA<ApiException>().having(
        (e) => e.message,
        'message',
        'email or password is incorrect',
      )),
    );
  });

  test('a non-JSON failure still raises a readable error', () async {
    final client = ApiClient(
      baseUrl: 'https://api.test',
      client: MockClient((_) async => http.Response('<html>502</html>', 502)),
    );

    expect(
      () => client.login(email: 'a@b.org', password: 'x'),
      throwsA(isA<ApiException>().having((e) => e.statusCode, 'status', 502)),
    );
  });

  test('me reads the account behind a bearer token', () async {
    late http.BaseRequest seen;
    final client = ApiClient(
      baseUrl: 'https://api.test',
      client: MockClient((request) async {
        seen = request;
        return http.Response(
          jsonEncode({
            'account': {'id': 'acct-1', 'email': 'a@b.org', 'name': 'A'}
          }),
          200,
        );
      }),
    );

    final account = await client.me('tok-1');

    expect(seen.headers['Authorization'], 'Bearer tok-1');
    expect(account.id, 'acct-1');
  });

  test('a capture carries coordinates and accuracy the server can route on',
      () {
    final meta = sampleCapture().toMeta();
    final location = meta['location'] as Map<String, dynamic>;

    expect(location['lat'], 19.2094);
    expect(location['lon'], 72.8348);
    // Accuracy decides whether a capture can be attributed to a contract at
    // all; rounding it away would silently degrade every match downstream.
    expect(location['accuracy_m'], 6.2);
    expect(meta['captured_at'], '2026-09-25T08:14:22.000Z');
  });

  test('a capture with no label omits the field kit block', () {
    expect(sampleCapture().toMeta().containsKey('label'), isFalse);
  });

  test('submit posts multipart with the photograph and the bearer token',
      () async {
    late http.BaseRequest seen;
    late String body;
    final client = ApiClient(
      baseUrl: 'https://api.test',
      client: MockClient((request) async {
        seen = request;
        body = request.body;
        return http.Response(
          jsonEncode({'report_id': 'rep-1', 'status': 'pending'}),
          202,
        );
      }),
    );

    final id = await client.submit(sampleCapture(), 'tok-1');

    expect(id, 'rep-1');
    expect(seen.url.toString(), 'https://api.test/v1/reports');
    expect(seen.headers['Authorization'], 'Bearer tok-1');
    expect(seen.headers['Idempotency-Key'], isNotEmpty);
    expect(body, contains('name="meta"'));
    expect(body, contains('name="media"'));
    expect(body, contains('filename="close.jpg"'));
  });

  test('two submissions of the same capture reuse one idempotency key',
      () async {
    final keys = <String>[];
    final client = ApiClient(
      baseUrl: 'https://api.test',
      client: MockClient((request) async {
        keys.add(request.headers['Idempotency-Key'] ?? '');
        return http.Response(jsonEncode({'report_id': 'rep-1'}), 202);
      }),
    );

    // A retry after a dropped connection must not create a second report.
    final capture = sampleCapture();
    await client.submit(capture, 'tok-1');
    await client.submit(capture, 'tok-1');

    expect(keys.first, keys.last);
  });

  test('a report that has not been enriched yet is still a report', () {
    // Enrichment runs after the photograph is safe, so a capture can exist
    // without a verdict. The app must render that, not treat it as broken.
    final d = ReportDetail.fromJson({'id': 'r1', 'status': 'pending'});

    expect(d.id, 'r1');
    expect(d.enriched, isFalse);
    expect(d.summary, 'Working on it…');
  });

  test('a fully enriched report reads back whole', () {
    final d = ReportDetail.fromJson({
      'id': 'r1',
      'status': 'pending',
      'classification': {
        'category': 'road_defect',
        'subcategory': 'pothole',
        'outcome': 'accepted',
        'confidence': 0.91,
        'rationale': 'Large depression in asphalt',
      },
      'ward': 'R/C',
      'authority': 'BMC',
      'contractor_name': 'M/s Example Infracon Pvt. Ltd',
      'road_name': 'Jain Mandir Road',
      'contract_source': 'bmc_roads_api',
    });

    expect(d.enriched, isTrue);
    expect(d.summary, 'pothole');
    expect(d.ward, 'R/C');
    expect(d.contractorName, 'M/s Example Infracon Pvt. Ltd');
  });

  test('reports are read from the list endpoint', () async {
    final client = ApiClient(
      baseUrl: 'https://api.test',
      client: MockClient((request) async {
        expect(request.headers['Authorization'], 'Bearer tok-1');
        return http.Response(
          jsonEncode({
            'reports': [
              {'id': 'r1', 'status': 'pending'},
              {'id': 'r2', 'status': 'pending'},
            ]
          }),
          200,
        );
      }),
    );

    final list = await client.reports('tok-1');

    expect(list.length, 2);
    expect(list.first.id, 'r1');
  });
}
