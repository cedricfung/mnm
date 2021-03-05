package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
)

//go:embed mnm.sh
var MNMSH []byte

//go:embed html/index.html
var TemplateIndex []byte

//go:embed html/token.html
var TemplateToken []byte

type Handler struct {
	db     *badger.DB
	mixin  *mixin.Client
	secret string
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
	log.Println(r)
	defer handlePanic(w, r)

	if r.URL.Path == "/auth" {
		hdr.handleAuth(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/in/") {
		if r.Method == "POST" {
			hdr.handleMessage(w, r)
		} else {
			hdr.handleScript(w, r)
		}
		return
	}

	r = hdr.setCurrentUser(w, r)
	if r == nil {
		return
	}

	if r.URL.Path == "/" {
		hdr.handleRoot(w, r)
		return
	}

	http.NotFound(w, r)
}

func (hdr *Handler) setCurrentUser(w http.ResponseWriter, r *http.Request) *http.Request {

	ac, err := r.Cookie("Authorization")
	if err == http.ErrNoCookie {
		oauth := fmt.Sprintf("https://mixin.one/oauth/authorize?client_id=%s&scope=PROFILE:READ", hdr.mixin.ClientID)
		http.Redirect(w, r, oauth, http.StatusTemporaryRedirect)
		return nil
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "authorization server error", http.StatusInternalServerError)
		return nil
	}
	me, err := mixin.UserMe(r.Context(), ac.Value)
	if mixin.IsErrorCodes(err, mixin.Unauthorized) {
		oauth := fmt.Sprintf("https://mixin.one/oauth/authorize?client_id=%s&scope=PROFILE:READ", hdr.mixin.ClientID)
		http.Redirect(w, r, oauth, http.StatusTemporaryRedirect)
		return nil
	}
	if err != nil {
		log.Println(err)
		http.Error(w, "authorization mixin error", http.StatusInternalServerError)
		return nil
	}
	return r.WithContext(context.WithValue(r.Context(), "ME", me))
}

func (hdr *Handler) getCurrentUser(ctx context.Context) *mixin.User {
	return ctx.Value("ME").(*mixin.User)
}

func (hdr *Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	me := hdr.getCurrentUser(r.Context())
	convs, err := hdr.listConversations(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "list conversations error", http.StatusInternalServerError)
		return
	}

	tpl, err := template.New("index").Parse(string(TemplateIndex))
	if err != nil {
		panic(err)
	}

	data := struct {
		Me            *mixin.User
		Conversations []*ConverstaionWithToken
	}{me, convs}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl.Execute(w, data)
}

func (hdr *Handler) handleAuth(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, _, err := mixin.AuthorizeToken(r.Context(), hdr.mixin.ClientID, hdr.secret, code, "")
	if err != nil {
		log.Println(err)
		http.Error(w, "authorization mixin error", http.StatusInternalServerError)
		return
	}
	ac := http.Cookie{
		Name:     "Authorization",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &ac)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (hdr *Handler) handleScript(w http.ResponseWriter, r *http.Request) {
	token, sh := getHookToken(r.URL.Path)
	cid, pid, err := hdr.readConvPartByToken(r.Context(), token)
	if err != nil {
		log.Println(err)
		http.Error(w, "read conversation error", http.StatusInternalServerError)
		return
	}
	if cid == "" || pid == "" {
		http.Error(w, "read conversation error", http.StatusForbidden)
		return
	}

	tpl, err := template.New("token").Parse(string(TemplateToken))
	if err != nil {
		panic(err)
	}
	script := string(MNMSH)
	script = strings.Replace(script, "MM-WEBHOOK-TOKEN", token, -1)
	if sh {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(script))
		return
	}

	data := struct {
		Script string
	}{script}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	tpl.Execute(w, data)
}

func (hdr *Handler) handleMessage(w http.ResponseWriter, r *http.Request) {
	token, _ := getHookToken(r.URL.Path)
	cid, pid, err := hdr.readConvPartByToken(r.Context(), token)
	if err != nil {
		log.Println(err)
		http.Error(w, "read conversation error", http.StatusInternalServerError)
		return
	}
	if cid == "" || pid == "" {
		http.Error(w, "read conversation error", http.StatusForbidden)
		return
	}
	var msg mixin.MessageRequest
	err = json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Println(err)
		http.Error(w, "parse message error", http.StatusBadRequest)
		return
	}
	mid, _ := uuid.NewV4()
	msg.ConversationID = cid
	msg.MessageID = mid.String()
	err = hdr.mixin.SendMessage(r.Context(), &msg)
	if err != nil {
		log.Println(err)
		http.Error(w, "send message error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
}

func getHookToken(path string) (string, bool) {
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		return "", false
	}
	if parts[1] != "in" {
		return "", false
	}
	parts = strings.Split(parts[2], ".")
	if len(parts) < 1 {
		return "", false
	}
	id := uuid.FromStringOrNil(parts[0])
	if id.String() != parts[0] {
		return "", false
	}
	return id.String(), len(parts) == 2 && parts[1] == "sh"
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
