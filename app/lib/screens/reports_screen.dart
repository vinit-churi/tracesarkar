

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../api.dart';
import '../theme.dart';
import '../widgets/verdict.dart';

/// Everything this person has sent, newest first.
class ReportsScreen extends StatefulWidget {
  const ReportsScreen({super.key, required this.client, required this.session});

  final ApiClient client;
  final Session session;

  @override
  State<ReportsScreen> createState() => _ReportsScreenState();
}

class _ReportsScreenState extends State<ReportsScreen> {
  List<ReportDetail>? _reports;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _error = null);
    try {
      final list = await widget.client.reports(widget.session.token);
      if (!mounted) return;
      setState(() => _reports = list);
    } on ApiException catch (e) {
      if (mounted) setState(() => _error = e.message);
    } catch (_) {
      if (mounted) {
        setState(() => _error = 'Could not reach the server.');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;

    return Scaffold(
      appBar: AppBar(title: const Text('My reports')),
      body: RefreshIndicator(
        onRefresh: _load,
        child: Builder(builder: (context) {
          if (_error != null) {
            return ListView(
              padding: const EdgeInsets.all(Tokens.gutter),
              children: [Text(_error!, style: text.bodyLarge)],
            );
          }
          if (_reports == null) {
            return const Center(child: CircularProgressIndicator());
          }
          if (_reports!.isEmpty) {
            return ListView(
              padding: const EdgeInsets.all(32),
              children: [
                Text(
                  'Nothing sent yet.',
                  style: text.titleMedium,
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                Text(
                  'Photograph a problem and it will appear here with what we '
                  'worked out about it.',
                  style: text.bodySmall,
                  textAlign: TextAlign.center,
                ),
              ],
            );
          }

          return ListView.separated(
            padding: const EdgeInsets.all(Tokens.gutter),
            itemCount: _reports!.length,
            separatorBuilder: (_, __) => const SizedBox(height: 10),
            itemBuilder: (_, i) => _ReportRow(
              report: _reports![i],
              client: widget.client,
              session: widget.session,
            ),
          );
        }),
      ),
    );
  }
}

class _ReportRow extends StatelessWidget {
  const _ReportRow({
    required this.report,
    required this.client,
    required this.session,
  });

  final ReportDetail report;
  final ApiClient client;
  final Session session;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;

    return InkWell(
      borderRadius: BorderRadius.circular(Tokens.cardRadius),
      onTap: () => Navigator.of(context).push(MaterialPageRoute(
        builder: (_) => ReportScreen(
          client: client, session: session, reportId: report.id),
      )),
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(Tokens.gutter),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(report.summary, style: text.labelLarge),
                    const SizedBox(height: 2),
                    Text(
                      [
                        if (report.ward != null && report.ward!.isNotEmpty)
                          '${report.authority ?? 'BMC'} · ${report.ward}',
                        if (report.createdAt != null)
                          report.createdAt!.substring(0, 10),
                      ].join(' · '),
                      style: text.bodySmall,
                    ),
                    if (report.contractorName != null &&
                        report.contractorName!.isNotEmpty) ...[
                      const SizedBox(height: 2),
                      Text('Under contract', style: text.bodySmall),
                    ],
                  ],
                ),
              ),
              const Icon(Icons.chevron_right, color: Tokens.ink45),
            ],
          ),
        ),
      ),
    );
  }
}

/// One capture, and everything concluded about it.
class ReportScreen extends StatefulWidget {
  const ReportScreen({
    super.key,
    required this.client,
    required this.session,
    required this.reportId,
  });

  final ApiClient client;
  final Session session;
  final String reportId;

  @override
  State<ReportScreen> createState() => _ReportScreenState();
}

class _ReportScreenState extends State<ReportScreen> {
  ReportDetail? _report;
  Uint8List? _photo;
  bool _photoFailed = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final r = await widget.client.report(widget.reportId, widget.session.token);
      if (!mounted) return;
      setState(() => _report = r);

