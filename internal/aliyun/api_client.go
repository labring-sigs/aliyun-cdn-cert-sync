package aliyun

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type APIClientConfig struct {
	Region           string
	CredentialSource string
	AccessKeyID      string
	AccessKeySecret  string
	CASEndpoint      string
	CDNEndpoint      string
}

type APIClient struct {
	cfg  APIClientConfig
	http *http.Client
}

func NewAPIClient(cfg APIClientConfig) (*APIClient, error) {
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, fmt.Errorf("%w: region is required", ErrTerminal)
	}
	if strings.TrimSpace(cfg.CredentialSource) == "" {
		return nil, fmt.Errorf("%w: credential source is required", ErrTerminal)
	}
	if strings.TrimSpace(cfg.CASEndpoint) == "" {
		return nil, fmt.Errorf("%w: cas endpoint is required", ErrTerminal)
	}
	if strings.TrimSpace(cfg.CDNEndpoint) == "" {
		return nil, fmt.Errorf("%w: cdn endpoint is required", ErrTerminal)
	}

	return &APIClient{
		cfg: cfg,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}, nil
}

func (c *APIClient) FindCertificateByFingerprint(ctx context.Context, fingerprint string) (Certificate, error) {
	if strings.TrimSpace(fingerprint) == "" {
		return Certificate{}, fmt.Errorf("%w: fingerprint is empty", ErrTerminal)
	}
	if err := ctx.Err(); err != nil {
		return Certificate{}, fmt.Errorf("%w: context done", ErrRetryable)
	}

	resp, err := c.callCAS(ctx, map[string]string{
		"Action": "DescribeUserCertificateList",
	})
	if err != nil {
		return Certificate{}, err
	}

	certs, parseErr := parseCASCertificateList(resp)
	if parseErr != nil {
		return Certificate{}, fmt.Errorf("%w: decode cas list response: %v", ErrRetryable, parseErr)
	}
	for _, item := range certs {
		if item.Fingerprint == fingerprint && item.ID != "" {
			return Certificate{
				ID:          item.ID,
				Fingerprint: item.Fingerprint,
			}, nil
		}
	}

	return Certificate{}, ErrNotFound
}

func (c *APIClient) UploadCertificate(ctx context.Context, certPEM, keyPEM, fingerprint string) (Certificate, error) {
	if strings.TrimSpace(certPEM) == "" || strings.TrimSpace(keyPEM) == "" || strings.TrimSpace(fingerprint) == "" {
		return Certificate{}, fmt.Errorf("%w: certificate fields are incomplete", ErrTerminal)
	}
	if err := ctx.Err(); err != nil {
		return Certificate{}, fmt.Errorf("%w: context done", ErrRetryable)
	}

	resp, err := c.callCAS(ctx, map[string]string{
		"Action":      "UploadUserCertificate",
		"Cert":        certPEM,
		"Key":         keyPEM,
		"Name":        "cdn-cert-sync",
		"Fingerprint": fingerprint,
	})
	if err != nil {
		return Certificate{}, err
	}

	certID, parseErr := parseCASUploadCertID(resp)
	if parseErr != nil {
		return Certificate{}, fmt.Errorf("%w: decode cas upload response: %v", ErrRetryable, parseErr)
	}
	if certID == "" {
		return Certificate{}, fmt.Errorf("%w: missing cert id in cas response", ErrRetryable)
	}

	return Certificate{
		ID:          certID,
		Fingerprint: fingerprint,
		CertPEM:     certPEM,
		KeyPEM:      keyPEM,
	}, nil
}

func (c *APIClient) UpdateDomainCertificate(ctx context.Context, domain, certificateID string) error {
	if strings.TrimSpace(domain) == "" || strings.TrimSpace(certificateID) == "" {
		return fmt.Errorf("%w: domain or certificate id is empty", ErrTerminal)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: context done", ErrRetryable)
	}

	_, err := c.callCDN(ctx, map[string]string{
		"Action":              "SetDomainServerCertificate",
		"DomainName":          domain,
		"ServerCertificateId": certificateID,
	})
	return err
}

func ClassifyError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrRetryable) {
		return ErrRetryable
	}
	return ErrTerminal
}

func (c *APIClient) callCAS(ctx context.Context, params map[string]string) ([]byte, error) {
	return c.callRPC(ctx, c.cfg.CASEndpoint, params)
}

func (c *APIClient) callCDN(ctx context.Context, params map[string]string) ([]byte, error) {
	return c.callRPC(ctx, c.cfg.CDNEndpoint, params)
}

