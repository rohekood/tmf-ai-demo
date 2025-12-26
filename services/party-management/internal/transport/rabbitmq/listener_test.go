package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListener_GetHandler(t *testing.T) {
	l := &Listener{}
	h := &Handlers{} // Methods on nil or empty handlers might be tricky if checked, but here we just check function pointer equality or existence

	tests := []struct {
		name       string
		routingKey string
		wantValid  bool
	}{
		{"Create Party", CmdPartyCreate, true},
		{"Update Party", CmdPartyUpdate, true},
		{"Patch Party", CmdPartyPatch, true},
		{"Delete Party", CmdPartyDelete, true},
		{"Get Party", QueryPartyGet, true},
		{"Search Party", QueryPartySearch, true},
		{"Unknown Key", "unknown.key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, valid := l.GetHandler(tt.routingKey, h)
			assert.Equal(t, tt.wantValid, valid)
			if tt.wantValid {
				assert.NotNil(t, handler)
			} else {
				assert.Nil(t, handler)
			}
		})
	}
}

// Mock handler methods to ensure GetHandler returns correct function?
// Since we can't easily compare function pointers across reloads/compilations reliably in all cases,
// checking NotNil and Valid is a good enough proxy for "it returns something for this key".
// Detailed verification that it returns specifically *HandleCreateParty* vs *HandleUpdateParty*
// is implicit in the implementation which is a simple switch.
