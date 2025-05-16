package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Thedrogon/blogbish/search-service/internal/models"
	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchRepository struct {
	client *elasticsearch.Client
}

func NewElasticsearchRepository(client *elasticsearch.Client) *ElasticsearchRepository {
	return &ElasticsearchRepository{client: client}
}

func (r *ElasticsearchRepository) IndexPost(ctx context.Context, post *models.SearchablePost) error {
	body, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post: %w", err)
	}

	res, err := r.client.Index(
		"posts",
		bytes.NewReader(body),
		r.client.Index.WithDocumentID(post.ID),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to index post: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index post: %s", res.String())
	}

	return nil
}

func (r *ElasticsearchRepository) IndexComment(ctx context.Context, comment *models.SearchableComment) error {
	body, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	res, err := r.client.Index(
		"comments",
		bytes.NewReader(body),
		r.client.Index.WithDocumentID(comment.ID),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to index comment: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index comment: %s", res.String())
	}

	return nil
}

func (r *ElasticsearchRepository) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	var query map[string]interface{}

	// Build the search query based on the request
	if req.Type == "post" || req.Type == "all" {
		query = buildPostQuery(req)
	} else if req.Type == "comment" {
		query = buildCommentQuery(req)
	}

	// Execute search
	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(getIndices(req.Type)...),
		r.client.Search.WithBody(bytes.NewReader(mustMarshal(query))),
		r.client.Search.WithFrom(req.From),
		r.client.Search.WithSize(req.Size),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return parseSearchResponse(result, req), nil
}

func (r *ElasticsearchRepository) Suggest(ctx context.Context, req *models.SuggestionRequest) (*models.SuggestionResponse, error) {
	query := buildSuggestionQuery(req)

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(getIndices(req.Type)...),
		r.client.Search.WithBody(bytes.NewReader(mustMarshal(query))),
		r.client.Search.WithSize(req.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute suggestion: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("suggestion failed: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return parseSuggestionResponse(result), nil
}

// Helper functions

func buildPostQuery(req *models.SearchRequest) map[string]interface{} {
	must := []map[string]interface{}{
		{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{"title^3", "content^2", "excerpt", "tags^2", "categories"},
				"type":   "best_fields",
			},
		},
	}

	if req.Status != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"status": req.Status,
			},
		})
	}

	if len(req.Tags) > 0 {
		must = append(must, map[string]interface{}{
			"terms": map[string]interface{}{
				"tags": req.Tags,
			},
		})
	}

	if req.Category != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"categories": req.Category,
			},
		})
	}

	sort := []map[string]interface{}{}
	if req.SortBy == "date" {
		sort = append(sort, map[string]interface{}{
			"created_at": map[string]interface{}{
				"order": req.SortOrder,
			},
		})
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
		"sort": sort,
	}
}

func buildCommentQuery(req *models.SearchRequest) map[string]interface{} {
	must := []map[string]interface{}{
		{
			"match": map[string]interface{}{
				"content": map[string]interface{}{
					"query":    req.Query,
					"operator": "and",
				},
			},
		},
	}

	if req.Status != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"status": req.Status,
			},
		})
	}

	sort := []map[string]interface{}{}
	if req.SortBy == "date" {
		sort = append(sort, map[string]interface{}{
			"created_at": map[string]interface{}{
				"order": req.SortOrder,
			},
		})
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
		"sort": sort,
	}
}

func buildSuggestionQuery(req *models.SuggestionRequest) map[string]interface{} {
	var field string
	switch req.Type {
	case "post":
		field = "title"
	case "comment":
		field = "content"
	case "tag":
		field = "tags"
	case "category":
		field = "categories"
	}

	return map[string]interface{}{
		"suggest": map[string]interface{}{
			"text": req.Query,
			"completion": map[string]interface{}{
				"field": field + "_suggest",
				"size":  req.Limit,
			},
		},
	}
}

func getIndices(searchType string) []string {
	switch searchType {
	case "post":
		return []string{"posts"}
	case "comment":
		return []string{"comments"}
	case "all":
		return []string{"posts", "comments"}
	default:
		return []string{"posts"}
	}
}

func parseSearchResponse(result map[string]interface{}, req *models.SearchRequest) *models.SearchResponse {
	hits := result["hits"].(map[string]interface{})
	total := int64(hits["total"].(map[string]interface{})["value"].(float64))

	response := &models.SearchResponse{
		Total: total,
		From:  req.From,
		Size:  req.Size,
	}

	if req.Type == "post" || req.Type == "all" {
		var posts []*models.SearchablePost
		for _, hit := range hits["hits"].([]interface{}) {
			source := hit.(map[string]interface{})["_source"]
			post := &models.SearchablePost{}
			sourceBytes, _ := json.Marshal(source)
			json.Unmarshal(sourceBytes, post)
			posts = append(posts, post)
		}
		response.Posts = posts
	}

	if req.Type == "comment" || req.Type == "all" {
		var comments []*models.SearchableComment
		for _, hit := range hits["hits"].([]interface{}) {
			source := hit.(map[string]interface{})["_source"]
			comment := &models.SearchableComment{}
			sourceBytes, _ := json.Marshal(source)
			json.Unmarshal(sourceBytes, comment)
			comments = append(comments, comment)
		}
		response.Comments = comments
	}

	return response
}

func parseSuggestionResponse(result map[string]interface{}) *models.SuggestionResponse {
	suggestions := make([]string, 0)
	if suggest, ok := result["suggest"].(map[string]interface{}); ok {
		for _, option := range suggest["completion"].([]interface{}) {
			text := option.(map[string]interface{})["text"].(string)
			suggestions = append(suggestions, text)
		}
	}
	return &models.SuggestionResponse{Suggestions: suggestions}
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
