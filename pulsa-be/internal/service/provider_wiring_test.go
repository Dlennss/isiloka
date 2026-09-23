package service

import (
	"testing"

	"pulsa2/chytron"
	"pulsa2/internal/provider"
	"pulsa2/internal/repository"
	"pulsa2/rajabiller"
)

func TestMemberTrxServiceRegistersChytronPAYClient(t *testing.T) {
	repo := repository.NewMemberTrxRepository(nil)
	chClient := chytron.New("https://example.invalid", "id", "pin", "user", "pass", 0)

	svc := NewMemberTrxService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, chClient, nil)
	client, ok := svc.Clients["chytron"]
	if !ok {
		t.Fatalf("chytron PAY client is not registered")
	}
	if _, ok := client.(*provider.ChytronAdapter); !ok {
		t.Fatalf("chytron PAY client has unexpected type %T", client)
	}
	if svc.CHClient == nil {
		t.Fatalf("chytron raw client is not retained on service")
	}
}

func TestMemberTrxServiceRegistersRajabillerPAYClient(t *testing.T) {
	repo := repository.NewMemberTrxRepository(nil)
	rjClient := rajabiller.New("https://example.invalid", "uid", "pin", 0)

	svc := NewMemberTrxService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, rjClient)
	client, ok := svc.Clients["rajabiller"]
	if !ok {
		t.Fatalf("rajabiller PAY client is not registered")
	}
	if _, ok := client.(*provider.RajabillerAdapter); !ok {
		t.Fatalf("rajabiller PAY client has unexpected type %T", client)
	}
	if svc.RJClient == nil {
		t.Fatalf("rajabiller raw client is not retained on service")
	}
}
