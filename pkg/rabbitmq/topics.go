package rabbitmq

// Query Topics
const (
	QueryInventoryResourceCapacity  = "query.inventory.resource.capacity"
	QueryGISGeographyCheck          = "query.gis.geography.check"
	EvtCartSessionUpdated           = "evt.cart.session.updated"
	EvtCartSessionPriced            = "evt.cart.session.priced"
	CmdCartPriceUpdate              = "cmd.cart.price.update"
	CmdInventoryResourceReserve     = "cmd.inventory.resource.reserve"
	CmdPaymentTransactionAuthorize  = "cmd.payment.transaction.authorize"
	CmdOrderManagementCreate        = "cmd.order.management.create"
	CmdInventoryResourceRelease     = "cmd.inventory.resource.release"
	CmdCartItemAdd                  = "cmd.cart.item.add"
	CmdQualEligibilityCheck         = "cmd.qual.eligibility.check"
	CmdOrderCheckoutSubmit          = "cmd.order.checkout.submit"
	EvtPaymentTransactionAuthorized = "evt.payment.transaction.authorized"
	EvtPaymentTransactionDeclined   = "evt.payment.transaction.declined"
	EvtInventoryResourceReserved    = "evt.inventory.resource.reserved"
	EvtInventoryResourceFailed      = "evt.inventory.resource.failed"
	EvtOrderManagementCreated       = "evt.order.management.created"
)
