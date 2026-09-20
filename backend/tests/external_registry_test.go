package tests

import (
	"testing"

	"dcisp/backend/internal/integrations"
	"dcisp/backend/internal/integrations/apiindonesia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji registry provider API eksternal generik: registrasi, pengambilan, duplikasi, dan pencatatan nama.
func TestExternalProviderRegistry(t *testing.T) {
	registry := integrations.NewRegistry()

	client := apiindonesia.NewClient("https://use.apiindonesia.id", "aip_live_test_key")
	assert.Equal(t, apiindonesia.ProviderName, client.Name())

	// 1. Registrasi provider baru berhasil
	require.NoError(t, registry.Register(client))

	// 2. Registrasi ganda dengan nama sama ditolak
	err := registry.Register(apiindonesia.NewClient("https://use.apiindonesia.id", "aip_live_other"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sudah terdaftar")

	// 3. Pengambilan provider terdaftar berhasil
	got, err := registry.Get(apiindonesia.ProviderName)
	require.NoError(t, err)
	assert.Equal(t, apiindonesia.ProviderName, got.Name())

	// 4. Pengambilan provider yang belum terdaftar ditolak
	_, err = registry.Get("provider-lain-yang-belum-ada")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak terdaftar")

	// 5. Daftar nama provider terurut
	assert.Equal(t, []string{apiindonesia.ProviderName}, registry.Names())
}
