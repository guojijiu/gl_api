package backends

import (
	"net/http"
	"strings"
	"testing"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
	intranetclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/client"
	cloudclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
)

func TestNewForPlatform_CloudPublic(t *testing.T) {
	cfg := &Config.AiGatewayConfig{}
	cfg.SetDefaults()
	api := cloudclient.NewAPI(cfg, http.DefaultClient)

	b, err := NewForPlatform(Requests.AiPlatformCloudPublic, api, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if b.PlatformID() != Requests.AiPlatformCloudPublic {
		t.Fatalf("PlatformID: got %q", b.PlatformID())
	}
	if _, ok := b.(*CloudGateway); !ok {
		t.Fatalf("expected *CloudGateway, got %T", b)
	}
}

func TestNewForPlatform_CloudPublic_NilAPI(t *testing.T) {
	_, err := NewForPlatform(Requests.AiPlatformCloudPublic, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), Requests.AiPlatformCloudPublic) || !strings.Contains(err.Error(), "平台 API 未初始化") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewForPlatform_CloudIntranet(t *testing.T) {
	cfg := &Config.AiGatewayConfig{}
	cfg.SetDefaults()
	api := intranetclient.NewAPI(cfg, http.DefaultClient)

	b, err := NewForPlatform(Requests.AiPlatformCloudIntranet, nil, api)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if b.PlatformID() != Requests.AiPlatformCloudIntranet {
		t.Fatalf("PlatformID: got %q", b.PlatformID())
	}
	if _, ok := b.(*CloudGateway); !ok {
		t.Fatalf("expected *CloudGateway, got %T", b)
	}
}

func TestNewForPlatform_CloudIntranet_NilAPI(t *testing.T) {
	_, err := NewForPlatform(Requests.AiPlatformCloudIntranet, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "平台 API 未初始化") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewForPlatform_ImageCompare(t *testing.T) {
	b, err := NewForPlatform(Requests.AiPlatformImageCompare, nil, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if b.PlatformID() != Requests.AiPlatformImageCompare {
		t.Fatalf("PlatformID: got %q", b.PlatformID())
	}
	if _, ok := b.(*Minimal); !ok {
		t.Fatalf("expected *Minimal, got %T", b)
	}
}

func TestNewForPlatform_Unknown(t *testing.T) {
	_, err := NewForPlatform("unknown_platform", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "未注册平台后端") {
		t.Fatalf("unexpected error: %v", err)
	}
}
