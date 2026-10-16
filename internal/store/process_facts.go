package store

import (
	"fmt"
	"strconv"
	"strings"
)

// A process fact is a published figure describing how the thing in the
// photograph is lawfully done — what permission it needs, what the permission
// costs, how long the authority has to decide an application for it.
//
// It is **not** a deadline. Forty-five days is the time the Tree Authority has
// to decide an application to fell a tree; it is not the time it has to answer
// a citizen who photographs one coming down (D160). Keeping the two in
// separate maps is what stops the second from being read as the first, and
// `deadlineKeyFor` deliberately knows nothing about these keys.
//
// The sentences state the process and nothing about any party. A resident
// draws their own conclusion from a fee beside a stump; the platform does not
// draw it for them, which is hard rule 3.
type processFact struct {
	key  string
	form string
}

var processFacts = map[string][]processFact{
	"tree felling": {
		{key: "tree.felling_permission_days",
			form: "Felling or transplanting a tree needs the Tree Authority's " +
				"permission, and the Authority has %s to decide an application."},
		{key: "tree.felling_processing_fee_paise",
			form: "The published processing fee is %s per tree."},
		{key: "tree.felling_deposit_paise",
			form: "A deposit of %s per tree is published for felling."},
	},
}

// processFactsFor returns the facts bound to one thing inside a category.
// Nothing is bound to a whole category: a fee for felling a tree says nothing
// about a mangrove, and a category-wide match would print it anyway.
func processFactsFor(_, subcategory string) []processFact {
	return processFacts[strings.ToLower(strings.TrimSpace(subcategory))]
}

// formatConstantValue renders a stored constant in the unit its key names.
//
// Money is stored in paise (hard rule: `int64` paise, never a float) and has
// to be shown in rupees, grouped the way a reader here expects — ₹6,000, not
// ₹6000.00 and certainly not 600000. A value that will not parse returns
// empty, and an empty value is dropped rather than printed raw.
func formatConstantValue(key, value string) string {
	value = strings.Trim(strings.TrimSpace(value), `"`)
	switch {
	case strings.HasSuffix(key, "_paise"):
		paise, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ""
		}
		rupees, remainder := paise/100, paise%100
		out := "₹" + groupIndian(rupees)
		if remainder != 0 {
			out += fmt.Sprintf(".%02d", remainder)
		}
		return out
	case strings.HasSuffix(key, "_days"):
		if _, err := strconv.Atoi(value); err != nil {
			return ""
		}
		return value + " days"
	case strings.HasSuffix(key, "_hours"):
		if _, err := strconv.Atoi(value); err != nil {
			return ""
		}
		return value + " hours"
	}
	return value
}

// groupIndian writes a whole number of rupees with Indian digit grouping:
// the last three digits, then pairs. 1000000 becomes 10,00,000.
func groupIndian(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}
	head, tail := s[:len(s)-3], s[len(s)-3:]
	var parts []string
	for len(head) > 2 {
		parts = append([]string{head[len(head)-2:]}, parts...)
		head = head[:len(head)-2]
	}
	if head != "" {
		parts = append([]string{head}, parts...)
	}
	out := strings.Join(parts, ",") + "," + tail
	if neg {
		return "-" + out
	}
	return out
}
