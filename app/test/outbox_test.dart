import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/outbox.dart';

Capture capture({String name = 'a.jpg', int size = 32}) => Capture(
      bytes: Uint8List.fromList(List.filled(size, 7)),
      filename: name,
      latitude: 19.2340,
      longitude: 72.8444,
      accuracyMetres: 6,
      capturedAt: DateTime.utc(2026, 10, 4, 9, 30),
      notes: 'near the corner',
    );

void main() {
  late Directory dir;
  late Outbox outbox;

  setUp(() async {
    dir = await Directory.systemTemp.createTemp('outbox');
    outbox = Outbox(directory: dir);
  });
  tearDown(() async => dir.delete(recursive: true));

  test('a capture that could not be sent is kept on disk', () async {
    await outbox.add(capture());

    // On disk, not in memory: a walk means the app is backgrounded, killed by
    // the system, and reopened. Anything held in RAM is gone by then, and the
    // photograph cannot be taken again.
    final reopened = Outbox(directory: dir);
    final waiting = await reopened.pending();

    expect(waiting, hasLength(1));
    expect(waiting.first.capture.latitude, 19.2340);
    expect(waiting.first.capture.accuracyMetres, 6);
    expect(waiting.first.capture.notes, 'near the corner');
    // The photograph itself has to survive, not just its metadata.
    expect(waiting.first.capture.bytes, hasLength(32));
  });

  test('the position and time are the ones from the field, not from the retry',
      () async {
    await outbox.add(capture());
    final waiting = await Outbox(directory: dir).pending();

    // A capture sent hours later must still say where and when it was taken.
    // Re-stamping it on upload would put every pothole at the bus stop home.
    expect(waiting.first.capture.capturedAt.toUtc(),
        DateTime.utc(2026, 10, 4, 9, 30));
  });

  test('a sent capture leaves the outbox', () async {
    await outbox.add(capture(name: 'a.jpg'));
    await outbox.add(capture(name: 'b.jpg'));

    var sent = 0;
    final result = await outbox.flush((c) async {
      sent++;
      return 'report-$sent';
    });

    expect(sent, 2);
    expect(result.sent, 2);
    expect(result.failed, 0);
    expect(await outbox.pending(), isEmpty);
    // The files must go too, or a day of walking fills the phone.
    expect(dir.listSync().whereType<File>().where(
        (f) => f.path.endsWith('.jpg')), isEmpty);
  });

  test('a capture that fails again stays, and the rest still go', () async {
    await outbox.add(capture(name: 'bad.jpg'));
    await outbox.add(capture(name: 'good.jpg'));

    final result = await outbox.flush((c) async {
      if (c.filename == 'bad.jpg') throw Exception('still no signal');
      return 'report-1';
    });

    expect(result.sent, 1);
    expect(result.failed, 1);
    final left = await outbox.pending();
    expect(left, hasLength(1));
    expect(left.first.capture.filename, 'bad.jpg');
  });

  test('the server rejecting a capture does not retry it forever', () async {
    await outbox.add(capture());

    // 4xx means this capture will never be accepted. Retrying it every time
    // the app opens blocks the queue behind it and achieves nothing.
    final result = await outbox.flush((c) async {
      throw ApiException(400, 'that capture is not acceptable');
    });

    expect(result.sent, 0);
    expect(result.rejected, 1);
    expect(await outbox.pending(), isEmpty,
        reason: 'a permanently rejected capture must not block the queue');
  });

  test('oldest first, so the walk uploads in the order it happened', () async {
    await outbox.add(capture(name: 'first.jpg'));
    await Future<void>.delayed(const Duration(milliseconds: 5));
    await outbox.add(capture(name: 'second.jpg'));

    final order = <String>[];
    await outbox.flush((c) async {
      order.add(c.filename);
      return 'r';
    });

    expect(order, ['first.jpg', 'second.jpg']);
  });

  test('a half-written entry does not break the queue', () async {
    await outbox.add(capture(name: 'good.jpg'));
    // The app being killed mid-write is the normal case, not the exotic one.
    File('${dir.path}/truncated.json').writeAsStringSync('{"not": ');

    final waiting = await outbox.pending();
    expect(waiting, hasLength(1));
    expect(waiting.first.capture.filename, 'good.jpg');
  });
}
