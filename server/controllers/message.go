package controllers

import (
	"net/http"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/verbeux-ai/whatsmiau/interfaces"
	"github.com/verbeux-ai/whatsmiau/lib/whatsmiau"
	"github.com/verbeux-ai/whatsmiau/server/dto"
	"github.com/verbeux-ai/whatsmiau/utils"
	"go.mau.fi/whatsmeow/types"
	"go.uber.org/zap"
)

type Message struct {
	repo      interfaces.InstanceRepository
	whatsmiau *whatsmiau.Whatsmiau
}

func NewMessages(repository interfaces.InstanceRepository, whatsmiau *whatsmiau.Whatsmiau) *Message {
	return &Message{
		repo:      repository,
		whatsmiau: whatsmiau,
	}
}

func (s *Message) SendText(ctx echo.Context) error {
	var request dto.SendTextRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	if err := validator.New().Struct(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid request body")
	}

	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid number format")
	}

	sendText := &whatsmiau.SendText{
		Text:       request.Text,
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
	}

	if request.Quoted != nil && len(request.Quoted.Key.Id) > 0 && len(request.Quoted.Message.Conversation) > 0 {
		sendText.QuoteMessage = request.Quoted.Message.Conversation
		sendText.QuoteMessageID = request.Quoted.Key.Id
	}

	c := ctx.Request().Context()
	if err := s.whatsmiau.ChatPresence(&whatsmiau.ChatPresenceRequest{
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
		Presence:   types.ChatPresenceComposing,
	}); err != nil {
		zap.L().Error("Whatsmiau.ChatPresence", zap.Error(err))
	} else {
		time.Sleep(time.Millisecond * time.Duration(request.Delay)) // TODO: create a more robust solution
	}

	res, err := s.whatsmiau.SendText(c, sendText)
	if err != nil {
		zap.L().Error("Whatsmiau.SendText failed", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to send text")
	}

	return ctx.JSON(http.StatusOK, dto.SendTextResponse{
		Key: dto.MessageResponseKey{
			RemoteJid: request.Number,
			FromMe:    true,
			Id:        res.ID,
		},
		Status: "sent",
		Message: dto.SendTextResponseMessage{
			Conversation: request.Text,
		},
		MessageType:      "conversation",
		MessageTimestamp: int(res.CreatedAt.Unix() / 1000),
		InstanceId:       request.InstanceID,
	})
}

func (s *Message) SendAudio(ctx echo.Context) error {
	var request dto.SendAudioRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	if err := validator.New().Struct(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid request body")
	}

	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid number format")
	}

	sendText := &whatsmiau.SendAudioRequest{
		AudioURL:   request.Audio,
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
		ViewOnce:   request.ViewOnce,
	}

	if request.Quoted != nil && len(request.Quoted.Key.Id) > 0 && len(request.Quoted.Message.Conversation) > 0 {
		sendText.QuoteMessage = request.Quoted.Message.Conversation
		sendText.QuoteMessageID = request.Quoted.Key.Id
	}

	c := ctx.Request().Context()
	if err := s.whatsmiau.ChatPresence(&whatsmiau.ChatPresenceRequest{
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
		Presence:   types.ChatPresenceComposing,
		Media:      types.ChatPresenceMediaAudio,
	}); err != nil {
		zap.L().Error("Whatsmiau.ChatPresence", zap.Error(err))
	} else {
		time.Sleep(time.Millisecond * time.Duration(request.Delay)) // TODO: create a more robust solution
	}

	res, err := s.whatsmiau.SendAudio(c, sendText)
	if err != nil {
		zap.L().Error("Whatsmiau.SendAudioRequest failed", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to send audio")
	}

	return ctx.JSON(http.StatusOK, dto.SendAudioResponse{
		Key: dto.MessageResponseKey{
			RemoteJid: request.Number,
			FromMe:    true,
			Id:        res.ID,
		},

		Status:           "sent",
		MessageType:      "audioMessage",
		MessageTimestamp: int(res.CreatedAt.Unix() / 1000),
		InstanceId:       request.InstanceID,
	})
}

