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
  name = "iam_for_lambda_${terraform.workspace}"
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

// policy that defines permissions for the lambda.
resource "aws_iam_policy" "lambda_policy" {
  name        = "dinghy_cache_bust_function_policy_${terraform.workspace}"
  path        = "/"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

// attach the policy to role
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_policy.arn
}


// Define the source for our Lambda.
resource "aws_lambda_function" "dinghy_cache_bust_function" {
  s3_bucket     = var.bucket_name
  s3_key        = "${var.lambda_key}.zip"
  function_name = "dinghy_cache_bust_function_${terraform.workspace}"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "main"

  runtime = "go1.x"


  environment {
    variables = {
      DINGHY_URL = var.dinghy_host
      VERSION    = var.lambda_key
    }
  }
}

// Creates a binding from Yeti's SNS topic to our Lambda so that we always
resource "aws_cloudwatch_log_group" "dinghy_cache_bust_function" {
  name              = "/aws/lambda/dinghy_cache_bust_function_${terraform.workspace}"
  retention_in_days = 30
}

// Creates a binding from Yeti's SNS topic to our Lambda so that we always
// recieve config updates.
resource "aws_sns_topic_subscription" "dinghy_yeti_integration" {
  topic_arn = var.notification_topic_arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.dinghy_cache_bust_function.arn
}
