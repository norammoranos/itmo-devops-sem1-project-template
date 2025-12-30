package download

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"log"
	"net/http"
	"strconv"

	"project_sem/usecases/prices"
)

type Handler struct {
	prices Prices
}

func New(p Prices) *Handler {
	return &Handler{prices: p}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := prices.Filter{
		StartDate: query.Get("start"),
		EndDate:   query.Get("end"),
	}

	if minStr := query.Get("min"); minStr != "" {
		if min, err := strconv.ParseFloat(minStr, 64); err == nil {
			filter.MinPrice = min
		}
	}

	if maxStr := query.Get("max"); maxStr != "" {
		if max, err := strconv.ParseFloat(maxStr, 64); err == nil {
			filter.MaxPrice = max
		}
	}

	priceList, err := h.prices.GetPrices(filter)
	if err != nil {
		log.Println("GetPrices error:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	csvData, err := createCSV(priceList)
	if err != nil {
		log.Println("createCSV error:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	zipData, err := createZip(csvData)
	if err != nil {
		log.Println("createZip error:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
	w.Write(zipData)
}

func createCSV(priceList []prices.Price) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	writer.Write([]string{"id", "name", "category", "price", "create_date"})

	for _, p := range priceList {
		writer.Write([]string{
			strconv.Itoa(p.ID),
			p.Name,
			p.Category,
			strconv.FormatFloat(p.Price, 'f', 2, 64),
			p.CreateDate,
		})
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

func createZip(csvData []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	f, err := writer.Create("data.csv")
	if err != nil {
		return nil, err
	}

	if _, err := f.Write(csvData); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
