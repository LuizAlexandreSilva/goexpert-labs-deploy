package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Mockando a resposta das APIs
const mockViaCEPResponse = `{
	"localidade": "São Paulo",
	"erro": "false"
}`
const mockViaCEPErrorResponse = `{
	"erro": "true"
}`
const mockWeatherAPIResponse = `{
	"current": {
		"temp_c": 25.0
	}
}`

// Mock do RoundTripper para simular requisições HTTP
type mockRoundTripper struct {
	responseMap map[string]string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	responseBody, exists := m.responseMap[req.URL.String()]
	if !exists {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"erro": "true"}`)),
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}, nil
}

func TestHandler(t *testing.T) {
	tests := []struct {
		name           string
		cep            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "CEP válido",
			cep:            "01001000",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"city":"São Paulo","temp_C":25,"temp_F":77,"temp_K":298.15}`,
		},
		{
			name:           "CEP inválido - letras",
			cep:            "abcdefgh",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Invalid zipcode",
		},
		{
			name:           "CEP inválido - menos dígitos",
			cep:            "12345",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Invalid zipcode",
		},
		{
			name:           "CEP inexistente",
			cep:            "99999999",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "can not find zipcode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar um mock do cliente HTTP
			mockClient := &http.Client{
				Transport: &mockRoundTripper{
					responseMap: map[string]string{
						fmt.Sprintf("http://viacep.com.br/ws/%s/json/", tt.cep): mockViaCEPResponse,
						fmt.Sprintf("http://viacep.com.br/ws/99999999/json/"): mockViaCEPErrorResponse,
						fmt.Sprintf("https://api.weatherapi.com/v1/current.json?q=São Paulo&key=906fd34420aa45c2a20174551251802"): mockWeatherAPIResponse,
					},
				},
			}

			// Criar uma requisição HTTP fake
			req := httptest.NewRequest("GET", fmt.Sprintf("/?cep=%s", tt.cep), nil)
			rec := httptest.NewRecorder()

			// Chamar a função handler com o mock do cliente HTTP
			handler(rec, req, mockClient)

			// Verificar o código de status
			if rec.Code != tt.expectedStatus {
				t.Errorf("esperado status %d, mas recebeu %d", tt.expectedStatus, rec.Code)
			}

			// Verificar o corpo da resposta
			body, _ := io.ReadAll(rec.Body)
			if !strings.Contains(string(body), tt.expectedBody) {
				t.Errorf("esperado corpo contendo %s, mas recebeu %s", tt.expectedBody, string(body))
			}
		})
	}
}
