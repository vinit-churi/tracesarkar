import 'package:flutter/material.dart';

import 'api.dart';
import 'screens/capture_screen.dart';
import 'screens/sign_in_screen.dart';
import 'session_store.dart';
import 'theme.dart';

void main() {
  runApp(const TraceSarkarApp());
}

class TraceSarkarApp extends StatefulWidget {
  const TraceSarkarApp({super.key, this.client, this.sessions});

  /// Injectable so a widget test can run the app against a fake server.
  final ApiClient? client;
  final SessionStore? sessions;

  @override
  State<TraceSarkarApp> createState() => _TraceSarkarAppState();
}

class _TraceSarkarAppState extends State<TraceSarkarApp> {
  late final ApiClient _client = widget.client ?? ApiClient();
  late final SessionStore _sessions = widget.sessions ?? SessionStore();

  Session? _session;
  bool _restoring = true;

  @override
  void initState() {
    super.initState();
    _restore();
  }

  Future<void> _restore() async {
    final stored = await _sessions.read();
    if (!mounted) return;
    setState(() {
      _session = stored;
      _restoring = false;
    });
  }

  Future<void> _signedIn(Session session) async {
    await _sessions.save(session);
    if (!mounted) return;
    setState(() => _session = session);
  }

  Future<void> _signOut() async {
    await _sessions.clear();
    if (!mounted) return;
    setState(() => _session = null);
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'TraceSarkar',
      debugShowCheckedModeBanner: false,
      theme: buildTheme(),
      home: _restoring
          ? const Scaffold(
              backgroundColor: Tokens.paper,
              body: Center(child: CircularProgressIndicator()),
            )
          : _session == null
              ? SignInScreen(client: _client, onSignedIn: _signedIn)
              : CaptureScreen(
                  client: _client,
                  session: _session!,
                  onSignOut: _signOut,
                ),
    );
  }
}
