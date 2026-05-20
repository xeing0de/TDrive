package tgclient 

import (
    "context"
    "fmt"
    "strings"
    "time"
	"path/filepath"

    "github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram"
    "github.com/gotd/td/telegram/query/dialogs"
    "github.com/gotd/td/telegram/query/messages"
    "github.com/gotd/td/tg"
	"github.com/gotd/td/telegram/uploader"
)

type GroupRef struct {
    Title      string
    ID         int64
    AccessHash int64
    Peer       tg.InputPeerClass
    Channel    tg.InputChannelClass
}

func randomID() int64 {
    return time.Now().UnixNano()
}

func CreateSupergroup(ctx context.Context, api *tg.Client, title string) (*GroupRef, error) {
	title = "[Disk]" + title
    updates, err := api.ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
        Megagroup: true,
		Forum: true,
        Title: title,
        About: "Private storage group",
    })
    if err != nil {
        return nil, fmt.Errorf("create supergroup: %w", err)
    }

	var chats []tg.ChatClass

    switch u := updates.(type) {
    case *tg.Updates:
        chats = u.Chats
    case *tg.UpdatesCombined:
        chats = u.Chats
    default:
        return nil, fmt.Errorf("unexpected updates type: %T", updates)
    }

    for _, chat := range chats {
        ch, ok := chat.(*tg.Channel)
        if !ok {
            continue
        }

        if ch.Title != title {
            continue
        }

        return &GroupRef{
            Title:      ch.Title,
            ID:         ch.ID,
            AccessHash: ch.AccessHash,
            Peer: &tg.InputPeerChannel{
            	ChannelID:  ch.ID,
            	AccessHash: ch.AccessHash,
            },
            Channel: &tg.InputChannel{
                ChannelID:  ch.ID,
                AccessHash: ch.AccessHash,
            },
        }, nil
    }

    return nil, fmt.Errorf("created group not found in updates")
}

func GetTopicMessageCount(
    ctx context.Context,
    api *tg.Client,
    peer tg.InputPeerClass,
    topicID int,
) (int, error) {
    count, err := messages.NewQueryBuilder(api).
        GetReplies(peer).
        MsgID(topicID).
        Count(ctx)
    if err != nil {
        return 0, err
    }

    return count, nil
}

func FindDiskGroups(ctx context.Context, api *tg.Client) ([]GroupRef, error) {
	const prefix = "[Disk]"

	groups := make([]GroupRef, 0)

	err := query.GetDialogs(api).
		BatchSize(50).
		ForEach(ctx, func(ctx context.Context, dlg dialogs.Elem) error {
			if dlg.Deleted() {
				return nil
			}

			peer, ok := dlg.Peer.(*tg.InputPeerChannel)
			if !ok {
				return nil
			}

			ch, ok := dlg.Entities.Channel(peer.ChannelID)
			if !ok {
				return nil
			}

			if !ch.Megagroup {
				return nil
			}

			if !strings.HasPrefix(ch.Title, prefix) {
				return nil
			}

			groups = append(groups, GroupRef{
				Title:      ch.Title,
				ID:         ch.ID,
				AccessHash: ch.AccessHash,
				Peer: &tg.InputPeerChannel{
					ChannelID:  ch.ID,
					AccessHash: ch.AccessHash,
				},
				Channel: &tg.InputChannel{
					ChannelID:  ch.ID,
					AccessHash: ch.AccessHash,
				},
			})

			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("get dialogs: %w", err)
	}

	return groups, nil
}

func CreateTopic(
    ctx context.Context,
    client *telegram.Client,
    channel tg.InputChannelClass,
    title string,
) (int, error) {
    var result tg.UpdatesBox

    req := &channelsCreateTopicRequest{
        Channel:  channel,
        Title:    title,
        RandomID: randomID(),
    }

    if err := client.Invoke(ctx, req, &result); err != nil {
        return 0, fmt.Errorf("create forum topic: %w", err)
    }

	topicID, ok := getCreatedTopicID(result.Updates)
    if !ok {
        return 0, fmt.Errorf("create topic: topic id not found in updates: %T", result.Updates)
    }

    return topicID, nil
}

func SendMessageToTopic(
    ctx context.Context,
    api *tg.Client,
    peer tg.InputPeerClass,
    topicID int,
    text string,
) error {
	_, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
        Peer: peer,
        ReplyTo: &tg.InputReplyToMessage{
            ReplyToMsgID: topicID,
        },
        Message:  text,
        RandomID: randomID(),
    })
    if err != nil {
        return fmt.Errorf("send message to topic: %w", err)
    }

    return nil
}

func DeleteTopic(
    ctx context.Context,
    client *telegram.Client,
    channel tg.InputChannelClass,
    topicID int,
) error {
    var result tg.MessagesAffectedHistory

    req := &channelsDeleteTopicHistoryRequest{
        Channel:  channel,
        TopMsgID: topicID,
    }

    if err := client.Invoke(ctx, req, &result); err != nil {
        return fmt.Errorf("delete forum topic: %w", err)
    }

    return nil
}

func SendFileToTopic(
    ctx context.Context,
    api *tg.Client,
    peer tg.InputPeerClass,
    filePath string,
    topicID int,
    caption string,
) error {
    upload, err := uploader.NewUploader(api).
    	WithPartSize(512 * 1024).
    	FromPath(ctx, filePath)    
	if err != nil {
        return fmt.Errorf("upload file: %w", err)
    }

    fileName := filepath.Base(filePath)

    _, err = api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
        Peer: peer,
        Media: &tg.InputMediaUploadedDocument{
            File:     upload,
            MimeType: "application/octet-stream",
            Attributes: []tg.DocumentAttributeClass{
                &tg.DocumentAttributeFilename{
                    FileName: fileName,
                },
            },
        },
        Message:  caption,
        RandomID: randomID(),
    })
    if err != nil {
        return fmt.Errorf("send media: %w", err)
    }

    return nil
}
