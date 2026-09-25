import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

import 'api.dart';

/// Keeps the signed-in session across launches.
///
/// The token is a bearer credential, so it lives in the platform's own store
/// rather than anywhere the app writes itself, and nothing else about the
/// account is cached — `/v1/auth/me` is the authority.
class SessionStore {
  static const _key = 'tracesarkar.session';

  Future<Session?> read() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_key);
    if (raw == null || raw.isEmpty) return null;

    try {
      final decoded = jsonDecode(raw) as Map<String, dynamic>;
      final token = decoded['token'] as String?;
      if (token == null || token.isEmpty) return null;
      return Session(
        token: token,
        account: Account.fromJson(
            (decoded['account'] as Map<String, dynamic>?) ?? const {}),
      );
    } catch (_) {
      // A session we cannot read is a session we do not have.
      await prefs.remove(_key);
      return null;
    }
  }

  Future<void> save(Session session) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(
      _key,
      jsonEncode({
        'token': session.token,
        'account': {
          'id': session.account.id,
          'email': session.account.email,
          'name': session.account.name,
        },
      }),
    );
  }

  Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_key);
  }
}
