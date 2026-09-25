import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/session_store.dart';

void main() {
  setUp(() => SharedPreferences.setMockInitialValues({}));

  test('nothing is stored before anyone signs in', () async {
    expect(await SessionStore().read(), isNull);
  });

  test('a saved session survives a restart', () async {
    await SessionStore().save(Session(
      token: 'tok-1',
      account: Account(id: 'acct-1', email: 'a@b.org', name: 'A'),
    ));

    // A fresh instance stands in for the next launch of the app.
    final restored = await SessionStore().read();

    expect(restored, isNotNull);
    expect(restored!.token, 'tok-1');
    expect(restored.account.email, 'a@b.org');
  });

  test('signing out leaves no token behind', () async {
    final store = SessionStore();
    await store.save(Session(
      token: 'tok-1',
      account: Account(id: 'acct-1', email: 'a@b.org', name: 'A'),
    ));

    await store.clear();

    expect(await store.read(), isNull);
  });
}
