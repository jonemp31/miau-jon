package whatsmiau

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type SendText struct {
	Text           string     `json:"text"`
	InstanceID     string     `json:"instance_id"`
	RemoteJID      *types.JID `json:"remote_jid"`
	QuoteMessageID string     `json:"quote_message_id"`
	QuoteMessage   string     `json:"quote_message"`
	Participant    *types.JID `json:"participant"`
}

type SendTextResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Whatsmiau) SendText(ctx context.Context, data *SendText) (*SendTextResponse, error) {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return nil, whatsmeow.ErrClientIsNil
	}

	//rJid := data.RemoteJID.ToNonAD().String()
	var extendedMessage *waE2E.ExtendedTextMessage
	if len(data.QuoteMessage) > 0 && len(data.QuoteMessageID) > 0 {
		extendedMessage = &waE2E.ExtendedTextMessage{
			//ContextInfo: &waE2E.ContextInfo{ // TODO: implement quoted message
			//	StanzaID:    &data.QuoteMessageID,
			//	Participant: &rJid,
			//	QuotedMessage: &waE2E.Message{
			//		Conversation: &data.QuoteMessage,
			//		ProtocolMessage: &waE2E.ProtocolMessage{
			//			Key: &waCommon.MessageKey{
			//				RemoteJID:   &rJid,
			//				FromMe:      &[]bool{true}[0],
			//				ID:          &data.QuoteMessageID,
			//				Participant: nil,
			//			},
			//		},
			//	},
			//},
		}
	}

	res, err := client.SendMessage(ctx, *data.RemoteJID, &waE2E.Message{
		Conversation:        &data.Text,
		ExtendedTextMessage: extendedMessage,
	})
	if err != nil {
		return nil, err
	}

	return &SendTextResponse{
		ID:        res.ID,
		CreatedAt: res.Timestamp,
	}, nil
}

type SendAudioRequest struct {
	AudioURL       string     `json:"text"`
	InstanceID     string     `json:"instance_id"`
	RemoteJID      *types.JID `json:"remote_jid"`
	QuoteMessageID string     `json:"quote_message_id"`
	QuoteMessage   string     `json:"quote_message"`
	Participant    *types.JID `json:"participant"`
	ViewOnce       bool       `json:"view_once"`
}

type SendAudioResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Whatsmiau) SendAudio(ctx context.Context, data *SendAudioRequest) (*SendAudioResponse, error) {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return nil, whatsmeow.ErrClientIsNil
	}

	var audioData []byte
	var waveForm []byte
	var secs float64
	var err error

	// Detectar se é base64 data URI ou URL
	if strings.HasPrefix(data.AudioURL, "data:audio/ogg") || strings.HasPrefix(data.AudioURL, "data:audio/opus") {
		// É base64 data URI - extrair e usar diretamente
		audioData, waveForm, secs, err = processBase64Audio(data.AudioURL)
		if err != nil {
			return nil, err
		}
	} else {
		// É URL - baixar e converter
		resAudio, err := s.getCtx(ctx, data.AudioURL)
		if err != nil {
			return nil, err
		}

		dataBytes, err := io.ReadAll(resAudio.Body)
		if err != nil {
			return nil, err
		}

		audioData, waveForm, secs, err = convertAudio(dataBytes, 64)
		if err != nil {
			return nil, err
		}
	}

	uploaded, err := client.Upload(ctx, audioData, whatsmeow.MediaAudio)
	if err != nil {
		return nil, err
	}

	audio := waE2E.AudioMessage{
		URL:           proto.String(uploaded.URL),
		Mimetype:      proto.String("audio/ogg; codecs=opus"),
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
		Seconds:       proto.Uint32(uint32(secs)),
		PTT:           proto.Bool(true),
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		DirectPath:    proto.String(uploaded.DirectPath),
		Waveform:      waveForm,
		ViewOnce:      proto.Bool(data.ViewOnce),
	}

	res, err := client.SendMessage(ctx, *data.RemoteJID, &waE2E.Message{
		AudioMessage: &audio,
	})
	if err != nil {
		return nil, err
	}

	return &SendAudioResponse{
		ID:        res.ID,
		CreatedAt: res.Timestamp,
	}, nil
}

type SendDocumentRequest struct {
	InstanceID string     `json:"instance_id"`
	MediaURL   string     `json:"media_url"`
	Caption    string     `json:"caption"`
	FileName   string     `json:"file_name"`
	RemoteJID  *types.JID `json:"remote_jid"`
	Mimetype   string     `json:"mimetype"`
}

type SendDocumentResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Whatsmiau) SendDocument(ctx context.Context, data *SendDocumentRequest) (*SendDocumentResponse, error) {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return nil, whatsmeow.ErrClientIsNil
	}

	resMedia, err := s.getCtx(ctx, data.MediaURL)
	if err != nil {
		return nil, err
	}

	dataBytes, err := io.ReadAll(resMedia.Body)
	if err != nil {
		return nil, err
	}

	uploaded, err := client.Upload(ctx, dataBytes, whatsmeow.MediaDocument)
	if err != nil {
		return nil, err
	}

	doc := waE2E.DocumentMessage{
		URL:           proto.String(uploaded.URL),
		Mimetype:      proto.String(data.Mimetype),
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
		MediaKey:      uploaded.MediaKey,
		FileName:      &data.FileName,
		FileEncSHA256: uploaded.FileEncSHA256,
		DirectPath:    proto.String(uploaded.DirectPath),
		Caption:       proto.String(data.Caption),
	}

	res, err := client.SendMessage(ctx, *data.RemoteJID, &waE2E.Message{
		DocumentMessage: &doc,
	})
	if err != nil {
		return nil, err
	}

	return &SendDocumentResponse{
		ID:        res.ID,
		CreatedAt: res.Timestamp,
	}, nil
}

type SendImageRequest struct {
	InstanceID string     `json:"instance_id"`
	MediaURL   string     `json:"media_url"`
	Caption    string     `json:"caption"`
	RemoteJID  *types.JID `json:"remote_jid"`
	Mimetype   string     `json:"mimetype"`
	ViewOnce   bool       `json:"view_once"`
}
type SendImageResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Whatsmiau) SendImage(ctx context.Context, data *SendImageRequest) (*SendImageResponse, error) {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return nil, whatsmeow.ErrClientIsNil
	}

	var dataBytes []byte
	var err error

	// Detectar se é base64 data URI ou URL
	if strings.HasPrefix(data.MediaURL, "data:image/") {
		// É base64 data URI - extrair e usar diretamente
		dataBytes, err = processBase64Image(data.MediaURL)
		if err != nil {
			return nil, err
		}
	} else {
		// É URL - baixar normalmente
		resMedia, err := s.getCtx(ctx, data.MediaURL)
		if err != nil {
			return nil, err
		}

		dataBytes, err = io.ReadAll(resMedia.Body)
		if err != nil {
			return nil, err
		}
	}

	uploaded, err := client.Upload(ctx, dataBytes, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	if data.Mimetype == "" {
		mimetype, err := extractMimetype(dataBytes, uploaded.URL)
		if err == nil {
			data.Mimetype = mimetype
		}
	}

	doc := waE2E.ImageMessage{
		URL:           proto.String(uploaded.URL),
		Mimetype:      proto.String(data.Mimetype),
		Caption:       proto.String(data.Caption),
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		DirectPath:    proto.String(uploaded.DirectPath),
		ViewOnce:      proto.Bool(data.ViewOnce),
	}

	res, err := client.SendMessage(ctx, *data.RemoteJID, &waE2E.Message{
		ImageMessage: &doc,
	})
	if err != nil {
		return nil, err
	}

	return &SendImageResponse{
		ID:        res.ID,
		CreatedAt: res.Timestamp,
	}, nil
}

type SendReactionRequest struct {
	InstanceID string     `json:"instance_id"`
	Reaction   string     `json:"reaction"`
	RemoteJID  *types.JID `json:"remote_jid"`
	MessageID  string     `json:"message_id"`
	FromMe     bool       `json:"from_me"`
}

type SendReactionResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Whatsmiau) SendReaction(ctx context.Context, data *SendReactionRequest) (*SendReactionResponse, error) {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return nil, whatsmeow.ErrClientIsNil
	}

	if len(data.Reaction) <= 0 {
		return nil, fmt.Errorf("empty reaction, len: %d", len(data.Reaction))
	}

	if len(data.MessageID) <= 0 {
		return nil, fmt.Errorf("invalid message_id")
	}

	if client.Store == nil || client.Store.ID == nil {
		return nil, fmt.Errorf("device is not connected")
	}

	sender := data.RemoteJID
	if data.FromMe {
		sender = client.Store.ID
	}

	doc := client.BuildReaction(*data.RemoteJID, *sender, data.MessageID, data.Reaction)
	res, err := client.SendMessage(ctx, *data.RemoteJID, doc)
	if err != nil {
		return nil, err
	}

	return &SendReactionResponse{
		ID:        res.ID,
		CreatedAt: res.Timestamp,
	}, nil
}

