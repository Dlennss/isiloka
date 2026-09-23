package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"pulsa2/internal/helper"
)

func (s *AuthService) LoginGoogle(ctx context.Context, email, nama, googleID, idToken string) (string, bool, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	nama = strings.TrimSpace(nama)
	googleID = strings.TrimSpace(googleID)
	idToken = strings.TrimSpace(idToken)
	if email == "" || googleID == "" || idToken == "" {
		return "", false, "", errors.New("email/google_id/id_token required")
	}

	info, err := verifyGoogleIDToken(ctx, idToken)
	if err != nil {
		return "", false, "", err
	}
	if strings.TrimSpace(info.Sub) == "" || info.Sub != googleID {
		return "", false, "", errors.New("google token subject mismatch")
	}
	if strings.TrimSpace(strings.ToLower(info.Email)) == "" || strings.TrimSpace(strings.ToLower(info.Email)) != email {
		return "", false, "", errors.New("google token email mismatch")
	}
	if !strings.EqualFold(strings.TrimSpace(info.EmailVerified), "true") {
		return "", false, "", errors.New("google email not verified")
	}

	row, err := s.repo.GetByGoogleSub(ctx, googleID)
	if err != nil {
		return "", false, "", err
	}

	created := false
	if row == nil {
		row, err = s.repo.GetByEmail(ctx, email)
		if err != nil {
			return "", false, "", err
		}
		if row != nil {
			if row.GoogleSub != "" && row.GoogleSub != googleID {
				return "", false, "", errors.New("email already linked to different google account")
			}
			if row.GoogleSub == "" {
				if err := s.repo.LinkGoogleSubByMemberID(ctx, row.ID, googleID); err != nil {
					return "", false, "", err
				}
				row.GoogleSub = googleID
			}
		}
	}

	if row == nil {
		row, err = s.repo.CreateGoogleMember(ctx, email, nama, googleID)
		if err != nil {
			return "", false, "", err
		}
		created = true
	}

	if !row.Aktif {
		return "", false, "", errors.New("member inactive")
	}

	role := helper.NormalizeRole(row.Role)
	if role == "" {
		role = helper.RoleUser
	}

	tok, err := helper.MakeJWT(s.jwtSecret, row.ID, role, 7*24*time.Hour)
	if err != nil {
		return "", false, "", err
	}
	return tok, created, role, nil
}

func verifyGoogleIDToken(ctx context.Context, idToken string) (*googleTokenInfo, error) {
	allowedClientIDs := collectAllowedGoogleClientIDs()
	if len(allowedClientIDs) == 0 {
		return nil, errors.New("GOOGLE_CLIENT_ID not configured")
	}

	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to verify google token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid google token")
	}

	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, errors.New("invalid google token response")
	}
	if !containsString(allowedClientIDs, strings.TrimSpace(info.Aud)) {
		return nil, errors.New("google token audience mismatch")
	}
	return &info, nil
}

func collectAllowedGoogleClientIDs() []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, 3)
	for _, raw := range []string{
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_WEB_CLIENT_ID"),
		os.Getenv("AUTH_GOOGLE_ID"),
		os.Getenv("GOOGLE_CLIENT_ID_LEGACY"),
		os.Getenv("GOOGLE_CLIENT_IDS"),
	} {
		for _, part := range strings.Split(raw, ",") {
			v := strings.TrimSpace(part)
			if v == "" {
				continue
			}
			if _, exists := seen[v]; exists {
				continue
			}
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

func containsString(items []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, item := range items {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}
