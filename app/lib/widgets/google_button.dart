library;

/// Google's sign-in button, which is a different thing on each platform.
///
/// The web build must use Google's own rendered button — Google Identity
/// Services refuses a sign-in started by anything else — while a device uses
/// its native account picker.
export 'google_button_io.dart'
    if (dart.library.js_interop) 'google_button_web.dart';
