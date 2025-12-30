package example

import (
	"fmt"
	"strings"
)

func ExamplePerformanceIssues() {
	// This will be flagged for performance issues
	result := ""
	for i := 0; i < 10000; i++ {
		result += fmt.Sprintf("item-%d,", i)
	}
	fmt.Println(result)
}

func inefficientStringBuilding(items []string) string {
	result := ""
	for _, item := range items {
		result = result + item + ","
	}
	return result
}

func betterStringBuilding(items []string) string {
	var builder strings.Builder
	for _, item := range items {
		builder.WriteString(item)
		builder.WriteString(",")
	}
	return builder.String()
}

func veryLongFunctionThatShouldBeBrokenDown(
	param1 string,
	param2 string,
	param3 string,
	param4 string,
	param5 string,
	param6 string,
	param7 string) (string, error) {

	if param1 == "" {
		if param2 == "" {
			if param3 == "" {
				if param4 == "" {
					if param5 == "" {
						if param6 == "" {
							if param7 == "" {
								return "", fmt.Errorf("all parameters are empty")
							}
							return param7, nil
						}
						return param6, nil
					}
					return param5, nil
				}
				return param4, nil
			}
			return param3, nil
		}
		return param2, nil
	}
	return param1, nil
}
