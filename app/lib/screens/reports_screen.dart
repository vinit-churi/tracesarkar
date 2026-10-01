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
