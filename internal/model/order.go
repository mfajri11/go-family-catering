package model

type Order struct {
	BaseOrderID   int64   `db:"base_order_id"` // unique per menu_id
	OrderID       int64   `db:"order_id"`      // to allow one user order multiple menu
	CustomerEmail string  `db:"customer_email"`
	Qty           int     `db:"qty"`
	MenuID        int64   `db:"menu_id"`
	MenuName      string  `db:"menu_name"` // menu's id
	Price         float32 `db:"price"`
	Status        int     `db:"status"` // 1 NEW, 2 PAID, 3 Cancelled
	CreatedAt     string  `db:"created_at"`
	UpdatedAt     string  `db:"updated_at"`
}

type OrderQuery struct {
	ID                  int64
	MenuNames           []string
	ExactMenuNamesMatch bool
	CustomerEmails      []string
	Qty                 int
	Price               float32
	Status              int
	MinPrice            float32
	MaxPrice            float32
	StartDay            string
	EndDay              string
}

type BaseOrder struct {
	ID     int
	Qty    int
	Name   string
	MenuId int64
	Status int
} //	@name	base_order

type BaseOrderRequest struct {
	Name string `json:"name" validate:"required"`
	Qty  int    `json:"qty" validate:"required,numeric"`
}

type SearchResponse struct {
	OrderID       int64   `json:"order_id,omitempty"`
	CustomerEmail string  `json:"customer_email,omitempty"`
	Qty           int     `json:"qty"`
	MenuName      string  `json:"menu_name"`
	MenuId        int64   `json:"menu_id,omitempty"`
	Price         float32 `json:"price"`
	Status        int     `json:"status"`
	CreatedAt     string  `json:"created_at"`
}

type SearchOrdersResponse struct {
	Orders     []*SearchResponse `json:"orders"`
	TotalPrice float32           `json:"total_price"`
}

type CreateOrderRequest struct {
	CustomerEmail string             `json:"customer_email" validate:"required,email"`
	Orders        []BaseOrderRequest `json:"orders"`
}

type ConfirmPaymentRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type CreateOrderResponse struct {
	OrderID       int64  `json:"order_id"`
	CustomerEmail string `json:"customer_email"`
	// Orders        []BaseOrder `json:"orders"`
	Message    string  `json:"message"`
	TotalPrice float32 `json:"total_price"`
}

// type UpdateOrderRequest struct {
// 	ID            int64
// 	CustomerEmail string
// 	Order
// }

type CancelUnpaidOrderResponse struct {
	Message             string `json:"message"`
	TotalOrderCancelled int64  `json:"total_order_cancelled"`
}

type OrderResponse struct {
	Order interface{} `json:"order"`
}
