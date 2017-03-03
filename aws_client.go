package main

type awsClient struct {
	ec2 *awsEc2Controller
}

func newAwsClient(dryRun bool) *awsClient {
	awsClient := &awsClient{
		ec2: newAWSEc2Controller(newAWSEc2Client(), dryRun),
	}
	return awsClient
}