// For evolution compatibility
func (s *Message) SendMedia(ctx echo.Context) error {
	var request dto.SendMediaRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	if err := validator.New().Struct(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid request body")
	}
	switch request.Mediatype {
	case "image":
		request.SendDocumentRequest.Mimetype = "image/png"
		return s.sendImage(ctx, request.SendDocumentRequest)
	case "video":
		return s.sendVideo(ctx, request.SendDocumentRequest)
	}

	return s.sendDocument(ctx, request.SendDocumentRequest)
}

func (s *Message) SendDocument(ctx echo.Context) error {
	var request dto.SendDocumentRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	if err := validator.New().Struct(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid request body")
	}

	return s.sendDocument(ctx, request)
}

func (s *Message) sendDocument(ctx echo.Context, request dto.SendDocumentRequest) error {
	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid number format")
	}

	sendData := &whatsmiau.SendDocumentRequest{
		InstanceID: request.InstanceID,
		MediaURL:   request.Media,
		Caption:    request.Caption,
		FileName:   request.FileName,
		RemoteJID:  jid,
		Mimetype:   request.Mimetype,
	}

	c := ctx.Request().Context()
	time.Sleep(time.Millisecond * time.Duration(request.Delay)) // TODO: create a more robust solution

	res, err := s.whatsmiau.SendDocument(c, sendData)
	if err != nil {
		zap.L().Error("Whatsmiau.SendDocument failed", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to send document")
	}

	return ctx.JSON(http.StatusOK, dto.SendDocumentResponse{
		Key: dto.MessageResponseKey{
			RemoteJid: request.Number,
			FromMe:    true,
			Id:        res.ID,
		},
		Status:           "sent",
		MessageType:      "documentMessage",
		MessageTimestamp: int(res.CreatedAt.Unix() / 1000),
		InstanceId:       request.InstanceID,
	})
}

func (s *Message) SendImage(ctx echo.Context) error {
	var request dto.SendDocumentRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	if err := validator.New().Struct(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid request body")
	}

	return s.sendImage(ctx, request)
}

func (s *Message) sendImage(ctx echo.Context, request dto.SendDocumentRequest) error {
	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid number format")
	}

	sendData := &whatsmiau.SendImageRequest{
		InstanceID: request.InstanceID,
		MediaURL:   request.Media,
		Caption:    request.Caption,
		RemoteJID:  jid,
		Mimetype:   request.Mimetype,
		ViewOnce:   request.ViewOnce,
	}

	c := ctx.Request().Context()
	time.Sleep(time.Millisecond * time.Duration(request.Delay)) // TODO: create a more robust solution

	res, err := s.whatsmiau.SendImage(c, sendData)
	if err != nil {
		zap.L().Error("Whatsmiau.SendDocument failed", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to send document")
	}

	return ctx.JSON(http.StatusOK, dto.SendDocumentResponse{
		Key: dto.MessageResponseKey{
			RemoteJid: request.Number,
			FromMe:    true,
			Id:        res.ID,
		},
		Status:           "sent",
		MessageType:      "imageMessage",
		MessageTimestamp: int(res.CreatedAt.Unix() / 1000),
		InstanceId:       request.InstanceID,
	})
}

