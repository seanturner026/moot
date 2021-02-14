resource "aws_dynamodb_table" "this" {
  name         = var.tags.name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"
  range_key    = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "RepoProvider"
    type = "S"
  }

  global_secondary_index {
    name            = var.global_secondary_index_name
    hash_key        = "SK"
    range_key       = "RepoProvider"
    projection_type = "ALL"
  }

  tags = var.tags
}
