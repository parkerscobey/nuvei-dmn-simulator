package targetsafe

import "testing"

func TestClassify_Localhost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url      string
		expected Classification
	}{
		{"http://localhost:3000/webhook", ClassificationLocal},
		{"http://127.0.0.1:3000/webhook", ClassificationLocal},
		{"http://[::1]:3000/webhook", ClassificationLocal},
		{"https://localhost:3000/webhook", ClassificationLocal},
		{"http://localhost/webhook", ClassificationLocal},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := Classify(tt.url, nil, nil)
			if result.Classification != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.url, result.Classification, tt.expected)
			}
			if !result.Allowed() {
				t.Errorf("Classify(%q) allowed=false, want true", tt.url)
			}
		})
	}
}

func TestClassify_LocalSuffixes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url      string
		expected Classification
	}{
		{"http://app.test/webhook", ClassificationLocal},
		{"http://app.local/webhook", ClassificationLocal},
		{"https://myservice.localhost/webhook", ClassificationLocal},
		{"http://service.local/webhook", ClassificationLocal},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := Classify(tt.url, nil, nil)
			if result.Classification != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.url, result.Classification, tt.expected)
			}
		})
	}
}

func TestClassify_PublicTestingNamesAreUntrustedWithoutProfile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url      string
		expected Classification
	}{
		{"https://app.staging/webhook", ClassificationUntrusted},
		{"https://app.preview/webhook", ClassificationUntrusted},
		{"https://app.dev/webhook", ClassificationUntrusted},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := Classify(tt.url, nil, nil)
			if result.Classification != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.url, result.Classification, tt.expected)
			}
		})
	}
}

func TestClassify_TrustedProfiles(t *testing.T) {
	t.Parallel()
	trusted := map[string]Profile{
		"staging-gw": {URL: "https://staging-gateway.example.com/webhook", Kind: "staging"},
		"demo-gw":    {URL: "https://demo.example.com/nuvei", Kind: "staging"},
		"my-sandbox": {URL: "https://sandbox.mycompany.com/nuvei", Kind: "sandbox"},
	}

	tests := []struct {
		url      string
		expected Classification
	}{
		{"https://staging-gateway.example.com/webhook", ClassificationStaging},
		{"https://staging-gateway.example.com:8080/nuvei", ClassificationStaging},
		{"https://demo.example.com/nuvei", ClassificationStaging},
		{"https://demo.example.com/different/path", ClassificationStaging},
		{"https://sandbox.mycompany.com/nuvei", ClassificationStaging},
		{"https://unknown.example.com/webhook", ClassificationUntrusted},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := Classify(tt.url, trusted, nil)
			if result.Classification != tt.expected {
				t.Errorf("Classify(%q) = %v, want %v", tt.url, result.Classification, tt.expected)
			}
		})
	}
}

func TestClassify_TrustedProductionHostedSandbox(t *testing.T) {
	t.Parallel()
	trusted := map[string]Profile{
		"prod-sandbox": {URL: "https://app.mycompany.com/nuvei", Kind: "production-hosted-sandbox"},
	}
	result := Classify("https://app.mycompany.com/nuvei", trusted, nil)
	if result.Classification != ClassificationProductionHostedSandbox {
		t.Errorf("Classify prod-sandbox profile = %v, want production-hosted-sandbox", result.Classification)
	}
	if !result.RequiresConfirm {
		t.Errorf("Classify prod-sandbox profile RequiresConfirm=false, want true")
	}
}

func TestClassify_TrustedProfileRequiresConfirm(t *testing.T) {
	t.Parallel()
	trusted := map[string]Profile{
		"prod-hosted-demo": {URL: "https://app.mycompany.com/nuvei", Kind: "demo", RequiresConfirm: true},
	}
	result := Classify("https://app.mycompany.com/nuvei", trusted, nil)
	if result.Classification != ClassificationProductionHostedSandbox {
		t.Errorf("Classify requires-confirm profile = %v, want production-hosted-sandbox", result.Classification)
	}
	if !result.RequiresConfirm {
		t.Errorf("Classify requires-confirm profile RequiresConfirm=false, want true")
	}
}

func TestClassify_UntrustedPublicHttps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url string
	}{
		{"https://example.com/webhook"},
		{"https://api.stripe.com/webhook"},
		{"https://prod-gateway.mycompany.com/nuvei"},
		{"https://shop.example.org/webhooks/nuvei"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := Classify(tt.url, nil, nil)
			if result.Classification != ClassificationUntrusted {
				t.Errorf("Classify(%q) = %v, want untrusted", tt.url, result.Classification)
			}
			if result.Allowed() {
				t.Errorf("Classify(%q) allowed=true, want false", tt.url)
			}
		})
	}
}

func TestClassify_DeniedSchemes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url string
	}{
		{"file:///etc/passwd"},
		{"javascript:alert(1)"},
		{"data:text/html,<script>alert(1)</script>"},
		{"ftp://example.com/file"},
		{"telnet://example.com"},
		{"ssh://example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := Classify(tt.url, nil, nil)
			if result.Classification != ClassificationDenied {
				t.Errorf("Classify(%q) = %v, want denied", tt.url, result.Classification)
			}
			if result.Allowed() {
				t.Errorf("Classify(%q) allowed=true, want false", tt.url)
			}
		})
	}
}

func TestClassify_CustomDeniedPrefixes(t *testing.T) {
	t.Parallel()
	denied := []string{"https://malicious.example.com/"}
	result := Classify("https://malicious.example.com/phishing", nil, denied)
	if result.Classification != ClassificationDenied {
		t.Errorf("Classify with custom denied = %v, want denied", result.Classification)
	}
}

func TestClassify_NonLoopbackHTTP(t *testing.T) {
	t.Parallel()
	result := Classify("http://192.168.1.100:3000/webhook", nil, nil)
	if result.Classification != ClassificationUntrusted {
		t.Errorf("Classify non-loopback HTTP = %v, want untrusted", result.Classification)
	}

	result = Classify("http://10.0.0.5/webhook", nil, nil)
	if result.Classification != ClassificationUntrusted {
		t.Errorf("Classify private-network HTTP = %v, want untrusted", result.Classification)
	}
}

func TestClassify_UnparseableURL(t *testing.T) {
	t.Parallel()
	tests := []string{
		"not-a-url",
		"://missing-scheme",
		"http://",
	}
	for _, url := range tests {
		t.Run(url, func(t *testing.T) {
			result := Classify(url, nil, nil)
			if result.Classification != ClassificationDenied {
				t.Errorf("Classify(%q) = %v, want denied", url, result.Classification)
			}
		})
	}
}

func TestResult_Allowed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		classif Classification
		allowed bool
	}{
		{ClassificationLocal, true},
		{ClassificationStaging, true},
		{ClassificationProductionHostedSandbox, true},
		{ClassificationUntrusted, false},
		{ClassificationDenied, false},
		{ClassificationUnknown, false},
	}
	for _, tt := range tests {
		r := Result{Classification: tt.classif, URL: "https://example.com"}
		if r.Allowed() != tt.allowed {
			t.Errorf("Result{Classification:%v}.Allowed() = %v, want %v", tt.classif, r.Allowed(), tt.allowed)
		}
	}
}