func (c *APIClient) callRPC(ctx context.Context, endpoint string, params map[string]string) ([]byte, error) {
	if strings.TrimSpace(c.cfg.AccessKeyID) == "" || strings.TrimSpace(c.cfg.AccessKeySecret) == "" {
		return nil, fmt.Errorf("%w: access key credentials are required", ErrTerminal)
	}

	version := "2018-01-29"
	if action := params["Action"]; action == "SetDomainServerCertificate" {
		version = "2018-05-10"
	}

	requestParams := map[string]string{
		"AccessKeyId":      c.cfg.AccessKeyID,
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   strconv.FormatInt(time.Now().UnixNano(), 10),
		"SignatureVersion": "1.0",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          version,
	}
	for k, v := range params {
		requestParams[k] = v
	}

	signature := signRPC(requestParams, c.cfg.AccessKeySecret)
	requestParams["Signature"] = signature

	form := url.Values{}
	for k, v := range requestParams {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrRetryable, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		if isNetworkRetryable(err) {
			return nil, fmt.Errorf("%w: network error: %v", ErrRetryable, err)
		}
		return nil, fmt.Errorf("%w: request failed: %v", ErrTerminal, err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("%w: read response: %v", ErrRetryable, readErr)
	}

	if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("%w: service status=%d", ErrRetryable, resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%w: request status=%d body=%s", ErrTerminal, resp.StatusCode, string(body))
	}

	if apiErr := parseRPCError(body); apiErr != nil {
		if isRPCRetryableCode(apiErr.Code) {
			return nil, fmt.Errorf("%w: code=%s message=%s request_id=%s", ErrRetryable, apiErr.Code, apiErr.Message, apiErr.RequestID)
		}
		return nil, fmt.Errorf("%w: code=%s message=%s request_id=%s", ErrTerminal, apiErr.Code, apiErr.Message, apiErr.RequestID)
	}
	return body, nil
}

func signRPC(params map[string]string, accessKeySecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	query := make(url.Values)
	for _, k := range keys {
		query.Set(k, params[k])
	}
	stringToSign := "POST&%2F&" + url.QueryEscape(query.Encode())

	mac := hmac.New(sha1.New, []byte(accessKeySecret+"&"))
	_, _ = mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func isNetworkRetryable(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded)
}

type rpcErrorEnvelope struct {
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	RequestID string `json:"RequestId"`
}

type casCertificate struct {
	ID          string `json:"CertId"`
	Fingerprint string `json:"Fingerprint"`
}

func parseRPCError(body []byte) *rpcErrorEnvelope {
	var payload rpcErrorEnvelope
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	if payload.Code == "" && payload.Message == "" {
		return nil
	}
	return &payload
}

func isRPCRetryableCode(code string) bool {
	c := strings.ToLower(strings.TrimSpace(code))
	return strings.Contains(c, "throttl") ||
		strings.Contains(c, "timeout") ||
		strings.Contains(c, "unavailable") ||
		strings.Contains(c, "internalerror") ||
		strings.Contains(c, "servicebusy")
}

func parseCASCertificateList(body []byte) ([]casCertificate, error) {
	var payload struct {
		CertificateList []casCertificate `json:"CertificateList"`
		Certificates    []casCertificate `json:"Certificates"`
		Data            struct {
			CertificateList []casCertificate `json:"CertificateList"`
			Certificates    []casCertificate `json:"Certificates"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	if len(payload.CertificateList) > 0 {
		return payload.CertificateList, nil
	}
	if len(payload.Certificates) > 0 {
		return payload.Certificates, nil
	}
	if len(payload.Data.CertificateList) > 0 {
		return payload.Data.CertificateList, nil
	}
	if len(payload.Data.Certificates) > 0 {
		return payload.Data.Certificates, nil
	}
	return nil, nil
}

func parseCASUploadCertID(body []byte) (string, error) {
	var payload struct {
		CertID        string `json:"CertId"`
		CertificateID string `json:"CertificateId"`
		Data          struct {
			CertID        string `json:"CertId"`
			CertificateID string `json:"CertificateId"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	switch {
	case payload.CertID != "":
		return payload.CertID, nil
	case payload.CertificateID != "":
		return payload.CertificateID, nil
	case payload.Data.CertID != "":
		return payload.Data.CertID, nil
	case payload.Data.CertificateID != "":
		return payload.Data.CertificateID, nil
	default:
		return "", nil
	}
}
