package tgclient 

import (
    "context"
    "fmt"
	"strings"

	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/telegram/query/messages"
    "github.com/gotd/td/tg"
)

type GroupRef struct {
    Title      string
    ID         int64
    AccessHash int64
    Peer       tg.InputPeerClass
    Channel    tg.InputChannelClass
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

func GetGeneralMessageCount(
    ctx context.Context,
    api *tg.Client,
    peer tg.InputPeerClass,
) (int, error) {
    count, err := messages.NewQueryBuilder(api).
        GetHistory(peer).
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
