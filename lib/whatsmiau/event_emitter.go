package whatsmiau

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-vcard"
	"github.com/google/uuid"
	cache "github.com/patrickmn/go-cache"
	"github.com/verbeux-ai/whatsmiau/models"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/proto/waHistorySync"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type emitter struct {
	url  string
	data any
}

func (s *Whatsmiau) getInstance(id string) *models.Instance {
	ctx, c := context.WithTimeout(context.Background(), time.Second*5)
	defer c()

	res, err := s.repo.List(ctx, id)
	if err != nil {
		zap.L().Panic("failed to get instanceCached by instance", zap.Error(err))
	}

	if len(res) == 0 {
		zap.L().Warn("no instanceCached found by instance", zap.String("instance", id))
		return nil
	}

	s.instanceCache.Set(id, res[0], cache.DefaultExpiration)

	return &res[0]
}

func (s *Whatsmiau) getInstanceCached(id string) *models.Instance {
	// Try to get from cache first
	if cached, found := s.instanceCache.Get(id); found {
		instance := cached.(models.Instance)
		return &instance
	}

	// Cache miss - fetch from database
	ctx, c := context.WithTimeout(context.Background(), time.Second*5)
	defer c()

	res, err := s.repo.List(ctx, id)
	if err != nil {
		zap.L().Panic("failed to get instanceCached by instance", zap.Error(err))
	}

	if len(res) == 0 {
		zap.L().Warn("no instanceCached found by instance", zap.String("instance", id))
		return nil
	}

	// Store in cache with automatic TTL (5min expiration, 10min cleanup)
	s.instanceCache.Set(id, res[0], cache.DefaultExpiration)

	return &res[0]
}

func (s *Whatsmiau) startEmitter() {
	for event := range s.emitter {
		s.sendWebhookWithRetry(event)
	}
}

// sendWebhookWithRetry envia webhook com retry automático (3 tentativas: 2s, 5s, 10s)
func (s *Whatsmiau) sendWebhookWithRetry(event emitter) {
	maxRetries := 3
	retryDelays := []time.Duration{
		2 * time.Second,  // 1ª tentativa após erro
		5 * time.Second,  // 2ª tentativa após erro
		10 * time.Second, // 3ª tentativa após erro
	}

	data, err := json.Marshal(event.data)
	if err != nil {
		zap.L().Error("failed to marshal event", zap.Error(err))
		return
	}

	var lastErr error
	var lastStatusCode int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Se não é a primeira tentativa, aguarda o delay
		if attempt > 0 {
			delay := retryDelays[attempt-1]
			zap.L().Warn("retrying webhook after delay",
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay),
				zap.String("url", event.url))
			time.Sleep(delay)
		}

		// Cria nova requisição para cada tentativa
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, event.url, bytes.NewReader(data))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			zap.L().Error("webhook request creation failed",
				zap.Int("attempt", attempt+1),
				zap.Error(err))
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		// Envia requisição
		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to send request: %w", err)
			zap.L().Error("webhook send failed",
				zap.Int("attempt", attempt+1),
				zap.Error(err),
				zap.String("url", event.url))
			continue
		}

		lastStatusCode = resp.StatusCode

		// Lê resposta
		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Sucesso!
		if resp.StatusCode == http.StatusOK {
			if attempt > 0 {
				zap.L().Info("webhook succeeded after retry",
					zap.Int("attempt", attempt+1),
					zap.String("url", event.url))
			}
			return
		}

		// Falhou mas temos a resposta
		if readErr != nil {
			lastErr = fmt.Errorf("status %d, failed to read response", resp.StatusCode)
		} else {
			lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
		}

		zap.L().Warn("webhook failed",
			zap.Int("attempt", attempt+1),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("response", string(respBody)),
			zap.String("url", event.url))
	}

	// Todas as tentativas falharam
	zap.L().Error("webhook failed after all retries",
		zap.Int("totalAttempts", maxRetries+1),
		zap.Int("lastStatusCode", lastStatusCode),
		zap.Error(lastErr),
		zap.String("url", event.url))
}

func (s *Whatsmiau) emit(body any, url string) {
	s.emitter <- emitter{url, body}
}

// emitConnectionUpdate é uma função helper para enviar webhooks de status de conexão
func (s *Whatsmiau) emitConnectionUpdate(id string, status string) {
	instance := s.getInstanceCached(id)
	if instance == nil {
		zap.L().Warn("no instance found for connection event", zap.String("instance", id))
		return
	}

	// Verifica se o usuário se inscreveu para este evento específico
	eventMap := make(map[string]bool)
	for _, event := range instance.Webhook.Events {
		eventMap[event] = true
	}

	if !eventMap["CONNECTION_UPDATE"] {
		// O usuário não quer receber este evento, então não fazemos nada.
		return
	}

	// Monta o payload do webhook
	payload := &WookEvent[WookConnectionUpdateData]{
		Instance: instance.ID,
		Data: &WookConnectionUpdateData{
			Status: status, // "open" ou "close"
		},
		DateTime: time.Now(),
		Event:    WookConnectionUpdate,
	}

	zap.L().Info("emitting connection update", zap.String("instance", id), zap.String("status", status))
	s.emit(payload, instance.Webhook.Url)
}

