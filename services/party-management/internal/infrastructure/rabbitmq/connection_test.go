package rabbitmq

import (
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestConnectionManager_Connect_DialError(t *testing.T) {
	dialFn := func(url string) (*amqp.Connection, error) {
		return nil, errors.New("mock dial error")
	}

	mgr := NewConnectionManager("amqp://guest:guest@localhost:5672/", dialFn)
	err := mgr.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock dial error")
}

func TestConnectionManager_Close(t *testing.T) {
	dialFn := func(url string) (*amqp.Connection, error) {
		return nil, errors.New("mock dial error")
	}

	mgr := NewConnectionManager("amqp://guest:guest@localhost:5672/", dialFn)
	err := mgr.Close()
	assert.NoError(t, err)
}

func TestConnectionManager_Getters(t *testing.T) {
	mgr := NewConnectionManager("amqp://guest:guest@localhost:5672/")
	assert.Nil(t, mgr.GetConnection())
	assert.Nil(t, mgr.GetChannel())

	notify := mgr.ReconnectNotify()
	assert.NotNil(t, notify)
}

func TestConnectionManager_HandleReconnect_ContextCancelled(t *testing.T) {
	dialFn := func(url string) (*amqp.Connection, error) {
		return nil, errors.New("mock dial error")
	}

	mgr := NewConnectionManager("amqp://guest:guest@localhost:5672/", dialFn)

	mgr.cancel()

	mgr.conn = nil

	_ = mgr.Close()
}
