package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	commondto "pulsa2/internal/dto/common"
	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
	"pulsa2/internal/service"
)

type AppAdAdminController struct {
	svc  *service.AppAdService
	base string
}

func NewAppAdAdminController(svc *service.AppAdService, base string) *AppAdAdminController {
	return &AppAdAdminController{svc: svc, base: strings.TrimRight(base, "/")}
}

func (h *AppAdAdminController) Handle(w http.ResponseWriter, r *http.Request) {
	base := h.base + "/app-ads"
	if r.URL.Path == base {
		h.handleCollection(w, r)
		return
	}

	id, ok := helper.ParseIDFromPath(r.URL.Path, base)
	if !ok {
		helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("not found"))
		return
	}
	h.handleItem(w, r, id)
}

func (h *AppAdAdminController) handleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := h.svc.List(r.Context(), helper.QueryString(r, "q"))
		if err != nil {
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
	case http.MethodPost:
		in, ok := h.decodeRequest(w, r, 0)
		if !ok {
			return
		}
		id, err := h.svc.Create(r.Context(), in)
		if err != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}

func (h *AppAdAdminController) handleItem(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		row, err := h.svc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("iklan not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "item": row})
	case http.MethodPut:
		in, ok := h.decodeRequest(w, r, id)
		if !ok {
			return
		}
		if err := h.svc.Update(r.Context(), in); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("iklan not found"))
				return
			}
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	case http.MethodDelete:
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helper.WriteJSON(w, http.StatusNotFound, commondto.MapError("iklan not found"))
				return
			}
			helper.WriteJSON(w, http.StatusInternalServerError, commondto.MapError(err.Error()))
			return
		}
		helper.WriteJSON(w, http.StatusOK, commondto.MapOK())
	default:
		helper.WriteJSON(w, http.StatusMethodNotAllowed, commondto.MapError("method not allowed"))
	}
}

func (h *AppAdAdminController) decodeRequest(w http.ResponseWriter, r *http.Request, id int64) (repository.AppAdUpsertInput, bool) {
	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") {
		return h.decodeMultipartRequest(w, r, id)
	}

	var req struct {
		Judul      string `json:"judul"`
		Keterangan string `json:"keterangan"`
		ImageURL   string `json:"image_url"`
		LinkURL    string `json:"link_url"`
		Urutan     int    `json:"urutan"`
		Aktif      bool   `json:"aktif"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("invalid json"))
		return repository.AppAdUpsertInput{}, false
	}
	return repository.AppAdUpsertInput{
		ID:         id,
		Judul:      req.Judul,
		Keterangan: req.Keterangan,
		ImageURL:   req.ImageURL,
		LinkURL:    req.LinkURL,
		Urutan:     req.Urutan,
		Aktif:      req.Aktif,
	}, true
}

func (h *AppAdAdminController) decodeMultipartRequest(w http.ResponseWriter, r *http.Request, id int64) (repository.AppAdUpsertInput, bool) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("multipart form tidak valid"))
		return repository.AppAdUpsertInput{}, false
	}

	imageURL := strings.TrimSpace(r.FormValue("image_url"))
	deleteOldImage := strings.TrimSpace(strings.ToLower(r.FormValue("delete_old_image"))) == "1"

	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		savedURL, saveErr := helper.SaveAppAdUpload(file, header)
		if saveErr != nil {
			helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError(saveErr.Error()))
			return repository.AppAdUpsertInput{}, false
		}
		imageURL = savedURL
		deleteOldImage = true
	} else if !errors.Is(err, http.ErrMissingFile) {
		helper.WriteJSON(w, http.StatusBadRequest, commondto.MapError("gagal membaca file gambar"))
		return repository.AppAdUpsertInput{}, false
	}

	if id > 0 && deleteOldImage {
		existing, getErr := h.svc.Get(r.Context(), id)
		if getErr == nil && existing != nil && existing.ImageURL != "" && existing.ImageURL != imageURL {
			_ = helper.DeleteManagedAppAdUpload(existing.ImageURL)
		}
	}

	urutan, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("urutan")))

	return repository.AppAdUpsertInput{
		ID:         id,
		Judul:      r.FormValue("judul"),
		Keterangan: r.FormValue("keterangan"),
		ImageURL:   imageURL,
		LinkURL:    r.FormValue("link_url"),
		Urutan:     urutan,
		Aktif:      strings.TrimSpace(strings.ToLower(r.FormValue("aktif"))) == "true" || r.FormValue("aktif") == "1" || strings.TrimSpace(strings.ToLower(r.FormValue("aktif"))) == "on",
	}, true
}