type SendVideoRequest struct {
	InstanceID string     `json:"instance_id"`
	MediaURL   string     `json:"media_url"`
	Caption    string     `json:"caption"`
	RemoteJID  *types.JID `json:"remote_jid"`
	Mimetype   string     `json:"mimetype"`
	ViewOnce   bool       `json:"view_once"`
}

type SendVideoResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Whatsmiau) SendVideo(ctx context.Context, data *SendVideoRequest) (*SendVideoResponse, error) {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return nil, whatsmeow.ErrClientIsNil
	}

	resMedia, err := s.getCtx(ctx, data.MediaURL)
	if err != nil {
		return nil, err
	}

	dataBytes, err := io.ReadAll(resMedia.Body)
	if err != nil {
		return nil, err
	}

	uploaded, err := client.Upload(ctx, dataBytes, whatsmeow.MediaVideo)
	if err != nil {
		return nil, err
	}

	if data.Mimetype == "" {
		mimetype, err := extractMimetype(dataBytes, uploaded.URL)
		if err != nil {
			data.Mimetype = "video/mp4" // fallback to mp4
		} else {
			data.Mimetype = mimetype
		}
	}

	video := waE2E.VideoMessage{
		URL:           proto.String(uploaded.URL),
		Mimetype:      proto.String(data.Mimetype),
		Caption:       proto.String(data.Caption),
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uploaded.FileLength),
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		DirectPath:    proto.String(uploaded.DirectPath),
		ViewOnce:      proto.Bool(data.ViewOnce),
	}

	res, err := client.SendMessage(ctx, *data.RemoteJID, &waE2E.Message{
		VideoMessage: &video,
	})
	if err != nil {
		return nil, err
	}

	return &SendVideoResponse{
		ID:        res.ID,
		CreatedAt: res.Timestamp,
	}, nil
}

// ============================================================================
// ⚠️ EXPERIMENTAL ENDPOINT - LABORATORY USE ONLY
// ============================================================================
// SendMissedCall simulates a missed call notification by sending a protocol message.
//
// ⚠️ WARNING: This is an EXPERIMENTAL feature for testing purposes only!
//
// RISKS:
// - May violate WhatsApp Terms of Service if used in production
// - WhatsApp may detect this as abuse and ban the account
// - Behavior may change without notice (not officially supported)
// - Should NOT be used for spam or unsolicited notifications
//
// RECOMMENDATION:
// - Use ONLY with development/test instances
// - Use ONLY with explicit user permission
// - Limit usage to 1 per conversation per day maximum
// - Monitor account status for any warnings
// - Have a backup plan (regular text message)
//
// This sends a ProtocolMessage with Type: CALL that tells the recipient's phone
// "a call with this ID just happened" - the phone displays it as "Missed Call"
// because there was no actual active call.
// ============================================================================

type SendMissedCallRequest struct {
	InstanceID string     `json:"instance_id"`
	RemoteJID  *types.JID `json:"remote_jid"`
	VideoCall  bool       `json:"video_call"` // true = video call, false = voice call
}

type SendMissedCallResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Whatsmiau) SendMissedCall(ctx context.Context, data *SendMissedCallRequest) (*SendMissedCallResponse, error) {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return nil, whatsmeow.ErrClientIsNil
	}

	// AVISO: Este código usa estruturas internas do protobuf WhatsApp
	// Pode violar os Termos de Serviço - use apenas para desenvolvimento/testes

	// Definir o tipo de chamada (voz ou vídeo)
	isVideo := proto.Bool(data.VideoCall)

	// CallOutcome = MISSED (valor 1 no enum)
	missedOutcome := waE2E.CallLogMessage_CallOutcome(1).Enum() // 1 = MISSED

	// CallType = REGULAR (chamada comum, não agendada)
	regularType := waE2E.CallLogMessage_REGULAR.Enum()

	// Duração 0 para chamada perdida
	duration := proto.Int64(0)

	// Criar a mensagem de log de chamada
	callLogMsg := &waE2E.CallLogMessage{
		IsVideo:      isVideo,
		CallOutcome:  missedOutcome,
		CallType:     regularType,
		DurationSecs: duration,
	}

	// Enviar a mensagem
	res, err := client.SendMessage(ctx, *data.RemoteJID, &waE2E.Message{
		CallLogMesssage: callLogMsg, // Nota: campo tem typo "Messsage" no protobuf original
	})
	if err != nil {
		return nil, err
	}

	return &SendMissedCallResponse{
		ID:        res.ID,
		CreatedAt: res.Timestamp,
	}, nil
}
