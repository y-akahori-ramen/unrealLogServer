package elastic

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/y-akahori-ramen/unrealLogServer/db"
)

var verbosityFilterNames = []string{
	"",
	"Warning",
	"Error",
	"Display",
	"Verbose",
	"VeryVerbose",
}

var machAllQuery = map[string]interface{}{
	"match_all": map[string]interface{}{},
}

func GetVerbosityFilterNames(verbosityFilter db.Verbosity) []string {
	names := []string{}
	for i := 0; i < db.VerbosityNum; i++ {
		if verbosityFilter&(db.Verbosity)(1<<i) > 0 {
			names = append(names, verbosityFilterNames[i])
		}
	}
	return names
}

func CreateFilter(filter db.Filter) []map[string]interface{} {
	filters := []map[string]interface{}{}

	verbosityFileterNames := GetVerbosityFilterNames(filter.Verbosity)
	if len(verbosityFileterNames) > 0 {
		verbosityFilter := map[string]interface{}{
			"terms": map[string]interface{}{
				"Verbosity": verbosityFileterNames,
			},
		}
		filters = append(filters, verbosityFilter)
	}

	if len(filter.Categories) > 0 {
		categoryFilter := map[string]interface{}{
			"terms": map[string]interface{}{
				"Category": filter.Categories,
			},
		}
		filters = append(filters, categoryFilter)
	}

	if len(filter.Hosts) > 0 {
		hostFilter := map[string]interface{}{
			"terms": map[string]interface{}{
				"Host": filter.Hosts,
			},
		}
		filters = append(filters, hostFilter)
	}

	if len(filter.Platforms) > 0 {
		platformFilter := map[string]interface{}{
			"terms": map[string]interface{}{
				"Platform": filter.Platforms,
			},
		}
		filters = append(filters, platformFilter)
	}

	if filter.FileOpenAtUnixMilli > 0 {
		fileOpenAtFilter := map[string]interface{}{
			"term": map[string]interface{}{
				"FileOpenAtUnixMilli": filter.FileOpenAtUnixMilli,
			},
		}
		filters = append(filters, fileOpenAtFilter)
	}

	from, to, existTimeRange := filter.GetTimeRange()
	if existTimeRange {
		timeFilter := map[string]interface{}{
			"range": map[string]interface{}{
				"FileOpenAtUnixMilli": map[string]interface{}{
					"gte": from.UnixMilli(),
					"lte": to.UnixMilli(),
				},
			},
		}
		filters = append(filters, timeFilter)
	}

	return filters
}

func CreateQuery(filter db.Filter) map[string]interface{} {
	filters := CreateFilter(filter)
	if len(filters) > 0 {
		return map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filters,
			},
		}
	} else {
		return machAllQuery
	}
}

func CreateCollapseQuery(collapseFieldName string, filter db.Filter) map[string]interface{} {
	query := map[string]interface{}{
		"query": CreateQuery(filter),
		"collapse": map[string]interface{}{
			"field": collapseFieldName,
		},
	}
	return query
}

func HandleError(res *esapi.Response) error {
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return fmt.Errorf("[Error]Error parsing the response body: %s", err)
		} else {
			return fmt.Errorf("[Error][%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}
	return nil
}