func (s *Whatsmiau) Handle(id string) whatsmeow.EventHandler {
	return func(evt any) {
		s.handlerSemaphore <- struct{}{}
		go func() {
			defer func() { <-s.handlerSemaphore }()
			instance := s.getInstanceCached(id)
			if instance == nil {
				zap.L().Warn("no instance found for event", zap.String("instance", id))
				return
			}

			eventMap := make(map[string]bool)
			for _, event := range instance.Webhook.Events {
				eventMap[event] = true
			}

			switch e := evt.(type) {
			case *events.LoggedOut:
				s.handleLoggedOut(id)
			case *events.Message:
				s.handleMessageEvent(id, instance, e, eventMap)
			case *events.Receipt:
				s.handleReceiptEvent(id, instance, e, eventMap)
			case *events.BusinessName:
				s.handleBusinessNameEvent(id, instance, e, eventMap)
			case *events.Contact:
				s.handleContactEvent(id, instance, e, eventMap)
			case *events.Picture:
				s.handlePictureEvent(id, instance, e, eventMap)
			case *events.HistorySync:
				s.handleHistorySyncEvent(id, instance, e, eventMap)
			case *events.GroupInfo:
				s.handleGroupInfoEvent(id, instance, e, eventMap)
			case *events.PushName:
				s.handlePushNameEvent(id, instance, e, eventMap)
			default:
				zap.L().Debug("unknown event", zap.String("type", fmt.Sprintf("%T", evt)), zap.Any("raw", evt))
			}
		}()
	}
}

func (s *Whatsmiau) handleLoggedOut(id string) {
	// Emitir webhook de desconexão
	s.emitConnectionUpdate(id, "close")

	client, ok := s.clients.Load(id)
	if ok {
		if err := s.deleteDeviceIfExists(context.Background(), client); err != nil {
			zap.L().Error("failed to delete device for instance", zap.String("instance", id), zap.Error(err))
			return
		}
	}

	s.clients.Delete(id)
}
func (s *Whatsmiau) handleMessageEvent(id string, instance *models.Instance, e *events.Message, eventMap map[string]bool) {
	if !eventMap["MESSAGES_UPSERT"] {
		return
	}

	if canIgnoreGroup(e, instance) {
		return
	}

	if canIgnoreMessage(e) {
		return
	}

	messageData := s.convertEventMessage(id, instance, e)
	if messageData == nil {
		zap.L().Error("failed to convert event", zap.String("id", id), zap.String("type", fmt.Sprintf("%T", e)), zap.Any("raw", e))
		return
	}

	messageData.InstanceId = instance.ID

	// Enviar receipt de entrega automaticamente (2 vistos cinza)
	if e.Info.IsFromMe == false && e.Info.Chat.Server != "broadcast" {
		go s.sendDeliveryReceipt(id, e)
	}

	// Marcar como lido após 8 segundos se ReadMessages estiver ativado
	if instance.ReadMessages && e.Info.IsFromMe == false && e.Info.Chat.Server != "broadcast" {
		go s.markAsReadAfterDelay(id, e, 8*time.Second)
	}

	dateTime := time.Unix(int64(messageData.MessageTimestamp), 0)
	wookMessage := &WookEvent[WookMessageData]{
		Instance: instance.ID,
		Data:     messageData,
		DateTime: dateTime,
		Event:    WookMessagesUpsert,
	}

	if wookMessage.Data.Message != nil && len(wookMessage.Data.Message.Base64) > 0 {
		b64Temp := wookMessage.Data.Message.Base64
		wookMessage.Data.Message.Base64 = ""
		zap.L().Debug("message event", zap.String("instance", id), zap.Any("data", wookMessage.Data))
		wookMessage.Data.Message.Base64 = b64Temp
	} else if wookMessage.Data.Message != nil {
		zap.L().Debug("message event", zap.String("instance", id), zap.Any("data", wookMessage.Data))
	}

	s.emit(wookMessage, instance.Webhook.Url)
}

func (s *Whatsmiau) handleReceiptEvent(id string, instance *models.Instance, e *events.Receipt, eventMap map[string]bool) {
	if !eventMap["MESSAGES_UPDATE"] {
		return
	}

	if canIgnoreGroup(e, instance) {
		return
	}

	data := s.convertEventReceipt(id, e)
	if data == nil {
		return
	}

	for _, event := range data {
		wookData := &WookEvent[WookMessageUpdateData]{
			Instance: instance.ID,
			Data:     &event,
			DateTime: e.Timestamp,
			Event:    WookMessagesUpdate,
		}

		s.emit(wookData, instance.Webhook.Url)
	}
}

