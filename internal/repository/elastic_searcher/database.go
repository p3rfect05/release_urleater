package elastic_searcher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
)

type Searcher struct {
	client *elastic.Client
	url    string
}

func NewSearcher(client *elastic.Client, url string) *Searcher {
	return &Searcher{client: client, url: url}
}

func (s *Searcher) SearchShortLinks(ctx context.Context, word string, limit int, offset int) ([]string, error) {
	type ShortLink struct {
		ShortURL string `json:"short_url"`
	}

	query := elastic.NewWildcardQuery("short_url.keyword", "*"+word+"*")
	fmt.Println("*"+word+"*", offset)
	// Выполняем запрос с учетом пагинации: offset и limit.
	searchResult, err := s.client.Search().
		Index("short_links").
		Query(query).
		From(offset).
		Size(limit).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing elastic query: %w", err)
	}

	var links []string
	fmt.Println(searchResult.Hits.Hits)
	for _, hit := range searchResult.Hits.Hits {
		var sl ShortLink
		if err := json.Unmarshal(hit.Source, &sl); err != nil {
			continue
		}
		links = append(links, sl.ShortURL)
	}

	fmt.Println(links)
	return links, nil
}

func (s *Searcher) AddShortLink(ctx context.Context, link string) error {
	doc := map[string]interface{}{
		"short_url": link,
	}

	_, err := s.client.Index().
		Index("short_links").
		BodyJson(doc).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("error getting short link: %w", err)
	}

	return nil
}

func (s *Searcher) DeleteShortLink(ctx context.Context, link string) error {
	query := elastic.NewTermQuery("short_url.keyword", link)

	res, err := s.client.DeleteByQuery().
		Index("short_links").
		Query(query).
		Refresh("true").
		Do(ctx)
	if err != nil {
		return fmt.Errorf("ошибка удаления короткой ссылки: %w", err)
	}

	if res.Deleted == 0 {
		return fmt.Errorf("документ не найден")
	}

	return nil
}
