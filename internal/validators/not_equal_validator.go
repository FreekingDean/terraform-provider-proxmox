package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NotEqual(notEqualPaths ...path.Expression) notEqualValidator {
	return notEqualValidator{notEqualPaths}
}

type notEqualValidator struct {
	notEqualPaths path.Expressions
}

// Description describes the validation in plain text formatting.
func (ne notEqualValidator) Description(_ context.Context) string {
	var attributePaths []string
	for _, p := range ne.notEqualPaths {
		attributePaths = append(attributePaths, p.String())
	}

	return fmt.Sprintf("value must not be equal to any of %s", strings.Join(attributePaths, " + "))
}

// MarkdownDescription describes the validation in Markdown formatting.
func (ne notEqualValidator) MarkdownDescription(ctx context.Context) string {
	return ne.Description(ctx)
}

func (ne notEqualValidator) ValidateInt64(ctx context.Context, request validator.Int64Request, response *validator.Int64Response) {

	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}
	// Ensure input path expressions resolution against the current attribute
	expressions := request.PathExpression.MergeExpressions(ne.notEqualPaths...)

	for _, expression := range expressions {
		matchedPaths, diags := request.Config.PathMatches(ctx, expression)
		response.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(request.Path) {
				continue
			}
			// Get the value
			var matchedValue attr.Value
			diags := request.Config.GetAttribute(ctx, mp, &matchedValue)
			response.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			if matchedValue.IsUnknown() {
				return
			}

			if matchedValue.IsNull() {
				continue
			}
			// We know there is a value, convert it to the expected type
			var foundAttrib types.Int64
			diags = tfsdk.ValueAs(ctx, matchedValue, &foundAttrib)
			response.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			if foundAttrib.ValueInt64() == request.ConfigValue.ValueInt64() {
				response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
					request.Path,
					ne.Description(ctx),
					fmt.Sprintf("%d", request.ConfigValue.ValueInt64()),
				))

			}
		}
	}
}