func (s *Whatsmiau) handleBusinessNameEvent(id string, instance *models.Instance, e *events.BusinessName, eventMap map[string]bool) {
	if !eventMap["CONTACTS_UPSERT"] {
		return
	}

	data := s.convertBusinessName(id, e)
	if data == nil {
		zap.L().Error("failed to convert business name", zap.String("id", id), zap.String("type", fmt.Sprintf("%T", e)), zap.Any("raw", e))
		return
	}

	wookData := &WookEvent[WookContactUpsertData]{
		Instance: instance.ID,
		Data:     &WookContactUpsertData{*data},
		DateTime: time.Now(),
		Event:    WookContactsUpsert,
	}

	s.emit(wookData, instance.Webhook.Url)
}

func (s *Whatsmiau) handleContactEvent(id string, instance *models.Instance, e *events.Contact, eventMap map[string]bool) {
	if !eventMap["CONTACTS_UPSERT"] {
		return
	}

	if canIgnoreGroup(e, instance) {
		return
	}

	data := s.convertContact(id, e)
	if data == nil {
		zap.L().Error("failed to convert contact", zap.String("id", id), zap.String("type", fmt.Sprintf("%T", e)), zap.Any("raw", e))
		return
	}

	wookData := &WookEvent[WookContactUpsertData]{
		Instance: instance.ID,
		Data:     &WookContactUpsertData{*data},
		DateTime: time.Now(),
		Event:    WookContactsUpsert,
	}

	s.emit(wookData, instance.Webhook.Url)
}

func (s *Whatsmiau) handlePictureEvent(id string, instance *models.Instance, e *events.Picture, eventMap map[string]bool) {
	if !eventMap["CONTACTS_UPSERT"] {
		return
	}

	data := s.convertPicture(id, e)
	if data == nil {
		return
	}

	wookData := &WookEvent[WookContactUpsertData]{
		Instance: instance.ID,
		Data:     &WookContactUpsertData{*data},
		DateTime: e.Timestamp,
		Event:    WookContactsUpsert,
	}

	s.emit(wookData, instance.Webhook.Url)
}

func (s *Whatsmiau) handleHistorySyncEvent(id string, instance *models.Instance, e *events.HistorySync, eventMap map[string]bool) {
	if !eventMap["CONTACTS_UPSERT"] {
		return
	}

	data := s.convertContactHistorySync(id, e.Data.GetPushnames(), e.Data.Conversations)
	if data == nil {
		return
	}

	wookData := &WookEvent[WookContactUpsertData]{
		Instance: instance.ID,
		Data:     &data,
		DateTime: time.Now(),
		Event:    WookContactsUpsert,
	}

	s.emit(wookData, instance.Webhook.Url)
}

func (s *Whatsmiau) handleGroupInfoEvent(id string, instance *models.Instance, e *events.GroupInfo, eventMap map[string]bool) {
	if !eventMap["CONTACTS_UPSERT"] {
		return
	}

	if instance.GroupsIgnore {
		return
	}

	data := s.convertGroupInfo(id, e)
	if data == nil {
		zap.L().Debug("failed to convert group info", zap.String("id", id), zap.String("type", fmt.Sprintf("%T", e)), zap.Any("raw", e))
		return
	}

	wookData := &WookEvent[WookContactUpsertData]{
		Instance: instance.ID,
		Data:     &WookContactUpsertData{*data},
		DateTime: time.Now(),
		Event:    WookContactsUpsert,
	}

	s.emit(wookData, instance.Webhook.Url)
}

func (s *Whatsmiau) handlePushNameEvent(id string, instance *models.Instance, e *events.PushName, eventMap map[string]bool) {
	if !eventMap["CONTACTS_UPSERT"] {
		return
	}

	if canIgnoreGroup(e, instance) {
		return
	}

	data := s.convertPushName(id, e)
	if data == nil {
		zap.L().Error("failed to convert pushname", zap.String("id", id), zap.String("type", fmt.Sprintf("%T", e)), zap.Any("raw", e))
		return
	}

	wookData := &WookEvent[WookContactUpsertData]{
		Instance: instance.ID,
		Data:     &WookContactUpsertData{*data},
		DateTime: time.Now(),
		Event:    WookContactsUpsert,
	}

	s.emit(wookData, instance.Webhook.Url)
}

