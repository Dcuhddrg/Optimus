package cloud

import (
	"context"
	"errors"
	"testing"
)

func TestAccountServiceWithoutAccessibleProjectsDoesNotQueryDB(t *testing.T) {
	service, err := NewCloudAccountService(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(context.Background(), nil, 1); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected scope denial, got %v", err)
	}
}
