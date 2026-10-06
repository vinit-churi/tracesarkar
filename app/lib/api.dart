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

/// What the platform concluded about one capture. Every part is optional:
/// enrichment runs after the photograph is safe, so a report may be stored
/// and not yet understood — and a report with a ward and no contract is a
/// complete answer, not a half-finished one.
class ReportDetail {
  ReportDetail({
    required this.id,
    required this.status,
    this.createdAt,
    this.category,
    this.subcategory,
    this.outcome,
    this.confidence,
    this.rationale,
    this.ward,
    this.authority,
    this.wardBasis,
    this.contractorName,
    this.roadName,
    this.contractBasis,
    this.contractSource,
    this.attributionConfidence,
    this.latitude,
    this.longitude,
    this.accuracyMetres,
  });

  factory ReportDetail.fromJson(Map<String, dynamic> json) {
    final c = json['classification'] as Map<String, dynamic>?;
    return ReportDetail(
      id: (json['id'] ?? '') as String,
      status: (json['status'] ?? '') as String,
      createdAt: json['created_at'] as String?,
      category: c?['category'] as String?,
      subcategory: c?['subcategory'] as String?,
      outcome: c?['outcome'] as String?,
      confidence: (c?['confidence'] as num?)?.toDouble(),
      rationale: c?['rationale'] as String?,
      ward: json['ward'] as String?,
      authority: json['authority'] as String?,
      wardBasis: json['ward_basis'] as String?,
      contractorName: json['contractor_name'] as String?,
      roadName: json['road_name'] as String?,
      contractBasis: json['contract_basis'] as String?,
      contractSource: json['contract_source'] as String?,
      attributionConfidence: json['attribution_confidence'] as String?,
      latitude: (json['lat'] as num?)?.toDouble(),
      longitude: (json['lon'] as num?)?.toDouble(),
      accuracyMetres: (json['accuracy_m'] as num?)?.toDouble(),
    );
  }

  final String id;
  final String status;
  final String? createdAt;
  final String? category;
  final String? subcategory;
  final String? outcome;
  final double? confidence;
  final String? rationale;
  final String? ward;
  final String? authority;
  final String? wardBasis;
  final String? contractorName;
  final String? roadName;
  final String? contractBasis;
  final String? contractSource;
  final String? attributionConfidence;

  /// Where the photograph was taken, and how far out that may be.
  ///
  /// Nullable rather than defaulted: 0,0 is a real place in the Gulf of Guinea
  /// and no capture was ever taken there. The accuracy travels with it because
  /// every confidence gate downstream reasons about it — a verdict that says
  /// "the point lies 9 m from this road" is unreadable without knowing the fix
  /// could be out by six.
  final double? latitude;
  final double? longitude;
  final double? accuracyMetres;

  bool get hasLocation => latitude != null && longitude != null;

  /// Whether the platform has finished thinking about this capture.
  bool get enriched => outcome != null && ward != null;

  /// A short line for a list row.
  String get summary {
    if (subcategory != null && subcategory!.isNotEmpty) return subcategory!;
    if (category != null && category!.isNotEmpty) {
      return category!.replaceAll('_', ' ');
    }
    return 'Working on it…';
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

  /// Reads one capture and everything concluded about it.
  Future<ReportDetail> report(String id, String token) async {
    final response = await _client.get(
      _url('/v1/reports/$id'),
      headers: {'Authorization': 'Bearer $token'},
    );
    final decoded = _decode(response);
    return ReportDetail.fromJson(decoded['report'] as Map<String, dynamic>);
  }

  /// Reads one capture's photograph.
  ///
  /// Fetched with the token rather than handed to an <img>: the archive bucket
  /// is not public, and this endpoint serves a capture only to the account that
  /// made it.
  Future<Uint8List> media(String id, String token) async {
    final response = await _client.get(
      _url('/v1/reports/$id/media'),
      headers: {'Authorization': 'Bearer $token'},
    );
    if (response.statusCode >= 400) {
      throw ApiException(response.statusCode, 'That photograph could not be loaded.');
    }
    return response.bodyBytes;
  }

  /// Reads the signed-in person's own captures, newest first.
  Future<List<ReportDetail>> reports(String token) async {
    final response = await _client.get(
      _url('/v1/reports'),
      headers: {'Authorization': 'Bearer $token'},
    );
    final decoded = _decode(response);
    final items = (decoded['reports'] as List<dynamic>? ?? const []);
    return items
        .map((e) => ReportDetail.fromJson(e as Map<String, dynamic>))
        .toList();
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
