package p2p

import (
	"context"
	"encoding/json"

	"github.com/libp2p/go-libp2p/core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const ChatRoomBufSize = 128

// Room represents a subscription to a single PubSub topic. Messages
// can be published to the topic with Room.Publish, and received
// messages are pushed to the Messages channel.
type Room struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages chan *ChatMessage

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string
}

// ChatMessage gets converted to/from JSON and sent in the body of pubsub messages.
type ChatMessage struct {
	Message    []byte
	SenderID   string
	SenderNick string
}

// JoinRoom tries to subscribe to the PubSub topic for the room name, returning
// a ChatRoom on success.
func JoinRoom(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, roomName string) (*Room, error) {
	// join the pubsub topic
	topic, err := ps.Join(topicName(roomName))
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	r := &Room{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		nick:     nickname,
		roomName: roomName,
		Messages: make(chan *ChatMessage, ChatRoomBufSize),
	}

	// start reading messages from the subscription in a loop
	go r.readLoop()
	return r, nil
}

// Publish sends a message to the pubsub topic.
func (r *Room) Publish(message []byte) error {
	m := ChatMessage{
		Message:    message,
		SenderID:   r.self.Pretty(),
		SenderNick: r.nick,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return r.topic.Publish(r.ctx, msgBytes)
}

func (r *Room) ListPeers() []peer.ID {
	return r.ps.ListPeers(topicName(r.roomName))
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (r *Room) readLoop() {
	for {
		msg, err := r.sub.Next(r.ctx)
		if err != nil {
			close(r.Messages)
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == r.self {
			continue
		}
		cm := new(ChatMessage)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel
		r.Messages <- cm
	}
}

func topicName(roomName string) string {
	return "chat-room:" + roomName
}

func (r *Room) handleEvents() {

	for {
		select {
		case input := <-data2Send:
			// when the user types in a line, publish it to the chat room and print to the message window
			err := r.Publish(input)
			if err != nil {
				printErr("publish error: %s", err)
			}
		case m := <-r.Messages:
			// when we receive a message from the chat room, print it to the message window
			Data2Handle <- m.Message

		case <-r.ctx.Done():
			return

			// case <-ui.doneCh:
			// 	return
		}
	}
}
