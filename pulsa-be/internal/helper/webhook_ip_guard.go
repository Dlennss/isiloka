package helper

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	providerIPMap     map[string]string
	providerIPMapOnce sync.Once
)

func loadProviderIPMap() {
	providerIPMap = make(map[string]string)

	envMap := map[string]string{
		"yuscom":       "YUSCOM_BASE_URL",
		"javapay":      "JAVAPAY_BASE_URL",
		"talentapay":   "TALENTA_BASE_URL",
		"multikom":     "MULTIKOM_BASE_URL",
		"sagaramobile": "SAGARA_BASE_URL",
		"minions":      "MINIONS_BASE_URL",
		"trionik":      "TRIONIK_BASE_URL",
		"ajs":          "AJS_BASE_URL",
		"gemilang":     "GEMILANG_BASE_URL",
		"smb":          "SMB_BASE_URL",
		"loketbayar":   "LOKETBAYAR_BASE_URL",
		"chytron":      "CHYTRON_BASE_URL",
		"rajabiller":   "RAJABILLER_BASE_URL",
		"pulsa24jam":   "PULSA24JAM_BASE_URL",
	}

	for provider, envKey := range envMap {
		raw := os.Getenv(envKey)
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		host := u.Hostname()
		if host != "" {
			// If host is a domain (not IP), resolve to IP
			if ip := net.ParseIP(host); ip == nil {
				if addrs, err := net.LookupHost(host); err == nil && len(addrs) > 0 {
					providerIPMap[provider] = strings.Join(addrs, ",")
				} else {
					providerIPMap[provider] = host
				}
			} else {
				providerIPMap[provider] = host
			}
		}
	}

	// Javapay: override/append callback IP. Javapay callback datang dari IP asli
	// server (bukan lewat Cloudflare), jadi DNS resolve saja tidak cukup.
	if jp := os.Getenv("JAVAPAY_CALLBACK_IP"); jp != "" {
		providerIPMap["javapay"] = strings.TrimSpace(jp)
	} else if existing, ok := providerIPMap["javapay"]; ok {
		// Fallback: append known javapay server IP jika env tidak diset
		providerIPMap["javapay"] = existing + ",160.19.166.109"
	}

	if lb := os.Getenv("LOKETBAYAR_CALLBACK_IP"); lb != "" {
		providerIPMap["loketbayar"] = strings.TrimSpace(lb)
	}
	if ch := os.Getenv("CHYTRON_CALLBACK_IP"); ch != "" {
		providerIPMap["chytron"] = strings.TrimSpace(ch)
	}
	if rj := os.Getenv("RAJABILLER_CALLBACK_IP"); rj != "" {
		providerIPMap["rajabiller"] = strings.TrimSpace(rj)
	} else if existing, ok := providerIPMap["rajabiller"]; ok {
		providerIPMap["rajabiller"] = existing + ",34.128.119.54,34.128.94.169"
	} else {
		providerIPMap["rajabiller"] = "34.128.119.54,34.128.94.169"
	}
	if p24 := os.Getenv("PULSA24JAM_CALLBACK_IP"); p24 != "" {
		providerIPMap["pulsa24jam"] = strings.TrimSpace(p24)
	}

	// Extra IPs from env (comma-separated): PROVIDER_EXTRA_IPS=yuscom:1.2.3.4,javapay:5.6.7.8
	if extra := os.Getenv("PROVIDER_EXTRA_IPS"); extra != "" {
		for _, entry := range strings.Split(extra, ",") {
			parts := strings.SplitN(strings.TrimSpace(entry), ":", 2)
			if len(parts) == 2 {
				p := strings.TrimSpace(strings.ToLower(parts[0]))
				ip := strings.TrimSpace(parts[1])
				if existing, ok := providerIPMap[p]; ok {
					providerIPMap[p] = existing + "," + ip
				} else {
					providerIPMap[p] = ip
				}
			}
		}
	}
}

// ProviderIPGuard creates a middleware that only allows requests from the provider's known IP.
// The IP is extracted from the provider's BASE_URL in .env.
func ProviderIPGuard(provider string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerIPMapOnce.Do(loadProviderIPMap)

		allowedRaw, ok := providerIPMap[strings.ToLower(strings.TrimSpace(provider))]
		if !ok || allowedRaw == "" {
			// No IP configured for this provider, allow (backward compat)
			next(w, r)
			return
		}

		clientIP := extractClientIP(r)

		// Check against all allowed IPs for this provider
		for _, allowed := range strings.Split(allowedRaw, ",") {
			allowed = strings.TrimSpace(allowed)
			if allowed == clientIP {
				next(w, r)
				return
			}
		}

		// Also allow localhost/loopback (nginx proxy)
		if clientIP == "127.0.0.1" || clientIP == "::1" {
			next(w, r)
			return
		}

		AppendProviderServiceLog("provider_callback_error.log",
			"BLOCKED webhook from unauthorized IP=%s provider=%s allowed=%s path=%s",
			clientIP, provider, allowedRaw, r.URL.Path)
		http.Error(w, `{"ok":false,"error":"forbidden"}`, http.StatusForbidden)
	}
}

func extractClientIP(r *http.Request) string {
	if cf := r.Header.Get("CF-Connecting-IP"); cf != "" {
		return strings.TrimSpace(cf)
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
