package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/provider"
)

func TestCustomEphemeralResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:          config.TestNameDirectory(),
				ProtoV6ProviderFactories: provider.GetTestProviders(),
				// The provider server starting and recognizing the ephemeral resource to
				// the point of calling its implementation logic is sufficient
				// to verify inclusion of the custom ephemeral resource.
				ExpectError: regexp.MustCompile(`Intentionally Missing Open Implementation`),
			},
		},
	})
}
