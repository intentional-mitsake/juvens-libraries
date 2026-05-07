package services

import (
	"fmt"
	"juvens-library/internal/models"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

func SearchBooks(query string) (models.Book, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	queries := strings.Split(query, " ") // split the query into individual words
	for i, q := range queries {
		queries[i] = q + "+"
	}
	query = strings.Join(queries, "")
	logger.Info("Formatted query", "query", query)
	url := fmt.Sprintf(`https://openlibrary.org/search.json?q=%s`, query)
	resp, err := client.Get(url)
	if err != nil {
		return models.Book{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.Book{}, fmt.Errorf("search failed with status code %d", resp.StatusCode)
	}
	logger.Info("Book search successful", "response", resp.Body)
	return models.Book{}, nil
}

/*API RESULT FORMAT:
    {
    "start": 0,
    "num_found": 629,
    "docs": [
        {...},
        {...},
        ...
        {...}]
	}
EACH DOCS IS:
	{
    "cover_i": 258027,
    "has_fulltext": true,
    "edition_count": 120,
    "title": "The Lord of the Rings",
    "author_name": [
        "J. R. R. Tolkien"
    ],
    "first_publish_year": 1954,
    "key": "OL27448W",
    "ia": [
        "returnofking00tolk_1",
        "lordofrings00tolk_1",
        "lordofrings00tolk_0",
    ],
    "author_key": [
        "OL26320A"
    ],
    "public_scan_b": true
}
*/
