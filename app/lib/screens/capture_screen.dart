import 'dart:async';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:image_picker/image_picker.dart';

import '../api.dart';
import '../outbox.dart';
import '../fixture_position.dart';
import '../theme.dart';
import '../widgets/verdict.dart';
import 'reports_screen.dart';

/// A capture that has reached the server, kept for this session so the person
/// can see what they sent.
class _Sent {
  _Sent(this.reportId, this.at);
  final String reportId;
  final DateTime at;
}

/// Where a capture was taken, and how that was established. A measured fix and
/// a build-time fixture are both positions, but they are not the same claim, so
/// the difference is carried here rather than inferred later.
class _Fix {
  _Fix({
    required this.latitude,
    required this.longitude,
    required this.accuracyMetres,
    required this.at,
    required this.isFixture,
  });

  factory _Fix.measured(Position p) => _Fix(
        latitude: p.latitude,
        longitude: p.longitude,
        accuracyMetres: (p.accuracy * 10).roundToDouble() / 10,
        at: p.timestamp,
        isFixture: false,
      );

  factory _Fix.fixture(FixturePosition f) => _Fix(
        latitude: f.latitude,
        longitude: f.longitude,
        accuracyMetres: f.accuracyMetres,
        at: DateTime.now().toUtc(),
        isFixture: true,
      );

  final double latitude;
  final double longitude;
  final double accuracyMetres;
  final DateTime at;
  final bool isFixture;
}

/// Turns whatever the platform threw into something a person can act on.
///
/// Without a position there is no capture, so this is the one error on the
/// platform that has to carry instructions rather than just a cause. "User
/// denied Geolocation" is what a browser says to a programmer: shown to a
/// person it reads as an accusation and tells them nothing about what to do
/// next.
String locationMessage(String raw) {
  final e = raw.toLowerCase();

  if (e.contains('denied') || e.contains('permission')) {
    return 'Location permission is switched off for this site. Tap the icon at '
        'the left of the address bar, allow Location, then Refresh. '
        'Without a position a photograph cannot be matched to a ward or a '
        'contract.';
  }
  if (e.contains('services are disabled') || e.contains('location service')) {
    return 'Location services are switched off on this device. Turn them on in '
        'system settings, then Refresh.';
  }
  if (e.contains('timeout')) {
    return 'No fix yet. Indoors a phone often cannot see enough satellites — '
        'step into the open sky and Refresh.';
  }
  return 'The position could not be read. Refresh to try again; if it keeps '
      'failing, step into the open sky.';
}

class CaptureScreen extends StatefulWidget {
  const CaptureScreen({
    super.key,
    required this.client,
    required this.session,
    required this.onSignOut,
    this.outbox,
  });

  /// Captures that could not be sent yet. A walk goes in and out of signal and
  /// a photograph of a pothole cannot be taken again later.
  final CaptureStore? outbox;

  final ApiClient client;
  final Session session;
  final Future<void> Function() onSignOut;

  @override
  State<CaptureScreen> createState() => _CaptureScreenState();
}

class _CaptureScreenState extends State<CaptureScreen> {
  final _picker = ImagePicker();
  final _notes = TextEditingController();

  Uint8List? _photo;
  String _filename = 'capture.jpg';
  _Fix? _position;
  String? _locationError;
  bool _locating = false;
  bool _sending = false;
  String? _error;
  late final CaptureStore _outbox = widget.outbox ?? Outbox();

  final List<_Sent> _sent = [];
  int _waiting = 0;
  ReportDetail? _verdict;
  bool _watching = false;

  @override
  void initState() {
    super.initState();
    _locate();
    // Anything the last walk could not send goes now, while there is signal.
    unawaited(_drain());
  }

  /// Sends whatever is waiting and reports what is left.
  Future<void> _drain() async {
    final result = await _outbox.flush(
      (c) => widget.client.submit(c, widget.session.token),
    );
    final left = await _outbox.count();
    if (!mounted) return;
    setState(() => _waiting = left);
    if (result.rejected > 0) {
      setState(() => _error =
          '${result.rejected} saved capture(s) were refused by the server and '
          'have been discarded.');
    }
  }

  @override
  void dispose() {
    _notes.dispose();
    super.dispose();
  }

