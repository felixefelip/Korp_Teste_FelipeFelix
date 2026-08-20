package model

type ProductRepository interface {
	GetProducts() ([]Product, error)
	GetProductByID(id int) (Product, error)
	CreateProduct(product Product) (int, error)
	UpdateProduct(product Product) error
}
