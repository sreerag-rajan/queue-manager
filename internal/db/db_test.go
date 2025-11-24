package db

import "testing"

func TestConnect_EmptyURI(t *testing.T) {
	_, err := Connect("")
	if err == nil {
		t.Fatalf("expected error for empty POSTGRES_URI")
	}
}


