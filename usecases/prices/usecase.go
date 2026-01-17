package prices

import (
	"context"
	"time"

	"project_sem/internal/infrastructure/database"
)

type Price struct {
	ID         int
	Name       string
	Category   string
	Price      float64
	CreateDate string
}

type Stats struct {
	TotalCount      int
	DuplicatesCount int
	TotalItems      int
	TotalCategories int
	TotalPrice      float64
}

type Filter struct {
	StartDate string
	EndDate   string
	MinPrice  float64
	MaxPrice  float64
}

type Usecase struct {
	storage Storage
}

func New(storage Storage) *Usecase {
	return &Usecase{storage: storage}
}

func (u *Usecase) SavePrices(prices []Price, totalCount int) (*Stats, error) {
	ctx := context.Background()

	uniquePrices, duplicatesInInput := removeDuplicates(prices)

	var stats Stats
	stats.TotalCount = totalCount

	err := u.storage.WithTransaction(func(conn database.Connection) error {
		duplicatesInDB := 0

		for _, p := range uniquePrices {
			var exists bool
			err := conn.QueryRow(ctx,
				"SELECT EXISTS(SELECT 1 FROM prices WHERE name = $1 AND category = $2 AND price = $3 AND create_date = $4)",
				p.Name, p.Category, p.Price, p.CreateDate,
			).Scan(&exists)
			if err != nil {
				return err
			}

			if exists {
				duplicatesInDB++
				continue
			}

			_, err = conn.Exec(ctx,
				"INSERT INTO prices (name, category, price, create_date) VALUES ($1, $2, $3, $4)",
				p.Name, p.Category, p.Price, p.CreateDate,
			)
			if err != nil {
				return err
			}
		}

		stats.DuplicatesCount = duplicatesInInput + duplicatesInDB
		stats.TotalItems = len(uniquePrices) - duplicatesInDB

		return conn.QueryRow(ctx, `
			SELECT COUNT(DISTINCT category), COALESCE(SUM(price), 0)
			FROM prices
		`).Scan(&stats.TotalCategories, &stats.TotalPrice)
	})

	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (u *Usecase) GetPrices(filter Filter) ([]Price, error) {
	ctx := context.Background()

	query := "SELECT id, name, category, price, create_date FROM prices WHERE 1=1"
	args := []any{}
	argNum := 1

	if filter.StartDate != "" {
		query += " AND create_date >= $" + itoa(argNum)
		args = append(args, filter.StartDate)
		argNum++
	}

	if filter.EndDate != "" {
		query += " AND create_date <= $" + itoa(argNum)
		args = append(args, filter.EndDate)
		argNum++
	}

	if filter.MinPrice > 0 {
		query += " AND price >= $" + itoa(argNum)
		args = append(args, filter.MinPrice)
		argNum++
	}

	if filter.MaxPrice > 0 {
		query += " AND price <= $" + itoa(argNum)
		args = append(args, filter.MaxPrice)
		argNum++
	}

	query += " ORDER BY id"

	rows, err := u.storage.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []Price
	for rows.Next() {
		var p Price
		var createDate time.Time
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &createDate); err != nil {
			return nil, err
		}
		p.CreateDate = createDate.Format("2006-01-02")
		prices = append(prices, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prices, nil
}

func removeDuplicates(prices []Price) ([]Price, int) {
	seen := make(map[string]bool)
	var result []Price
	duplicates := 0

	for _, p := range prices {
		key := priceKey(p)
		if seen[key] {
			duplicates++
			continue
		}
		seen[key] = true
		result = append(result, p)
	}

	return result, duplicates
}

func priceKey(p Price) string {
	return p.Name + "|" + p.Category + "|" + ftoa(p.Price) + "|" + p.CreateDate
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	negative := i < 0
	if negative {
		i = -i
	}
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	if negative {
		result = "-" + result
	}
	return result
}

func ftoa(f float64) string {
	intPart := int(f)
	fracPart := int((f - float64(intPart)) * 100)
	if fracPart < 0 {
		fracPart = -fracPart
	}
	frac := itoa(fracPart)
	if len(frac) == 1 {
		frac = "0" + frac
	}
	return itoa(intPart) + "." + frac
}

