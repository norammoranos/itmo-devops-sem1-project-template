package download

import "project_sem/usecases/prices"

type Prices interface {
	GetPrices(filter prices.Filter) ([]prices.Price, error)
}