  /// A capture without coordinates cannot be routed to a ward or matched to a
  /// contract, so the fix is taken before the photograph is sent, not after.
  Future<void> _locate() async {
    setState(() {
      _locating = true;
      _locationError = null;
    });

    // A build that supplies a fixture does not ask the device at all: the
    // point is to work where the device cannot.
    final fixture = FixturePosition.configured;
    if (fixture != null) {
      setState(() {
        _position = _Fix.fixture(fixture);
        _locating = false;
      });
      return;
    }

    try {
      var permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
      }
      if (permission == LocationPermission.denied ||
          permission == LocationPermission.deniedForever) {
        throw 'Location permission is needed to route a report.';
      }

      final position = await Geolocator.getCurrentPosition(
        locationSettings: const LocationSettings(
          accuracy: LocationAccuracy.best,
          timeLimit: Duration(seconds: 20),
        ),
      );
      if (!mounted) return;
      setState(() => _position = _Fix.measured(position));
    } catch (e) {
      if (mounted) setState(() => _locationError = locationMessage(e.toString()));
    } finally {
      if (mounted) setState(() => _locating = false);
    }
  }

  Future<void> _pick(ImageSource source) async {
    try {
      final file = await _picker.pickImage(
        source: source,
        maxWidth: 2048,
        imageQuality: 85,
      );
      if (file == null) return;
      final bytes = await file.readAsBytes();
      if (!mounted) return;
      setState(() {
        _photo = bytes;
        _filename = file.name.isEmpty ? 'capture.jpg' : file.name;
        _error = null;
      });
    } catch (e) {
      if (mounted) setState(() => _error = 'Could not read the photograph: $e');
    }
  }

  Future<void> _send() async {
    final photo = _photo;
    final position = _position;
    if (photo == null || position == null) return;

    setState(() {
      _sending = true;
      _error = null;
    });
    try {
      final reportId = await widget.client.submit(
        Capture(
          bytes: photo,
          filename: _filename,
          latitude: position.latitude,
          longitude: position.longitude,
          accuracyMetres: position.accuracyMetres,
          capturedAt: position.at,
          notes: _notes.text.trim().isEmpty ? null : _notes.text.trim(),
        ),
        widget.session.token,
      );
      if (!mounted) return;
      setState(() {
        _sent.insert(0, _Sent(reportId, DateTime.now()));
        _photo = null;
        _notes.clear();
        _verdict = null;
      });
      // Enrichment runs after the capture is safe, so the answer arrives a
      // few seconds later. Watch for it rather than making the person guess.
      unawaited(_watchVerdict(reportId));
    } on ApiException catch (e) {
      // The server answered and refused. Saving it would only queue something
      // that will be refused again.
      if (mounted) setState(() => _error = e.message);
    } catch (_) {
      // The server could not be reached. The photograph cannot be taken again,
      // so it is written down and sent when there is signal.
      await _outbox.add(Capture(
        bytes: photo,
        filename: _filename,
        latitude: position.latitude,
        longitude: position.longitude,
        accuracyMetres: position.accuracyMetres,
        capturedAt: position.at,
        notes: _notes.text.trim().isEmpty ? null : _notes.text.trim(),
      ));
      final left = await _outbox.count();
      if (mounted) {
        setState(() {
          _waiting = left;
          _photo = null;
          _notes.clear();
          _verdict = null;
          _error = 'No signal. Saved — it will send itself when you are back '
              'online. Keep walking.';
        });
      }
    } finally {
      if (mounted) setState(() => _sending = false);
    }
  }

  /// Polls until the platform has worked out what the capture is.
  ///
  /// It gives up quietly: the report is safe and visible under My reports
  /// either way, and a spinner that never stops is worse than a short wait
  /// that ends.
  Future<void> _watchVerdict(String reportId) async {
    if (_watching) return;
    setState(() => _watching = true);
    try {
      for (var attempt = 0; attempt < 12; attempt++) {
        await Future<void>.delayed(const Duration(seconds: 3));
        if (!mounted) return;
        try {
          final r = await widget.client.report(reportId, widget.session.token);
          if (!mounted) return;
          setState(() => _verdict = r);
          if (r.enriched) return;
        } catch (_) {
          // A failed poll is not a failed capture.
        }
      }
    } finally {
      if (mounted) setState(() => _watching = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    final ready = _photo != null && _position != null && !_sending;

    return Scaffold(
      appBar: AppBar(
        title: const Text('TraceSarkar'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).push(MaterialPageRoute(
              builder: (_) => ReportsScreen(
                  client: widget.client, session: widget.session),
            )),
            child: const Text('My reports',
                style: TextStyle(color: Tokens.accent)),
          ),
          TextButton(
            onPressed: widget.onSignOut,
            child: const Text('Sign out',
                style: TextStyle(color: Tokens.ink70)),
          ),
        ],
      ),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 560),
            child: ListView(
              padding: const EdgeInsets.all(Tokens.gutter),
              children: [
                Text(
                  'Signed in as ${widget.session.account.email}',
                  style: text.bodySmall,
                ),
                const SizedBox(height: 16),
                _LocationCard(
                  position: _position,
                  locating: _locating,
                  error: _locationError,
                  onRetry: _locate,
                ),
                const SizedBox(height: 16),
                _PhotoCard(
                  photo: _photo,
                  onCamera: () => _pick(ImageSource.camera),
                  onGallery: () => _pick(ImageSource.gallery),
                  onClear: () => setState(() => _photo = null),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _notes,
                  maxLines: 3,
                  decoration: const InputDecoration(
                    labelText: 'What is in the photograph? (optional)',
                    alignLabelWithHint: true,
                  ),
                ),
                if (_error != null) ...[
                  const SizedBox(height: 16),
                  Text(_error!,
                      style: text.bodyLarge?.copyWith(color: Tokens.alert)),
                ],
                const SizedBox(height: 24),
                FilledButton(
                  onPressed: ready ? _send : null,
                  child: _sending
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                              strokeWidth: 2, color: Tokens.paper),
                        )
                      : const Text('Send capture'),
                ),
                if (_waiting > 0) ...[
                  const SizedBox(height: 12),
                  _WaitingNote(count: _waiting, onRetry: _drain),
                ],
                const SizedBox(height: 8),
                Text(
                  'Sending records the photograph and where it was taken. '
                  'Nothing is filed with any authority; any complaint or RTI '
                  'is a draft you review and submit yourself.',
                  style: text.bodySmall,
                ),
                if (_verdict != null) ...[
                  const SizedBox(height: 24),
                  Text('What we worked out', style: text.titleMedium),
                  const SizedBox(height: 10),
                  VerdictCard(report: _verdict!),
                  const SizedBox(height: 10),
                  OutlinedButton(
                    onPressed: () => Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => ReportsScreen(
                          client: widget.client, session: widget.session),
                      ),
                    ),
                    child: const Text('See all my reports'),
                  ),
                ],
                if (_sent.isNotEmpty) ...[
                  const SizedBox(height: 32),
                  Text('Sent in this session', style: text.titleMedium),
                  const SizedBox(height: 12),
                  for (final sent in _sent) ...[
                    _SentCard(sent),
                    const SizedBox(height: 8),
                  ],
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _LocationCard extends StatelessWidget {
  const _LocationCard({
    required this.position,
    required this.locating,
    required this.error,
    required this.onRetry,
  });

  final _Fix? position;
  final bool locating;
  final String? error;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;

    Widget body;
    if (locating && position == null) {
      body = Text('Finding your position…', style: text.bodyLarge);
    } else if (error != null && position == null) {
      body = Text(error!, style: text.bodyLarge?.copyWith(color: Tokens.alert));
    } else if (position == null) {
      body = Text('No position yet.', style: text.bodyLarge);
    } else {
      final accuracy = position!.accuracyMetres;
      // The same bands the field kit uses, so a capture from either surface
      // means the same thing.
      final (tint, ink, label) = position!.isFixture
          ? (Tokens.cautionWash, Tokens.caution, 'Fixture position')
          : accuracy <= 15
              ? (Tokens.confirmWash, Tokens.confirm, 'Good fix')
              : accuracy <= 50
                  ? (Tokens.cautionWash, Tokens.caution, 'Rough fix')
                  : (Tokens.alertWash, Tokens.alert, 'Poor fix');

      body = Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: tint,
                  borderRadius: BorderRadius.circular(Tokens.chipRadius),
                ),
                child: Text(
                    position!.isFixture
                        ? label
                        : '$label · ±${accuracy.toStringAsFixed(0)} m',
                    style: text.labelLarge?.copyWith(color: ink)),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            '${position!.latitude.toStringAsFixed(5)}, '
            '${position!.longitude.toStringAsFixed(5)}',
            style: text.bodyLarge?.copyWith(
                fontFeatures: const [FontFeature.tabularFigures()]),
          ),
          const SizedBox(height: 4),
          Text(
            position!.isFixture
                ? 'This position was supplied by the build, not measured by '
                    'this device. It is fine for a demonstration and worthless '
                    'as evidence.'
                : 'Accuracy decides whether this can be matched to a ward and '
                    'a contract. Exact coordinates stay private to your '
                    'account.',
            style: text.bodySmall,
          ),
        ],
      );
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(Tokens.gutter),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Location', style: text.titleMedium),
                TextButton(
                  onPressed: locating ? null : onRetry,
                  child: Text(locating ? 'Locating…' : 'Refresh',
                      style: const TextStyle(color: Tokens.accent)),
                ),
              ],
            ),
            const SizedBox(height: 4),
            body,
          ],
        ),
      ),
    );
  }
}

