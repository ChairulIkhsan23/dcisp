package people

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Sumber data eksternal institusi yang didukung.
const (
	InstitutionSourceManual       = "MANUAL"
	InstitutionSourceAPIKampus    = "API_KAMPUS"
	InstitutionSourceAPISekolah   = "API_SEKOLAH"
	InstitutionSourceAPIINDONESIA = "API_INDONESIA"
)

// Respons external DTO dari API Indonesia untuk direktori kampus (GET /api/v1/kampus).
type APIIndonesiaKampus struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ShortName     *string `json:"short_name"`
	Jenis         string  `json:"jenis"`
	Kelompok      string  `json:"kelompok"`
	ProvinceID    string  `json:"province_id"`
	RegencyID     string  `json:"regency_id"`
	Address       *string `json:"address"`
	PostalCode    *string `json:"postal_code"`
	Website       *string `json:"website"`
	Accreditation *string `json:"accreditation"`
	IsActive      *int    `json:"is_active"`
	ProvinceName  *string `json:"province_name"`
	RegencyName   *string `json:"regency_name"`
}

// Respons external DTO dari API Indonesia untuk direktori sekolah (GET /api/v1/sekolah).
type APIIndonesiaSekolah struct {
	NPSN          string  `json:"npsn"`
	Name          string  `json:"name"`
	Jenis         string  `json:"jenis"`
	Status        string  `json:"status"`
	ProvinceID    string  `json:"province_id"`
	RegencyID     string  `json:"regency_id"`
	DistrictID    *string `json:"district_id"`
	VillageID     *string `json:"village_id"`
	Address       *string `json:"address"`
	PostalCode    *string `json:"postal_code"`
	Accreditation *string `json:"accreditation"`
	Phone         *string `json:"phone"`
	Email         *string `json:"email"`
	Website       *string `json:"website"`
}

// Amplop paginasi generik dari API Indonesia.
type APIIndonesiaMeta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

// Klien HTTP khusus untuk Public API API Indonesia (https://use.apiindonesia.id).
type APIIndonesiaClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Menginisialisasi klien HTTP API Indonesia dengan timeout ketat 10 detik.
func NewAPIIndonesiaClient(baseURL, apiKey string) *APIIndonesiaClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if trimmed == "" {
		trimmed = "https://use.apiindonesia.id"
	}
	return &APIIndonesiaClient{
		baseURL: trimmed,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Memeriksa ketersediaan kredensial API Indonesia sebelum melakukan permintaan eksternal.
func (c *APIIndonesiaClient) requireAPIKey() error {
	if strings.TrimSpace(c.apiKey) == "" {
		return fmt.Errorf("kredensial API Indonesia belum dikonfigurasi (API_INDONESIA_KEY kosong)")
	}
	return nil
}

// Mengeksekusi permintaan GET terautentikasi ke API Indonesia dan mengembalikan body mentah.
func (c *APIIndonesiaClient) doGet(ctx context.Context, path string, query url.Values) ([]byte, int, error) {
	if err := c.requireAPIKey(); err != nil {
		return nil, 0, err
	}
	fullURL := c.baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal membangun permintaan API Indonesia: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghubungi API Indonesia: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("gagal membaca respons API Indonesia: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, resp.StatusCode, mapAPIIndonesiaHTTPError(resp.StatusCode, body)
	}
	return body, resp.StatusCode, nil
}

// Memetakan kode status HTTP API Indonesia menjadi pesan kesalahan yang aman bagi klien.
func mapAPIIndonesiaHTTPError(statusCode int, body []byte) error {
	var errEnvelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &errEnvelope)
	detail := strings.TrimSpace(errEnvelope.Error.Message)
	if detail == "" {
		detail = strings.TrimSpace(string(body))
	}
	if len(detail) > 300 {
		detail = detail[:300]
	}
	switch statusCode {
	case 400:
		return fmt.Errorf("permintaan ke API Indonesia tidak valid: %s", detail)
	case 401:
		return fmt.Errorf("otentikasi API Indonesia gagal (kunci tidak valid atau dicabut)")
	case 402:
		return fmt.Errorf("kuota bulanan API Indonesia telah habis")
	case 403:
		return fmt.Errorf("akun API Indonesia ditangguhkan")
	case 404:
		return fmt.Errorf("data institusi tidak ditemukan pada API Indonesia")
	case 429:
		return fmt.Errorf("batas laju API Indonesia terlampaui, silakan coba lagi sesaat")
	default:
		if statusCode >= 500 {
			return fmt.Errorf("layanan API Indonesia sedang bermasalah, silakan coba lagi nanti")
		}
		return fmt.Errorf("API Indonesia merespons status %d", statusCode)
	}
}

