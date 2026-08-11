# Every page is walked, so data holds the whole collection. A filter narrows
# it; first or after ask for one page instead.
data "netskopebwan_controller_operators" "example" {
  filter = "name eq \"example\""
}
