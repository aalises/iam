package security

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.skypicker.com/platform/security/iam/security/secrets"
)

type mockedSecretManager struct {
	mock.Mock
}

func (s *mockedSecretManager) GetAppToken(app string) (string, error) {
	if app == "SERVICENAME" {
		return "valid token", nil
	}
	return "", errors.New("wrong token bro")
}

func (s *mockedSecretManager) GetSetting(app string) (string, error) {
	if app == "SERVICENAME" {
		return "valid token", nil
	}
	return "", errors.New("wrong token bro")
}

func createFakeManager() secrets.SecretManager {
	return &mockedSecretManager{}
}

func TestGetServiceName(t *testing.T) {
	tests := map[string]string{
		"balkan":                            "BALKAN",
		"BALKAN/4704b82 (Kiwi.com sandbox)": "BALKAN",
		"balkan/1.42.1 (Kiwi.com sandbox)":  "BALKAN",
		"balkan-graphql/1.42.1":             "BALKAN-GRAPHQL",
		"balkan_graphql/1.42.1":             "BALKAN_GRAPHQL",
		"balkan graphql/1.42.1":             "BALKAN_GRAPHQL",
	}

	for test, expected := range tests {
		res, err := GetServiceName(test)
		assert.Equal(t, expected, res)
		assert.Equal(t, err, nil)
	}

	res, err := GetServiceName("")
	assert.Equal(t, "", res)
	assert.Error(t, err)
}

func TestCheckServiceName(t *testing.T) {
	tests := map[string]bool{
		"balkan":            false,
		"balkan_PROD1-test": false,
		"balkan%2f../":      true,
		"balkan/../":        true,
		"":                  true,
		"balkan$":           true,
	}

	for input, shouldError := range tests {
		if shouldError {
			assert.Error(t, CheckServiceName(input))
		} else {
			assert.NoError(t, CheckServiceName(input))
		}
	}
}

func TestCheckAuth(t *testing.T) {
	m := createFakeManager()

	req, _ := http.NewRequest("GET", "http://example.com/", nil)
	err := checkAuth(req, m)
	assert.Error(t, err, "Should error on missing email")

	req, _ = http.NewRequest("GET", "http://example.com/?email=email@example.com", nil)
	err = checkAuth(req, m)
	assert.Error(t, err, "Should error on missing User-Agent")
	req.Header.Set("User-Agent", "serviceName/version (Kiwi.com environment)")

	err = checkAuth(req, m)
	assert.Error(t, err, "Should error on missing Authorization header")
	req.Header.Set("Authorization", "invalid token")

	err = checkAuth(req, m)
	assert.Error(t, err, "Should error on invalid token")
	req.Header.Set("Authorization", "valid token")

	err = checkAuth(req, m)
	assert.NoError(t, err, "Should not error on valid request token")
}
