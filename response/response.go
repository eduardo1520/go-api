package response

type Message struct {
	Message string `json:"message"`
}

type Product struct {
	ID    int     `json:"id_product"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
