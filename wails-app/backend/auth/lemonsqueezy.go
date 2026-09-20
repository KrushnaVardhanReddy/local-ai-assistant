package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	LemonSqueezyAPIBase = "https://api.lemonsqueezy.com/v1"
)

type licenseResponse struct {
	Activated    bool   `json:"activated"`
	Error        string `json:"error"`
	Instance     *struct {
		ID string `json:"id"`
	} `json:"instance"`
	Meta *struct {
		StoreID int `json:"store_id"`
	} `json:"meta"`
}

type validateResponse struct {
	Valid bool `json:"valid"`
	Error string `json:"error"`
	Meta  *struct {
		StoreID int `json:"store_id"`
	} `json:"meta"`
	LicenseKey *struct {
		Status string `json:"status"`
	} `json:"license_key"`
}

type deactivateResponse struct {
	Deactivated bool   `json:"deactivated"`
	Error       string `json:"error"`
}

// ValidateLicenseKey makes a POST request to LemonSqueezy license activation endpoint.
func ValidateLicenseKey(licenseKey, machineID, instanceName string) (bool, string, error) {
	apiURL := LemonSqueezyAPIBase + "/licenses/activate"
	data := url.Values{}
	data.Set("license_key", licenseKey)
	data.Set("instance_name", instanceName)

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return false, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("failed to read response: %w", err)
	}

	var lr licenseResponse
	if err := json.Unmarshal(body, &lr); err != nil {
		return false, "", fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode == http.StatusOK && lr.Activated {
		var activationID string
		if lr.Instance != nil {
			activationID = lr.Instance.ID
		}
		return true, activationID, nil
	}

	errMsg := lr.Error
	if errMsg == "" {
		errMsg = "unknown error"
	}
	return false, "", fmt.Errorf("activation failed: %s", errMsg)
}

// DeactivateLicenseKey makes a POST request to LemonSqueezy license deactivate endpoint.
func DeactivateLicenseKey(licenseKey, instanceID string) error {
	apiURL := LemonSqueezyAPIBase + "/licenses/deactivate"
	data := url.Values{}
	data.Set("license_key", licenseKey)
	data.Set("instance_id", instanceID)

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var dr deactivateResponse
	if err := json.Unmarshal(body, &dr); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode == http.StatusOK && dr.Deactivated {
		return nil
	}

	errMsg := dr.Error
	if errMsg == "" {
		errMsg = "unknown error"
	}
	return fmt.Errorf("deactivation failed: %s", errMsg)
}

// CheckLicenseKeyStatus checks the status of the license key.
func CheckLicenseKeyStatus(licenseKey string) (string, error) {
	apiURL := LemonSqueezyAPIBase + "/licenses/validate"
	data := url.Values{}
	data.Set("license_key", licenseKey)

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var vr validateResponse
	if err := json.Unmarshal(body, &vr); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if vr.LicenseKey != nil && vr.LicenseKey.Status != "" {
		return vr.LicenseKey.Status, nil
	}

	if vr.Error != "" {
		return "not_found", fmt.Errorf("validation failed: %s", vr.Error)
	}

	return "unknown", nil
}