// parseWAMessage converts a raw waE2E.Message into our internal representation.
// It only inspects the content of the protobuf message itself –
// media upload (URL/Base64 generation) is handled later by the caller.
func (s *Whatsmiau) parseWAMessage(m *waE2E.Message) (string, *WookMessageRaw, *waE2E.ContextInfo) {
	var messageType string
	raw := &WookMessageRaw{}
	var ci *waE2E.ContextInfo

	// === Prioritize action-like messages ===
	if r := m.GetReactionMessage(); r != nil {
		messageType = "reactionMessage"
		reactionKey := &WookKey{}
		if rk := r.GetKey(); rk != nil {
			reactionKey.RemoteJid = rk.GetRemoteJID()
			reactionKey.FromMe = rk.GetFromMe()
			reactionKey.Id = rk.GetID()
			reactionKey.Participant = rk.GetParticipant()
		}
		raw.ReactionMessage = &ReactionMessageRaw{
			Text:              r.GetText(),
			SenderTimestampMs: i64(r.GetSenderTimestampMS()),
			Key:               reactionKey,
		}
	} else if lr := m.GetListResponseMessage(); lr != nil {
		messageType = "listResponseMessage"
		listType := lr.GetListType().String()
		var selectedRowID string
		if ssr := lr.GetSingleSelectReply(); ssr != nil {
			selectedRowID = ssr.GetSelectedRowID()
		}
		raw.ListResponseMessage = &WookListMessageRaw{
			ListType: listType,
			SingleSelectReply: &WookListMessageRawListSingleSelectReply{
				SelectedRowId: selectedRowID,
			},
		}
	} else if img := m.GetImageMessage(); img != nil {
		messageType = "imageMessage"
		ci = img.GetContextInfo()
		raw.ImageMessage = &WookImageMessageRaw{
			Url:               img.GetURL(),
			Mimetype:          img.GetMimetype(),
			FileSha256:        b64(img.GetFileSHA256()),
			FileLength:        u64(img.GetFileLength()),
			Height:            int(img.GetHeight()),
			Width:             int(img.GetWidth()),
			Caption:           img.GetCaption(),
			MediaKey:          b64(img.GetMediaKey()),
			FileEncSha256:     b64(img.GetFileEncSHA256()),
			DirectPath:        img.GetDirectPath(),
			MediaKeyTimestamp: i64(img.GetMediaKeyTimestamp()),
			JpegThumbnail:     b64(img.GetJPEGThumbnail()),
			ViewOnce:          img.GetViewOnce(),
		}
	} else if aud := m.GetAudioMessage(); aud != nil {
		messageType = "audioMessage"
		ci = aud.GetContextInfo()
		raw.AudioMessage = &WookAudioMessageRaw{
			Url:               aud.GetURL(),
			Mimetype:          aud.GetMimetype(),
			FileSha256:        b64(aud.GetFileSHA256()),
			FileLength:        u64(aud.GetFileLength()),
			Seconds:           int(aud.GetSeconds()),
			Ptt:               aud.GetPTT(),
			MediaKey:          b64(aud.GetMediaKey()),
			FileEncSha256:     b64(aud.GetFileEncSHA256()),
			DirectPath:        aud.GetDirectPath(),
			MediaKeyTimestamp: i64(aud.GetMediaKeyTimestamp()),
			Waveform:          b64(aud.GetWaveform()),
			ViewOnce:          aud.GetViewOnce(),
		}
	} else if doc := m.GetDocumentMessage(); doc != nil {
		messageType = "documentMessage"
		ci = doc.GetContextInfo()
		raw.DocumentMessage = &WookDocumentMessageRaw{
			Url:               doc.GetURL(),
			Mimetype:          doc.GetMimetype(),
			Title:             doc.GetTitle(),
			FileSha256:        b64(doc.GetFileSHA256()),
			FileLength:        u64(doc.GetFileLength()),
			PageCount:         int(doc.GetPageCount()),
			MediaKey:          b64(doc.GetMediaKey()),
			FileName:          doc.GetFileName(),
			FileEncSha256:     b64(doc.GetFileEncSHA256()),
			DirectPath:        doc.GetDirectPath(),
			MediaKeyTimestamp: i64(doc.GetMediaKeyTimestamp()),
			ContactVcard:      doc.GetContactVcard(),
			JpegThumbnail:     b64(doc.GetJPEGThumbnail()),
			Caption:           doc.GetCaption(),
		}
	} else if video := m.GetVideoMessage(); video != nil {
		messageType = "videoMessage"
		raw.VideoMessage = &WookVideoMessageRaw{
			Url:           video.GetURL(),
			Mimetype:      video.GetMimetype(),
			Caption:       video.GetCaption(),
			FileSha256:    b64(video.GetFileSHA256()),
			FileLength:    u64(video.GetFileLength()),
			Seconds:       video.GetSeconds(),
			MediaKey:      b64(video.GetMediaKey()),
			FileEncSha256: b64(video.GetFileEncSHA256()),
			JPEGThumbnail: b64(video.GetJPEGThumbnail()),
			GIFPlayback:   video.GetGifPlayback(),
		}
		ci = video.GetContextInfo()
	} else if contact := m.GetContactMessage(); contact != nil {
		card, err := vcard.NewDecoder(strings.NewReader(contact.GetVcard())).Decode()
		if err != nil {
			zap.L().Error("decode card error", zap.Error(err))
		}

		messageType = "contactMessage"
		raw.ContactMessage = &ContactMessageRaw{
			VCard:        contact.GetVcard(),
			DisplayName:  contact.GetDisplayName(),
			DecodedVcard: card,
		}
		ci = contact.GetContextInfo()
	} else if contactArray := m.GetContactsArrayMessage(); contactArray != nil {
		messageType = "contactsArrayMessage"
		var contacts []ContactMessageRaw
		for _, contact := range contactArray.Contacts {
			card, err := vcard.NewDecoder(strings.NewReader(contact.GetVcard())).Decode()
			if err != nil {
				zap.L().Error("decode card error", zap.Error(err))
			}

			contacts = append(contacts, ContactMessageRaw{
				VCard:        contact.GetVcard(),
				DisplayName:  contact.GetDisplayName(),
				DecodedVcard: card,
			})
		}
		raw.ContactsArrayMessage = &ContactsArrayMessageRaw{
			DisplayName: contactArray.GetDisplayName(),
			Contacts:    contacts,
		}
		ci = contactArray.GetContextInfo()
	} else if conv := strings.TrimSpace(m.GetConversation()); conv != "" {
		messageType = "conversation"
		raw.Conversation = conv
	} else if et := m.GetExtendedTextMessage(); et != nil && len(et.GetText()) > 0 {
		messageType = "conversation"
		raw.Conversation = et.GetText()
		ci = et.GetContextInfo()
	} else {
		messageType = "unknown"
	}

	return messageType, raw, ci
}

