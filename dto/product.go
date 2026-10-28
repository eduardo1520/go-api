package dto

type ProductRequest struct {
	Name  string  `json:"name" binding:"required,min=1,max=120"`
	Price float64 `json:"price" binding:"required,gt=0"`
}
