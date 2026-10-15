package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// EscalationContact is an office a complaint can be taken to when the desk
// below it has not acted, with a published way to reach it.
//
// Every field here describes a post, never a person. The source names the
// individual holding each one and prints their personal mobile; neither is
// stored. A post outlives whoever holds it, and several of those entries
// already carry a retirement date.
type EscalationContact struct {
	Designation   string
	OfficeEmail   string
	ControlRoom   string
	OfficeAddress string
	Zone          string
	SourceRef     string
	RetrievedAt   time.Time
}

// Escalation is the ladder above a ward desk: the ward's Assistant
// Commissioner, then the Deputy Municipal Commissioner for the zone that ward
// belongs to.
type Escalation struct {
	Ward EscalationContact
	Zone EscalationContact
}

// EscalationFor returns both rungs above a ward desk in one round trip.
//
// A ward with no row is not an error: it means the register has not been
// extended to that authority yet, and the caller gets zero values it can test
// for. An error here is a database failure, nothing else.
func (d *DB) EscalationFor(ctx context.Context, authority, ward string) (Escalation, error) {
	var e Escalation

	err := d.pool.QueryRow(ctx, `
		SELECT COALESCE(w.designation, ''), COALESCE(w.office_email, ''),
		       COALESCE(w.control_room, ''), COALESCE(w.office_address, ''),
		       COALESCE(w.zone, ''), COALESCE(w.source_ref, ''),
		       COALESCE(w.retrieved_at, 'epoch'::timestamptz),
		       COALESCE(z.designation, ''), COALESCE(z.office_email, ''),
		       COALESCE(z.zone, ''), COALESCE(z.source_ref, ''),
		       COALESCE(z.retrieved_at, 'epoch'::timestamptz)
		  FROM ward_escalation_contacts w
		  LEFT JOIN ward_escalation_contacts z
		         ON z.authority = w.authority
		        AND z.tier      = 'zone'
		        AND z.zone      = w.zone
		 WHERE w.authority = $1
		   AND w.ward      = $2
		   AND w.tier      = 'ward'
		   AND w.effective_from <= CURRENT_DATE`, authority, ward).
		Scan(&e.Ward.Designation, &e.Ward.OfficeEmail, &e.Ward.ControlRoom,
			&e.Ward.OfficeAddress, &e.Ward.Zone, &e.Ward.SourceRef,
			&e.Ward.RetrievedAt,
			&e.Zone.Designation, &e.Zone.OfficeEmail, &e.Zone.Zone,
			&e.Zone.SourceRef, &e.Zone.RetrievedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Escalation{}, nil
		}
		return Escalation{}, fmt.Errorf("escalation for %s %s: %w", authority, ward, err)
	}
	return e, nil
}

// Department is one row of BMC's published department list: that the
// department exists, and where its Section 4 handbook is.
//
// It is a register, not a routing rule. Nothing here says which department
// answers for a given complaint.
type Department struct {
	NameEN      string
	NameMR      string
	RTIURL      string
	DocsURL     string
	SourceRef   string
	RetrievedAt time.Time
}

// Departments lists every department in the register, in English name order.
//
// Prefer RTIURL over DocsURL: BMC's Docs links frequently return the portal's
// homepage rather than the department's documents, because the inner view
// fails to render. The RTI link is the one that reliably resolves.
func (d *DB) Departments(ctx context.Context) ([]Department, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT name_en, COALESCE(name_mr, ''), COALESCE(rti_url, ''),
		       COALESCE(docs_url, ''), source_ref, retrieved_at
		  FROM bmc_departments
		 ORDER BY name_en`)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	defer rows.Close()

	var out []Department
	for rows.Next() {
		var dp Department
		if err := rows.Scan(&dp.NameEN, &dp.NameMR, &dp.RTIURL, &dp.DocsURL,
			&dp.SourceRef, &dp.RetrievedAt); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		out = append(out, dp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read departments: %w", err)
	}
	return out, nil
}
