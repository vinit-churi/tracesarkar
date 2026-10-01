import 'package:flutter/material.dart';

/// On a device with its own account picker, an ordinary button is right.
Widget googleSignInButton({required VoidCallback onPressed}) {
  return OutlinedButton.icon(
    onPressed: onPressed,
    icon: const Icon(Icons.account_circle_outlined, size: 20),
    label: const Text('Continue with Google'),
  );
}
