import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/widgets/verdict.dart';

// The first capture taken past Dahisar came back from jurisdiction with no
// ward — "outside every ward boundary we hold" — and the app read the missing
// ward as "still working", so it showed a spinner for a day. Outside the map is
// an answer, not a pending state.
const outsideBasis = 'This point is outside every ward boundary we hold. '
    'Greater Mumbai is not the whole metropolitan region, so it may belong '
    'to a neighbouring corporation.';

Map<String, dynamic> outsideJson() => {
      'id': 'r1',
      'status': 'pending',
      'created_at': '2026-10-06T12:49:00Z',
      'classification': {
        'category': 'road_defect',
        'subcategory': 'missing manhole cover',
        'outcome': 'needs_confirmation',
        'confidence': 0.5,
      },
      'ward_confidence': 'none',
      'ward_basis': outsideBasis,
      'needs_question': true,
    };

void main() {
  test('a point outside every ward is finished, not pending', () {
    final d = ReportDetail.fromJson(outsideJson());
    expect(d.enriched, isTrue,
        reason: 'jurisdiction answered; the answer is that we hold no ward here');
    expect(d.outsideWards, isTrue);
  });

  test('a classified capture whose jurisdiction has not run is still pending',
      () {
    final json = outsideJson()
      ..remove('ward_confidence')
      ..remove('ward_basis')
      ..remove('needs_question');
    final d = ReportDetail.fromJson(json);
    expect(d.enriched, isFalse);
    expect(d.outsideWards, isFalse);
  });

  test('a capture with a ward is not outside the wards', () {
    final d = ReportDetail.fromJson(outsideJson()
      ..['ward'] = 'R/C'
      ..['authority'] = 'BMC'
      ..['ward_confidence'] = 'high');
    expect(d.enriched, isTrue);
    expect(d.outsideWards, isFalse);
  });

  testWidgets('the verdict says the point is outside the wards we cover',
      (tester) async {
    await tester.pumpWidget(MaterialApp(
      home: Scaffold(
        body: SingleChildScrollView(
          child: VerdictCard(report: ReportDetail.fromJson(outsideJson())),
        ),
      ),
    ));

    expect(find.byType(CircularProgressIndicator), findsNothing,
        reason: 'a finished answer must not look like work in progress');
    expect(find.textContaining('Outside the wards we cover'), findsOneWidget);
    expect(find.textContaining('outside every ward boundary we hold'),
        findsOneWidget,
        reason: 'the basis travels with the answer');

    // "About half of the ward's roads are covered" is a claim about a BMC
    // ward. Saying it about a point in no ward we hold would be wrong.
    expect(find.textContaining('half of the ward'), findsNothing);
  });
}
