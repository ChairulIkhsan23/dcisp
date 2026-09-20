package integrations

import (
	"fmt"
	"sort"
	"sync"
)

// Provider adalah kontrak generik bagi setiap penyedia API eksternal
// yang terdaftar pada aplikasi (contoh: api-indonesia, raja-ongkir, dsb).
type Provider interface {
	Name() string
}

// Registry menyimpan dan mendistribusikan instance provider API eksternal
// berdasarkan nama uniknya agar modul domain tidak menginstansiasi klien secara langsung.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// Membuat registry provider API eksternal baru yang kosong.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Mendaftarkan instance provider API eksternal ke dalam registry.
func (r *Registry) Register(p Provider) error {
	if p == nil {
		return fmt.Errorf("provider API eksternal tidak boleh nil")
	}
	if p.Name() == "" {
		return fmt.Errorf("nama provider API eksternal wajib diisi")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[p.Name()]; exists {
		return fmt.Errorf("provider API eksternal '%s' sudah terdaftar", p.Name())
	}
	r.providers[p.Name()] = p
	return nil
}

// Mengambil instance provider API eksternal berdasarkan nama yang terdaftar.
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider API eksternal '%s' tidak terdaftar", name)
	}
	return p, nil
}

// Mengembalikan daftar nama seluruh provider API eksternal yang terdaftar secara terurut.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