func (s *Whatsmiau) convertContactHistorySync(id string, event []*waHistorySync.Pushname, conversations []*waHistorySync.Conversation) WookContactUpsertData {
	resultMap := make(map[string]WookContact)
	for _, pushName := range event {

		if len(pushName.GetPushname()) == 0 {
			continue
		}

		if dt := strings.Split(pushName.GetPushname(), "@"); len(dt) == 2 && (dt[1] == "g.us" || dt[1] == "s.whatsapp.net") {
			return nil
		}

		jid, err := types.ParseJID(pushName.GetID())
		if err != nil {
			zap.L().Error("failed to parse jid", zap.String("pushname", pushName.GetPushname()))
			return nil
		}

		jidParsed, lid := s.GetJidLid(context.Background(), id, jid)

		resultMap[jidParsed] = WookContact{
			RemoteJid:  jidParsed,
			PushName:   pushName.GetPushname(),
			InstanceId: id,
			RemoteLid:  lid,
		}
	}

	for _, conversation := range conversations {
		name := conversation.GetName()
		if len(name) == 0 {
			name = conversation.GetDisplayName()
		}
		if len(name) == 0 {
			name = conversation.GetUsername()
		}
		if len(name) == 0 {
			continue
		}
		if dt := strings.Split(name, "@"); len(dt) == 2 && (dt[1] == "g.us" || dt[1] == "s.whatsapp.net") {
			return nil
		}

		jid, err := types.ParseJID(conversation.GetID())
		if err != nil {
			zap.L().Error("failed to parse jid", zap.String("name", conversation.GetName()))
			return nil
		}
		jidParsed, lid := s.GetJidLid(context.Background(), id, jid)

		resultMap[conversation.GetID()] = WookContact{
			RemoteJid:  jidParsed,
			PushName:   name,
			InstanceId: id,
			RemoteLid:  lid,
		}
	}

	var result []WookContact
	for _, c := range resultMap {
		jid, err := types.ParseJID(c.RemoteJid)
		if err != nil {
			continue
		}

		url, _, err := s.getPic(id, jid)
		if err != nil {
			zap.L().Error("failed to get pic", zap.Error(err))
		}

		c.ProfilePicUrl = url
		result = append(result, c)
	}

	return result
}

