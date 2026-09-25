import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/main.dart';
import 'package:tracesarkar_app/session_store.dart';

ApiClient clientReturning(http.Response Function(http.Request) reply) =>
    ApiClient(
      baseUrl: 'https://api.test',
      client: MockClient((request) async => reply(request)),
    );

http.Response signedIn(String email) => http.Response(
      jsonEncode({
        'token': 'tok-1',
        'account': {'id': 'acct-1', 'email': email, 'name': ''},
      }),
      200,
    );

Future<void> signIn(WidgetTester tester, {String password = 'correct-horse'}) async {
  await tester.enterText(
      find.widgetWithText(TextFormField, 'Email'), 'a@b.org');
  await tester.enterText(
      find.widgetWithText(TextFormField, 'Password'), password);
  await tester.tap(find.widgetWithText(FilledButton, 'Sign in'));
  await tester.pumpAndSettle();
}

void main() {
  setUp(() => SharedPreferences.setMockInitialValues({}));

  testWidgets('an unsigned-in launch asks for credentials', (tester) async {
    await tester.pumpWidget(TraceSarkarApp(
      client: clientReturning((_) => signedIn('a@b.org')),
      sessions: SessionStore(),
    ));
    await tester.pumpAndSettle();

    expect(find.text('TraceSarkar'), findsOneWidget);
    expect(find.widgetWithText(FilledButton, 'Sign in'), findsOneWidget);
    expect(find.text('Send capture'), findsNothing);
  });

  testWidgets('signing in reaches the capture screen', (tester) async {
    await tester.pumpWidget(TraceSarkarApp(
      client: clientReturning((_) => signedIn('a@b.org')),
      sessions: SessionStore(),
    ));
    await tester.pumpAndSettle();

    await signIn(tester);

    expect(find.text('Send capture'), findsOneWidget);
    expect(find.text('Signed in as a@b.org'), findsOneWidget);
  });

  testWidgets('a rejected sign-in shows the reason and stays put',
      (tester) async {
    await tester.pumpWidget(TraceSarkarApp(
      client: clientReturning((_) => http.Response(
            jsonEncode({
              'error': {'message': 'email or password is incorrect'}
            }),
            401,
          )),
      sessions: SessionStore(),
    ));
    await tester.pumpAndSettle();

    await signIn(tester, password: 'wrong');

    expect(find.text('email or password is incorrect'), findsOneWidget);
    expect(find.text('Send capture'), findsNothing);
  });

  testWidgets('a session saved earlier opens straight into capture',
      (tester) async {
    await SessionStore().save(Session(
      token: 'tok-1',
      account: Account(id: 'acct-1', email: 'a@b.org', name: 'A'),
    ));

    await tester.pumpWidget(TraceSarkarApp(
      client: clientReturning((_) => signedIn('a@b.org')),
      sessions: SessionStore(),
    ));
    await tester.pumpAndSettle();

    expect(find.text('Send capture'), findsOneWidget);
  });

  testWidgets('signing out returns to the sign-in screen', (tester) async {
    await SessionStore().save(Session(
      token: 'tok-1',
      account: Account(id: 'acct-1', email: 'a@b.org', name: 'A'),
    ));
    await tester.pumpWidget(TraceSarkarApp(
      client: clientReturning((_) => signedIn('a@b.org')),
      sessions: SessionStore(),
    ));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Sign out'));
    await tester.pumpAndSettle();

    expect(find.widgetWithText(FilledButton, 'Sign in'), findsOneWidget);
    // The token must be gone, not merely off screen.
    expect(await SessionStore().read(), isNull);
  });

  testWidgets('the capture screen never implies the platform files anything',
      (tester) async {
    await SessionStore().save(Session(
      token: 'tok-1',
      account: Account(id: 'acct-1', email: 'a@b.org', name: 'A'),
    ));
    await tester.pumpWidget(TraceSarkarApp(
      client: clientReturning((_) => signedIn('a@b.org')),
      sessions: SessionStore(),
    ));
    await tester.pumpAndSettle();

    // Hard rule 1: nothing is auto-filed, and no control may suggest it is.
    for (final forbidden in ['File complaint', 'File RTI', 'Submit to BMC']) {
      expect(find.text(forbidden), findsNothing);
    }
    // The note sits under the send button, which is below the fold on a small
    // screen — exactly where someone reaching for the button will read it.
    await tester.scrollUntilVisible(
      find.textContaining('Nothing is filed with any authority'),
      200,
      scrollable: find.byType(Scrollable).first,
    );
    expect(
      find.textContaining('Nothing is filed with any authority'),
      findsOneWidget,
    );
  });
}
