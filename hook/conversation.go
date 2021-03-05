package main

import (
	"context"
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
	"github.com/gofrs/uuid"
)

const (
	DbPrefixConversationMeta        = "CONV#META#"
	DbPrefixConversationParticipant = "CONV#PART#"
	DbPrefixToken                   = "TOKEN#"
)

func (hdr *Handler) addParticipant(ctx context.Context, cid, pid string) error {
	return hdr.db.Update(func(txn *badger.Txn) error {
		return addConvPart(txn, cid, pid)
	})
}

func (hdr *Handler) removeParticipant(ctx context.Context, cid, pid string) error {
	return hdr.db.Update(func(txn *badger.Txn) error {
		key := keyConvPart(cid, pid)
		it, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		tk, err := it.ValueCopy(nil)
		if err != nil {
			return err
		}
		err = txn.Delete(tk)
		if err != nil {
			return err
		}
		return txn.Delete(key)
	})
}

func (hdr *Handler) refreshConversation(ctx context.Context, id string) error {
	conv, err := hdr.mixin.ReadConversation(ctx, id)
	if err != nil {
		return err
	}

	return hdr.db.Update(func(txn *badger.Txn) error {
		key := keyConvMeta(id)
		b, err := json.Marshal(conv)
		if err != nil {
			return err
		}
		err = txn.Set(key, b)
		if err != nil {
			return err
		}
		for _, p := range conv.Participants {
			err := addConvPart(txn, id, p.UserID)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func addConvPart(txn *badger.Txn, cid, pid string) error {
	cp := keyConvPart(cid, pid)
	_, err := txn.Get(cp)
	if err == nil {
		return nil
	}
	if err != badger.ErrKeyNotFound {
		return err
	}

	token := generateRandomToken()
	tk := keyToken(token)
	return txn.Set(cp, tk)
}

func keyToken(token string) []byte {
	return append([]byte(DbPrefixToken), token...)
}

func keyConvPart(cid, pid string) []byte {
	key := append([]byte(DbPrefixConversationParticipant), pid...)
	return append(key, cid...)
}

func keyConvMeta(cid string) []byte {
	return append([]byte(DbPrefixConversationMeta), cid...)
}

func generateRandomToken() string {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return id.String()
}