func (s *Whatsmiau) convertEventMessage(id string, instance *models.Instance, evt *events.Message) *WookMessageData {
	ctx, c := context.WithTimeout(context.Background(), time.Second*60)
	defer c()

	client, ok := s.clients.Load(id)
	if !ok {
		zap.L().Warn("no client for event", zap.String("id", id))
		return nil
	}

	if evt == nil || evt.Message == nil {
		return nil
	}

	jid, lid := s.GetJidLid(ctx, id, evt.Info.Chat)
	senderJid, _ := s.GetJidLid(ctx, id, evt.Info.Sender)

	// Always unwrap to work with the real content
	e := evt.UnwrapRaw()
	m := e.Message

	// Build the key
	key := &WookKey{
		RemoteJid:   jid,
		RemoteLid:   lid,
		FromMe:      e.Info.IsFromMe,
		Id:          e.Info.ID,
		Participant: senderJid,
	}

	// Determine status
	status := "received"
	if e.Info.IsFromMe {
		status = "sent"
	}

	// Timestamp
	ts := e.Info.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	// Convert the WA protobuf message into our internal raw structure
	messageType, raw, ci := s.parseWAMessage(m)

	// Upload media (URL / Base64) when needed
	switch messageType {
	case "imageMessage":
		if img := m.GetImageMessage(); img != nil {
			raw.MediaURL, raw.Base64 = s.uploadMessageFile(ctx, instance, client, img, img.GetMimetype(), "")
		}
	case "audioMessage":
		if aud := m.GetAudioMessage(); aud != nil {
			raw.MediaURL, raw.Base64 = s.uploadMessageFile(ctx, instance, client, aud, aud.GetMimetype(), "")
		}
	case "documentMessage":
		if doc := m.GetDocumentMessage(); doc != nil {
			raw.MediaURL, raw.Base64 = s.uploadMessageFile(ctx, instance, client, doc, doc.GetMimetype(), doc.GetFileName())
		}
	case "videoMessage":
		if vid := m.GetVideoMessage(); vid != nil {
			raw.MediaURL, raw.Base64 = s.uploadMessageFile(ctx, instance, client, vid, vid.GetMimetype(), "")
		}
	}

	// Map MessageContextInfo (quoted, mentions, disappearing mode, external ad reply)
	var messageContext WookMessageContextInfo
	if ci != nil {
		messageContext.EphemeralSettingTimestamp = i64(ci.GetEphemeralSettingTimestamp())
		messageContext.StanzaId = ci.GetStanzaID()
		messageContext.Participant = ci.GetParticipant()
		messageContext.Expiration = int(ci.GetExpiration())
		messageContext.MentionedJid = ci.GetMentionedJID()
		messageContext.ConversionSource = ci.GetConversionSource()
		messageContext.ConversionData = b64(ci.GetConversionData())
		messageContext.ConversionDelaySeconds = int(ci.GetConversionDelaySeconds())
		messageContext.EntryPointConversionSource = ci.GetEntryPointConversionSource()
		messageContext.EntryPointConversionApp = ci.GetEntryPointConversionApp()
		messageContext.EntryPointConversionDelaySeconds = int(ci.GetEntryPointConversionDelaySeconds())
		messageContext.TrustBannerAction = ci.GetTrustBannerAction()

		if dm := ci.GetDisappearingMode(); dm != nil {
			messageContext.DisappearingMode = &ContextInfoDisappearingMode{
				Initiator:     dm.GetInitiator().String(),
				Trigger:       dm.GetTrigger().String(),
				InitiatedByMe: dm.GetInitiatedByMe(),
			}
		}

		if ear := ci.GetExternalAdReply(); ear != nil {
			messageType = "conversation"
			messageContext.ExternalAdReply = &WookMessageContextInfoExternalAdReply{
				Title:                 ear.GetTitle(),
				Body:                  ear.GetBody(),
				MediaType:             ear.GetMediaType().String(),
				ThumbnailUrl:          ear.GetThumbnailURL(),
				Thumbnail:             b64(ear.GetThumbnail()),
				SourceType:            ear.GetSourceType(),
				SourceId:              ear.GetSourceID(),
				SourceUrl:             ear.GetSourceURL(),
				ContainsAutoReply:     ear.GetContainsAutoReply(),
				RenderLargerThumbnail: ear.GetRenderLargerThumbnail(),
				ShowAdAttribution:     ear.GetShowAdAttribution(),
				CtwaClid:              ear.GetCtwaClid(),
			}
		}

		if qm := ci.GetQuotedMessage(); qm != nil {
			_, qmRaw, _ := s.parseWAMessage(qm)
			messageContext.QuotedMessage = qmRaw
		}
	}

	return &WookMessageData{
		Key:              key,
		PushName:         strings.TrimSpace(e.Info.PushName),
		Status:           status,
		Message:          raw,
		ContextInfo:      &messageContext,
		MessageType:      messageType,
		MessageTimestamp: int(ts.Unix()),
		InstanceId:       id,
		Source:           "whatsapp",
	}
}

func (s *Whatsmiau) convertEventReceipt(id string, evt *events.Receipt) []WookMessageUpdateData {
	var status WookMessageUpdateStatus
	switch evt.Type {
	case types.ReceiptTypeRead:
		status = MessageStatusRead
	case types.ReceiptTypeDelivered:
		status = MessageStatusDeliveryAck
	default:
		return nil
	}

	chatJid, chatLid := s.GetJidLid(context.Background(), id, evt.Chat)
	participantJid, _ := s.GetJidLid(context.Background(), id, evt.Sender)

	var result []WookMessageUpdateData
	for _, messageID := range evt.MessageIDs {
		result = append(result, WookMessageUpdateData{
			MessageId:   messageID,
			KeyId:       messageID,
			RemoteJid:   chatJid,
			RemoteLid:   chatLid,
			FromMe:      evt.IsFromMe,
			Participant: participantJid,
			Status:      status,
			InstanceId:  id,
		})
	}

	return result
}

func (s *Whatsmiau) uploadMessageFile(ctx context.Context, instance *models.Instance, client *whatsmeow.Client, fileMessage whatsmeow.DownloadableMessage, mimetype, fileName string) (string, string) {
	var (
		b64Result string
		urlResult string
		ext       string
	)

	tmpFile, err := os.CreateTemp("", "file-*")
	if err != nil {
		panic(err)
	}

	defer os.Remove(tmpFile.Name())
	if err := client.DownloadToFile(ctx, fileMessage, tmpFile); err != nil {
		zap.L().Error("failed to download image", zap.Error(err))
		return "", ""
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		zap.L().Error("failed to seek image", zap.Error(err))
	}

	ext = extractExtFromFile(fileName, mimetype, tmpFile)
	if instance.Webhook.Base64 != nil && *instance.Webhook.Base64 {
		data, err := io.ReadAll(tmpFile)
		if err != nil {
			zap.L().Error("failed to read image", zap.Error(err))
		} else {
			b64Result = base64.StdEncoding.EncodeToString(data)
		}
	}
	if s.fileStorage != nil {
		if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
			zap.L().Error("failed to seek image", zap.Error(err))
		}

		urlResult, _, err = s.fileStorage.Upload(ctx, uuid.NewString()+"."+ext, mimetype, tmpFile)
		if err != nil {
			zap.L().Error("failed to upload image", zap.Error(err))
		}
	}

	return urlResult, b64Result
}

