package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccTorKeys() string {
	return fmt.Sprintf(`
resource "tor_keys" "test" {

}
`)
}

func TestAccTorKeysResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create / Read testing
			{
				Config: testAccTorKeys(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("tor_keys.test", "public_key"),
					resource.TestCheckResourceAttrSet("tor_keys.test", "private_key"),
					resource.TestCheckResourceAttrSet("tor_keys.test", "address"),
				),
			},
		},
	})
}
