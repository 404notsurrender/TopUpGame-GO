package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type VIPResellerService interface {
	GetGameFeatures() ([]VIPProduct, error)
	CreateOrder(order VIPOrder) (*VIPOrderResponse, error)
	CheckStatus(orderID string) (*VIPStatusResponse, error)
}

type VIPProduct struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Stock       int     `json:"stock"`
}

type VIPOrder struct {
	GameID     string `json:"game_id"`
	GameServer string `json:"game_server"`
	ProductSKU string `json:"product_sku"`
}

type VIPOrderResponse struct {
	OrderID     string  `json:"order_id"`
	Status      string  `json:"status"`
	TotalPrice  float64 `json:"total_price"`
	CreatedAt   string  `json:"created_at"`
}

type VIPStatusResponse struct {
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	GameID      string `json:"game_id"`
	GameServer  string `json:"game_server"`
	ProductSKU  string `json:"product_sku"`
	UpdatedAt   string `json:"updated_at"`
}

type vipResellerService struct {
	baseURL string
	apiKey  string
	userID  string
	client  *http.Client
}

func NewVIPResellerService(baseURL, apiKey, userID string) VIPResellerService {
	return &vipResellerService{
		baseURL: baseURL,
		apiKey:  apiKey,
		userID:  userID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *vipResellerService) GetGameFeatures() ([]VIPProduct, error) {
	req, err := http.NewRequest("GET", s.baseURL+"/game-feature", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("User-ID", s.userID)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var response struct {
		Status   string       `json:"status"`
		Message  string       `json:"message"`
		Products []VIPProduct `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return response.Products, nil
}

func (s *vipResellerService) CreateOrder(order VIPOrder) (*VIPOrderResponse, error) {
	jsonData, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order data: %v", err)
	}

	req, err := http.NewRequest("POST", s.baseURL+"/order", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("User-ID", s.userID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var response struct {
		Status   string           `json:"status"`
		Message  string           `json:"message"`
		Order    VIPOrderResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response.Order, nil
}

func (s *vipResellerService) CheckStatus(orderID string) (*VIPStatusResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/status/%s", s.baseURL, orderID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("User-ID", s.userID)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var response struct {
		Status   string            `json:"status"`
		Message  string            `json:"message"`
		Data     VIPStatusResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response.Data, nil
}
