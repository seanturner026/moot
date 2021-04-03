# resource "aws_sfn_state_machine" "this" {
#   name       = var.tags.name
#   role_arn   = aws_iam_role.iam_for_sfn.arn
#   type       = "STANDARD"
#   definition = templatefile("${path.root}/assets/sfn.json", { onboarding_lambda_arn = aws_lambda_function.this["admin"].arn })

#   logging_configuration {
#     log_destination        = "${aws_cloudwatch_log_group.log_group_for_sfn.arn}:*"
#     include_execution_data = true
#     level                  = "ERROR"
#   }

#   tags = var.tags
# }
