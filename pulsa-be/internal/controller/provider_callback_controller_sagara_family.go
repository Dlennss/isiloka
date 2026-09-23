package controller

import (
	"net/http"

	"pulsa2/internal/helper"
)

func (h *ProviderCallbackController) CallbackSagara(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appendProviderLog("sagaramobile", r, []byte(r.URL.RawQuery))
		status, out := h.svc.ProcessSagaraCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
		helper.WriteJSON(w, status, out)
	case http.MethodPost:
		raw := readProviderCallbackBody(r)
		appendProviderLog("sagaramobile", r, raw)
		q := mergeCallbackQuery(raw, r.URL.Query())
		status, out := h.svc.ProcessSagaraCallback(r.Context(), string(raw), q)
		helper.WriteJSON(w, status, out)
	default:
		writeMethodNotAllowed(w)
	}
}

func (h *ProviderCallbackController) CallbackMinions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appendProviderLog("minions", r, []byte(r.URL.RawQuery))
		status, out := h.svc.ProcessMinionsCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
		helper.WriteJSON(w, status, out)
	case http.MethodPost:
		raw := readProviderCallbackBody(r)
		appendProviderLog("minions", r, raw)
		q := mergeCallbackQuery(raw, r.URL.Query())
		status, out := h.svc.ProcessMinionsCallback(r.Context(), string(raw), q)
		helper.WriteJSON(w, status, out)
	default:
		writeMethodNotAllowed(w)
	}
}

func (h *ProviderCallbackController) CallbackTrionik(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appendProviderLog("trionik", r, []byte(r.URL.RawQuery))
		status, out := h.svc.ProcessTrionikCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
		helper.WriteJSON(w, status, out)
	case http.MethodPost:
		raw := readProviderCallbackBody(r)
		appendProviderLog("trionik", r, raw)
		q := mergeCallbackQuery(raw, r.URL.Query())
		status, out := h.svc.ProcessTrionikCallback(r.Context(), string(raw), q)
		helper.WriteJSON(w, status, out)
	default:
		writeMethodNotAllowed(w)
	}
}

func (h *ProviderCallbackController) CallbackAJS(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appendProviderLog("ajs", r, []byte(r.URL.RawQuery))
		status, out := h.svc.ProcessAJSCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
		helper.WriteJSON(w, status, out)
	case http.MethodPost:
		raw := readProviderCallbackBody(r)
		appendProviderLog("ajs", r, raw)
		q := mergeCallbackQuery(raw, r.URL.Query())
		status, out := h.svc.ProcessAJSCallback(r.Context(), string(raw), q)
		helper.WriteJSON(w, status, out)
	default:
		writeMethodNotAllowed(w)
	}
}

func (h *ProviderCallbackController) CallbackGemilang(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appendProviderLog("gemilang", r, []byte(r.URL.RawQuery))
		status, out := h.svc.ProcessGemilangCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
		helper.WriteJSON(w, status, out)
	case http.MethodPost:
		raw := readProviderCallbackBody(r)
		appendProviderLog("gemilang", r, raw)
		q := mergeCallbackQuery(raw, r.URL.Query())
		status, out := h.svc.ProcessGemilangCallback(r.Context(), string(raw), q)
		helper.WriteJSON(w, status, out)
	default:
		writeMethodNotAllowed(w)
	}
}

func (h *ProviderCallbackController) CallbackSMB(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appendProviderLog("smb", r, []byte(r.URL.RawQuery))
		status, out := h.svc.ProcessSMBCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
		helper.WriteJSON(w, status, out)
	case http.MethodPost:
		raw := readProviderCallbackBody(r)
		appendProviderLog("smb", r, raw)
		q := mergeCallbackQuery(raw, r.URL.Query())
		status, out := h.svc.ProcessSMBCallback(r.Context(), string(raw), q)
		helper.WriteJSON(w, status, out)
	default:
		writeMethodNotAllowed(w)
	}
}

func (h *ProviderCallbackController) CallbackLoketBayar(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		appendProviderLog("loketbayar", r, []byte(r.URL.RawQuery))
		status, out := h.svc.ProcessLoketBayarCallback(r.Context(), r.URL.RawQuery, r.URL.Query())
		helper.WriteJSON(w, status, out)
	case http.MethodPost:
		raw := readProviderCallbackBody(r)
		appendProviderLog("loketbayar", r, raw)
		q := mergeCallbackQuery(raw, r.URL.Query())
		status, out := h.svc.ProcessLoketBayarCallback(r.Context(), string(raw), q)
		helper.WriteJSON(w, status, out)
	default:
		writeMethodNotAllowed(w)
	}
}
