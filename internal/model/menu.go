package model

type Menu struct {
	ID         int64   `db:"id"`
	Name       string  `db:"name"`
	Price      float32 `db:"price"`
	Categories string  `db:"categories"`
}

type MenuQuery struct {
	IDs             []int64
	Names           []string
	ExactNamesMatch bool
	MaxPrice        float32 // optional for search query
	MinPrice        float32 // idem
	// Price           float32 `db:"price"`
	Categories string `db:"categories"`
}

type CreateMenuRequest struct {
	Name       string  `json:"name" validate:"required,max=250"`
	Price      float32 `json:"price" validate:"required,gte=0.05"`
	Categories string  `json:"categories"`
} //	@name	create-update_menu_request

type CreateMenuResponse struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Price      float32 `json:"price"`
	Categories string  `json:"categories"`
} //	@name	create-get-update_menu_response

type GetMenuResponse = CreateMenuResponse

type UpdateMenuRequest = CreateMenuRequest
type UpdateMenuResponse = CreateMenuResponse

type MenuResponse struct {
	Menu interface{} `json:"menu"`
} //	@name	menu_response
