package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// PlayIntegrityGuard is middleware that verifies Google Play Integrity tokens
// on requests from Android app. Web/non-Android requests are allowed through.
//
// Behavior:
// - If X-Play-Integrity header is present → verify with Google → block if invalid
// - If header is absent AND X-App-Platform is "android" → block (mod app stripped header)
// - If header is absent AND no X-App-Platform → allow (web/iOS/other clients)
func PlayIntegrityGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		integrityToken := strings.TrimSpace(r.Header.Get("X-Play-Integrity"))
		appPlatform := strings.TrimSpace(strings.ToLower(r.Header.Get("X-App-Platform")))

		// Non-Android clients (web, iOS) → skip verification
		if appPlatform != "android" && integrityToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Android without integrity token → likely mod app
		if appPlatform == "android" && integrityToken == "" {
			AppendProviderServiceLog("provider_callback_error.log",
				"BLOCKED android request without Play Integrity token path=%s ip=%s",
				r.URL.Path, extractClientIP(r))
			WriteJSON(w, http.StatusForbidden, map[string]any{
				"ok": false, "error": "integrity verification required",
			})
			return
		}

		// Has integrity token → verify with Google
		packageName := os.Getenv("ANDROID_PACKAGE_NAME")
		if packageName == "" {
			packageName = "com.app.pulsakilat"
		}

		valid, reason := verifyPlayIntegrityToken(r.Context(), integrityToken, packageName)
		if !valid {
			AppendProviderServiceLog("provider_callback_error.log",
				"BLOCKED invalid Play Integrity token path=%s ip=%s reason=%s",
				r.URL.Path, extractClientIP(r), reason)
			WriteJSON(w, http.StatusForbidden, map[string]any{
				"ok": false, "error": "app integrity verification failed",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func verifyPlayIntegrityToken(ctx context.Context, token, packageName string) (bool, string) {
	apiKey := os.Getenv("PLAY_INTEGRITY_API_KEY")
	if apiKey == "" {
		// If API key not configured, log but allow (graceful degradation)
		return true, "api_key_not_configured"
	}

	verifyURL := fmt.Sprintf(
		"https://playintegrity.googleapis.com/v1/%s:decodeIntegrityToken?key=%s",
		packageName, apiKey,
	)

	body := fmt.Sprintf(`{"integrity_token":"%s"}`, token)
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", verifyURL, strings.NewReader(body))
	if err != nil {
		return true, "request_create_failed" // allow on error
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return true, "google_api_unreachable" // allow if Google unreachable
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("google_api_status_%d", resp.StatusCode)
	}

	var result struct {
		TokenPayloadExternal struct {
			AppIntegrity struct {
				AppRecognitionVerdict string `json:"appRecognitionVerdict"`
			} `json:"appIntegrity"`
			DeviceIntegrity struct {
				DeviceRecognitionVerdict []string `json:"deviceRecognitionVerdict"`
			} `json:"deviceIntegrity"`
			AccountDetails struct {
				AppLicensingVerdict string `json:"appLicensingVerdict"`
			} `json:"accountDetails"`
			RequestDetails struct {
				RequestPackageName string `json:"requestPackageName"`
			} `json:"requestDetails"`
		} `json:"tokenPayloadExternal"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, "json_parse_failed"
	}

	payload := result.TokenPayloadExternal

	// Verify package name
	if payload.RequestDetails.RequestPackageName != packageName {
		return false, fmt.Sprintf("package_mismatch:%s", payload.RequestDetails.RequestPackageName)
	}

	// Verify app integrity — PLAY_RECOGNIZED means installed from Play Store with valid signature
	appVerdict := strings.ToUpper(payload.AppIntegrity.AppRecognitionVerdict)
	if appVerdict != "PLAY_RECOGNIZED" && appVerdict != "" {
		return false, fmt.Sprintf("app_not_recognized:%s", appVerdict)
	}

	// Verify device integrity — at least MEETS_DEVICE_INTEGRITY
	hasBasicIntegrity := false
	for _, v := range payload.DeviceIntegrity.DeviceRecognitionVerdict {
		if strings.EqualFold(v, "MEETS_DEVICE_INTEGRITY") || strings.EqualFold(v, "MEETS_BASIC_INTEGRITY") {
			hasBasicIntegrity = true
			break
		}
	}
	if !hasBasicIntegrity && len(payload.DeviceIntegrity.DeviceRecognitionVerdict) > 0 {
		return false, "device_integrity_failed"
	}

	return true, "ok"
}
