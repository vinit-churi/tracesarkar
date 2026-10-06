import 'dart:typed_data';

import 'package:flutter/material.dart';

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
      // The photograph is the report. It is fetched separately because the
      // archive is not public and the bytes need the token.
      final bytes =
          await widget.client.media(widget.reportId, widget.session.token);
      if (!mounted) return;
      setState(() => _photo = bytes);
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
                _PhotoAndPlace(report: _report!, photo: _photo),
                const SizedBox(height: 16),
                VerdictCard(report: _report!),
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
  const _PhotoAndPlace({required this.report, required this.photo});

  final ReportDetail report;
  final Uint8List? photo;

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
            child: photo == null
                ? Container(
                    color: Tokens.evidence,
                    alignment: Alignment.center,
                    child: const SizedBox(
                      height: 22,
                      width: 22,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  )
                : Image.memory(photo!, fit: BoxFit.cover),
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