// Parameter pencarian kampus pada API Indonesia.
type SearchKampusParams struct {
	Query    string
	Province string
	Regency  string
	Type     string
	Group    string
	Page     int
	PerPage  int
}

// Hasil pencarian kampus beserta metadata paginasi.
type SearchKampusResult struct {
	Items []APIIndonesiaKampus `json:"items"`
	Meta  APIIndonesiaMeta     `json:"meta"`
}

// Mencari direktori kampus pada API Indonesia berdasarkan nama, wilayah, jenis, dan kelompok.
func (c *APIIndonesiaClient) SearchKampus(ctx context.Context, params SearchKampusParams) (*SearchKampusResult, error) {
	query := url.Values{}
	if strings.TrimSpace(params.Query) != "" {
		query.Set("q", strings.TrimSpace(params.Query))
	}
	if params.Province != "" {
		query.Set("province", params.Province)
	}
	if params.Regency != "" {
		query.Set("regency", params.Regency)
	}
	if params.Type != "" {
		query.Set("type", params.Type)
	}
	if params.Group != "" {
		query.Set("group", params.Group)
	}
	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 || perPage > 200 {
		perPage = 20
	}
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("per_page", fmt.Sprintf("%d", perPage))

	body, _, err := c.doGet(ctx, "/api/v1/kampus", query)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data []APIIndonesiaKampus `json:"data"`
		Meta APIIndonesiaMeta     `json:"meta"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("respons kampus API Indonesia tidak valid: %w", err)
	}
	return &SearchKampusResult{Items: envelope.Data, Meta: envelope.Meta}, nil
}

// Mengambil detail satu kampus dari API Indonesia berdasarkan ID eksternal.
func (c *APIIndonesiaClient) GetKampusDetail(ctx context.Context, externalID string) (*APIIndonesiaKampus, error) {
	if strings.TrimSpace(externalID) == "" {
		return nil, fmt.Errorf("ID kampus eksternal wajib diisi")
	}
	body, _, err := c.doGet(ctx, "/api/v1/kampus/"+url.PathEscape(externalID), nil)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data APIIndonesiaKampus `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("respons detail kampus API Indonesia tidak valid: %w", err)
	}
	if strings.TrimSpace(envelope.Data.ID) == "" {
		return nil, fmt.Errorf("data kampus tidak ditemukan pada API Indonesia")
	}
	return &envelope.Data, nil
}

// Parameter pencarian sekolah pada API Indonesia.
type SearchSekolahParams struct {
	Query    string
	Province string
	Regency  string
	Jenis    string
	Status   string
	Page     int
	PerPage  int
}

// Hasil pencarian sekolah beserta metadata paginasi.
type SearchSekolahResult struct {
	Items []APIIndonesiaSekolah `json:"items"`
	Meta  APIIndonesiaMeta      `json:"meta"`
}

