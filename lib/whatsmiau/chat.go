package whatsmiau

import (
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
	"golang.org/x/net/context"
)

type ReadMessageRequest struct {
	MessageIDs []string   `json:"message_ids"`
	InstanceID string     `json:"instance_id"`
	RemoteJID  *types.JID `json:"remote_jid"`
	Sender     *types.JID `json:"sender"`
}

func (s *Whatsmiau) ReadMessage(data *ReadMessageRequest) error {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return whatsmeow.ErrClientIsNil
	}

	sender := *data.RemoteJID
	if data.Sender != nil {
		sender = *data.Sender
	}

	return client.MarkRead(context.TODO(), data.MessageIDs, time.Now(), *data.RemoteJID, sender)
}

type ChatPresenceRequest struct {
	InstanceID string                  `json:"instance_id"`
	RemoteJID  *types.JID              `json:"remote_jid"`
	Presence   types.ChatPresence      `json:"presence"`
	Media      types.ChatPresenceMedia `json:"media"`
}

func (s *Whatsmiau) ChatPresence(data *ChatPresenceRequest) error {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return whatsmeow.ErrClientIsNil
	}

	return client.SendChatPresence(context.TODO(), *data.RemoteJID, data.Presence, data.Media)
}

type NumberExistsRequest struct {
	InstanceID string   `json:"instance_id"`
	Numbers    []string `json:"numbers"`
}

type NumberExistsResponse []Exists

type Exists struct {
	Exists bool   `json:"exists"`
	Jid    string `json:"jid"`
	Lid    string `json:"lid"`
	Number string `json:"number"`
}

func (s *Whatsmiau) NumberExists(ctx context.Context, data *NumberExistsRequest) (NumberExistsResponse, error) {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return nil, whatsmeow.ErrClientIsNil
	}

	resp, err := client.IsOnWhatsApp(context.TODO(), data.Numbers)
	if err != nil {
		return nil, err
	}

	var results []Exists
	for _, item := range resp {
		jid, lid := s.GetJidLid(ctx, data.InstanceID, item.JID)

		results = append(results, Exists{
			Exists: item.IsIn,
			Jid:    jid,
			Lid:    lid,
			Number: item.Query,
		})
	}

	return results, nil
}

// DeleteChatRequest define os parâmetros necessários para deletar um chat
type DeleteChatRequest struct {
	InstanceID string     `json:"instance_id"`
	RemoteJID  *types.JID `json:"remote_jid"`
}

// DeleteChat deleta um chat (remove completamente da lista de conversas)
func (s *Whatsmiau) DeleteChat(ctx context.Context, data *DeleteChatRequest) error {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return whatsmeow.ErrClientIsNil
	}

	// Usa app state para deletar o chat
	patch := appstate.BuildDeleteChat(*data.RemoteJID, time.Now(), nil)
	return client.SendAppState(ctx, patch)
}

// ArchiveChatRequest define os parâmetros necessários para arquivar/desarquivar um chat
type ArchiveChatRequest struct {
	InstanceID string     `json:"instance_id"`
	RemoteJID  *types.JID `json:"remote_jid"`
	Archive    bool       `json:"archive"` // true = arquivar, false = desarquivar
}

// ArchiveChat arquiva ou desarquiva um chat
func (s *Whatsmiau) ArchiveChat(ctx context.Context, data *ArchiveChatRequest) error {
	client, ok := s.clients.Load(data.InstanceID)
	if !ok {
		return whatsmeow.ErrClientIsNil
	}

	// Usa app state para arquivar/desarquivar
	patch := appstate.BuildArchive(*data.RemoteJID, data.Archive, time.Now(), nil)
	return client.SendAppState(ctx, patch)
}
