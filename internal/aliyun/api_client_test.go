package aliyun

import (
	"context"
	"errors"
	"testing"

	casopenapi "github.com/alibabacloud-go/cas-20200407/v4/client"
	cdnopenapi "github.com/alibabacloud-go/cdn-20180510/v8/client"
)

type stubCASClient struct {
	listFn   func(request *casopenapi.ListUserCertificateOrderRequest) (*casopenapi.ListUserCertificateOrderResponse, error)
	uploadFn func(request *casopenapi.UploadUserCertificateRequest) (*casopenapi.UploadUserCertificateResponse, error)
	deleteFn func(request *casopenapi.DeleteUserCertificateRequest) (*casopenapi.DeleteUserCertificateResponse, error)
}

func (s *stubCASClient) ListUserCertificateOrder(request *casopenapi.ListUserCertificateOrderRequest) (*casopenapi.ListUserCertificateOrderResponse, error) {
	return s.listFn(request)
}

func (s *stubCASClient) UploadUserCertificate(request *casopenapi.UploadUserCertificateRequest) (*casopenapi.UploadUserCertificateResponse, error) {
	return s.uploadFn(request)
}

func (s *stubCASClient) DeleteUserCertificate(request *casopenapi.DeleteUserCertificateRequest) (*casopenapi.DeleteUserCertificateResponse, error) {
	return s.deleteFn(request)
}

type stubCDNClient struct {
	setFn func(request *cdnopenapi.SetCdnDomainSSLCertificateRequest) (*cdnopenapi.SetCdnDomainSSLCertificateResponse, error)
}

func (s *stubCDNClient) SetCdnDomainSSLCertificate(request *cdnopenapi.SetCdnDomainSSLCertificateRequest) (*cdnopenapi.SetCdnDomainSSLCertificateResponse, error) {
	return s.setFn(request)
}

func TestAPIClientFindCertificateByFingerprint(t *testing.T) {
	client := &APIClient{
		cas: &stubCASClient{
			listFn: func(request *casopenapi.ListUserCertificateOrderRequest) (*casopenapi.ListUserCertificateOrderResponse, error) {
				if request == nil {
					t.Fatalf("unexpected nil request")
				}
				if request.Keyword != nil {
					t.Fatalf("expected keyword to be unset, got %#v", request.Keyword)
				}
				if request.ResourceGroupId == nil || *request.ResourceGroupId != "rg-1" {
					t.Fatalf("unexpected resource group: %#v", request.ResourceGroupId)
				}
				return &casopenapi.ListUserCertificateOrderResponse{
					Body: &casopenapi.ListUserCertificateOrderResponseBody{
						ShowSize: teaInt64(100),
						CertificateOrderList: []*casopenapi.ListUserCertificateOrderResponseBodyCertificateOrderList{
							{
								CertificateId: teaInt64(42),
								Fingerprint:   teaString("fp-1"),
							},
						},
					},
				}, nil
			},
		},
	}

	cert, err := client.FindCertificateByFingerprint(context.Background(), "fp-1", "rg-1")
	if err != nil {
		t.Fatalf("FindCertificateByFingerprint returned error: %v", err)
	}
	if cert.ID != "42" {
		t.Fatalf("expected id 42, got %q", cert.ID)
	}
}

func TestAPIClientFindCertificateByFingerprintNormalizesDelimiters(t *testing.T) {
	client := &APIClient{
		cas: &stubCASClient{
			listFn: func(request *casopenapi.ListUserCertificateOrderRequest) (*casopenapi.ListUserCertificateOrderResponse, error) {
				if request == nil {
					t.Fatalf("unexpected nil request")
				}
				if request.Keyword != nil {
					t.Fatalf("expected keyword to be unset, got %#v", request.Keyword)
				}
				if request.ResourceGroupId == nil || *request.ResourceGroupId != "rg-1" {
					t.Fatalf("unexpected resource group: %#v", request.ResourceGroupId)
				}
				return &casopenapi.ListUserCertificateOrderResponse{
					Body: &casopenapi.ListUserCertificateOrderResponseBody{
						ShowSize: teaInt64(100),
						CertificateOrderList: []*casopenapi.ListUserCertificateOrderResponseBodyCertificateOrderList{
							{
								CertificateId: teaInt64(42),
								Fingerprint:   teaString("aa:bb:cc"),
							},
						},
					},
				}, nil
			},
		},
	}

	cert, err := client.FindCertificateByFingerprint(context.Background(), "AA:BB:CC", "rg-1")
	if err != nil {
		t.Fatalf("FindCertificateByFingerprint returned error: %v", err)
	}
	if cert.ID != "42" {
		t.Fatalf("expected id 42, got %q", cert.ID)
	}
}

