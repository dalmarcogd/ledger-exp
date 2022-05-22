package main

import (
	"context"
	"log"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/internal/transactions"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
)

func main() {
	t, err := tracer.New("localhost:55681", "ledger-exp", "local", "1.0.0")
	if err != nil {
		log.Panic(err)
	}

	ctx := context.Background()

	defer func() {
		err := t.Stop(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, span := t.Span(ctx)
	defer span.End()

	trx1, err := transactions.NewTransaction(
		transactions.NilHash,
		accounts.Account{},
		accounts.Account{},
		0,
	)
	if err != nil {
		log.Panic(err)
	}

	trx2, err := transactions.NewTransaction(
		trx1.Hash,
		accounts.Account{},
		accounts.Account{},
		2,
	)
	if err != nil {
		log.Panic(err)
	}

	trx3, err := transactions.NewTransaction(
		trx2.Hash,
		accounts.Account{},
		accounts.Account{},
		3,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Previous Hash trx1: %x\n", trx1.PrevHash)
	log.Printf("Amount trx1: %f\n", trx1.Amount)
	log.Printf("Hash trx1: %x\n", trx1.Hash)
	log.Printf("Validate Hash trx1: %v\n", trx1.Validate(trx1.PrevHash))

	log.Printf("Previous Hash trx2: %x\n", trx2.PrevHash)
	log.Printf("Amount trx2: %f\n", trx2.Amount)
	log.Printf("Hash trx2: %x\n", trx2.Hash)
	log.Printf("Validate Hash trx2: %v\n", trx2.Validate(trx2.PrevHash))

	log.Printf("Previous Hash trx3: %x\n", trx3.PrevHash)
	log.Printf("Amount trx3go: %f\n", trx3.Amount)
	log.Printf("Hash trx3: %x\n", trx3.Hash)
	log.Printf("Validate Hash trx3: %v\n", trx3.Validate(trx3.PrevHash))

	var h transactions.Hash
	copy(h[:], "334b2e5b1110efaf0b1b0a2889b3edccd59701d41a273dac162f173e24db74b3")
	log.Printf("Hash trx4: %x\n", h)
	log.Printf("Validate Hash trx4: %v\n", transactions.Transaction{Hash: h}.Validate(transactions.NilHash))
}
