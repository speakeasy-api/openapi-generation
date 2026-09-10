package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
)

func TestCustomListResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// terraform-plugin-testing 1.14.0 requires apply before query.
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				// This does not need to actually apply anything.
				ExpectError: regexp.MustCompile(`Intentionally Missing Create Implementation`),
			},
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				Query:                    true,
				// The provider server starting and recognizing the resource to
				// the point of calling its implementation logic is sufficient
				// to verify inclusion of the custom resource.
				ExpectError: regexp.MustCompile(`Intentionally Missing List Implementation`),
			},
		},
	})
}
