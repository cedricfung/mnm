package main

import (
	"context"
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
)

const (
	DbPrefixConversationMeta        = "CONV#META#"
	DbPrefixConversationParticipant = "CONV#PART#"
	DbPrefixToken                   = "TOKEN#"
)

type ConverstaionWithToken struct {
	Token string
	Meta  *mixin.Conversation
}

func (hdr *Handler) readConvPartByToken(ctx context.Context, token string) (string, string, error) {
	txn := hdr.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(keyToken(token))
	if err == badger.ErrKeyNotFound {
		return "", "", nil
	} else if err != nil {
		return "", "", err
	}
	pc, err := item.ValueCopy(nil)
	if err != nil {
		return "", "", err
	}
	pc = pc[len(DbPrefixConversationParticipant):]
	return string(pc[36:]), string(pc[:36]), nil
}

func (hdr *Handler) listConversations(ctx context.Context) ([]*ConverstaionWithToken, error) {
	txn := hdr.db.NewTransaction(false)
	defer txn.Discard()

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	me := hdr.getCurrentUser(ctx)
	convs := make([]*ConverstaionWithToken, 0)
	prefix := keyConvPart("", me.UserID)
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		tk, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}
		token := string(tk[len(DbPrefixToken):])

		cid := item.Key()[len(DbPrefixConversationParticipant)+len(me.UserID):]
		_, err = txn.Get(keyConvPart(string(cid), hdr.mixin.ClientID))
		if err == badger.ErrKeyNotFound {
			continue
		} else if err != nil {
			return nil, err
		}

		item, err = txn.Get(keyConvMeta(string(cid)))
		if err != nil {
			return nil, err
		}
		cb, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}
		var conv mixin.Conversation
		err = json.Unmarshal(cb, &conv)
		if err != nil {
			return nil, err
		}

		convs = append(convs, &ConverstaionWithToken{
			Token: token,
			Meta:  &conv,
		})
	}
	return convs, nil
}

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
	err = txn.Set(cp, tk)
	if err != nil {
		return err
	}
	return txn.Set(tk, cp)
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
