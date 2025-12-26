package rabbitmq

const (
	CommandExchange    = "tmf.customer"
	EventExchange      = "tmf.events"
	DeadLetterExchange = "tmf.dlx"
	EventQueue         = "customer.events"
	CustomerQueue      = "customer.commands"
	DeadLetterQueue    = "customer.dlq"

	// Customer Events
	EvtCustomerCreated     = "evt.customer.created"
	EvtCustomerUpdated     = "evt.customer.updated"
	EvtCustomerStateChange = "evt.customer.stateChange"
	EvtCustomerDeleted     = "evt.customer.deleted"

	// Subscribed Events
	EvtPartyUpdated = "evt.party.updated"
	EvtPartyDeleted = "evt.party.deleted"
	
	// Saga events/commands
	EvtPartyDeletionInitiated = "evt.party.deletion_initiated"
	CmdPartyFinalizeDeletion  = "cmd.party.finalize_deletion"
	CmdPartyCancelDeletion    = "cmd.party.cancel_deletion"
)
