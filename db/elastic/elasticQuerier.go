package elastic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/y-akahori-ramen/unrealLogServer/db"
)

type ElasticQuerier struct {
	client *elasticsearch.Client
	index  string
}

func NewElasticQuerier(index string, config elasticsearch.Config) (*ElasticQuerier, error) {
	es, err := elasticsearch.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &ElasticQuerier{index: index, client: es}, nil
}

func (q *ElasticQuerier) searchCollapseValues(ctx context.Context, collapseFieldName string, filter db.Filter, from int, size int, sortKey ...string) ([]string, error) {
	values := []string{}

	query := CreateCollapseQuery(collapseFieldName, filter)
	res, err := q.client.Search(
		q.client.Search.WithContext(ctx),
		q.client.Search.WithIndex(q.index),
		q.client.Search.WithBody(esutil.NewJSONReader(&query)),
		q.client.Search.WithSource("false"),
		q.client.Search.WithFrom(from),
		q.client.Search.WithSize(size),
		q.client.Search.WithSort(sortKey...),
	)
	if err != nil {
		return values, err
	}
	defer res.Body.Close()

	err = HandleError(res)
	if err != nil {
		return values, err
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return values, fmt.Errorf("[Error]Error parsing the response body: %s", err)
	}

	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		fields := hit.(map[string]interface{})["fields"].(map[string]interface{})
		for _, categoryValue := range fields[collapseFieldName].([]interface{}) {
			values = append(values, categoryValue.(string))
		}
	}

	return values, nil
}

func (q *ElasticQuerier) searchAllCollapseValues(ctx context.Context, collapseFieldName string, filter db.Filter, sortKey ...string) ([]string, error) {
	const step = 100
	values := []string{}

	for from := 0; ; from += step {
		receivedValues, err := q.searchCollapseValues(ctx, collapseFieldName, filter, from, step, sortKey...)
		if err != nil {
			return nil, err
		}
		if len(receivedValues) == 0 {
			break
		} else {
			values = append(values, receivedValues...)
			if len(receivedValues) < step {
				break
			}
		}
	}

	return values, nil
}

// getLog ログを時刻の昇順で取得しlogHandlerに渡す
func (q *ElasticQuerier) getLog(ctx context.Context, logHandler db.LogHandler, filter db.Filter, searchAfter, size int) (int, int, error) {
	// Elasticsearchではfromとsizeによる指定は10000を超える場合にエラーとなり、大量のドキュメントを取得する場合にはsearchAfterを使用することが推奨されている
	// https://www.elastic.co/guide/en/elasticsearch/reference/current/paginate-search-results.html
	// ログの行数が10000を超えることは十分にありえるためsearchafterによるデータ取得を行う
	query := map[string]interface{}{
		"query":        CreateQuery(filter),
		"search_after": []interface{}{searchAfter},
	}

	res, err := q.client.Search(
		q.client.Search.WithContext(ctx),
		q.client.Search.WithIndex(q.index),
		q.client.Search.WithBody(esutil.NewJSONReader(&query)),
		q.client.Search.WithSource("true"),
		q.client.Search.WithSize(size),
		q.client.Search.WithSort("@timestamp:asc"),
	)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return 0, 0, fmt.Errorf("[Error]Error parsing the response body: %s", err)
	}

	var nextSearchAfter int
	var logCount int

	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		category := source["Category"].(string)
		verbosity := source["Verbosity"].(string)
		log := source["Log"].(string)
		nextSearchAfter = (int)(hit.(map[string]interface{})["sort"].([]interface{})[0].(float64))

		err = logHandler(db.LogData{Category: category, Verbosity: verbosity, Log: log})
		if err != nil {
			return logCount, 0, err
		}
		logCount++
	}

	return logCount, nextSearchAfter, nil
}

func (q *ElasticQuerier) GetLog(ctx context.Context, logHandler db.LogHandler, filter db.Filter) error {
	const step = 1000
	searchAfter := 0
	for {
		logCount, nextSearchAfter, err := q.getLog(ctx, logHandler, filter, searchAfter, step)
		if err != nil {
			return err
		}
		if logCount < step {
			break
		}
		searchAfter = nextSearchAfter
	}

	return nil
}

func (q *ElasticQuerier) GetHosts(ctx context.Context, filter db.Filter) ([]string, error) {
	return q.searchAllCollapseValues(ctx, "Host", filter, "Host:asc")
}

func (q *ElasticQuerier) GetPlatforms(ctx context.Context, filter db.Filter) ([]string, error) {
	return q.searchAllCollapseValues(ctx, "Platform", filter, "Platform:asc")
}

func (q *ElasticQuerier) GetCategories(ctx context.Context, id db.LogId) ([]string, error) {
	return q.searchAllCollapseValues(ctx, "Category", db.NewFilterFromLogID(id), "Category:asc")
}

func (q *ElasticQuerier) GetVerbosities(ctx context.Context, id db.LogId) ([]string, error) {
	return q.searchAllCollapseValues(ctx, "Verbosity", db.NewFilterFromLogID(id), "Verbosity:asc")
}

func (q *ElasticQuerier) GetIds(ctx context.Context, filter db.Filter, from int, size int) ([]db.LogId, error) {
	ids := []db.LogId{}

	query := CreateCollapseQuery("LogID", filter)

	res, err := q.client.Search(
		q.client.Search.WithContext(ctx),
		q.client.Search.WithIndex(q.index),
		q.client.Search.WithBody(esutil.NewJSONReader(&query)),
		q.client.Search.WithSource("true"),
		q.client.Search.WithFrom(from),
		q.client.Search.WithSize(size),
		q.client.Search.WithSort("FileOpenAtUnixMilli:desc"),
	)
	if err != nil {
		return ids, err
	}
	defer res.Body.Close()

	err = HandleError(res)
	if err != nil {
		return ids, err
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return ids, fmt.Errorf("[Error]Error parsing the response body: %s", err)
	}

	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		host := source["Host"].(string)
		platform := source["Platform"].(string)
		fileOpenAtUnixMilli := (int64)(source["FileOpenAtUnixMilli"].(float64))
		id := db.LogId{Host: host, Platform: platform, FileOpenAtUnixMilli: fileOpenAtUnixMilli}
		ids = append(ids, id)
	}

	return ids, nil
}
