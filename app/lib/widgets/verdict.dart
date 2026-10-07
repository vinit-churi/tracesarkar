import 'package:flutter/material.dart';

import '../api.dart';
import '../theme.dart';

/// What the platform concluded about one capture.
///
/// It renders honestly at every stage. A capture that has only been stored,
/// one classified but not yet routed, and one with no contract for its stretch
/// are all complete states rather than errors — roughly half of Borivali has
/// no contract data, so "we don't have that" is the common answer and must not
/// read as a failure.
class VerdictCard extends StatelessWidget {
  const VerdictCard({super.key, required this.report});

  final ReportDetail report;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;

    if (!report.enriched) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(Tokens.gutter),
          child: Row(
            children: [
              const SizedBox(
                height: 18,
                width: 18,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  'Saved. Working out where this is and who looks after it.',
                  style: text.bodyLarge,
                ),
              ),
            ],
          ),
        ),
      );
    }

    final hasContract = report.canNameContractor;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(Tokens.gutter),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _OutcomeChip(outcome: report.outcome ?? ''),
            const SizedBox(height: 10),
            Text(report.summary, style: text.titleMedium),
            if (report.rationale != null && report.rationale!.isNotEmpty) ...[
              const SizedBox(height: 4),
              Text(report.rationale!, style: text.bodySmall),
            ],
            if (report.confidence != null) ...[
              const SizedBox(height: 6),
              Text(
                'Classified automatically · '
                '${(report.confidence! * 100).round()}% confident · '
                'change it if this is wrong',
                style: text.bodySmall,
              ),
            ],
            if (report.ward != null && report.ward!.isNotEmpty) ...[
              const SizedBox(height: 14),
              _Statement(
                label: 'Who looks after this',
                body: '${report.authority ?? 'BMC'} · ${report.ward} ward',
                note: report.wardBasis,
              ),
            ],
            if (report.outsideWards) ...[
              const SizedBox(height: 14),
              _Statement(
                label: 'Who looks after this',
                body: 'Outside the wards we cover',
                note: report.wardBasis,
              ),
            ],
            const SizedBox(height: 12),
            // Contract coverage is a statement about a BMC ward's roads. For a
            // point in no ward we hold there is nothing true to say about it.
            if (report.outsideWards)
              const SizedBox.shrink()
            else if (hasContract)
              _Statement(
                label: 'This stretch is under contract',
                body: report.contractorName!,
                note: [
                  if (report.roadName != null && report.roadName!.isNotEmpty)
                    report.roadName!,
                  'Source · ${report.contractSource!} · '
                      'retrieved ${_day(report.contractRetrievedAt!)}',
                ].join(' · '),
              )
            else
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Tokens.evidence,
                  borderRadius: BorderRadius.circular(Tokens.chipRadius),
                ),
                child: Text(
                  'We don’t have contract data for this stretch yet. '
                  'About half of the ward’s roads are covered so far.',
                  style: text.bodySmall,
                ),
              ),
          ],
        ),
      ),
    );
  }
}

const _months = [
  'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
  'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
];

/// "5 Oct 2026", in India time — the date a reader in Mumbai would give it.
String _day(DateTime t) {
  final ist = t.toUtc().add(const Duration(hours: 5, minutes: 30));
  return '${ist.day} ${_months[ist.month - 1]} ${ist.year}';
}

class _Statement extends StatelessWidget {
  const _Statement({required this.label, required this.body, this.note});

  final String label;
  final String body;
  final String? note;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: const BoxDecoration(
        color: Tokens.evidence,
        border: Border(left: BorderSide(color: Tokens.accent, width: 3)),
        borderRadius: BorderRadius.only(
          topRight: Radius.circular(Tokens.chipRadius),
          bottomRight: Radius.circular(Tokens.chipRadius),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label.toUpperCase(),
            style: text.bodySmall?.copyWith(
              color: Tokens.accent,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.8,
              fontSize: 10,
            ),
          ),
          const SizedBox(height: 4),
          Text(body, style: text.bodyLarge),
          if (note != null && note!.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(note!, style: text.bodySmall),
          ],
        ],
      ),
    );
  }
}

class _OutcomeChip extends StatelessWidget {
  const _OutcomeChip({required this.outcome});

  final String outcome;

  @override
  Widget build(BuildContext context) {
    // Colour never carries meaning alone: every state has an icon and a label.
    final (Color bg, Color fg, String label, IconData icon) = switch (outcome) {
      'accepted' => (
          Tokens.confirmWash,
          Tokens.confirm,
          'Recorded',
          Icons.check_circle_outline
        ),
      'needs_confirmation' => (
          Tokens.cautionWash,
          Tokens.caution,
          'Needs a quick check',
          Icons.help_outline
        ),
      'needs_retake' => (
          Tokens.cautionWash,
          Tokens.caution,
          'Another photograph would help',
          Icons.photo_camera_outlined
        ),
      'not_yet_covered' => (
          Tokens.evidence,
          Tokens.ink70,
          'Not covered here yet',
          Icons.info_outline
        ),
      'out_of_scope' => (
          Tokens.evidence,
          Tokens.ink70,
          'Outside what we cover',
          Icons.info_outline
        ),
      _ => (Tokens.evidence, Tokens.ink70, 'Working on it', Icons.info_outline),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(Tokens.chipRadius),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: fg),
          const SizedBox(width: 6),
          Text(
            label,
            style: Theme.of(context).textTheme.labelLarge?.copyWith(color: fg),
          ),
        ],
      ),
    );
  }
}
