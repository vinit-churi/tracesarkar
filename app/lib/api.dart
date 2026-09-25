import 'dart:convert';
import 'dart:typed_data';

import 'package:http/http.dart' as http;

/// Where the backend lives. Override at build time:
///   flutter build web --dart-define=API_BASE=https://api.example.org
const apiBase = String.fromEnvironment(
  'API_BASE',
  defaultValue: 'http://localhost:8080',
);

/// Google sign-in is optional; without a client id the button is hidden and
/// email/password still works.
const googleClientId = String.fromEnvironment('GOOGLE_CLIENT_ID');

class ApiException implements Exception {
  ApiException(this.statusCode, this.message);
  final int statusCode;
  final String message;

  @override
  String toString() => message;
}

class Account {
  Account({required this.id, required this.email, required this.name});

  factory Account.fromJson(Map<String, dynamic> json) => Account(
        id: (json['id'] ?? '') as String,
        email: (json['email'] ?? '') as String,
        name: (json['name'] ?? '') as String,
      );

  final String id;
  final String email;
  final String name;
}

class Session {
  Session({required this.token, required this.account});
  final String token;
  final Account account;
}

/// A capture on its way to the server: the photograph, where it was taken, and
/// how accurate that position is.
class Capture {
  Capture({
    required this.bytes,
    required this.filename,
    required this.latitude,
    required this.longitude,
    required this.accuracyMetres,
    required this.capturedAt,
    this.role = 'close',
    this.label,
    this.ward,
    this.conditions = const [],
    this.notes,
  });

  final Uint8List bytes;
  final String filename;
  final double latitude;
  final double longitude;
  final double accuracyMetres;
  final DateTime capturedAt;
  final String role;
  final String? label;
  final String? ward;
  final List<String> conditions;
  final String? notes;

  Map<String, dynamic> toMeta() {
    final meta = <String, dynamic>{
      'location': {
        'lat': latitude,
        'lon': longitude,
        'accuracy_m': accuracyMetres,
      },
      'captured_at': capturedAt.toUtc().toIso8601String(),
    };
    if (label != null) {
      meta['label'] = {
        'frame_type': role,
        'label': label,
        'conditions': conditions,
        'ward_ground_truth': ward ?? '',
        'notes': notes ?? '',
      };
    }
    return meta;
  }
}

class ApiClient {
  ApiClient({http.Client? client, this.baseUrl = apiBase})
      : _client = client ?? http.Client();

  final http.Client _client;
  final String baseUrl;

  Uri _url(String path) => Uri.parse('$baseUrl$path');

  Future<Session> register({
    required String email,
    required String password,
    String name = '',
  }) =>
      _authCall('/v1/auth/register', {
        'email': email,
        'password': password,
        'name': name,
      });

  Future<Session> login({required String email, required String password}) =>
      _authCall('/v1/auth/login', {'email': email, 'password': password});

  Future<Session> signInWithGoogle(String idToken) =>
      _authCall('/v1/auth/google', {'id_token': idToken});

  Future<Session> _authCall(String path, Map<String, dynamic> body) async {
    final response = await _client.post(
      _url(path),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode(body),
    );
    final decoded = _decode(response);
    return Session(
      token: decoded['token'] as String,
      account: Account.fromJson(decoded['account'] as Map<String, dynamic>),
    );
  }

  Future<Account> me(String token) async {
    final response = await _client.get(
      _url('/v1/auth/me'),
      headers: {'Authorization': 'Bearer $token'},
    );
    final decoded = _decode(response);
    return Account.fromJson(decoded['account'] as Map<String, dynamic>);
  }

  /// Uploads a capture. The server stores the photograph before replying, so a
  /// 202 here means the capture is safe even though nothing has enriched it yet.
  Future<String> submit(Capture capture, String token) async {
    final request = http.MultipartRequest('POST', _url('/v1/reports'))
      ..headers['Authorization'] = 'Bearer $token'
      ..headers['Idempotency-Key'] =
          '${capture.capturedAt.microsecondsSinceEpoch}-${capture.bytes.length}'
      ..fields['meta'] = jsonEncode(capture.toMeta())
      ..fields['role'] = capture.role
      ..files.add(http.MultipartFile.fromBytes(
        'media',
        capture.bytes,
        filename: capture.filename,
      ));

    final streamed = await _client.send(request);
    final response = await http.Response.fromStream(streamed);
    final decoded = _decode(response);
    return decoded['report_id'] as String;
  }

  Map<String, dynamic> _decode(http.Response response) {
    Map<String, dynamic> body;
    try {
      body = jsonDecode(response.body) as Map<String, dynamic>;
    } catch (_) {
      throw ApiException(response.statusCode,
          'The server replied with ${response.statusCode}.');
    }
    if (response.statusCode >= 400) {
      final error = body['error'];
      final message = error is Map<String, dynamic>
          ? (error['message'] as String? ?? 'Something went wrong.')
          : 'Something went wrong.';
      throw ApiException(response.statusCode, message);
    }
    return body;
  }
}
