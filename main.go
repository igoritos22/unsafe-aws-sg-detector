package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func handler() {
	svc := ec2.New(session.New())
	var maxResults int64
	maxResults = 500
	var nextToken string

	for {
		descSGInput := &ec2.DescribeSecurityGroupsInput{
			MaxResults: aws.Int64(maxResults),
		}

		if nextToken != "" {
			descSGInput.NextToken = aws.String(nextToken)
		}

		descSGOut, err := svc.DescribeSecurityGroups(descSGInput)
		if err != nil {
			log.Println("Unable to get security groups. Err:", err)
			os.Exit(1)
		}

		for _, sg := range descSGOut.SecurityGroups {
			for _, ingress := range sg.IpPermissions {
				if (ingress.FromPort != nil && *ingress.FromPort == 80 && *ingress.ToPort == 80) ||
					(ingress.ToPort != nil && *ingress.FromPort == 443 && *ingress.ToPort == 443) {
					continue
				}
				if len(ingress.IpRanges) > 0 {
					for _, ipRange := range ingress.IpRanges {
						if *ipRange.CidrIp == "0.0.0.0/0" {

							var fromPort, toPort int64
							if ingress.FromPort != nil {
								fromPort = *ingress.FromPort
							}

							if ingress.ToPort != nil {
								toPort = *ingress.ToPort
							}

							log.Printf(`Exposed security group. ID: %s , Name: %s, FromPort: %d, ToPort: %d`, *sg.GroupId, *sg.GroupName, fromPort, toPort)

							strFromPort := strconv.FormatInt(int64(fromPort), 10)
							strToPort := strconv.FormatInt(int64(toPort), 10)

							webhookUrl := os.Getenv("SLACK_WEBHOOK")

							attachment1 := slack.Attachment{}
							attachment1.AddField(slack.Field{Title: "SG Name:", Value: *sg.GroupName}).AddField(slack.Field{Title: "FromPort", Value: strFromPort}).AddField(slack.Field{Title: "ToPort", Value: strToPort}).AddField(slack.Field{Title: "Allow CIDR block:", Value: "0.0.0.0/0"})
							attachment1.AddAction(slack.Action{Type: "button", Text: "AWS console login", Url: "https://aws.amazon.com/pt/console/", Style: "primary"})
							payload := slack.Payload{
								Text:        "Unsafe rule ingress detected for sg: " + *sg.GroupId,
								Username:    "mordor-eye",
								Channel:     "C000000012W",
								Attachments: []slack.Attachment{attachment1},
							}
							err := slack.Send(webhookUrl, "", payload)
							if len(err) > 0 {
								fmt.Printf("error: %s\n", err)
							}
							if err != nil {
								log.Fatal(err)
							}
						}
					}
				}
			}
		}

		if descSGOut.NextToken != nil {
			nextToken = *descSGOut.NextToken
		} else {
			break
		}
	}
}

func main() {
	lambda.Start(handler)
}
