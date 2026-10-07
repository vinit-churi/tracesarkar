import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tracesarkar_app/api.dart';
import 'package:tracesarkar_app/widgets/verdict.dart';

// Hard rule 2: a contractor's name is shown with its source and the time that
// source was retrieved, or it is not shown at all.
Map<String, dynamic> roadJson({bool retrieved = true, bool source = true}) => {
      'id': 'r1',
      'status': 'pending',
      'classification': {
        'category': 'road_defect',
        'subcategory': 'pothole',
        'outcome': 'accepted',
        'confidence': 0.9,
      },
      'ward': 'R/C',
      'authority': 'BMC',
      'ward_confidence': 'high',
      'contractor_name': 'M/s Example Infracon Pvt. Ltd',
      'road_name': 'S.V. Road',
      if (source) 'contract_source': 'bmc-roads-api',
      if (retrieved) 'contract_retrieved_at': '2026-10-05T03:00:00Z',
    };

Future<void> pump(WidgetTester tester, Map<String, dynamic> json) =>
    tester.pumpWidget(MaterialApp(
      home: Scaffold(
        body: SingleChildScrollView(
          child: VerdictCard(report: ReportDetail.fromJson(json)),
        ),
      ),
    ));

void main() {
  testWidgets('the contractor is shown with its source and retrieval date',
      (tester) async {
    await pump(tester, roadJson());
    expect(find.text('M/s Example Infracon Pvt. Ltd'), findsOneWidget);
    expect(find.textContaining('Source · bmc-roads-api'), findsOneWidget);
    expect(find.textContaining('retrieved 5 Oct 2026'), findsOneWidget);
  });

  testWidgets('no retrieval time, no name', (tester) async {
    await pump(tester, roadJson(retrieved: false));
    expect(find.text('M/s Example Infracon Pvt. Ltd'), findsNothing);
  });

  testWidgets('no source, no name', (tester) async {
    await pump(tester, roadJson(source: false));
    expect(find.text('M/s Example Infracon Pvt. Ltd'), findsNothing);
  });
}
