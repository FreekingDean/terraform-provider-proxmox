package utils

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NormalizeMap(in map[string]interface{}) (map[string]types.String, diag.Diagnostics) {
	var diags diag.Diagnostics
	normalizedMap := make(map[string]types.String)
	for k, v := range in {
		val := ""
		switch t := v.(type) {
		case string:
			val = t
		case float64, float32:
			val = strings.TrimRight(
				strings.TrimRight(fmt.Sprintf("%.2f", 100.0), "0"),
				".",
			)
		default:
			diags = append(diags, diag.NewErrorDiagnostic(
				fmt.Sprintf("Could not convert unkown type %T", v),
				fmt.Sprintf("Attempted to convert response[%s] from %T to string", k, v),
			))
		}
		normalizedMap[k] = types.StringValue(val)
	}
	return normalizedMap, diags
}
