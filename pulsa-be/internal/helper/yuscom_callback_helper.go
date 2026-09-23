package helper

import (
	"regexp"
	"strconv"
	"strings"
)

var reYuscomSaldoAfter = regexp.MustCompile(`(?i)\bSaldo\s+([0-9\.\,]+)\s*-\s*([0-9\.\,]+)\s*=\s*([0-9\.\,]+)`)
var reStockSaldoAfter = regexp.MustCompile(`(?i)\b(?:Saldo|Stok(?:\s+Pulsa)?)\s+([0-9\.\,]+)\s*-\s*([0-9\.\,]+)\s*=\s*([0-9\.\,]+)`)
var reYuscomTicket = regexp.MustCompile(`\bT#([0-9]{3,})\b`)

func ParseSaldoAfterFromYuscomMsg(msg string) (int64, bool) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return 0, false
	}
	m := reYuscomSaldoAfter.FindStringSubmatch(msg)
	if len(m) < 4 {
		return 0, false
	}
	s := strings.TrimSpace(m[3])
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func ParseSaldoAfterFromStockMsg(msg string) (int64, bool) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return 0, false
	}
	m := reStockSaldoAfter.FindStringSubmatch(msg)
	if len(m) < 4 {
		return 0, false
	}
	s := strings.TrimSpace(m[3])
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func ParseTicketFromYuscomMsg(msg string) string {
	m := reYuscomTicket.FindStringSubmatch(msg)
	if len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}