func TestAPIClientFindCertificateByFingerprintNotFound(t *testing.T) {
	client := &APIClient{
		cas: &stubCASClient{
			listFn: func(request *casopenapi.ListUserCertificateOrderRequest) (*casopenapi.ListUserCertificateOrderResponse, error) {
				return &casopenapi.ListUserCertificateOrderResponse{
					Body: &casopenapi.ListUserCertificateOrderResponseBody{
						ShowSize:             teaInt64(100),
						CertificateOrderList: []*casopenapi.ListUserCertificateOrderResponseBodyCertificateOrderList{},
					},
				}, nil
			},
		},
	}

	_, err := client.FindCertificateByFingerprint(context.Background(), "missing", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAPIClientUploadCertificate(t *testing.T) {
	client := &APIClient{
		cfg: APIClientConfig{ResourceGroupID: "rg-1"},
		cas: &stubCASClient{
			uploadFn: func(request *casopenapi.UploadUserCertificateRequest) (*casopenapi.UploadUserCertificateResponse, error) {
				if request == nil || request.Cert == nil || request.Key == nil || request.Name == nil {
					t.Fatalf("unexpected upload request: %#v", request)
				}
				if request.ResourceGroupId == nil || *request.ResourceGroupId != "rg-1" {
					t.Fatalf("unexpected resource group: %#v", request.ResourceGroupId)
				}
				return &casopenapi.UploadUserCertificateResponse{
					Body: &casopenapi.UploadUserCertificateResponseBody{CertId: teaInt64(99)},
				}, nil
			},
			deleteFn: func(request *casopenapi.DeleteUserCertificateRequest) (*casopenapi.DeleteUserCertificateResponse, error) {
				return &casopenapi.DeleteUserCertificateResponse{}, nil
			},
		},
	}

	cert, err := client.UploadCertificate(context.Background(), "cert", "key", "fp-99")
	if err != nil {
		t.Fatalf("UploadCertificate returned error: %v", err)
	}
	if cert.ID != "99" {
		t.Fatalf("expected id 99, got %q", cert.ID)
	}
}

func TestAPIClientDeleteCertificate(t *testing.T) {
	client := &APIClient{
		cas: &stubCASClient{
			listFn: func(request *casopenapi.ListUserCertificateOrderRequest) (*casopenapi.ListUserCertificateOrderResponse, error) {
				return &casopenapi.ListUserCertificateOrderResponse{}, nil
			},
			uploadFn: func(request *casopenapi.UploadUserCertificateRequest) (*casopenapi.UploadUserCertificateResponse, error) {
				return &casopenapi.UploadUserCertificateResponse{}, nil
			},
			deleteFn: func(request *casopenapi.DeleteUserCertificateRequest) (*casopenapi.DeleteUserCertificateResponse, error) {
				if request == nil || request.CertId == nil || *request.CertId != 123 {
					t.Fatalf("unexpected delete request: %#v", request)
				}
				return &casopenapi.DeleteUserCertificateResponse{}, nil
			},
		},
	}

	if err := client.DeleteCertificate(context.Background(), "123"); err != nil {
		t.Fatalf("DeleteCertificate returned error: %v", err)
	}
}

func TestAPIClientUpdateDomainCertificate(t *testing.T) {
	client := &APIClient{
		cfg: APIClientConfig{Region: "cn-hangzhou"},
		cdn: &stubCDNClient{
			setFn: func(request *cdnopenapi.SetCdnDomainSSLCertificateRequest) (*cdnopenapi.SetCdnDomainSSLCertificateResponse, error) {
				if request == nil || request.DomainName == nil || *request.DomainName != "www.example.com" {
					t.Fatalf("unexpected request: %#v", request)
				}
				if request.CertId == nil || *request.CertId != 123 {
					t.Fatalf("unexpected cert id: %#v", request.CertId)
				}
				if request.CertType == nil || *request.CertType != "cas" {
					t.Fatalf("unexpected cert type: %#v", request.CertType)
				}
				if request.SSLProtocol == nil || *request.SSLProtocol != "on" {
					t.Fatalf("unexpected ssl protocol: %#v", request.SSLProtocol)
				}
				return &cdnopenapi.SetCdnDomainSSLCertificateResponse{}, nil
			},
		},
	}

	if err := client.UpdateDomainCertificate(context.Background(), "www.example.com", "123"); err != nil {
		t.Fatalf("UpdateDomainCertificate returned error: %v", err)
	}
}

func TestClassifyError(t *testing.T) {
	if got := ClassifyError(fmtErrorWrap(ErrRetryable)); !errors.Is(got, ErrRetryable) {
		t.Fatalf("expected retryable, got %v", got)
	}
	if got := ClassifyError(errors.New("boom")); !errors.Is(got, ErrTerminal) {
		t.Fatalf("expected terminal, got %v", got)
	}
}

func teaInt64(v int64) *int64      { return &v }
func teaString(v string) *string   { return &v }
func fmtErrorWrap(err error) error { return errors.Join(err) }
