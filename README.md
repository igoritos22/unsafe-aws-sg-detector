# unsafe-aws-sg-detector
This is a function lambda to detect unsafe security groups ingress rule in AWS enviroment. The function are triggered by a security group event and there is a filter_pattern that detects this changes in cloudtrail logs.

For this context, unsafe ingress rule are defined by rules that exposes the traffic to 0.0.0.0/0 CIDR blocks with exception to HTTP and HTTPS services.

The flow example:


![imagem](https://user-images.githubusercontent.com/73206099/169702454-ffa1c514-bfe8-49e3-8f77-1a6412db64bb.png)


The filter pattern to get any changes in Security Groups:
```json
{($.eventName = AuthorizeSecurityGroupIngress) || ($.eventName = AuthorizeSecurityGroupEgress) || ($.eventName = RevokeSecurityGroupIngress) || ($.eventName = RevokeSecurityGroupEgress) || ($.eventName = CreateSecurityGroup) || ($.eventName = DeleteSecurityGroup)}
```
## Make changes in lambda function
To change anything in fuction lambda you need change the main.go file. After changes you need also build the new zip package to upload in function lambda. The commando to setup the .zip package is:

```console
CGO_ENABLED=0 go build -o main -ldflags '-w' main.go && zip archive.zip main
```
## Notifications
This function don't use any SNS endpoint to publish the findings. There is a external lib that sends an post message directly to slack channel. 

**CAUTION**: Expose the slack webhook channel can make the attackers explore your notification environment, sending malicious link, malwares and other threats.

## Using with Terraform
You can deploy this function with terraform modules. Below you can see a simple example how to do it.

Example of usage:
```hcl
locals {
  environment = "labs"
  region      = "us-east-1"
}
provider "aws" {
  region  = local.region
  profile = local.environment
}
module "aws_cloudtrail" {
  source             = "git@github.com:igoritos22/modules-terraform.git//aws-cloudtrail"
  environment        = local.environment
  s3_bucket_name     = "my-logs"
  s3_bucket_arn      = "arn:aws:s3:::my-logs"
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
  timeout                     = 59 # minutes
  memory_size                 = 128
  slack_webhook               = "https://hooks.slack.com/services/1XXX" #avoid expose slack webhook in plain-text
  cloud_watch_logs_group_arn  = module.aws_cloudtrail.cloudwatch_log_group_arn
  cloud_watch_logs_group_name = module.aws_cloudtrail.cloudwatch_log_group_name
  filter_name                 = "unsafee_ingress_rule"
  filter_pattern              = "{ ($.eventName = AuthorizeSecurityGroupIngress) || ($.eventName = AuthorizeSecurityGroupEgress) || ($.eventName = RevokeSecurityGroupIngress) || ($.eventName = RevokeSecurityGroupEgress) || ($.eventName = CreateSecurityGroup) || ($.eventName = DeleteSecurityGroup)}"
}"
```

## Slack Notification
The notification in slack channel has this layout:

![imagem](https://user-images.githubusercontent.com/73206099/169703053-4e7d6bbb-c9bf-40dd-99bf-8a19eb7f8995.png)

To customize that notification you need change the slack block in main.go file:

```golang
							webhookUrl := os.Getenv("SLACK_WEBHOOK")

							attachment1 := slack.Attachment{}
							attachment1.AddField(slack.Field{Title: "SG Name:", Value: *sg.GroupName}).AddField(slack.Field{Title: "FromPort", Value: strFromPort}).AddField(slack.Field{Title: "ToPort", Value: strToPort}).AddField(slack.Field{Title: "Allow CIDR block:", Value: "0.0.0.0/0"})
							attachment1.AddAction(slack.Action{Type: "button", Text: "AWS console login", Url: "https://aws.amazon.com/pt/console/", Style: "primary"})
							payload := slack.Payload{
								Text:        "Unsafe rule ingress detected for sg: " + *sg.GroupId,
								Username:    "mordor-eye",
								Channel:     "C000000012W",
								Attachments: []slack.Attachment{attachment1},
```
## Contributing
Feel free to contribute or make a issue. This function are made of worker to worker.