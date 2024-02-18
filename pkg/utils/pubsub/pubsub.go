// MIT License
//
// Copyright (c) 2024 Alain Gilbert
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package pubsub

import (
	"context"
	"sync"
	"time"
)

// Topic type
type Topic comparable

// PubSub contains and manage the map of topics -> subscribers
type PubSub[T Topic, V any] struct {
	sync.RWMutex
	m        map[T]map[*Sub[T, V]]struct{}
	buffered int
}

// Config for pubsub
type Config struct {
	Buffered int
}

// NewPubSub creates a new PubSub
func NewPubSub[V any](cfg *Config) *PubSub[string, V] {
	return NewPubSubTopic[string, V](cfg)
}

// NewPubSubTopic creates a new PubSub with customizable Topic type
func NewPubSubTopic[T Topic, V any](cfg *Config) *PubSub[T, V] {
	ps := PubSub[T, V]{}
	ps.m = make(map[T]map[*Sub[T, V]]struct{})
	if cfg != nil {
		ps.buffered = cfg.Buffered
	} else {
		ps.buffered = 10
	}
	return &ps
}

// Subscribe to one or more topics
func (p *PubSub[T, V]) Subscribe(topics ...T) *Sub[T, V] {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan Payload[T, V], p.buffered)
	s := &Sub[T, V]{topics: topics, ch: ch, ctx: ctx, cancel: cancel, p: p}
	p.addSubscriber(s)
	return s
}

// Publish a message on a given topic
func (p *PubSub[T, V]) Publish(topic T, msg V) { p.publish(topic, msg) }

// Pub shortcut for Publish
func (p *PubSub[T, V]) Pub(topic T, msg V) { p.Publish(topic, msg) }

func (p *PubSub[T, V]) publish(topic T, msg V) {
	p.RLock()
	defer p.RUnlock()
	for s := range p.m[topic] {
		s.publish(Payload[T, V]{topic, msg})
	}
}

func (p *PubSub[T, V]) addSubscriber(s *Sub[T, V]) {
	p.Lock()
	defer p.Unlock()
	for _, topic := range s.topics {
		if p.m[topic] == nil {
			p.m[topic] = make(map[*Sub[T, V]]struct{})
		}
		p.m[topic][s] = struct{}{}
	}
}

func (p *PubSub[T, V]) removeSubscriber(s *Sub[T, V]) {
	p.Lock()
	defer p.Unlock()
	for _, topic := range s.topics {
		delete(p.m[topic], s)
	}
}

// Payload topic and message
type Payload[T Topic, V any] struct {
	Topic T
	Msg   V
}

// Sub subscriber will receive messages published on a Topic in his ch
type Sub[T Topic, V any] struct {
	topics []T                // Topics subscribed to
	ch     chan Payload[T, V] // Receives messages in this channel
	ctx    context.Context
	cancel context.CancelFunc
	p      *PubSub[T, V]
}

func (s *Sub[T, V]) ReceiveContext(ctx context.Context) (topic T, msg V, err error) {
	select {
	case p := <-s.ch:
		return p.Topic, p.Msg, nil
	case <-ctx.Done():
		return topic, msg, ctx.Err()
	case <-s.ctx.Done():
		return topic, msg, s.ctx.Err()
	}
}

// ReceiveTimeout returns a message received on the channel or timeout
func (s *Sub[T, V]) ReceiveTimeout(timeout time.Duration) (topic T, msg V, err error) {
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()
	return s.ReceiveContext(ctx)
}

// Receive returns a message
func (s *Sub[T, V]) Receive() (topic T, msg V, err error) {
	return s.ReceiveContext(s.ctx)
}

// ReceiveCh returns the receive channel
func (s *Sub[T, V]) ReceiveCh() <-chan Payload[T, V] {
	return s.ch
}

// Close will remove the subscriber from the Topic subscribers
func (s *Sub[T, V]) Close() {
	s.cancel()
	s.p.removeSubscriber(s)
}

// publish a message to the subscriber channel
func (s *Sub[T, V]) publish(p Payload[T, V]) {
	select {
	case s.ch <- p:
	default:
	}
}
