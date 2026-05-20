package credentials

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	EnvironmentTest = "test"
	EnvironmentProd = "prod"

	TestGetSessionTokenURL = "https://ppp-test.nuvei.com/ppp/api/v1/getSessionToken.do"
	ProdGetSessionTokenURL = "https://secure.safecharge.com/ppp/api/v1/getSessionToken.do"

	MethodGetSessionToken = "getSessionToken"
)

var ErrNotVerified = errors.New("merchant profile has not been verified in this session")

type Profile struct {
	Environment       string
	MerchantID        string
	MerchantSiteID    string
	MerchantSecretKey string
}

type Client struct {
	HTTPClient *http.Client
	Clock      func() time.Time
	NewID      func() (string, error)
	Endpoints  Endpoints
	Cache      *SessionCache
}

type Endpoints struct {
	Test string
	Prod string
}

type SessionCache struct {
	mu           sync.Mutex
	verified     map[string]Verification
	fingerprints map[string]struct{}
}

type Verification struct {
	Method          string
	Environment     string
	Endpoint        string
	MerchantID      string
	MerchantSiteID  string
	ClientRequestID string
	SessionToken    string
	VerifiedAt      time.Time
	Cached          bool
}

type getSessionTokenRequest struct {
	MerchantID      string `json:"merchantId"`
	MerchantSiteID  string `json:"merchantSiteId"`
	ClientRequestID string `json:"clientRequestId"`
	TimeStamp       string `json:"timeStamp"`
	Checksum        string `json:"checksum"`
}

type getSessionTokenResponse struct {
	SessionToken    string `json:"sessionToken"`
	InternalRequest int64  `json:"internalRequestId"`
	Status          string `json:"status"`
	ErrCode         int    `json:"errCode"`
	Reason          string `json:"reason"`
	MerchantID      string `json:"merchantId"`
	MerchantSiteID  string `json:"merchantSiteId"`
	Version         string `json:"version"`
	ClientRequestID string `json:"clientRequestId"`
}

func NewClient(cache *SessionCache) *Client {
	return &Client{Cache: cache}
}

func NewSessionCache() *SessionCache {
	return &SessionCache{
		verified:     map[string]Verification{},
		fingerprints: map[string]struct{}{},
	}
}

