/// A position supplied at build time instead of read from the device.
///
/// This exists for one reason: a laptop in a meeting room often cannot get a
/// GPS fix, and without a position there is nothing to send. It is opt-in per
/// build — `--dart-define=DEMO_POSITION=19.2094,72.8348` — and the screen
/// labels it, because a capture whose position was typed in must never be
/// mistaken for one that was measured.
class FixturePosition {
  const FixturePosition._(this.latitude, this.longitude);

  final double latitude;
  final double longitude;

  /// Deliberately worse than any real fix, so it sorts and renders as the
  /// least trustworthy thing on the screen.
  double get accuracyMetres => 9999;

  bool get isFixture => true;

  /// Parses the build-time setting. Anything malformed yields null: a
  /// half-understood coordinate is worse than no coordinate.
  static FixturePosition? parse(String raw) {
    final parts = raw.split(',');
    if (parts.length != 2) return null;

    final lat = double.tryParse(parts[0].trim());
    final lon = double.tryParse(parts[1].trim());
    if (lat == null || lon == null) return null;
    if (lat < -90 || lat > 90 || lon < -180 || lon > 180) return null;

    return FixturePosition._(lat, lon);
  }

  /// The setting for this build, or null when none was given.
  static FixturePosition? get configured =>
      parse(const String.fromEnvironment('DEMO_POSITION'));
}
