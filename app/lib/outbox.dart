import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:path_provider/path_provider.dart';

import 'api.dart';

/// What one flush did.
class FlushResult {
  const FlushResult({this.sent = 0, this.failed = 0, this.rejected = 0});

  /// Accepted by the server and removed.
  final int sent;

  /// Could not be reached. Still queued, to try again.
  final int failed;

  /// Refused by the server and removed, because retrying cannot help.
  final int rejected;

  bool get quiet => sent == 0 && failed == 0 && rejected == 0;
}

/// One capture waiting to be sent.
class Queued {
  Queued({required this.id, required this.capture, required this.queuedAt});
  final String id;
  final Capture capture;
  final DateTime queuedAt;
}

/// Somewhere a capture can wait until it can be sent.
///
/// An interface because a widget test cannot touch the disk: `pumpAndSettle`
/// runs in a fake-async zone where a real file operation's future never
/// completes, so a screen tested against the real one deadlocks rather than
/// fails. [Outbox] is the real implementation and has its own tests.
abstract class CaptureStore {
  Future<void> add(Capture capture);
  Future<int> count();
  Future<FlushResult> flush(Future<String> Function(Capture) send);
}

/// Captures that could not be sent yet, kept on disk until they are.
///
/// A walk through Borivali goes in and out of signal, and an upload is a few
/// megabytes over whatever is left of it. Holding a failed capture in memory
/// loses it the moment the person takes the next photograph, backgrounds the
/// app, or the system reclaims it — and unlike every other kind of lost
/// request, a photograph of a pothole cannot be taken again later. The walk is
/// over.
///
/// So a capture that cannot be sent is written down, with the position and time
/// it was actually taken, and sent when there is signal.
class Outbox implements CaptureStore {
  Outbox({Directory? directory}) : _given = directory;

  final Directory? _given;
  Directory? _dir;

  Future<Directory> _directory() async {
    if (_given != null) return _given;
    if (_dir != null) return _dir!;
    final base = await getApplicationDocumentsDirectory();
    final dir = Directory('${base.path}/outbox');
    if (!await dir.exists()) await dir.create(recursive: true);
    return _dir = dir;
  }

  /// Writes a capture down. The photograph goes to its own file: a few
  /// megabytes of base64 inside a JSON document is slower to write and far
  /// more likely to be truncated by the app being killed mid-write.
  @override
  Future<void> add(Capture capture) async {
    final dir = await _directory();
    final id = '${DateTime.now().microsecondsSinceEpoch}';

    await File('${dir.path}/$id.jpg').writeAsBytes(capture.bytes, flush: true);

    final meta = <String, dynamic>{
      'id': id,
      'queued_at': DateTime.now().toUtc().toIso8601String(),
      'filename': capture.filename,
      'latitude': capture.latitude,
      'longitude': capture.longitude,
      'accuracy_m': capture.accuracyMetres,
      // The time it was taken, never the time it was sent. A capture uploaded
      // on the bus home must still say it was photographed on the street.
      'captured_at': capture.capturedAt.toUtc().toIso8601String(),
      'role': capture.role,
      'notes': capture.notes,
    };
    // The JSON is written last. Until it exists the entry is not read, so a
    // kill between the two leaves a stray image rather than a half-entry.
    await File('${dir.path}/$id.json')
        .writeAsString(jsonEncode(meta), flush: true);
  }

  /// Everything waiting, oldest first — the order the walk happened in.
  Future<List<Queued>> pending() async {
    final dir = await _directory();
    if (!await dir.exists()) return [];

    final entries = <Queued>[];
    final files = dir
        .listSync()
        .whereType<File>()
        .where((f) => f.path.endsWith('.json'))
        .toList()
      ..sort((a, b) => a.path.compareTo(b.path));

    for (final file in files) {
      try {
        final meta = jsonDecode(await file.readAsString()) as Map<String, dynamic>;
        final id = meta['id'] as String;
        final image = File('${dir.path}/$id.jpg');
        if (!await image.exists()) {
          // Metadata without its photograph is not a capture. Drop it rather
          // than carry an entry that can never be sent.
          await file.delete();
          continue;
        }
        entries.add(Queued(
          id: id,
          queuedAt: DateTime.parse(meta['queued_at'] as String),
          capture: Capture(
            bytes: Uint8List.fromList(await image.readAsBytes()),
            filename: meta['filename'] as String,
            latitude: (meta['latitude'] as num).toDouble(),
            longitude: (meta['longitude'] as num).toDouble(),
            accuracyMetres: (meta['accuracy_m'] as num).toDouble(),
            capturedAt: DateTime.parse(meta['captured_at'] as String),
            role: (meta['role'] as String?) ?? 'close',
            notes: meta['notes'] as String?,
          ),
        ));
      } catch (_) {
        // A truncated or unreadable entry must not take the rest of the queue
        // with it. The app being killed mid-write is the normal case.
        continue;
      }
    }
    return entries;
  }

  /// How many captures are waiting, without reading the photographs back.
  @override
  Future<int> count() async {
    final dir = await _directory();
    if (!await dir.exists()) return 0;
    return dir
        .listSync()
        .whereType<File>()
        .where((f) => f.path.endsWith('.json'))
        .length;
  }

  /// Tries to send everything waiting.
  ///
  /// A capture the server refuses outright is removed: a 4xx will be a 4xx
  /// every time, and retrying it on every app launch blocks the queue behind it
  /// for no gain. A capture that could not reach the server stays.
  @override
  Future<FlushResult> flush(Future<String> Function(Capture) send) async {
    final waiting = await pending();
    if (waiting.isEmpty) return const FlushResult();

    var sent = 0, failed = 0, rejected = 0;
    for (final item in waiting) {
      try {
        await send(item.capture);
        await _remove(item.id);
        sent++;
      } on ApiException catch (e) {
        if (e.statusCode >= 400 && e.statusCode < 500) {
          await _remove(item.id);
          rejected++;
        } else {
          failed++;
        }
      } catch (_) {
        failed++;
      }
    }
    return FlushResult(sent: sent, failed: failed, rejected: rejected);
  }

  Future<void> _remove(String id) async {
    final dir = await _directory();
    for (final path in ['${dir.path}/$id.json', '${dir.path}/$id.jpg']) {
      final file = File(path);
      if (await file.exists()) await file.delete();
    }
  }
}
