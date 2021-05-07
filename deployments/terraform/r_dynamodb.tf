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

  tags = var.tags
}

resource "random_uuid" "this" {
  count = var.enable_admin_user_creation ? 1 : 0
}

resource "aws_dynamodb_table_item" "this" {
  count      = var.enable_admin_user_creation ? 1 : 0
  depends_on = [null_resource.create_admin_user]

  table_name = aws_dynamodb_table.this.name
  hash_key   = aws_dynamodb_table.this.hash_key
  range_key  = aws_dynamodb_table.this.range_key

  item = templatefile(
    "${path.root}/assets/dynamodb_put_item_input.json",
    {
      admin_user_email = var.admin_user_email
      uud              = random_uuid.this[0].id
    }
  )
}
