package k8snode

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type mockAWSEc2Client struct {
	ec2iface.EC2API
	ThrowError error
}

func (c mockAWSEc2Client) TerminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	result := &ec2.TerminateInstancesOutput{}
	for _, instance := range input.InstanceIds {
		inst := &ec2.InstanceStateChange{
			InstanceId:    instance,
			CurrentState:  &ec2.InstanceState{Name: aws.String("shutting-down"), Code: aws.Int64(32)},
			PreviousState: &ec2.InstanceState{Name: aws.String("running"), Code: aws.Int64(16)},
		}
		result.TerminatingInstances = append(result.TerminatingInstances, inst)
	}
	return result, c.ThrowError
}

func TestTerminateInstances(t *testing.T) {
	tests := []struct {
		ExpectedError error
		InstanceID    string
	}{
		{ExpectedError: nil, InstanceID: "i-0ed48177c77a0acfb"},
		{ExpectedError: errors.New("Dummy error"), InstanceID: "i-0ed48177c77a0acfb"},
	}
	for _, tc := range tests {
		mockSvc := mockAWSEc2Client{ThrowError: tc.ExpectedError}
		client := AWSEc2Controller{client: mockSvc, dryRun: false}

		if err := client.TerminateInstance(tc.InstanceID); err != tc.ExpectedError {
			t.Errorf("TerminateInstance returned: %s\nExpecting: %s", err, tc.ExpectedError)
		}
	}

}
