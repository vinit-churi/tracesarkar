import 'package:flutter/material.dart';
import 'package:google_sign_in_web/google_sign_in_web.dart';
import 'package:google_sign_in_web/web_only.dart' as web;

/// On the web, Google requires its own rendered button — Google Identity
/// Services refuses a sign-in that anything else started. The token arrives
/// through the authentication event stream, not from this widget.
Widget googleSignInButton({required VoidCallback onPressed}) {
  // The rendered button is a platform view with no intrinsic size in Flutter,
  // so it collapses to nothing unless it is given one. 44 matches the height
  // of our own buttons; `large` is the tallest GSI offers and still fits.
  return SizedBox(
    height: 44,
    child: web.renderButton(
      configuration: GSIButtonConfiguration(
        type: GSIButtonType.standard,
        theme: GSIButtonTheme.outline,
        size: GSIButtonSize.large,
        text: GSIButtonText.continueWith,
        shape: GSIButtonShape.rectangular,
        logoAlignment: GSIButtonLogoAlignment.left,
      ),
    ),
  );
}