      // Fetched separately because the archive is not public and the bytes
      // need the token — and in its own try, because a photograph that will
      // not load must not take the verdict with it. Everything the platform
      // worked out is still worth reading without it.
      try {
        final bytes =
            await widget.client.media(widget.reportId, widget.session.token);
        if (!mounted) return;
        setState(() => _photo = bytes);
      } catch (_) {
        if (mounted) setState(() => _photoFailed = true);
      }
    } on ApiException catch (e) {
      if (mounted) setState(() => _error = e.message);
    } catch (_) {
      if (mounted) setState(() => _error = 'Could not reach the server.');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Report')),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 560),
          child: ListView(
            padding: const EdgeInsets.all(Tokens.gutter),
            children: [
              if (_error != null)
                Text(_error!, style: Theme.of(context).textTheme.bodyLarge)
              else if (_report == null)
                const Center(child: CircularProgressIndicator())
              else ...[
                _PhotoAndPlace(
                    report: _report!, photo: _photo, failed: _photoFailed),
                const SizedBox(height: 16),
                VerdictCard(report: _report!),
                if (_report!.nextStep != null) ...[
                  const SizedBox(height: 16),
                  _NextStepCard(step: _report!.nextStep!),
                ],
                const SizedBox(height: 12),
                Text(
                  'Nothing has been filed with any authority. Any complaint or '
                  'RTI is a draft you review and submit yourself.',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}


/// The photograph, and where it was taken.
///
/// Shown above the verdict because it is what the person recognises — a list
/// of conclusions about a capture you cannot see is hard to trust and
/// impossible to correct. The position is theirs and this screen is theirs,
/// so it is exact here; public surfaces coarsen it (hard rule 4).
class _PhotoAndPlace extends StatelessWidget {
  const _PhotoAndPlace({
    required this.report,
    required this.photo,
    this.failed = false,
  });

  final ReportDetail report;
  final Uint8List? photo;
  final bool failed;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(Tokens.cardRadius),
          child: AspectRatio(
            aspectRatio: 4 / 3,
            child: photo != null
                ? Image.memory(photo!, fit: BoxFit.cover)
                : Container(
                    color: Tokens.evidence,
                    alignment: Alignment.center,
                    child: failed
                        ? const Padding(
                            padding: EdgeInsets.all(16),
                            child: Text(
                              'The photograph could not be loaded. Everything '
                              'below was still worked out from it.',
                              textAlign: TextAlign.center,
                              style: TextStyle(color: Tokens.ink45, fontSize: 13),
                            ),
                          )
                        : const SizedBox(
                            height: 22,
                            width: 22,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                  ),
          ),
        ),
        if (report.hasLocation) ...[
          const SizedBox(height: 8),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Icon(Icons.place_outlined, size: 16, color: Tokens.ink45),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  '${report.latitude!.toStringAsFixed(5)}, '
                  '${report.longitude!.toStringAsFixed(5)}'
                  '${report.accuracyMetres != null ? ' · ±${report.accuracyMetres!.round()} m' : ''}',
                  style: text.bodySmall?.copyWith(color: Tokens.ink45),
                ),
              ),
            ],
          ),
        ],
      ],
    );
  }
}


/// What to do about this capture.
///
/// The thing that separates this from an app that collects complaints and
/// forwards them: it names the desk, gives the words, and says what obliges
/// them once told. Nothing here sends anything — hard rule 1 — so the action
/// is Copy, and the person sends it.
class _NextStepCard extends StatelessWidget {
  const _NextStepCard({required this.step});

  final NextStep step;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Tokens.surface,
        borderRadius: BorderRadius.circular(Tokens.cardRadius),
        border: Border.all(color: Tokens.hairline),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('What to do next', style: text.titleMedium),
          const SizedBox(height: 10),

          if (!step.available) ...[
            Text(step.whatHappensNext,
                style: text.bodyLarge?.copyWith(color: Tokens.ink70)),
          ] else ...[
            for (final c in step.channels) ...[
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Icon(Icons.arrow_forward, size: 16, color: Tokens.accent),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(c.name,
                              style: text.bodyLarge?.copyWith(fontWeight: FontWeight.w600)),
                          Text(c.how,
                              style: text.bodySmall?.copyWith(color: Tokens.ink70)),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ],

            const SizedBox(height: 6),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Tokens.evidence,
                borderRadius: BorderRadius.circular(Tokens.chipRadius),
              ),
              child: Text(step.text ?? '', style: text.bodySmall),
            ),
            const SizedBox(height: 10),
            OutlinedButton.icon(
              onPressed: () => Clipboard.setData(ClipboardData(text: step.text ?? '')),
              icon: const Icon(Icons.copy, size: 18),
              label: const Text('Copy the complaint text'),
            ),

            const SizedBox(height: 14),
            Text(step.whatHappensNext,
                style: text.bodyLarge?.copyWith(color: Tokens.ink70)),
            if (step.deadlineCitation != null &&
                step.deadlineCitation!.isNotEmpty) ...[
              const SizedBox(height: 8),
              // The source travels with the constant (hard rule 2).
              Text('Source · ${step.deadlineCitation}',
                  style: text.bodySmall?.copyWith(color: Tokens.ink45)),
            ],
          ],
        ],
      ),
    );
  }
}
