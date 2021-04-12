// Allows for things like CloudWatch logging and metrics.
data "aws_iam_policy" "AWSLambdaBasicExecutionRole" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

// Gives permissions to the Yeti topic to invoke our Lambda.
resource "aws_lambda_permission" "with_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.dinghy_cache_bust_function.function_name
  principal     = "sns.amazonaws.com"
  source_arn = var.notification_topic_arn
}

// Gives our Lambda permission to assume roles and attach the execution role.
resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"
  managed_policy_arns = [ data.aws_iam_policy.AWSLambdaBasicExecutionRole.arn ]
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

// Define the source for our Lambda.
resource "aws_lambda_function" "dinghy_cache_bust_function" {
  filename      = "../cachelambda/lambda_function_payload.zip"
  function_name = "dinghy_cache_bust_function"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "main"

  # The filebase64sha256() function is available in Terraform 0.11.12 and later
  source_code_hash = filebase64sha256("../cachelambda/lambda_function_payload.zip")

  runtime = "go1.x"
}

// Creates a binding from Yeti's SNS topic to our Lambda so that we always
// recieve config updates.
resource "aws_sns_topic_subscription" "dinghy_yeti_integration" {
  topic_arn = var.notification_topic_arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.dinghy_cache_bust_function.arn
}
