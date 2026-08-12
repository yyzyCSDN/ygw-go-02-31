package report

import (
	"sort"

	"example.com/eventledger/model"
)

type Summary struct {
	Total       int
	Active      int
	Inactive    int
	TotalAmount int64
	MinAmount   int64
	MaxAmount   int64
	ByStatus    map[string]int
	ByTag       map[string]int
}

func Build(values []model.EventRecord) Summary {
	result := Summary{ByStatus: map[string]int{}, ByTag: map[string]int{}}
	for index, value := range values {
		result.Total++
		if value.Active {
			result.Active++
		} else {
			result.Inactive++
		}
		result.TotalAmount += value.Amount
		if index == 0 || value.Amount < result.MinAmount {
			result.MinAmount = value.Amount
		}
		if index == 0 || value.Amount > result.MaxAmount {
			result.MaxAmount = value.Amount
		}
		result.ByStatus[value.Status]++
		for _, tag := range model.NormalizeTags(value.Tags) {
			result.ByTag[tag]++
		}
	}
	return result
}

func AverageAmount(summary Summary) float64 {
	if summary.Total == 0 {
		return 0
	}
	return float64(summary.TotalAmount) / float64(summary.Total)
}

func TopTags(summary Summary, limit int) []string {
	if limit <= 0 {
		return []string{}
	}
	type pair struct {
		name  string
		count int
	}
	pairs := make([]pair, 0, len(summary.ByTag))
	for name, count := range summary.ByTag {
		pairs = append(pairs, pair{name: name, count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].name < pairs[j].name
	})
	if limit > len(pairs) {
		limit = len(pairs)
	}
	result := make([]string, limit)
	for index := range result {
		result[index] = pairs[index].name
	}
	return result
}

func Statuses(summary Summary) []string {
	result := make([]string, 0, len(summary.ByStatus))
	for status := range summary.ByStatus {
		result = append(result, status)
	}
	sort.Strings(result)
	return result
}
