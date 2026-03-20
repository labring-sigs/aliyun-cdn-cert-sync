package aliyun

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	casopenapi "github.com/alibabacloud-go/cas-20200407/v4/client"
	cdnopenapi "github.com/alibabacloud-go/cdn-20180510/v8/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	tea "github.com/alibabacloud-go/tea/tea"
)

type APIClientConfig struct {
	Region           string
	CredentialSource string
	AccessKeyID      string
	AccessKeySecret  string
	CASEndpoint      string
	CDNEndpoint      string
	ResourceGroupID  string
}

type casSDK interface {
	ListUserCertificateOrder(request *casopenapi.ListUserCertificateOrderRequest) (*casopenapi.ListUserCertificateOrderResponse, error)
	UploadUserCertificate(request *casopenapi.UploadUserCertificateRequest) (*casopenapi.UploadUserCertificateResponse, error)
	DeleteUserCertificate(request *casopenapi.DeleteUserCertificateRequest) (*casopenapi.DeleteUserCertificateResponse, error)
}

type cdnSDK interface {
	SetCdnDomainSSLCertificate(request *cdnopenapi.SetCdnDomainSSLCertificateRequest) (*cdnopenapi.SetCdnDomainSSLCertificateResponse, error)
}

type APIClient struct {
	cfg APIClientConfig
	cas casSDK
	cdn cdnSDK
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
	if strings.TrimSpace(cfg.AccessKeyID) == "" || strings.TrimSpace(cfg.AccessKeySecret) == "" {
		return nil, fmt.Errorf("%w: access key credentials are required", ErrTerminal)
	}

	casConfig := &openapi.Config{}
	casConfig.SetAccessKeyId(cfg.AccessKeyID)
	casConfig.SetAccessKeySecret(cfg.AccessKeySecret)
	casConfig.SetRegionId(cfg.Region)
	casConfig.SetEndpoint(cfg.CASEndpoint)
	casConfig.SetProtocol("HTTPS")

	casClient, err := casopenapi.NewClient(casConfig)
	if err != nil {
		return nil, wrapSDKError("init cas client", err)
	}

	cdnConfig := &openapi.Config{}
	cdnConfig.SetAccessKeyId(cfg.AccessKeyID)
	cdnConfig.SetAccessKeySecret(cfg.AccessKeySecret)
	cdnConfig.SetRegionId(cfg.Region)
	cdnConfig.SetEndpoint(cfg.CDNEndpoint)
	cdnConfig.SetProtocol("HTTPS")

	cdnClient, err := cdnopenapi.NewClient(cdnConfig)
	if err != nil {
		return nil, wrapSDKError("init cdn client", err)
	}

	return &APIClient{cfg: cfg, cas: casClient, cdn: cdnClient}, nil
}

func (c *APIClient) FindCertificateByFingerprint(ctx context.Context, fingerprint string, resourceGroupID string) (Certificate, error) {
	if strings.TrimSpace(fingerprint) == "" {
		return Certificate{}, fmt.Errorf("%w: fingerprint is empty", ErrTerminal)
	}
	if err := ctx.Err(); err != nil {
		return Certificate{}, fmt.Errorf("%w: context done", ErrRetryable)
	}

	normalizedFingerprint := normalizeFingerprint(fingerprint)
	log.Printf("normalized fingerprint is %v", normalizedFingerprint)
	page := int64(1)
	for {
		request := &casopenapi.ListUserCertificateOrderRequest{}
		request.SetResourceGroupId(resourceGroupID)
		request.SetCurrentPage(page)
		request.SetShowSize(100)
		request.SetOrderType("UPLOAD")

		response, err := c.cas.ListUserCertificateOrder(request)
		if err != nil {
			return Certificate{}, wrapSDKError("list certificates", err)
		}
		if response == nil || response.Body == nil {
			return Certificate{}, fmt.Errorf("%w: empty cas certificate list response", ErrRetryable)
		}

		items := response.Body.CertificateOrderList
		for _, item := range items {
			if item == nil {
				continue
			}
			log.Printf("fingerprint %v found", tea.StringValue(item.Fingerprint))
			if normalizeFingerprint(tea.StringValue(item.Fingerprint)) == normalizedFingerprint && item.CertificateId != nil {
				return Certificate{
					ID:          strconv.FormatInt(tea.Int64Value(item.CertificateId), 10),
					Fingerprint: fingerprint,
				}, nil
			}
		}

		showSize := tea.Int64Value(response.Body.ShowSize)
		if showSize <= 0 || int64(len(items)) < showSize {
			break
		}
		page++
	}

	return Certificate{}, ErrNotFound
}

