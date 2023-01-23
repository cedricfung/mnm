package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/fox-one/mixin-sdk-go"
)

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
		hdr.refreshConversation(ctx, msg.ConversationID)
		return nil
	}

	data, err := base64.RawURLEncoding.DecodeString(msg.Data)
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
		err = hdr.refreshConversation(ctx, msg.ConversationID)
	}

	if err != nil {
		log.Println(msg, cp, err)
	}
	return nil
}