func (c *Client) Verify(ctx context.Context, profile Profile) (Verification, error) {
	if err := ValidateProfile(profile); err != nil {
		return Verification{}, err
	}

	cache := c.cache()
	if verification, ok := cache.lookup(profile); ok {
		verification.Cached = true
		return verification, nil
	}

	endpoint, err := c.endpoint(profile.Environment)
	if err != nil {
		return Verification{}, err
	}

	now := c.now().UTC()
	clientRequestID, err := c.newClientRequestID()
	if err != nil {
		return Verification{}, err
	}

	requestBody := getSessionTokenRequest{
		MerchantID:      profile.MerchantID,
		MerchantSiteID:  profile.MerchantSiteID,
		ClientRequestID: clientRequestID,
		TimeStamp:       now.Format("20060102150405"),
	}
	requestBody.Checksum = getSessionTokenChecksum(profile, requestBody.ClientRequestID, requestBody.TimeStamp)

	body, err := json.Marshal(requestBody)
	if err != nil {
		return Verification{}, fmt.Errorf("encode getSessionToken request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Verification{}, fmt.Errorf("build getSessionToken request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return Verification{}, fmt.Errorf("call Nuvei getSessionToken: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Verification{}, fmt.Errorf("read Nuvei getSessionToken response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Verification{}, fmt.Errorf("Nuvei getSessionToken returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(responseData)))
	}

	var response getSessionTokenResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return Verification{}, fmt.Errorf("decode Nuvei getSessionToken response: %w", err)
	}

	if response.Status != "SUCCESS" {
		return Verification{}, fmt.Errorf("Nuvei credential verification failed: status=%s errCode=%d reason=%s", response.Status, response.ErrCode, response.Reason)
	}
	if response.SessionToken == "" {
		return Verification{}, fmt.Errorf("Nuvei credential verification failed: missing sessionToken")
	}
	if response.MerchantID != "" && response.MerchantID != profile.MerchantID {
		return Verification{}, fmt.Errorf("Nuvei credential verification returned unexpected merchantId")
	}
	if response.MerchantSiteID != "" && response.MerchantSiteID != profile.MerchantSiteID {
		return Verification{}, fmt.Errorf("Nuvei credential verification returned unexpected merchantSiteId")
	}

	verification := Verification{
		Method:          MethodGetSessionToken,
		Environment:     profile.Environment,
		Endpoint:        endpoint,
		MerchantID:      profile.MerchantID,
		MerchantSiteID:  profile.MerchantSiteID,
		ClientRequestID: clientRequestID,
		SessionToken:    response.SessionToken,
		VerifiedAt:      now,
	}
	cache.store(profile, verification)

	return verification, nil
}

func ValidateProfile(profile Profile) error {
	switch profile.Environment {
	case EnvironmentTest, EnvironmentProd:
	default:
		return fmt.Errorf("environment must be test or prod")
	}
	if profile.MerchantID == "" {
		return fmt.Errorf("merchant ID is required")
	}
	if profile.MerchantSiteID == "" {
		return fmt.Errorf("merchant site ID is required")
	}
	if profile.MerchantSecretKey == "" {
		return fmt.Errorf("merchant secret key is required")
	}

	return nil
}

func (c *Client) IsVerified(profile Profile) bool {
	return c.cache().isVerified(profile)
}

func (c *Client) RequireVerified(profile Profile) error {
	if c.IsVerified(profile) {
		return nil
	}

	return ErrNotVerified
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) cache() *SessionCache {
	if c != nil && c.Cache != nil {
		return c.Cache
	}
	return defaultSessionCache
}

func (c *Client) now() time.Time {
	if c != nil && c.Clock != nil {
		return c.Clock()
	}
	return time.Now()
}

func (c *Client) newClientRequestID() (string, error) {
	if c != nil && c.NewID != nil {
		return c.NewID()
	}
	return randomClientRequestID()
}

func (c *Client) endpoint(environment string) (string, error) {
	endpoints := DefaultEndpoints()
	if c != nil {
		if c.Endpoints.Test != "" {
			endpoints.Test = c.Endpoints.Test
		}
		if c.Endpoints.Prod != "" {
			endpoints.Prod = c.Endpoints.Prod
		}
	}

	switch environment {
	case EnvironmentTest:
		return endpoints.Test, nil
	case EnvironmentProd:
		return endpoints.Prod, nil
	default:
		return "", fmt.Errorf("environment must be test or prod")
	}
}

func DefaultEndpoints() Endpoints {
	return Endpoints{Test: TestGetSessionTokenURL, Prod: ProdGetSessionTokenURL}
}

func (c *SessionCache) lookup(profile Profile) (Verification, bool) {
	if c == nil {
		return Verification{}, false
	}

	fingerprint := profileFingerprint(profile)
	c.mu.Lock()
	defer c.mu.Unlock()

	verification, ok := c.verified[fingerprint]
	return verification, ok
}

func (c *SessionCache) store(profile Profile, verification Verification) {
	if c == nil {
		return
	}

	fingerprint := profileFingerprint(profile)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.verified[fingerprint] = verification
	c.fingerprints[fingerprint] = struct{}{}
}

func (c *SessionCache) isVerified(profile Profile) bool {
	if c == nil {
		return false
	}

	fingerprint := profileFingerprint(profile)
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.fingerprints[fingerprint]
	return ok
}

func getSessionTokenChecksum(profile Profile, clientRequestID, timestamp string) string {
	source := profile.MerchantID + profile.MerchantSiteID + clientRequestID + timestamp + profile.MerchantSecretKey
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])
}

func profileFingerprint(profile Profile) string {
	source := profile.Environment + "\x00" + profile.MerchantID + "\x00" + profile.MerchantSiteID + "\x00" + profile.MerchantSecretKey
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])
}

func randomClientRequestID() (string, error) {
	data := make([]byte, 8)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("generate clientRequestId: %w", err)
	}

	return "nuvei-dmn-simulator-verify-" + hex.EncodeToString(data), nil
}

var defaultSessionCache = NewSessionCache()