func (s *Whatsmiau) convertContact(id string, evt *events.Contact) *WookContact {
	url, _, err := s.getPic(id, evt.JID)
	if err != nil {
		zap.L().Error("failed to get pic", zap.Error(err))
	}

	name := evt.Action.GetFirstName()
	if name == "" {
		name = evt.Action.GetFullName()
	}
	if name == "" {
		name = evt.Action.GetUsername()
	}
	if name == "" {
		return nil
	}

	if dt := strings.Split(name, "@"); len(dt) == 2 && (dt[1] == "g.us" || dt[1] == "s.whatsapp.net") {
		return nil
	}

	jid, lid := s.GetJidLid(context.Background(), id, evt.JID)
	return &WookContact{
		RemoteJid:     jid,
		RemoteLid:     lid,
		PushName:      name,
		ProfilePicUrl: url,
		InstanceId:    id,
	}
}

func (s *Whatsmiau) convertGroupInfo(id string, evt *events.GroupInfo) *WookContact {
	url, _, err := s.getPic(id, evt.JID)
	if err != nil {
		zap.L().Error("failed to get pic", zap.Error(err))
	}

	if evt.Name == nil || len(evt.Name.Name) == 0 {
		return nil
	}

	if dt := strings.Split(evt.Name.Name, "@"); len(dt) == 2 && (dt[1] == "g.us" || dt[1] == "s.whatsapp.net") {
		return nil
	}

	jid, lid := s.GetJidLid(context.Background(), id, evt.JID)

	return &WookContact{
		RemoteJid:     jid,
		PushName:      evt.Name.Name,
		ProfilePicUrl: url,
		InstanceId:    id,
		RemoteLid:     lid,
	}
}

func (s *Whatsmiau) convertPushName(id string, evt *events.PushName) *WookContact {
	url, _, err := s.getPic(id, evt.JID)
	if err != nil {
		zap.L().Error("failed to get pic", zap.Error(err))
	}

	name := evt.NewPushName
	if len(name) == 0 {
		name = evt.OldPushName
	}

	if name == "" {
		return nil
	}

	if dt := strings.Split(name, "@"); len(dt) == 2 && (dt[1] == "g.us" || dt[1] == "s.whatsapp.net") {
		return nil
	}

	jid, lid := s.GetJidLid(context.Background(), id, evt.JID)

	return &WookContact{
		RemoteJid:     jid,
		PushName:      evt.NewPushName,
		InstanceId:    id,
		ProfilePicUrl: url,
		RemoteLid:     lid,
	}
}

func (s *Whatsmiau) convertPicture(id string, evt *events.Picture) *WookContact {
	url, b64, err := s.getPic(id, evt.JID)
	if err != nil {
		zap.L().Error("failed to get pic", zap.Error(err))
	}

	if len(url) <= 0 {
		return nil
	}

	jid, lid := s.GetJidLid(context.Background(), id, evt.JID)

	return &WookContact{
		RemoteJid:     jid,
		InstanceId:    id,
		Base64Pic:     b64,
		ProfilePicUrl: url,
		RemoteLid:     lid,
	}
}

func (s *Whatsmiau) convertBusinessName(id string, evt *events.BusinessName) *WookContact {
	url, b64, err := s.getPic(id, evt.JID)
	if err != nil {
		zap.L().Error("failed to get pic", zap.Error(err))
	}

	name := evt.NewBusinessName
	if name == "" {
		name = evt.OldBusinessName
	}
	if name == "" && evt.Message != nil {
		name = evt.Message.PushName
	}
	if name == "" && evt.Message != nil && evt.Message.VerifiedName != nil && evt.Message.VerifiedName.Details != nil {
		name = evt.Message.VerifiedName.Details.GetVerifiedName()
	}

	if dt := strings.Split(name, "@"); len(dt) == 2 && (dt[1] == "g.us" || dt[1] == "s.whatsapp.net") {
		return nil
	}

	jid, lid := s.GetJidLid(context.Background(), id, evt.JID)

	return &WookContact{
		RemoteJid:     jid,
		InstanceId:    id,
		Base64Pic:     b64,
		ProfilePicUrl: url,
		PushName:      name,
		RemoteLid:     lid,
	}
}

