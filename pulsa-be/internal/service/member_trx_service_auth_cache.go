package service

import (
	"context"
	"crypto/sha256"
	"time"

	"pulsa2/internal/repository"
)

type authCacheKey struct {
	apiKeySum [32]byte
}

type authCacheEntry struct {
	expiresAt time.Time
	auth      repository.MemberAuth
}

type ipAllowCacheKey struct {
	memberID int64
	clientIP string
}

const memberAuthCacheTTL = 10 * time.Second

func (h *MemberTrxService) authByAPIKeyCached(ctx context.Context, apiKey string) (*repository.MemberAuth, error) {
	if h == nil || h.MemberRepo == nil {
		return nil, repository.ErrUnauthorized
	}

	key := authCacheKey{apiKeySum: sha256.Sum256([]byte(apiKey))}
	now := time.Now()

	h.authMu.Lock()
	if h.authCache != nil {
		if entry, ok := h.authCache[key]; ok && now.Before(entry.expiresAt) {
			auth := entry.auth
			h.authMu.Unlock()
			return &auth, nil
		}
	}
	h.authMu.Unlock()

	auth, err := h.MemberRepo.AuthByAPIKey(ctx, apiKey)
	if err != nil || auth == nil {
		return auth, err
	}

	h.authMu.Lock()
	if h.authCache == nil {
		h.authCache = make(map[authCacheKey]authCacheEntry)
	}
	h.authCache[key] = authCacheEntry{expiresAt: now.Add(memberAuthCacheTTL), auth: *auth}
	h.cleanupAuthCachesLocked(now)
	h.authMu.Unlock()

	return auth, nil
}

func (h *MemberTrxService) isIPAllowedForMemberCached(ctx context.Context, memberID int64, clientIP string) (bool, error) {
	if h == nil || h.MemberRepo == nil {
		return false, repository.ErrForbidden
	}

	key := ipAllowCacheKey{memberID: memberID, clientIP: clientIP}
	now := time.Now()

	h.authMu.Lock()
	if h.ipAllowCache != nil {
		if expiresAt, ok := h.ipAllowCache[key]; ok && now.Before(expiresAt) {
			h.authMu.Unlock()
			return true, nil
		}
	}
	h.authMu.Unlock()

	allowed, err := h.MemberRepo.IsIPAllowedForMember(ctx, memberID, clientIP)
	if err != nil || !allowed {
		return allowed, err
	}

	h.authMu.Lock()
	if h.ipAllowCache == nil {
		h.ipAllowCache = make(map[ipAllowCacheKey]time.Time)
	}
	h.ipAllowCache[key] = now.Add(memberAuthCacheTTL)
	h.cleanupAuthCachesLocked(now)
	h.authMu.Unlock()

	return true, nil
}

func (h *MemberTrxService) cleanupAuthCachesLocked(now time.Time) {
	if len(h.authCache) >= 10000 {
		for key, entry := range h.authCache {
			if !now.Before(entry.expiresAt) {
				delete(h.authCache, key)
			}
		}
	}
	if len(h.ipAllowCache) >= 10000 {
		for key, expiresAt := range h.ipAllowCache {
			if !now.Before(expiresAt) {
				delete(h.ipAllowCache, key)
			}
		}
	}
}
