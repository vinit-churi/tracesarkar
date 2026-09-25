import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:image_picker/image_picker.dart';

import '../api.dart';
import '../theme.dart';

/// A capture that has reached the server, kept for this session so the person
/// can see what they sent.
class _Sent {
  _Sent(this.reportId, this.at);
  final String reportId;
  final DateTime at;
}

class CaptureScreen extends StatefulWidget {
  const CaptureScreen({
    super.key,
    required this.client,
    required this.session,
    required this.onSignOut,
  });

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
  Position? _position;
  String? _locationError;
  bool _locating = false;
  bool _sending = false;
  String? _error;
  final List<_Sent> _sent = [];

  @override
  void initState() {
    super.initState();
    _locate();
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
      setState(() => _position = position);
    } catch (e) {
      if (mounted) setState(() => _locationError = e.toString());
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
          accuracyMetres:
              (position.accuracy * 10).roundToDouble() / 10,
          capturedAt: position.timestamp,
          notes: _notes.text.trim().isEmpty ? null : _notes.text.trim(),
        ),
        widget.session.token,
      );
      if (!mounted) return;
      setState(() {
        _sent.insert(0, _Sent(reportId, DateTime.now()));
        _photo = null;
        _notes.clear();
      });
    } on ApiException catch (e) {
      if (mounted) setState(() => _error = e.message);
    } catch (_) {
      if (mounted) {
        setState(() => _error =
            'Could not reach the server. The capture has not been sent; try again.');
      }
    } finally {
      if (mounted) setState(() => _sending = false);
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
                const SizedBox(height: 8),
                Text(
                  'Sending records the photograph and where it was taken. '
                  'Nothing is filed with any authority; any complaint or RTI '
                  'is a draft you review and submit yourself.',
                  style: text.bodySmall,
                ),
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

  final Position? position;
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
      final accuracy = position!.accuracy;
      // The same bands the field kit uses, so a capture from either surface
      // means the same thing.
      final (tint, ink, label) = accuracy <= 15
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
                child: Text('$label · ±${accuracy.toStringAsFixed(0)} m',
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
            'Accuracy decides whether this can be matched to a ward and a '
            'contract. Exact coordinates stay private to your account.',
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