class _PhotoCard extends StatelessWidget {
  const _PhotoCard({
    required this.photo,
    required this.onCamera,
    required this.onGallery,
    required this.onClear,
  });

  final Uint8List? photo;
  final VoidCallback onCamera;
  final VoidCallback onGallery;
  final VoidCallback onClear;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(Tokens.gutter),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Photograph', style: text.titleMedium),
            const SizedBox(height: 12),
            if (photo != null) ...[
              // The photograph is the hero: full width, 4:3.
              ClipRRect(
                borderRadius: BorderRadius.circular(Tokens.chipRadius),
                child: AspectRatio(
                  aspectRatio: 4 / 3,
                  child: Image.memory(photo!, fit: BoxFit.cover),
                ),
              ),
              const SizedBox(height: 12),
              OutlinedButton(
                onPressed: onClear,
                child: const Text('Replace photograph'),
              ),
            ] else ...[
              Text(
                'One clear photograph of the problem itself.',
                style: text.bodySmall,
              ),
              const SizedBox(height: 12),
              FilledButton.icon(
                onPressed: onCamera,
                icon: const Icon(Icons.photo_camera_outlined),
                label: const Text('Take a photograph'),
              ),
              const SizedBox(height: 8),
              OutlinedButton.icon(
                onPressed: onGallery,
                icon: const Icon(Icons.image_outlined),
                label: const Text('Choose an existing one'),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _SentCard extends StatelessWidget {
  const _SentCard(this.sent);
  final _Sent sent;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(Tokens.gutter),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.check_circle_outline,
                    size: 20, color: Tokens.confirm),
                const SizedBox(width: 8),
                Text('Capture recorded',
                    style: text.labelLarge?.copyWith(color: Tokens.confirm)),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              sent.reportId,
              style: text.bodySmall?.copyWith(
                fontFamily: 'monospace',
                fontFeatures: const [FontFeature.tabularFigures()],
              ),
            ),
            const SizedBox(height: 4),
            Text(
              'Stored. Jurisdiction and contract matching have not run yet.',
              style: text.bodySmall,
            ),
          ],
        ),
      ),
    );
  }
}


/// What is still waiting to reach the server.
///
/// Shown rather than hidden: someone who has walked for three hours needs to
/// see that nothing is stuck, and a queue that drains silently is
/// indistinguishable from one that has lost the photographs.
class _WaitingNote extends StatelessWidget {
  const _WaitingNote({required this.count, required this.onRetry});

  final int count;
  final Future<void> Function() onRetry;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Tokens.cautionWash,
        borderRadius: BorderRadius.circular(Tokens.chipRadius),
        border: Border.all(color: Tokens.caution.withValues(alpha: 0.3)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.schedule, size: 20, color: Tokens.caution),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              count == 1
                  ? '1 capture waiting to send. It is saved on this phone.'
                  : '$count captures waiting to send. They are saved on this '
                      'phone.',
              style: text.bodyLarge?.copyWith(color: Tokens.caution),
            ),
          ),
          TextButton(
            onPressed: onRetry,
            child: const Text('Send now',
                style: TextStyle(color: Tokens.caution)),
          ),
        ],
      ),
    );
  }
}
