package main

import "testing"

func TestAwsClient(t *testing.T) {
	awsClient := newAwsClient(true)
	if awsClient.ec2 == nil {
		t.Failed()
	}
}
