import 'package:flutter/material.dart';

/// Design tokens from docs/02-product/11-screen-spec.md §2.
///
/// The palette is party-neutral by requirement: saffron, green and blue are
/// party-coded in Maharashtra, so none of them may be the dominant brand
/// colour. Plum is the single accent. `confirm` green is reserved for
/// citizen-confirmed and successful capture, and never fills a surface.
abstract final class Tokens {
  static const ink = Color(0xFF141518);
  static const ink70 = Color(0xFF4A4D55);
  static const ink45 = Color(0xFF7C8089);
  static const paper = Color(0xFFFBFAF8);
  static const surface = Color(0xFFFFFFFF);
  static const hairline = Color(0xFFE3E1DC);
  static const accent = Color(0xFF6B2C4F);
  static const accentWash = Color(0xFFF5EAF0);
  static const alert = Color(0xFFA33A1F);
  static const alertWash = Color(0xFFFBEDE8);
  static const caution = Color(0xFF8A6410);
  static const cautionWash = Color(0xFFFAF3E2);
  static const confirm = Color(0xFF1F6A4D);
  static const confirmWash = Color(0xFFEAF3EF);
  static const evidence = Color(0xFFF2F0EB);

  // §2.3 — 4 pt base grid.
  static const gutter = 16.0;
  static const cardRadius = 12.0;
  static const chipRadius = 8.0;
  static const buttonRadius = 10.0;
  static const minTarget = 48.0;
}

ThemeData buildTheme() {
  const text = TextTheme(
    // §2.2. No webfont on the capture path — it must render on 2G, so this is
    // the system stack.
    displaySmall: TextStyle(
        fontSize: 28, height: 34 / 28, letterSpacing: -0.56, color: Tokens.ink),
    titleMedium: TextStyle(
        fontSize: 20, height: 26 / 20, fontWeight: FontWeight.w600, color: Tokens.ink),
    bodyLarge: TextStyle(fontSize: 16, height: 24 / 16, color: Tokens.ink),
    labelLarge: TextStyle(
        fontSize: 14, height: 20 / 14, fontWeight: FontWeight.w500, color: Tokens.ink),
    bodySmall: TextStyle(fontSize: 13, height: 18 / 13, color: Tokens.ink45),
  );

  final base = ThemeData(
    useMaterial3: true,
    colorScheme: ColorScheme.fromSeed(
      seedColor: Tokens.accent,
      primary: Tokens.accent,
      surface: Tokens.surface,
      error: Tokens.alert,
    ),
    scaffoldBackgroundColor: Tokens.paper,
    textTheme: text,
  );

  return base.copyWith(
    appBarTheme: const AppBarTheme(
      backgroundColor: Tokens.paper,
      surfaceTintColor: Colors.transparent,
      foregroundColor: Tokens.ink,
      elevation: 0,
      centerTitle: false,
    ),
    // §2.3: elevation is borders, not shadows.
    cardTheme: CardThemeData(
      color: Tokens.surface,
      elevation: 0,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        side: const BorderSide(color: Tokens.hairline),
        borderRadius: BorderRadius.circular(Tokens.cardRadius),
      ),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: Tokens.ink,
        foregroundColor: Tokens.paper,
        minimumSize: const Size.fromHeight(Tokens.minTarget),
        textStyle: text.labelLarge?.copyWith(fontSize: 16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(Tokens.buttonRadius),
        ),
      ),
    ),
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        foregroundColor: Tokens.ink,
        minimumSize: const Size.fromHeight(Tokens.minTarget),
        side: const BorderSide(color: Tokens.hairline),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(Tokens.buttonRadius),
        ),
      ),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: Tokens.surface,
      contentPadding:
          const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
      border: OutlineInputBorder(
        borderSide: const BorderSide(color: Tokens.hairline),
        borderRadius: BorderRadius.circular(Tokens.buttonRadius),
      ),
      enabledBorder: OutlineInputBorder(
        borderSide: const BorderSide(color: Tokens.hairline),
        borderRadius: BorderRadius.circular(Tokens.buttonRadius),
      ),
      focusedBorder: OutlineInputBorder(
        borderSide: const BorderSide(color: Tokens.accent, width: 2),
        borderRadius: BorderRadius.circular(Tokens.buttonRadius),
      ),
      labelStyle: const TextStyle(color: Tokens.ink70),
    ),
    dividerTheme: const DividerThemeData(
      color: Tokens.hairline,
      thickness: 1,
      space: 1,
    ),
  );
}
