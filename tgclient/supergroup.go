package tgclient 

import (
    "context"
    "fmt"

    "github.com/gotd/td/telegram"
    "github.com/gotd/td/tg"
)

func CreateSupergroup(ctx context.Context, client *telegram.Client) error {
    api := tg.NewClient(client)

    updates, err := api.ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
        Megagroup: true,
		Forum: true,
        Title: "TDrive Storage",
        About: "Private storage group",
    })
    if err != nil {
        return fmt.Errorf("create supergroup: %w", err)
    }

    _ = updates

    return nil
}
