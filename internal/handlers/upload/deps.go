package upload

import "project_sem/usecases/prices"

type Prices interface {
	SavePrices(prices []prices.Price, totalCount int) (*prices.Stats, error)
}
