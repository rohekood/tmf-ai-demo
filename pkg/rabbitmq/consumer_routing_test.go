package rabbitmq

import (
	"context"
	"testing"
)

func TestTopicMatches(t *testing.T) {
	tests := []struct {
		name       string
		bindingKey string
		routingKey string
		want       bool
	}{
		{name: "exact match", bindingKey: "cmd.cart.item.add", routingKey: "cmd.cart.item.add", want: true},
		{name: "single wildcard", bindingKey: "cmd.cart.*.add", routingKey: "cmd.cart.item.add", want: true},
		{name: "single wildcard mismatch", bindingKey: "cmd.cart.*.add", routingKey: "cmd.cart.item.fast.add", want: false},
		{name: "hash wildcard", bindingKey: "evt.catalog.#", routingKey: "evt.catalog.offering.updated", want: true},
		{name: "hash wildcard empty tail", bindingKey: "evt.catalog.#", routingKey: "evt.catalog", want: true},
		{name: "mismatch", bindingKey: "query.cart.get", routingKey: "query.cart.session.get", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := topicMatches(tc.bindingKey, tc.routingKey)
			if got != tc.want {
				t.Fatalf("topicMatches(%q, %q) = %v, want %v", tc.bindingKey, tc.routingKey, got, tc.want)
			}
		})
	}
}

func TestHandlerForRoutingKey_PrefersExactThenWildcard(t *testing.T) {
	exactCalled := false
	wildcardCalled := false

	exactHandler := func(context.Context, []byte) error {
		exactCalled = true
		return nil
	}
	wildcardHandler := func(context.Context, []byte) error {
		wildcardCalled = true
		return nil
	}

	consumer := &rabbitConsumer{}
	consumer.handlers = []subscription{
		{topic: "query.cart.#", handler: wildcardHandler},
		{topic: "query.cart.get", handler: exactHandler},
	}

	h, ok := consumer.handlerForRoutingKey("query.cart.get")
	if !ok {
		t.Fatal("expected handler for query.cart.get")
	}
	if err := h(context.Background(), nil); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if !exactCalled {
		t.Fatal("expected exact handler to be called")
	}
	if wildcardCalled {
		t.Fatal("did not expect wildcard handler to be called")
	}

	exactCalled = false
	wildcardCalled = false

	h, ok = consumer.handlerForRoutingKey("query.cart.session.get")
	if !ok {
		t.Fatal("expected wildcard handler for query.cart.session.get")
	}
	if err := h(context.Background(), nil); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if exactCalled {
		t.Fatal("did not expect exact handler to be called")
	}
	if !wildcardCalled {
		t.Fatal("expected wildcard handler to be called")
	}
}
