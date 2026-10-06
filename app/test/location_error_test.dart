import 'package:flutter_test/flutter_test.dart';
import 'package:tracesarkar_app/screens/capture_screen.dart';

// "User denied Geolocation" is what the browser says to a programmer. Shown to
// a person it reads as an accusation and tells them nothing about what to do,
// and the capture screen is unusable until it is resolved — so this is the one
// error message on the platform that has to carry instructions.
void main() {
  test('a denied permission explains how to undo it', () {
    final m = locationMessage('User denied Geolocation');
    expect(m.toLowerCase(), isNot(contains('user denied')));
    expect(m, contains('permission'));
    // It must say where the control is, not just that one exists.
    expect(m.toLowerCase(), anyOf(contains('address bar'), contains('settings')));
  });

  test('a timeout is told apart from a refusal', () {
    // Different cause, different remedy: one needs a tap, the other needs sky.
    final m = locationMessage('TimeoutException after 0:00:15.000000');
    expect(m.toLowerCase(), contains('open sky'));
    expect(m, isNot(contains('permission')));
  });

  test('switched-off location services are told apart from both', () {
    final m = locationMessage('Location services are disabled.');
    expect(m.toLowerCase(), contains('location services'));
  });

  test('anything unrecognised still says something a person can act on', () {
    final m = locationMessage('PlatformException(42, null, null, null)');
    expect(m, isNot(contains('PlatformException')));
    expect(m, isNot(contains('42')));
    expect(m.length, greaterThan(20));
  });
}