func normalizeFingerprint(fingerprint string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(fingerprint), ":", ""))
}

func (c *APIClient) UploadCertificate(ctx context.Context, certPEM, keyPEM, fingerprint string) (Certificate, error) {
	if strings.TrimSpace(certPEM) == "" || strings.TrimSpace(keyPEM) == "" || strings.TrimSpace(fingerprint) == "" {
		return Certificate{}, fmt.Errorf("%w: certificate fields are incomplete", ErrTerminal)
	}
	if err := ctx.Err(); err != nil {
		return Certificate{}, fmt.Errorf("%w: context done", ErrRetryable)
	}

	request := &casopenapi.UploadUserCertificateRequest{}
	request.SetName(certificateName(fingerprint))
	request.SetCert(certPEM)
	request.SetKey(keyPEM)

	response, err := c.cas.UploadUserCertificate(request)
	if err != nil {
		return Certificate{}, wrapSDKError("upload certificate", err)
	}
	if response == nil || response.Body == nil || response.Body.CertId == nil {
		return Certificate{}, fmt.Errorf("%w: missing cert id in cas response", ErrRetryable)
	}

	return Certificate{
		ID:          strconv.FormatInt(tea.Int64Value(response.Body.CertId), 10),
		Fingerprint: fingerprint,
		CertPEM:     certPEM,
		KeyPEM:      keyPEM,
	}, nil
}

func (c *APIClient) DeleteCertificate(ctx context.Context, certificateID string) error {
	if strings.TrimSpace(certificateID) == "" {
		return fmt.Errorf("%w: certificate id is empty", ErrTerminal)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: context done", ErrRetryable)
	}

	certID, err := strconv.ParseInt(certificateID, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid certificate id %q", ErrTerminal, certificateID)
	}

	request := &casopenapi.DeleteUserCertificateRequest{}
	request.SetCertId(certID)

	_, err = c.cas.DeleteUserCertificate(request)
	if err != nil {
		return wrapSDKError("delete certificate", err)
	}
	return nil
}

func (c *APIClient) UpdateDomainCertificate(ctx context.Context, domain, certificateID string) error {
	if strings.TrimSpace(domain) == "" || strings.TrimSpace(certificateID) == "" {
		return fmt.Errorf("%w: domain or certificate id is empty", ErrTerminal)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: context done", ErrRetryable)
	}

	certID, err := strconv.ParseInt(certificateID, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid certificate id %q", ErrTerminal, certificateID)
	}

	request := &cdnopenapi.SetCdnDomainSSLCertificateRequest{}
	request.SetDomainName(domain)
	request.SetSSLProtocol("on")
	request.SetCertType("cas")
	request.SetCertId(certID)
	request.SetCertName(certificateName(certificateID))
	request.SetCertRegion(c.cfg.Region)

	_, err = c.cdn.SetCdnDomainSSLCertificate(request)
	if err != nil {
		return wrapSDKError("update domain certificate", err)
	}
	return nil
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

func certificateName(suffix string) string {
	suffix = strings.TrimSpace(suffix)
	if suffix == "" {
		return "cdn-cert-sync"
	}
	if len(suffix) > 32 {
		suffix = suffix[:32]
	}
	return "cdn-cert-sync-" + suffix
}

func wrapSDKError(action string, err error) error {
	if err == nil {
		return nil
	}
	if isSDKRetryable(err) {
		return fmt.Errorf("%w: %s: %v", ErrRetryable, action, err)
	}
	return fmt.Errorf("%w: %s: %v", ErrTerminal, action, err)
}

func isSDKRetryable(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "throttl") ||
		strings.Contains(message, "timeout") ||
		strings.Contains(message, "timed out") ||
		strings.Contains(message, "temporary") ||
		strings.Contains(message, "unavailable") ||
		strings.Contains(message, "internalerror") ||
		strings.Contains(message, "servicebusy") ||
		strings.Contains(message, "too many requests") ||
		strings.Contains(message, "connection reset") ||
		strings.Contains(message, "connection refused") ||
		strings.Contains(message, "eof")
}
