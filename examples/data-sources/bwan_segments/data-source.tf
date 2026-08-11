# Without `first` or `after`, a list data source walks every page, so `data` holds
# the whole collection.
data "bwan_segments" "all" {}

# Setting `first` or `after` asks for exactly one page instead, and `page_info`
# carries the cursor for the next one.
data "bwan_segments" "first_page" {
  first = 25
  sort  = ["name"]
}

output "segment_names" {
  value = [for segment in data.bwan_segments.all.data : segment.name]
}
