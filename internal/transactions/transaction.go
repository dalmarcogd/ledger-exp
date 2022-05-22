package transactions

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"

	"github.com/dalmarcogd/blockchain-exp/internal/accounts"
	"github.com/google/uuid"
)

type Hash [32]byte

var NilHash Hash = [32]byte{}

type Amount float64

func (a Amount) Bytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, a)
	return buf.Bytes()
}

type Transaction struct {
	ID       uuid.UUID        `json:"id"`
	Hash     Hash             `json:"hash"`
	PrevHash Hash             `json:"prev_hash"`
	From     accounts.Account `json:"from"`
	To       accounts.Account `json:"to"`
	Amount   Amount           `json:"amount"`
}

func NewTransaction(
	prevHash Hash,
	from accounts.Account,
	to accounts.Account,
	amount Amount,
) (Transaction, error) {
	transaction := Transaction{
		ID:       uuid.New(),
		PrevHash: prevHash,
		From:     from,
		To:       to,
		Amount:   amount,
	}

	serialize, err := transaction.serialize()
	if err != nil {
		return Transaction{}, err
	}

	hash := deriveHash(transaction.PrevHash, serialize)
	transaction.Hash = hash

	return transaction, nil
}

func (t Transaction) serialize() ([]byte, error) {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err := enc.Encode(t)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func (t Transaction) Validate(prevHash Hash) bool {
	tCopy := t
	tCopy.Hash = NilHash
	serialize, err := tCopy.serialize()
	if err != nil {
		return false
	}

	hash := deriveHash(prevHash, serialize)
	return hash == t.Hash
}

//func deserialize(data []byte) (Transaction, error) {
//	var transaction Transaction
//
//	decoder := gob.NewDecoder(bytes.NewReader(data))
//	err := decoder.Decode(&transaction)
//	if err != nil {
//		return Transaction{}, err
//	}
//	return transaction, nil
//}

func deriveHash(prevHash Hash, trxSerialized []byte) Hash {
	return sha256.Sum256(
		bytes.Join(
			[][]byte{
				prevHash[:],
				trxSerialized,
			},
			[]byte{},
		),
	)
}