// Mencari direktori sekolah pada API Indonesia berdasarkan nama, NPSN, wilayah, dan jenjang.
func (c *APIIndonesiaClient) SearchSekolah(ctx context.Context, params SearchSekolahParams) (*SearchSekolahResult, error) {
	query := url.Values{}
	if strings.TrimSpace(params.Query) != "" {
		query.Set("q", strings.TrimSpace(params.Query))
	}
	if params.Province != "" {
		query.Set("provinsi_id", params.Province)
	}
	if params.Regency != "" {
		query.Set("kabupaten_id", params.Regency)
	}
	if params.Jenis != "" {
		query.Set("jenis", params.Jenis)
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 || perPage > 200 {
		perPage = 20
	}
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("per_page", fmt.Sprintf("%d", perPage))

	body, _, err := c.doGet(ctx, "/api/v1/sekolah", query)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data []APIIndonesiaSekolah `json:"data"`
		Meta APIIndonesiaMeta      `json:"meta"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("respons sekolah API Indonesia tidak valid: %w", err)
	}
	return &SearchSekolahResult{Items: envelope.Data, Meta: envelope.Meta}, nil
}

// Mengambil detail satu sekolah dari API Indonesia berdasarkan NPSN.
func (c *APIIndonesiaClient) GetSekolahDetail(ctx context.Context, npsn string) (*APIIndonesiaSekolah, error) {
	if strings.TrimSpace(npsn) == "" {
		return nil, fmt.Errorf("NPSN sekolah wajib diisi")
	}
	body, _, err := c.doGet(ctx, "/api/v1/sekolah/"+url.PathEscape(npsn), nil)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Data APIIndonesiaSekolah `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("respons detail sekolah API Indonesia tidak valid: %w", err)
	}
	if strings.TrimSpace(envelope.Data.NPSN) == "" {
		return nil, fmt.Errorf("data sekolah tidak ditemukan pada API Indonesia")
	}
	return &envelope.Data, nil
}

// Memetakan respons kampus eksternal menjadi draf institusi internal tanpa menyimpan kredensial eksternal.
func MapKampusToInstitutionDraft(ext *APIIndonesiaKampus) *Institution {
	inst := &Institution{
		Name:   ext.Name,
		Source: strPtr(InstitutionSourceAPIKampus),
	}
	externalID := "kampus:" + ext.ID
	inst.ExternalID = &externalID
	parts := []string{}
	if ext.Address != nil && strings.TrimSpace(*ext.Address) != "" {
		parts = append(parts, strings.TrimSpace(*ext.Address))
	}
	if ext.RegencyName != nil && strings.TrimSpace(*ext.RegencyName) != "" {
		parts = append(parts, strings.TrimSpace(*ext.RegencyName))
	}
	if ext.ProvinceName != nil && strings.TrimSpace(*ext.ProvinceName) != "" {
		parts = append(parts, strings.TrimSpace(*ext.ProvinceName))
	}
	if len(parts) > 0 {
		joined := strings.Join(parts, ", ")
		inst.Address = &joined
	}
	if ext.Website != nil && strings.TrimSpace(*ext.Website) != "" {
		website := strings.TrimSpace(*ext.Website)
		inst.Email = nil
		_ = website
	}
	meta := map[string]string{
		"short_name":  stringOrEmpty(ext.ShortName),
		"jenis":       ext.Jenis,
		"kelompok":    ext.Kelompok,
		"province_id": ext.ProvinceID,
		"regency_id":  ext.RegencyID,
	}
	_ = meta
	return inst
}

// Memetakan respons sekolah eksternal menjadi draf institusi internal.
func MapSekolahToInstitutionDraft(ext *APIIndonesiaSekolah) *Institution {
	inst := &Institution{
		Name:   ext.Name,
		Source: strPtr(InstitutionSourceAPISekolah),
	}
	externalID := "sekolah:" + ext.NPSN
	inst.ExternalID = &externalID
	if ext.Address != nil {
		inst.Address = ext.Address
	}
	if ext.Phone != nil {
		inst.Phone = ext.Phone
	}
	if ext.Email != nil {
		inst.Email = ext.Email
	}
	return inst
}

// Mengembalikan pointer string untuk nilai opsional.
func strPtr(s string) *string {
	copied := s
	return &copied
}

// Mengembalikan string kosong apabila pointer bernilai nil.
func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
