resource "aws_iam_role" "this" {
  for_each = local.lambdas

  name               = "${var.name}_${each.key}"
  assume_role_policy = data.aws_iam_policy_document.role.json
  tags               = var.tags
}

resource "aws_iam_role_policy" "this" {
  for_each = local.lambdas

  name   = "${var.name}_${each.key}"
  role   = aws_iam_role.this[each.key].name
  policy = data.aws_iam_policy_document.policy[each.key].json
}

resource "aws_iam_role_policy_attachment" "this" {
  for_each = local.lambdas

  role       = aws_iam_role.this[each.key].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
