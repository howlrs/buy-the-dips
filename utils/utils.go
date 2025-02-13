package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func Split(env string) []string {
	s := strings.Split(env, ".")
	if len(s) != 2 {
		return nil
	}
	return []string{s[0], s[1]}
}

func GetInnerData(result any) map[string]any {
	data, ok := result.(map[string]any)
	if !ok {
		return nil
	}

	return data
}

func GetInnerList(result any) []map[string]any {
	data := GetInnerData(result)
	if data == nil {
		return nil
	}

	interdata, ok := data["list"].([]any)
	if !ok {
		return nil
	}

	list := make([]map[string]any, len(interdata))
	for i, v := range interdata {
		list[i] = v.(map[string]any)
	}

	// length
	TestLog(true, "length: %d", len(list))

	return list
}

func DipsRatio(numbers []float64) float64 {
	latest, prev := numbers[0], numbers[len(numbers)-1]
	return (latest - prev) / prev
}

func ToString(isMarket bool) string {
	if isMarket {
		return "Market"
	}
	return "Limit"
}

func ToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func TestLog(isTest bool, format string, a ...any) {
	if !isTest {
		return
	}
	s := fmt.Sprintf("[test mode] %s", format)
	log.Debug().Msgf(s, a...)
}
