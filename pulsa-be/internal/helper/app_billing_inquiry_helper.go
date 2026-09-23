package helper

import (
	"regexp"
	"strconv"
	"strings"
)

type AppBillingInquiryParsed struct {
	Provider        string
	ProviderStatus  string
	ProviderMessage string
	DisplayMessage  string
	CustomerName    string
	UsageLabel      string
	MeterType       string
	PeriodLabel     string
	MeterRange      string
	BillAmount      int64
	PenaltyAmount   int64
	AdminAmount     int64
	TotalAmount     int64
	TransactionTime string
	CanPay          bool
}

var (
	reBillingBillAmount  = regexp.MustCompile(`(?i)(?:data\s+tag|tagihan)\s*:\s*rp?\s*([\d\.]+)`)
	reBillingTagAmount   = regexp.MustCompile(`(?i)rp\s+tag\s*:\s*rp?\s*([\d\.]+)`)
	reBillingPenalty     = regexp.MustCompile(`(?i)denda\s*:\s*rp?\s*([\d\.]+)`)
	reBillingAdminAmount = regexp.MustCompile(`(?i)\badm(?:in)?\s*:\s*rp?\s*([\d\.]+)`)
	reBillingTotalAmount = regexp.MustCompile(`(?i)total\s+tag(?:ihan)?\s*:\s*rp?\s*([\d\.]+)`)
	reBillingUsage       = regexp.MustCompile(`(?i)pemakaian\s*:\s*([^|/\.]+)`)
	reBillingPLNName     = regexp.MustCompile(`(?i)data\s+cust\s*:\s*nama\s*:\s*([^/\.]+)`)
	reBillingPLNUsage    = regexp.MustCompile(`(?i)(TD\s*:\s*[^/]+/BLTH\s*:[^/]+/St\s*MTR\s*:[^,\.]+)`)
	reBillingPLNMeter    = regexp.MustCompile(`(?i)TD\s*:\s*([^/]+)`)
	reBillingPLNPeriod   = regexp.MustCompile(`(?i)BLTH\s*:\s*([^/]+)`)
	reBillingPLNStand    = regexp.MustCompile(`(?i)St\s*MTR\s*:\s*([^,/.]+)`)
	reBillingPLNTime     = regexp.MustCompile(`(?i)Waktu\s+Trx\s*:\s*([^/]+)`)
	reBillingFailedMsg   = regexp.MustCompile(`(?i)\bGAGAL\.\s*([^@]+)`)
)

func parsePLNTokenCheckInfo(msg string) (summary string, customerName string, meterType string, power string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "", "", "", ""
	}
	idx := strings.Index(strings.ToUpper(msg), "SN/REF:")
	if idx < 0 {
		return "", "", "", ""
	}
	tail := strings.TrimSpace(msg[idx+len("SN/Ref:"):])
	if tail == "" {
		return "", "", "", ""
	}
	for _, marker := range []string{".Sukses Trx", ".SUKSES TRX", " Sukses Trx", " SUKSES TRX", " Saldo ", " SALDO "} {
		if cut := strings.Index(strings.ToUpper(tail), strings.ToUpper(marker)); cut > 0 {
			tail = strings.TrimSpace(tail[:cut])
			break
		}
	}
	tail = strings.Trim(tail, " .")
	if tail == "" {
		return "", "", "", ""
	}
	parts := strings.Split(tail, "/")
	if len(parts) < 4 {
		return tail, "", "", ""
	}
	customerName = strings.TrimSpace(parts[1])
	meterType = strings.TrimSpace(parts[2])
	power = strings.Trim(strings.TrimSpace(parts[3]), ".")
	return strings.Join([]string{
		strings.TrimSpace(parts[0]),
		customerName,
		meterType,
		power,
	}, "/"), customerName, meterType, power
}

