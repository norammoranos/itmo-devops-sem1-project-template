package upload

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"project_sem/usecases/prices"
)

var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

type Handler struct {
	prices Prices
}

func New(p Prices) *Handler {
	return &Handler{prices: p}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	archiveType := r.URL.Query().Get("type")
	if archiveType == "" {
		archiveType = "zip"
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var csvData []byte
	switch archiveType {
	case "zip":
		csvData, err = extractFromZip(fileContent)
	case "tar":
		csvData, err = extractFromTar(fileContent)
	default:
		http.Error(w, "unsupported archive type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	priceList, totalCount, err := parseCSV(csvData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stats, err := h.prices.SavePrices(priceList, totalCount)
	if err != nil {
		log.Println("SavePrices error:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		TotalCount:      stats.TotalCount,
		DuplicatesCount: stats.DuplicatesCount,
		TotalItems:      stats.TotalItems,
		TotalCategories: stats.TotalCategories,
		TotalPrice:      stats.TotalPrice,
	})
}

func extractFromZip(data []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".csv") {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}

	return nil, fmt.Errorf("csv file not found")
}

func extractFromTar(data []byte) ([]byte, error) {
	reader := tar.NewReader(bytes.NewReader(data))

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(header.Name, ".csv") {
			return io.ReadAll(reader)
		}
	}

	return nil, fmt.Errorf("csv file not found")
}

func parseCSV(data []byte) ([]prices.Price, int, error) {
	reader := csv.NewReader(bytes.NewReader(data))

	if _, err := reader.Read(); err != nil {
		return nil, 0, err
	}

	var result []prices.Price
	totalCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}

		totalCount++

		if len(record) != 5 {
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil || id <= 0 {
			continue
		}

		name := strings.TrimSpace(record[1])
		if name == "" {
			continue
		}

		category := strings.TrimSpace(record[2])
		if category == "" {
			continue
		}

		price, err := strconv.ParseFloat(record[3], 64)
		if err != nil || price < 0 {
			continue
		}

		createDate := strings.TrimSpace(record[4])
		if !dateRegex.MatchString(createDate) {
			continue
		}

		result = append(result, prices.Price{
			ID:         id,
			Name:       name,
			Category:   category,
			Price:      price,
			CreateDate: createDate,
		})
	}

	return result, totalCount, nil
}
