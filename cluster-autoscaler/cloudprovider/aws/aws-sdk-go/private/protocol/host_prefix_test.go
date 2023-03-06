//go:build go1.7
// +build go1.7

package protocol

import (
	"net/http"
	"net/url"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/request"
)

func TestHostPrefixBuilder(t *testing.T) {
	cases := map[string]struct {
		URLHost  string
		ReqHost  string
		Prefix   string
		LabelsFn func() map[string]string
		Disabled bool

		ExpectURLHost string
		ExpectReqHost string
	}{
		"no labels": {
			URLHost:       "service.region.samsungspc.com",
			Prefix:        "data-",
			ExpectURLHost: "data-service.region.samsungspc.com",
		},
		"with labels": {
			URLHost: "service.region.samsungspc.com",
			Prefix:  "{first}-{second}.",
			LabelsFn: func() map[string]string {
				return map[string]string{
					"first":  "abc",
					"second": "123",
				}
			},
			ExpectURLHost: "abc-123.service.region.samsungspc.com",
		},
		"with host prefix disabled": {
			Disabled: true,
			URLHost:  "service.region.samsungspc.com",
			Prefix:   "{first}-{second}.",
			LabelsFn: func() map[string]string {
				return map[string]string{
					"first":  "abc",
					"second": "123",
				}
			},
			ExpectURLHost: "service.region.samsungspc.com",
		},
		"with duplicate labels": {
			URLHost: "service.region.samsungspc.com",
			Prefix:  "{first}-{second}-{first}.",
			LabelsFn: func() map[string]string {
				return map[string]string{
					"first":  "abc",
					"second": "123",
				}
			},
			ExpectURLHost: "abc-123-abc.service.region.samsungspc.com",
		},
		"with unbracketed labels": {
			URLHost: "service.region.samsungspc.com",
			Prefix:  "first-{second}.",
			LabelsFn: func() map[string]string {
				return map[string]string{
					"first":  "abc",
					"second": "123",
				}
			},
			ExpectURLHost: "first-123.service.region.samsungspc.com",
		},
		"with req host": {
			URLHost:       "service.region.samsungspc.com:1234",
			ReqHost:       "service.region.samsungspc.com",
			Prefix:        "data-",
			ExpectURLHost: "data-service.region.samsungspc.com:1234",
			ExpectReqHost: "data-service.region.samsungspc.com",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			builder := HostPrefixBuilder{
				Prefix: c.Prefix, LabelsFn: c.LabelsFn,
			}
			req := &request.Request{
				Config: aws.Config{
					DisableEndpointHostPrefix: aws.Bool(c.Disabled),
				},
				HTTPRequest: &http.Request{
					Host: c.ReqHost,
					URL: &url.URL{
						Host: c.URLHost,
					},
				},
			}

			builder.Build(req)
			if e, a := c.ExpectURLHost, req.HTTPRequest.URL.Host; e != a {
				t.Errorf("expect URL host %v, got %v", e, a)
			}
			if e, a := c.ExpectReqHost, req.HTTPRequest.Host; e != a {
				t.Errorf("expect request host %v, got %v", e, a)
			}
		})
	}
}
