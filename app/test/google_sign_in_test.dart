import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/google_auth.dart';
import 'package:tracesarkar_app/main.dart';
import 'package:tracesarkar_app/screens/sign_in_screen.dart';
import 'package:tracesarkar_app/session_store.dart';

/// Stands in for Google. Hands back whatever token the test wants, when the
/// button is pressed, through the same callback the real flow uses.
class FakeGoogle implements GoogleAuthFlow {
  FakeGoogle({this.token = 'google-id-token', this.available = true});

  final String? token;
  @override
  final bool available;

  void Function(String)? _onToken;
  int starts = 0;
  int signIns = 0;
  int signOuts = 0;

  @override
  Future<void> start(void Function(String idToken) onIdToken) async {
    starts++;
    _onToken = onIdToken;
  }

  @override
  Future<void> signIn() async {
    signIns++;
    if (token != null) _onToken?.call(token!);
  }

  @override
  Future<void> signOut() async => signOuts++;
}

Future<void> pumpSignIn(
  WidgetTester tester, {
  required http.Client client,
  required GoogleAuthFlow google,
  void Function(Session)? onSignedIn,
}) async {
  await tester.pumpWidget(MaterialApp(
    home: SignInScreen(
      client: ApiClient(baseUrl: 'https://api.test', client: client),
      google: google,
      onSignedIn: (s) async => onSignedIn?.call(s),
    ),
  ));
  await tester.pumpAndSettle();
}

void main() {
  setUp(() => SharedPreferences.setMockInitialValues({}));

  testWidgets('without a client id, no Google button is offered',
      (tester) async {
    await pumpSignIn(
      tester,
      client: MockClient((_) async => http.Response('{}', 500)),
      google: FakeGoogle(available: false),
    );

    expect(find.textContaining('Google'), findsNothing);
    // Email and password must still work on their own.
    expect(find.widgetWithText(FilledButton, 'Sign in'), findsOneWidget);
  });

  testWidgets('a Google token is exchanged with our own backend',
      (tester) async {
    String? sentTo;
    String? sentToken;
    Session? signedIn;

    await pumpSignIn(
      tester,
      client: MockClient((request) async {
        sentTo = request.url.path;
        sentToken = (jsonDecode(request.body) as Map)['id_token'] as String?;
        return http.Response(
          jsonEncode({
            'token': 'session-tok',
            'account': {'id': 'acct-1', 'email': 'a@b.org', 'name': 'A'},
          }),
          200,
        );
      }),
      google: FakeGoogle(token: 'google-id-token'),
      onSignedIn: (s) => signedIn = s,
    );

    await tester.tap(find.textContaining('Continue with Google'));
    await tester.pumpAndSettle();

    // The ID token goes to our server, which is the only thing that may
    // decide whether it is genuine.
    expect(sentTo, '/v1/auth/google');
    expect(sentToken, 'google-id-token');
    expect(signedIn?.token, 'session-tok');
  });

  testWidgets('a rejected Google token shows the reason and stays put',
      (tester) async {
    Session? signedIn;
    await pumpSignIn(
      tester,
      client: MockClient((_) async => http.Response(
            jsonEncode({
              'error': {'message': 'that Google account is not allowed here'}
            }),
            401,
          )),
      google: FakeGoogle(token: 'stale-token'),
      onSignedIn: (s) => signedIn = s,
    );

    await tester.tap(find.textContaining('Continue with Google'));
    await tester.pumpAndSettle();

    expect(find.text('that Google account is not allowed here'),
        findsOneWidget);
    expect(signedIn, isNull);
  });

  testWidgets('a cancelled Google sign-in is silent, not an error',
      (tester) async {
    // Someone who closes the account picker has not failed at anything.
    await pumpSignIn(
      tester,
      client: MockClient((_) async => http.Response('{}', 500)),
      google: FakeGoogle(token: null),
    );

    await tester.tap(find.textContaining('Continue with Google'));
    await tester.pumpAndSettle();

    expect(find.byIcon(Icons.error_outline), findsNothing);
    expect(find.textContaining('Could not reach'), findsNothing);
  });

  testWidgets('signing out of the app signs out of Google too', (tester) async {
    // Otherwise the next launch picks the Google session straight back up and
    // signing out appears not to have worked.
    await SessionStore().save(Session(
      token: 'tok-1',
      account: Account(id: 'acct-1', email: 'a@b.org', name: 'A'),
    ));
    final google = FakeGoogle();

    await tester.pumpWidget(TraceSarkarApp(
      client: ApiClient(
        baseUrl: 'https://api.test',
        client: MockClient((_) async => http.Response('{}', 500)),
      ),
      sessions: SessionStore(),
      google: google,
    ));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Sign out'));
    await tester.pumpAndSettle();

    expect(google.signOuts, 1);
  });

  testWidgets('a session Google already holds is picked up without a tap',
      (tester) async {
    final google = FakeGoogle();
    await pumpSignIn(
      tester,
      client: MockClient((_) async => http.Response('{}', 500)),
      google: google,
    );

    // start() is what subscribes to Google's own event stream, which is how a
    // browser session already signed in arrives.
    expect(google.starts, 1);
  });
}
