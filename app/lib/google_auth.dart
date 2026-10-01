import 'dart:async';

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:google_sign_in/google_sign_in.dart';

import 'api.dart';

/// Google sign-in, reduced to the one thing this platform needs from it: an ID
/// token to hand to our own backend.
///
/// A token from here is a claim, not a fact. The backend verifies it against
/// Google's published keys and checks the issuer and audience before it will
/// mint a session. Nothing on the client is trusted.
abstract class GoogleAuthFlow {
  /// Whether this build was configured with a client id. Without one the
  /// button is hidden rather than shown and broken.
  bool get available;

  /// Subscribes to sign-ins. [onIdToken] fires for each one — including a
  /// session the browser or device already holds, which arrives without
  /// anyone pressing anything.
  Future<void> start(void Function(String idToken) onIdToken);

  /// Starts an interactive sign-in. On the web this does nothing, because
  /// Google Identity Services requires its own rendered button to begin the
  /// flow; the token still arrives through [start].
  Future<void> signIn();

  /// Ends the Google session. Clearing our own session without this leaves
  /// Google's in place, and the next launch picks it straight back up — so
  /// signing out would appear not to have worked.
  Future<void> signOut();
}

class GoogleAuth implements GoogleAuthFlow {
  GoogleAuth._();

  static final GoogleAuth instance = GoogleAuth._();

  bool _started = false;
  StreamSubscription<GoogleSignInAuthenticationEvent>? _sub;

  @override
  bool get available => googleClientId.isNotEmpty;

  @override
  Future<void> start(void Function(String idToken) onIdToken) async {
    if (_started || !available) return;
    _started = true;

    // The same web OAuth client serves both, under different names: the web
    // plugin asserts serverClientId is null, and Android refuses to mint an ID
    // token without it.
    await GoogleSignIn.instance.initialize(
      clientId: kIsWeb ? googleClientId : null,
      serverClientId: kIsWeb ? null : googleClientId,
    );

    _sub = GoogleSignIn.instance.authenticationEvents.listen((event) {
      if (event is GoogleSignInAuthenticationEventSignIn) {
        final token = event.user.authentication.idToken;
        if (token != null && token.isNotEmpty) onIdToken(token);
      }
    });

    // Returns a session already held, without prompting. Silent by design:
    // someone who did not ask to sign in is not interrupted.
    unawaited(GoogleSignIn.instance.attemptLightweightAuthentication());
  }

  @override
  Future<void> signIn() async {
    if (!available) return;
    // Web has no such sheet — renderButton() drives the flow there.
    if (!GoogleSignIn.instance.supportsAuthenticate()) return;
    await GoogleSignIn.instance.authenticate();
  }

  @override
  Future<void> signOut() async {
    if (!available) return;
    await GoogleSignIn.instance.signOut();
  }

  void dispose() {
    _sub?.cancel();
    _sub = null;
    _started = false;
  }
}