func (s *Message) SendReaction(ctx echo.Context) error {
	var request dto.SendReactionRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	if err := validator.New().Struct(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid request body")
	}

	jid, err := numberToJid(request.Key.RemoteJid)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid number format")
	}

	var emojiRegex = regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`)
	if !emojiRegex.MatchString(request.Reaction) {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid reaction, must be a emoji")
	}

	sendReaction := &whatsmiau.SendReactionRequest{
		InstanceID: request.InstanceID,
		Reaction:   request.Reaction,
		RemoteJID:  jid,
		MessageID:  request.Key.Id,
		FromMe:     request.Key.FromMe,
	}

	c := ctx.Request().Context()
	res, err := s.whatsmiau.SendReaction(c, sendReaction)
	if err != nil {
		zap.L().Error("Whatsmiau.SendReaction failed", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to send reaction")
	}

	return ctx.JSON(http.StatusOK, dto.SendReactionResponse{
		Key: dto.MessageResponseKey{
			RemoteJid: request.Key.RemoteJid,
			FromMe:    true,
			Id:        res.ID,
		},
		Status:           "sent",
		MessageType:      "reactionMessage",
		MessageTimestamp: int(res.CreatedAt.UnixMicro() / 1000),
		InstanceId:       request.InstanceID,
	})
}

func (s *Message) SendVideo(ctx echo.Context) error {
	var request dto.SendVideoRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	if err := validator.New().Struct(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid request body")
	}

	return s.sendVideo(ctx, request)
}

func (s *Message) sendVideo(ctx echo.Context, request interface{}) error {
	var (
		jid        *types.JID
		instanceID string
		videoURL   string
		caption    string
		mimetype   string
		number     string
		delay      int
		viewOnce   bool
		err        error
	)

	// Support both SendVideoRequest and SendDocumentRequest (for SendMedia compatibility)
	switch req := request.(type) {
	case dto.SendVideoRequest:
		number = req.Number
		instanceID = req.InstanceID
		videoURL = req.Video
		caption = req.Caption
		mimetype = req.Mimetype
		delay = req.Delay
		viewOnce = req.ViewOnce
	case dto.SendDocumentRequest:
		number = req.Number
		instanceID = req.InstanceID
		videoURL = req.Media
		caption = req.Caption
		mimetype = req.Mimetype
		delay = req.Delay
		viewOnce = req.ViewOnce
	default:
		return utils.HTTPFail(ctx, http.StatusBadRequest, nil, "invalid request type")
	}

	jid, err = numberToJid(number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid number format")
	}

	sendData := &whatsmiau.SendVideoRequest{
		InstanceID: instanceID,
		MediaURL:   videoURL,
		Caption:    caption,
		RemoteJID:  jid,
		Mimetype:   mimetype,
		ViewOnce:   viewOnce,
	}

	c := ctx.Request().Context()
	time.Sleep(time.Millisecond * time.Duration(delay)) // TODO: create a more robust solution

	res, err := s.whatsmiau.SendVideo(c, sendData)
	if err != nil {
		zap.L().Error("Whatsmiau.SendVideo failed", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to send video")
	}

	return ctx.JSON(http.StatusOK, dto.SendVideoResponse{
		Key: dto.MessageResponseKey{
			RemoteJid: number,
			FromMe:    true,
			Id:        res.ID,
		},
		Status:           "sent",
		MessageType:      "videoMessage",
		MessageTimestamp: int(res.CreatedAt.Unix() / 1000),
		InstanceId:       instanceID,
	})
}

// SendMissedCall simula uma notificação de chamada perdida
// AVISO: Este endpoint é EXPERIMENTAL e pode violar os Termos de Serviço do WhatsApp
// Use apenas em ambientes de desenvolvimento/teste e por sua conta e risco
func (s *Message) SendMissedCall(ctx echo.Context) error {
	var request dto.SendMissedCallRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	if err := validator.New().Struct(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid request body")
	}

	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusBadRequest, err, "invalid number format")
	}

	missedCallData := &whatsmiau.SendMissedCallRequest{
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
		VideoCall:  request.VideoCall,
	}

	res, err := s.whatsmiau.SendMissedCall(ctx.Request().Context(), missedCallData)
	if err != nil {
		zap.L().Error("Whatsmiau.SendMissedCall failed", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to send missed call notification")
	}

	return ctx.JSON(http.StatusOK, dto.SendMissedCallResponse{
		Key: dto.MessageResponseKey{
			RemoteJid: request.Number,
			FromMe:    true,
			Id:        res.ID,
		},
		Status:           "sent",
		MessageType:      "missedCallNotification",
		MessageTimestamp: int(res.CreatedAt.Unix() / 1000),
		InstanceId:       request.InstanceID,
	})
}
