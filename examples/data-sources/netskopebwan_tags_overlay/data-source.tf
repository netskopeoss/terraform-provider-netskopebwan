# Generated from the provider's schema: the arguments this object requires, with
# placeholders to fill in. Remove these two lines to maintain the example by hand.
# Every page is walked, so data holds the whole collection. A filter narrows
# it; first or after ask for one page instead.
data "netskopebwan_tags_overlay" "example" {
  filter = "name eq \"example\""
}
