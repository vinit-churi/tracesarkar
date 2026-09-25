import 'package:flutter_test/flutter_test.dart';
import 'package:tracesarkar_app/fixture_position.dart';

void main() {
  test('no fixture is configured by default', () {
    expect(FixturePosition.parse(''), isNull);
  });

  test('a configured fixture is used for the position', () {
    final fixture = FixturePosition.parse('19.2094,72.8348');

    expect(fixture, isNotNull);
    expect(fixture!.latitude, 19.2094);
    expect(fixture.longitude, 72.8348);
  });

  test('a fixture carries an accuracy that cannot pass for a real fix', () {
    // A fixture must never look better than a real GPS fix, or someone will
    // read a demo capture as a located one.
    final fixture = FixturePosition.parse('19.2094,72.8348')!;

    expect(fixture.accuracyMetres, greaterThan(50));
    expect(fixture.isFixture, isTrue);
  });

  test('malformed settings are ignored rather than guessed at', () {
    for (final bad in ['19.2094', 'north,west', '19.2094,72.8348,extra', ',']) {
      expect(FixturePosition.parse(bad), isNull, reason: bad);
    }
  });

  test('coordinates outside the world are refused', () {
    expect(FixturePosition.parse('91.0,72.8'), isNull);
    expect(FixturePosition.parse('19.2,181.0'), isNull);
  });
}
