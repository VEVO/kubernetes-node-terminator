package k8snode

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/golang/glog"
)

type AWSEc2Controller struct {
	client  ec2iface.EC2API
	filters []*ec2.Filter
	dryRun  bool
}

func NewAWSEc2Controller(dryRun bool, region string) Provider {

	return &AWSEc2Controller{
		client: ec2.New(session.New(&aws.Config{Region: aws.String(region)})),
		dryRun: dryRun,
	}
}

func (c *AWSEc2Controller) TerminateInstance(instance string) error {
	glog.Infof("Terminating instance %s\n", instance)

	params := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instance),
		},
		DryRun: aws.Bool(c.dryRun),
	}
	_, err := c.client.TerminateInstances(params)
	return err
}
