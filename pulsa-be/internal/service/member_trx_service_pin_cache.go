package service

import (
	"crypto/sha256"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type pinCacheKey struct {
	memberID int64
	pinHash  string
	pinSum   [32]byte
}

const memberPinCacheTTL = 5 * time.Minute

func (h *MemberTrxService) verifyMemberPIN(memberID int64, pinHash string, pin string) bool {
	if h == nil || memberID <= 0 || pinHash == "" || pin == "" {
		return false
	}
	key := pinCacheKey{memberID: memberID, pinHash: pinHash, pinSum: sha256.Sum256([]byte(pin))}
	now := time.Now()
	if h.isPINCacheHit(key, now) {
		return true
	}

	actual, _ := h.pinLocks.LoadOrStore(key, new(sync.Mutex))
	mu := actual.(*sync.Mutex)
	mu.Lock()
	defer func() {
		mu.Unlock()
		h.pinLocks.Delete(key)
	}()

	now = time.Now()
	if h.isPINCacheHit(key, now) {
		return true
	}
	if bcrypt.CompareHashAndPassword([]byte(pinHash), []byte(pin)) != nil {
		return false
	}
	h.storePINCacheHit(key, now.Add(memberPinCacheTTL), now)
	return true
}

func (h *MemberTrxService) isPINCacheHit(key pinCacheKey, now time.Time) bool {
	h.pinMu.Lock()
	defer h.pinMu.Unlock()
	if h.pinCache == nil {
		return false
	}
	expiresAt, ok := h.pinCache[key]
	return ok && now.Before(expiresAt)
}

func (h *MemberTrxService) storePINCacheHit(key pinCacheKey, expiresAt time.Time, now time.Time) {
	h.pinMu.Lock()
	defer h.pinMu.Unlock()
	if h.pinCache == nil {
		h.pinCache = make(map[pinCacheKey]time.Time)
	}
	h.pinCache[key] = expiresAt
	if len(h.pinCache) < 10000 {
		return
	}
	for k, exp := range h.pinCache {
		if !now.Before(exp) {
			delete(h.pinCache, k)
		}
	}
}
