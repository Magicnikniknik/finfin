package orders

import (
	"testing"
)

func TestBuildRequestHashDeterministic(t *testing.T) {
	a := buildRequestHash("a", "b", "c")
	b := buildRequestHash("a", "b", "c")
	if a != b {
		t.Fatalf("expected equal hashes, got %s and %s", a, b)
	}
}

func TestBuildRequestHashDifferentPayload(t *testing.T) {
	a := buildRequestHash("a", "b", "c")
	b := buildRequestHash("a", "c", "b")
	if a == b {
		t.Fatalf("expected different hashes for different payload ordering")
	}
}

func TestDecodeCachedResponseErrorsOnNil(t *testing.T) {
	_, err := decodeCachedResponse[ReserveOrderResult](nil)
	if err == nil {
		t.Fatal("expected error for nil record")
	}
}
