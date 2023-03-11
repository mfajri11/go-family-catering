package service

import "family-catering/internal/model"

func newOwnerResponse(owner *model.Owner) *model.GetOwnerResponse {
	res := &model.GetOwnerResponse{
		Id:          owner.Id,
		Name:        owner.Name,
		Email:       owner.Email,
		PhoneNumber: owner.PhoneNumber,
	}

	if owner.DateOfBirth.Valid {
		res.DateOfBirth = owner.DateOfBirth.String
	}

	return res
}

func newOwnerCreateResponse(owner *model.Owner) *model.CreateOwnerResponse {
	res := &model.CreateOwnerResponse{
		Name:        owner.Name,
		Email:       owner.Email,
		Password:    owner.Password,
		PhoneNumber: owner.PhoneNumber,
	}
	return res
}

func newListOwnersResponse(owners []*model.Owner) []*model.GetOwnerResponse {
	if len(owners) == 0 {
		return []*model.GetOwnerResponse{}
	}

	ress := make([]*model.GetOwnerResponse, 0, len(owners))

	for _, owner := range owners {
		res := newOwnerResponse(owner)
		ress = append(ress, res)

	}

	return ress

}

// menu
func newMenuResponse(menu *model.Menu) *model.GetMenuResponse {
	menuResp := &model.GetMenuResponse{
		ID:         menu.ID,
		Name:       menu.Name,
		Price:      menu.Price,
		Categories: menu.Categories,
	}

	return menuResp
}

func newMenusResponse(menus []*model.Menu) []*model.GetMenuResponse {
	if len(menus) == 0 {
		return []*model.GetMenuResponse{}
	}

	ress := make([]*model.GetMenuResponse, 0, len(menus))

	for _, menu := range menus {
		res := newMenuResponse(menu)
		ress = append(ress, res)
	}

	return ress
}
