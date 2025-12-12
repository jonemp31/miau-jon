package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

var client *Client

// Initialize cria uma nova instância do cliente Supabase
func Initialize(url, apiKey string) {
	client = &Client{
		baseURL: url,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
	zap.L().Info("Supabase client initialized", zap.String("url", url))
}

// Get retorna a instância do cliente Supabase
func Get() *Client {
	return client
}

// IsEnabled verifica se o Supabase está configurado
func IsEnabled() bool {
	return client != nil && client.baseURL != "" && client.apiKey != ""
}

// Upsert insere ou atualiza um registro na tabela do Supabase
func (c *Client) Upsert(table string, data interface{}) error {
	if !IsEnabled() {
		zap.L().Debug("Supabase not enabled, skipping upsert")
		return nil
	}

	url := fmt.Sprintf("%s/rest/v1/%s", c.baseURL, table)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase error: status=%d body=%s", resp.StatusCode, string(body))
	}

	zap.L().Debug("Supabase upsert successful", zap.String("table", table))
	return nil
}

// Delete deleta um registro da tabela do Supabase
func (c *Client) Delete(table, column, value string) error {
	if !IsEnabled() {
		zap.L().Debug("Supabase not enabled, skipping delete")
		return nil
	}

	url := fmt.Sprintf("%s/rest/v1/%s?%s=eq.%s", c.baseURL, table, column, value)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase error: status=%d body=%s", resp.StatusCode, string(body))
	}

	zap.L().Debug("Supabase delete successful", zap.String("table", table))
	return nil
}

// Query busca registros na tabela do Supabase
func (c *Client) Query(table, filter string) ([]map[string]interface{}, error) {
	if !IsEnabled() {
		return nil, fmt.Errorf("supabase not enabled")
	}

	url := fmt.Sprintf("%s/rest/v1/%s?%s", c.baseURL, table, filter)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("supabase error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// CallRPC chama uma função RPC (stored procedure) no Supabase
func (c *Client) CallRPC(functionName string, params map[string]interface{}) error {
	if !IsEnabled() {
		zap.L().Debug("Supabase not enabled, skipping RPC call")
		return nil
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/%s", c.baseURL, functionName)

	jsonData, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute RPC request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase RPC error: status=%d body=%s", resp.StatusCode, string(body))
	}

	zap.L().Debug("Supabase RPC call successful", zap.String("function", functionName))
	return nil
}
