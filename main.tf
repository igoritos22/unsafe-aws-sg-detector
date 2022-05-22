locals {
  environment = "labs"
}
provider "aws" {
  region  = "us-east-1"
  profile = local.environment
}
module "aws_cloudtrail" {
  source             = "git@github.com:igoritos22/modules-terraform.git//aws-cloudtrail"
  environment        = local.environment
  s3_bucket_name     = "my-labs-cloudtrail-logs"
  s3_bucket_arn      = "arn:aws:s3:::my-labs-cloudtrail-logs"
  log_retention_days = 90
}
module "aws_lambda" {
  source                      = "git@github.com:igoritos22/modules-terraform.git//aws-lambda"
  environment                 = local.environment
  function_name               = "unsafe_sg_ingress_rule"
  filename                    = "archive.zip"
  lambda_role_name            = "lambdaExecutionRole"
  source_code_hash            = filebase64sha256("archive.zip")
  handler                     = "main"
  runtime                     = "go1.x"
  timeout                     = 59
  memory_size                 = 128
  slack_webhook               = "https://hooks.slack.com/services/1XXX"
  cloud_watch_logs_group_arn  = module.aws_cloudtrail.cloudwatch_log_group_arn
  cloud_watch_logs_group_name = module.aws_cloudtrail.cloudwatch_log_group_name
  filter_name                 = "unsafee_ingress_rule"
  filter_pattern              = "{ ($.eventName = AuthorizeSecurityGroupIngress) || ($.eventName = AuthorizeSecurityGroupEgress) || ($.eventName = RevokeSecurityGroupIngress) || ($.eventName = RevokeSecurityGroupEgress) || ($.eventName = CreateSecurityGroup) || ($.eventName = DeleteSecurityGroup)}"
}