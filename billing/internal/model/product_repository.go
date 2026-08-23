package model

const (
	ProductCreatedKey = "product.created"
	ProductUpdatedKey = "product.updated"
	ProductDeletedKey = "product.deleted"
)

type ProductRepository interface {
	GetProducts() ([]Product, error)
	UpsertProduct(product Product) error
	DeactivateProduct(inventoryID int) error
}
