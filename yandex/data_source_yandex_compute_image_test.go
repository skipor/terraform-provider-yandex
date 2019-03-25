package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceComputeImage_byID(t *testing.T) {
	t.Parallel()

	family := "ubuntu-1804-lts"
	name := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckComputeImageDestroy,
			testAccCheckComputeDiskDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomImageConfig(family, name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_image.source", "image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"name", name),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.source",
						"id"),
					testAccCheckCreatedAtAttr("data.yandex_compute_image.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeImage_byName(t *testing.T) {
	t.Parallel()

	family := "ubuntu-1804-lts"
	name := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckComputeImageDestroy,
			testAccCheckComputeDiskDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomImageConfig(family, name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_image.source", "image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"name", name),
					resource.TestCheckResourceAttr("data.yandex_compute_image.source",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.source",
						"id"),
					testAccCheckCreatedAtAttr("data.yandex_compute_image.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeImage_StandardByFamily(t *testing.T) {
	t.Parallel()

	family := "ubuntu-1804-lts"
	re := regexp.MustCompile("ubuntu")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceStandardImageByFamily(family),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_image.by_family", "image_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_image.by_family",
						"family", family),
					resource.TestCheckResourceAttrSet("data.yandex_compute_image.by_family",
						"id"),
					resource.TestMatchResourceAttr("data.yandex_compute_image.by_family",
						"name", re),
					testAccCheckCreatedAtAttr("data.yandex_compute_image.by_family"),
				),
			},
		},
	})
}

func testAccDataSourceCustomImageResourceConfig(family, name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_image" "image" {
  family        = "%s"
  name          = "%s"
  source_disk   = "${yandex_compute_disk.disk.id}"
  min_disk_size = 10
  os_type       = "linux"
}

resource "yandex_compute_disk" "disk" {
  name = "%s-disk"
  zone = "ru-central1-a"
  size = 4
}
`, family, name, name)
}

const computeImageDataByIDConfig = `
data "yandex_compute_image" "source" {
  image_id = "${yandex_compute_image.image.id}"
}
`

const computeImageDataByNameConfig = `
data "yandex_compute_image" "source" {
  name = "${yandex_compute_image.image.name}"
}
`

func testAccDataSourceCustomImageConfig(family, name string, useID bool) string {
	if useID {
		return testAccDataSourceCustomImageResourceConfig(family, name) + computeImageDataByIDConfig
	}

	return testAccDataSourceCustomImageResourceConfig(family, name) + computeImageDataByNameConfig
}

func testAccDataSourceStandardImageByFamily(family string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "by_family" {
  family = "%s"
}
`, family)
}
