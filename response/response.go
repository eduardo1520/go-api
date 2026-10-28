package response

import "fmt"

type Message struct {
	Message string `json:"message"`
}

type Product struct {
	ID    int    `json:"id_product"`
	Name  string `json:"name"`
	Price string `json:"price"`
}

func FormatPrice(price float64) string {
	return fmt.Sprintf("%.2f", price)
}
