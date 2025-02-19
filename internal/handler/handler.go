package handler

// Handler holds all HTTP handlers
type Handler struct {
	userHandler    *UserHandler
	productHandler *ProductHandler
	orderHandler   *OrderHandler
	messageHandler *MessageHandler
}

// New creates a new Handler
func New(
	userRepo UserRepository,
	productRepo ProductRepository,
	orderRepo OrderRepository,
	producer MessageProducer,
	cache CacheRepository,
) *Handler {
	return &Handler{
		userHandler:    NewUserHandler(userRepo, cache, producer),
		productHandler: NewProductHandler(productRepo),
		orderHandler:   NewOrderHandler(orderRepo),
		messageHandler: NewMessageHandler(producer),
	}
}

// GetUserHandler returns the user handler
func (h *Handler) GetUserHandler() *UserHandler {
	return h.userHandler
}

// GetProductHandler returns the product handler
func (h *Handler) GetProductHandler() *ProductHandler {
	return h.productHandler
}

// GetOrderHandler returns the order handler
func (h *Handler) GetOrderHandler() *OrderHandler {
	return h.orderHandler
}

// GetMessageHandler returns the message handler
func (h *Handler) GetMessageHandler() *MessageHandler {
	return h.messageHandler
}
