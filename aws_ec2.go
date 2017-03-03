package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
)

type awsEc2 interface {
	terminateInstances(*ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)
}

type awsEc2Client struct {
	session *ec2.EC2
}

type awsEc2Controller struct {
	client  awsEc2
	filters []*ec2.Filter
	dryRun  bool
}

func newAWSEc2Client() awsEc2 {
	return &awsEc2Client{
		session: ec2.New(session.New()),
	}
}

func newAWSEc2Controller(awsEc2Client awsEc2, dryRyn bool) *awsEc2Controller {
	return &awsEc2Controller{
		client: awsEc2Client,
		dryRun: dryRun,
	}
}

func (e awsEc2Client) terminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	return e.session.TerminateInstances(input)
}

func (c *awsEc2Controller) terminateInstance(instance string) (*ec2.TerminateInstancesOutput, error) {
	var resp *ec2.TerminateInstancesOutput
	var err error

	glog.V(4).Infof("Terminating instance %s\n", instance)

	params := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instance),
		},
		DryRun: aws.Bool(c.dryRun),
	}
	resp, err = c.client.terminateInstances(params)
	return resp, err
}
