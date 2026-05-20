package tgclient

import (
    "fmt"

    "github.com/gotd/td/bin"
    "github.com/gotd/td/tg"
)

const channelsCreateTopicTypeID = 0xf40c0224
const channelsDeleteTopicHistoryTypeID = 0x34435f2d

type channelsCreateTopicRequest struct {
    Flags       bin.Fields
    Channel     tg.InputChannelClass
    Title       string
    IconColor   int
    IconEmojiID int64
    RandomID    int64
    SendAs      tg.InputPeerClass
}

type channelsDeleteTopicHistoryRequest struct {
    Channel  tg.InputChannelClass
    TopMsgID int
}

func (r *channelsCreateTopicRequest) TypeID() uint32 {
    return channelsCreateTopicTypeID
}

func (r *channelsCreateTopicRequest) TypeName() string {
    return "channels.createTopic"
}

func (r *channelsCreateTopicRequest) SetFlags() {
    if r.IconColor != 0 {
        r.Flags.Set(0)
    }

    if r.SendAs != nil {
        r.Flags.Set(2)
    }

    if r.IconEmojiID != 0 {
        r.Flags.Set(3)
    }
}

func (r *channelsCreateTopicRequest) Encode(b *bin.Buffer) error {
    if r == nil {
        return fmt.Errorf("can't encode channels.createTopic#f40c0224 as nil")
    }

    b.PutID(channelsCreateTopicTypeID)

    return r.EncodeBare(b)
}

func (r *channelsCreateTopicRequest) EncodeBare(b *bin.Buffer) error {
    if r == nil {
        return fmt.Errorf("can't encode channels.createTopic#f40c0224 as nil")
    }

    r.SetFlags()

    if err := r.Flags.Encode(b); err != nil {
        return fmt.Errorf("encode channels.createTopic flags: %w", err)
    }

    if r.Channel == nil {
        return fmt.Errorf("encode channels.createTopic: channel is nil")
    }

    if err := r.Channel.Encode(b); err != nil {
        return fmt.Errorf("encode channels.createTopic channel: %w", err)
    }

    b.PutString(r.Title)

    if r.Flags.Has(0) {
        b.PutInt(r.IconColor)
    }

    if r.Flags.Has(3) {
        b.PutLong(r.IconEmojiID)
    }

    b.PutLong(r.RandomID)

    if r.Flags.Has(2) {
        if r.SendAs == nil {
            return fmt.Errorf("encode channels.createTopic: send_as is nil")
        }

        if err := r.SendAs.Encode(b); err != nil {
            return fmt.Errorf("encode channels.createTopic send_as: %w", err)
        }
    }

    return nil
}

func getCreatedTopicID(updates tg.UpdatesClass) (int, bool) {
    switch u := updates.(type) {
    case *tg.Updates:
        return getCreatedTopicIDFromUpdateList(u.Updates)

    case *tg.UpdatesCombined:
        return getCreatedTopicIDFromUpdateList(u.Updates)

    case *tg.UpdateShort:
        return getCreatedTopicIDFromUpdate(u.Update)

    case *tg.UpdateShortSentMessage:
        return u.ID, true

    default:
        return 0, false
    }
}

func getCreatedTopicIDFromUpdateList(updates []tg.UpdateClass) (int, bool) {
    for _, update := range updates {
        topicID, ok := getCreatedTopicIDFromUpdate(update)
        if ok {
            return topicID, true
        }
    }

    return 0, false
}

func getCreatedTopicIDFromUpdate(update tg.UpdateClass) (int, bool) {
    switch u := update.(type) {
    case *tg.UpdateNewChannelMessage:
        return getCreatedTopicIDFromMessage(u.Message)

    case *tg.UpdateNewMessage:
        return getCreatedTopicIDFromMessage(u.Message)

    default:
        return 0, false
    }
}

func getCreatedTopicIDFromMessage(message tg.MessageClass) (int, bool) {
    msg, ok := message.(*tg.MessageService)
    if !ok {
        return 0, false
    }

    _, ok = msg.Action.(*tg.MessageActionTopicCreate)
    if !ok {
        return 0, false
    }

    return msg.ID, true
}

func (r *channelsDeleteTopicHistoryRequest) TypeID() uint32 {
    return channelsDeleteTopicHistoryTypeID
}

func (r *channelsDeleteTopicHistoryRequest) TypeName() string {
    return "channels.deleteTopicHistory"
}

func (r *channelsDeleteTopicHistoryRequest) Encode(b *bin.Buffer) error {
    if r == nil {
        return fmt.Errorf("can't encode channels.deleteTopicHistory#34435f2d as nil")
    }

    b.PutID(channelsDeleteTopicHistoryTypeID)

    return r.EncodeBare(b)
}

func (r *channelsDeleteTopicHistoryRequest) EncodeBare(b *bin.Buffer) error {
    if r == nil {
        return fmt.Errorf("can't encode channels.deleteTopicHistory#34435f2d as nil")
    }

    if r.Channel == nil {
        return fmt.Errorf("encode channels.deleteTopicHistory: channel is nil")
    }

    if err := r.Channel.Encode(b); err != nil {
        return fmt.Errorf("encode channels.deleteTopicHistory channel: %w", err)
    }

    b.PutInt(r.TopMsgID)

    return nil
}
