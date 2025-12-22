package rabbitmq

const (
	CommandExchange = "tmf.commands"
	EventExchange   = "tmf.events"
	EventQueue      = "customer.events"
	CustomerQueue   = "customer.commands"

	// Customer Events
	EvtCustomerCreated     = "evt.customer.created"
	EvtCustomerUpdated     = "evt.customer.updated"
	EvtCustomerStateChange = "evt.customer.stateChange"
	EvtCustomerDeleted     = "evt.customer.deleted"

	// Subscribed Events
	EvtPartyUpdated = "evt.party.updated"
	EvtPartyDeleted = "evt.party.deleted"
)