// sendDeliveryReceipt envia automaticamente o receipt de entrega (2 vistos cinza)
func (s *Whatsmiau) sendDeliveryReceipt(id string, e *events.Message) {
	client, ok := s.clients.Load(id)
	if !ok || client == nil {
		return
	}

	err := client.MarkRead(context.TODO(), []types.MessageID{e.Info.ID}, e.Info.Timestamp, e.Info.Chat, e.Info.Sender, types.ReceiptTypeDelivered)
	if err != nil {
		zap.L().Debug("failed to send delivery receipt",
			zap.String("instance", id),
			zap.String("chat", e.Info.Chat.String()),
			zap.Error(err))
	}
}

// markAsReadAfterDelay marca mensagem como lida após delay (2 vistos azuis)
func (s *Whatsmiau) markAsReadAfterDelay(id string, e *events.Message, delay time.Duration) {
	time.Sleep(delay)

	client, ok := s.clients.Load(id)
	if !ok || client == nil {
		return
	}

	err := client.MarkRead(context.TODO(), []types.MessageID{e.Info.ID}, time.Now(), e.Info.Chat, e.Info.Sender)
	if err != nil {
		zap.L().Debug("failed to mark message as read",
			zap.String("instance", id),
			zap.String("chat", e.Info.Chat.String()),
			zap.Error(err))
	}
}

func (s *Whatsmiau) getPic(id string, jid types.JID) (string, string, error) {
	client, ok := s.clients.Load(id)
	if !ok || client == nil {
		zap.L().Warn("no client for event", zap.String("id", id))
		return "", "", fmt.Errorf("no client for event %s", id)
	}

	pic, err := client.GetProfilePictureInfo(context.TODO(), jid, &whatsmeow.GetProfilePictureParams{
		Preview:     true,
		IsCommunity: false,
	})
	if err != nil {
		return "", "", nil
	}

	if pic == nil {
		return "", "", err
	}

	res, err := s.httpClient.Get(pic.URL)
	if err != nil {
		zap.L().Error("get profile picture error", zap.String("id", id), zap.Error(err))
		return "", "", err
	}

	picRaw, err := io.ReadAll(res.Body)
	if err != nil {
		zap.L().Error("get profile picture error", zap.String("id", id), zap.Error(err))
		return "", "", err
	}

	return pic.URL, base64.StdEncoding.EncodeToString(picRaw), nil
}

// keepAlwaysOnlineManager gerencia presence de todas as instâncias de forma centralizada
// Otimizado para alta escala: 15min interval + batching com 20 workers
func (s *Whatsmiau) keepAlwaysOnlineManager() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	zap.L().Info("AlwaysOnline manager started", zap.Int("interval_minutes", 15))

	// Processar imediatamente na inicialização
	s.processAlwaysOnlineInstances()

	for range ticker.C {
		s.processAlwaysOnlineInstances()
	}
}

// processAlwaysOnlineInstances processa todas as instâncias AlwaysOnline em lotes
func (s *Whatsmiau) processAlwaysOnlineInstances() {
	var instanceIDs []string

	// Coletar todos os IDs das instâncias AlwaysOnline
	s.alwaysOnlineIDs.Range(func(id string, enabled bool) bool {
		if enabled {
			instanceIDs = append(instanceIDs, id)
		}
		return true
	})

	if len(instanceIDs) == 0 {
		return
	}

	zap.L().Info("Processing AlwaysOnline batch",
		zap.Int("total_instances", len(instanceIDs)))

	// Semáforo para limitar workers concorrentes (20 workers simultâneos)
	semaphore := make(chan struct{}, 20)
	var wg sync.WaitGroup

	for _, id := range instanceIDs {
		wg.Add(1)
		go func(instanceID string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			client, ok := s.clients.Load(instanceID)
			if !ok || client == nil {
				// Cliente não existe mais - remover do tracking
				s.alwaysOnlineIDs.Delete(instanceID)
				return
			}

			if !client.IsConnected() || !client.IsLoggedIn() {
				zap.L().Debug("Skipping AlwaysOnline - client not connected",
					zap.String("instance", instanceID))
				return
			}

			// Enviar presence "available" (online)
			err := client.SendPresence(context.TODO(), types.PresenceAvailable)
			if err != nil {
				zap.L().Debug("Failed to send presence in batch",
					zap.String("instance", instanceID),
					zap.Error(err))
			}
		}(id)
	}

	wg.Wait()
	zap.L().Info("AlwaysOnline batch completed",
		zap.Int("processed", len(instanceIDs)))
}

// enableAlwaysOnline ativa AlwaysOnline para uma instância
func (s *Whatsmiau) enableAlwaysOnline(id string) {
	s.alwaysOnlineIDs.Store(id, true)
	zap.L().Info("AlwaysOnline enabled", zap.String("instance", id))
}

// disableAlwaysOnline desativa AlwaysOnline para uma instância
func (s *Whatsmiau) disableAlwaysOnline(id string) {
	s.alwaysOnlineIDs.Delete(id)
	zap.L().Info("AlwaysOnline disabled", zap.String("instance", id))
}
