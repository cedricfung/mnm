package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/fox-one/mixin-sdk-go"
)

type Handler struct {
	db    *badger.DB
	mixin *mixin.Client
}

func NewServer(hdr *Handler, port int) *http.Server {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      hdr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return server
}

func handlePanic(w http.ResponseWriter, r *http.Request) {
	rcv := recover()
	if rcv == nil {
		return
	}
	err := fmt.Sprint(rcv)
	log.Println(err)
}

func (hdr *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer handlePanic(w, r)

	if r.URL.Path == "/" {
		hdr.handleRoot(w, r)
		return
	}

	if r.URL.Path == "/auth" {
		hdr.handleAuth(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/me/") {
		hdr.handleScript(w, r)
		return
	}

	http.NotFound(w, r)
}

func (hdr *Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
}

func (hdr *Handler) handleAuth(w http.ResponseWriter, r *http.Request) {
}

func (hdr *Handler) handleScript(w http.ResponseWriter, r *http.Request) {
}

func (hdr *Handler) OnAckReceipt(ctx context.Context, msg *mixin.MessageView, userID string) error {
	return nil
}

type systemConversationPayload struct {
	Action        string `json:"action"`
	ParticipantId string `json:"participant_id"`
	UserId        string `json:"user_id,omitempty"`
	Role          string `json:"role,omitempty"`
}

func (hdr *Handler) OnMessage(ctx context.Context, msg *mixin.MessageView, userID string) error {
	if msg.Category != mixin.MessageCategorySystemConversation {
		return nil
	}

	data, err := base64.URLEncoding.DecodeString(msg.Data)
	if err != nil {
		log.Println(msg, err)
		return nil
	}

	var cp systemConversationPayload
	err = json.Unmarshal(data, &cp)
	if err != nil {
		log.Println(msg, err)
		return nil
	}

	switch cp.Action {
	case "ADD":
		if cp.ParticipantId == hdr.mixin.ClientID {
			err = hdr.refreshConversation(ctx, msg.ConversationID)
		} else {
			err = hdr.addParticipant(ctx, msg.ConversationID, cp.ParticipantId)
		}
	case "REMOVE":
		err = hdr.removeParticipant(ctx, msg.ConversationID, cp.ParticipantId)
	case "UPDATE":
		err = hdr.refreshConversation(ctx, msg.ConversationID)
	default:
		return nil
	}

	if err != nil {
		log.Println(msg, cp, err)
	}
	return nil
}