func ParseAppBillingInquiry(provider, providerStatus, providerMessage string) *AppBillingInquiryParsed {
	msg := strings.TrimSpace(providerMessage)
	if msg == "" {
		return nil
	}

	out := &AppBillingInquiryParsed{
		Provider:        strings.TrimSpace(strings.ToLower(provider)),
		ProviderStatus:  strings.TrimSpace(strings.ToLower(providerStatus)),
		ProviderMessage: msg,
		DisplayMessage:  msg,
	}

	if summary, customerName, meterType, power := parsePLNTokenCheckInfo(msg); summary != "" {
		out.DisplayMessage = summary
		out.CustomerName = customerName
		out.MeterType = meterType
		out.MeterRange = power
		out.UsageLabel = summary
	}

	if idx := strings.Index(strings.ToUpper(msg), "SN/REF:"); idx >= 0 {
		tail := strings.TrimSpace(msg[idx+len("SN/Ref:"):])
		if slashIdx := strings.Index(tail, "/"); slashIdx > 0 {
			name := strings.TrimSpace(tail[:slashIdx])
			if name != "" && out.CustomerName == "" {
				out.CustomerName = name
			}
		}
	}
	if out.CustomerName == "" {
		if match := reBillingPLNName.FindStringSubmatch(msg); len(match) > 1 {
			out.CustomerName = strings.TrimSpace(match[1])
		}
	}

	if match := reBillingUsage.FindStringSubmatch(msg); len(match) > 1 {
		out.UsageLabel = strings.TrimSpace(match[1])
	}
	if out.UsageLabel == "" {
		if match := reBillingPLNUsage.FindStringSubmatch(msg); len(match) > 1 {
			out.UsageLabel = strings.TrimSpace(match[1])
		}
	}
	if match := reBillingPLNMeter.FindStringSubmatch(msg); len(match) > 1 {
		out.MeterType = strings.TrimSpace(match[1])
	}
	if match := reBillingPLNPeriod.FindStringSubmatch(msg); len(match) > 1 {
		out.PeriodLabel = strings.TrimSpace(match[1])
	}
	if match := reBillingPLNStand.FindStringSubmatch(msg); len(match) > 1 {
		out.MeterRange = strings.TrimSpace(match[1])
	}
	if match := reBillingPLNTime.FindStringSubmatch(msg); len(match) > 1 {
		out.TransactionTime = strings.TrimSpace(match[1])
	}
	if match := reBillingBillAmount.FindStringSubmatch(msg); len(match) > 1 {
		out.BillAmount = parseBillingAmount(match[1])
	}
	if match := reBillingTagAmount.FindStringSubmatch(msg); len(match) > 1 {
		out.BillAmount = parseBillingAmount(match[1])
	}
	if match := reBillingPenalty.FindStringSubmatch(msg); len(match) > 1 {
		out.PenaltyAmount = parseBillingAmount(match[1])
	}
	if match := reBillingAdminAmount.FindStringSubmatch(msg); len(match) > 1 {
		out.AdminAmount = parseBillingAmount(match[1])
	}
	if match := reBillingTotalAmount.FindStringSubmatch(msg); len(match) > 1 {
		out.TotalAmount = parseBillingAmount(match[1])
	}
	if out.TotalAmount <= 0 && (out.BillAmount > 0 || out.AdminAmount > 0) {
		out.TotalAmount = out.BillAmount + out.AdminAmount
	}
	if out.ProviderStatus == "success" && out.TotalAmount > 0 {
		out.CanPay = true
	}
	if !out.CanPay {
		if idx := strings.Index(strings.ToLower(msg), "tagihan sudah terbayar"); idx >= 0 {
			out.DisplayMessage = "Tagihan sudah terbayar"
			return out
		}
		if match := reBillingFailedMsg.FindStringSubmatch(msg); len(match) > 1 {
			display := strings.TrimSpace(match[1])
			if idx := strings.Index(strings.ToLower(display), "saldo"); idx >= 0 {
				display = strings.TrimSpace(display[:idx])
			}
			display = strings.Trim(display, " .")
			if display != "" {
				out.DisplayMessage = display
			}
		}
	}
	return out
}

func parseBillingAmount(raw string) int64 {
	clean := strings.ReplaceAll(strings.TrimSpace(raw), ".", "")
	clean = strings.ReplaceAll(clean, ",", "")
	if clean == "" {
		return 0
	}
	n, _ := strconv.ParseInt(clean, 10, 64)
	return n
}
