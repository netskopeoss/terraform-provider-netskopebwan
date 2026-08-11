# Every page is walked, so data holds the whole collection. A filter narrows
# it; first or after ask for one page instead.
data "netskopebwan_webroot_categories" "example" {
  filter = "name eq \"example\""
}
