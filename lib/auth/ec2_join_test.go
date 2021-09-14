package auth

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gravitational/teleport/api/types"
	check "gopkg.in/check.v1"
)

const (
	testIID = `MIAGCSqGSIb3DQEHAqCAMIACAQExDzANBglghkgBZQMEAgEFADCABgkqhkiG9w0BBwGggCSABIIB
23sKICAiYWNjb3VudElkIiA6ICIyNzg1NzYyMjA0NTMiLAogICJhcmNoaXRlY3R1cmUiIDogIng4
Nl82NCIsCiAgImF2YWlsYWJpbGl0eVpvbmUiIDogInVzLXdlc3QtMmEiLAogICJiaWxsaW5nUHJv
ZHVjdHMiIDogbnVsbCwKICAiZGV2cGF5UHJvZHVjdENvZGVzIiA6IG51bGwsCiAgIm1hcmtldHBs
YWNlUHJvZHVjdENvZGVzIiA6IG51bGwsCiAgImltYWdlSWQiIDogImFtaS0wZmE5ZTFmNjQxNDJj
ZGUxNyIsCiAgImluc3RhbmNlSWQiIDogImktMDc4NTE3Y2E4YTcwYTFkZGUiLAogICJpbnN0YW5j
ZVR5cGUiIDogInQyLm1lZGl1bSIsCiAgImtlcm5lbElkIiA6IG51bGwsCiAgInBlbmRpbmdUaW1l
IiA6ICIyMDIxLTA5LTAzVDIxOjI1OjQ0WiIsCiAgInByaXZhdGVJcCIgOiAiMTAuMC4wLjIwOSIs
CiAgInJhbWRpc2tJZCIgOiBudWxsLAogICJyZWdpb24iIDogInVzLXdlc3QtMiIsCiAgInZlcnNp
b24iIDogIjIwMTctMDktMzAiCn0AAAAAAAAxggIvMIICKwIBATBpMFwxCzAJBgNVBAYTAlVTMRkw
FwYDVQQIExBXYXNoaW5ndG9uIFN0YXRlMRAwDgYDVQQHEwdTZWF0dGxlMSAwHgYDVQQKExdBbWF6
b24gV2ViIFNlcnZpY2VzIExMQwIJALZL3lrQCSTMMA0GCWCGSAFlAwQCAQUAoIGYMBgGCSqGSIb3
DQEJAzELBgkqhkiG9w0BBwEwHAYJKoZIhvcNAQkFMQ8XDTIxMDkwMzIxMjU0N1owLQYJKoZIhvcN
AQk0MSAwHjANBglghkgBZQMEAgEFAKENBgkqhkiG9w0BAQsFADAvBgkqhkiG9w0BCQQxIgQgCH2d
JiKmdx9uhxlm8ObWAvFOhqJb7k79+DW/T3ezwVUwDQYJKoZIhvcNAQELBQAEggEANWautigs/qZ6
w8g5/EfWsAFj8kHgUD+xqsQ1HDrBUx3IQ498NMBZ78379B8RBfuzeVjbaf+yugov0fYrDbGvSRRw
myy49TfZ9gdlpWQXzwSg3OPMDNToRoKw00/LQjSxcTCaPP4vMDEIjYMUqZ3i4uWYJJJ0Lb7fDMDk
Anu7yHolVfbnvIAuZe8lGpc7ofCSBG5wulm+/pqzO25YPMH1cLEvOadE+3N2GxK6gRTLJoE98rsm
LDp6OuU/b2QfaxU0ec6OogdtSJto/URI0/ygHmNAzBis470A29yh5nVwm6AkY4krjPsK7uiBIRhs
lr5x0X6+ggQfF2BKAJ/BRcAHNgAAAAAAAA==`
	testAWSAccount = "278576220453"
	testAWSRegion  = "us-west-2"
)

type ec2ClientNoInstance struct{}
type ec2ClientNotRunning struct{}
type ec2ClientRunning struct{}

func (c ec2ClientNoInstance) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{}, nil
}

func (c ec2ClientNotRunning) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []ec2types.Reservation{
			{
				Instances: []ec2types.Instance{
					{
						InstanceId: &params.InstanceIds[0],
						State: &ec2types.InstanceState{
							Name: ec2types.InstanceStateNameTerminated,
						},
					},
				},
			},
		},
	}, nil
}

func (c ec2ClientRunning) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []ec2types.Reservation{
			{
				Instances: []ec2types.Instance{
					{
						InstanceId: &params.InstanceIds[0],
						State: &ec2types.InstanceState{
							Name: ec2types.InstanceStateNameRunning,
						},
					},
				},
			},
		},
	}, nil
}

func (s *AuthSuite) TestSimplifiedNodeJoin(c *check.C) {
	/*
		token, err := types.NewProvisionTokenFromSpec(
			"test_token",
			time.Now().Add(time.Minute),
			types.ProvisionTokenSpecV2{
				Roles: []types.SystemRole{types.RoleNode},
				Allow: []*types.TokenRule{
					&types.TokenRule{
						AWSAccount: testAWSAccount,
						AWSRegions: []string{testAWSRegion},
						AWSRole:    "asdf",
					},
				},
			})
		c.Assert(err, check.IsNil)

		err = s.a.UpsertToken(context.Background(), token)
		c.Assert(err, check.IsNil)

		ctx := context.Background()

		ec2Client := ec2ClientRunning{}
		ctx = context.WithValue(ctx, ec2ClientKey{}, ec2Client)

		err = s.a.CheckEC2Request(ctx, RegisterUsingTokenRequest{
			Token:               "test_token",
			HostID:              "278576220453-i-078517ca8a70a1dde",
			NodeName:            "node_name",
			Role:                types.RoleNode,
			EC2IdentityDocument: []byte(testIID),
		})
		c.Assert(err, check.IsNil)
	*/

	testCases := []struct {
		desc       string
		tokenRules []*types.TokenRule
		ec2Client  ec2Client
		request    RegisterUsingTokenRequest
	}{
		{
			desc: "basic",
			tokenRules: []*types.TokenRule{
				&types.TokenRule{
					AWSAccount: testAWSAccount,
					AWSRegions: []string{testAWSRegion},
					AWSRole:    "asdf",
				},
			},
			ec2Client: ec2ClientRunning{},
			request: RegisterUsingTokenRequest{
				Token:               "test_token",
				HostID:              "278576220453-i-078517ca8a70a1dde",
				NodeName:            "node_name",
				Role:                types.RoleNode,
				EC2IdentityDocument: []byte(testIID),
			},
		},
	}
	for _, tc := range testCases {
		token, err := types.NewProvisionTokenFromSpec(
			"test_token",
			time.Now().Add(time.Minute),
			types.ProvisionTokenSpecV2{
				Roles: []types.SystemRole{types.RoleNode},
				Allow: tc.tokenRules,
			})
		c.Assert(err, check.IsNil)

		err = s.a.UpsertToken(context.Background(), token)
		c.Assert(err, check.IsNil)

		ctx := context.WithValue(context.Background(), ec2ClientKey{}, tc.ec2Client)

		err = s.a.CheckEC2Request(ctx, tc.request)
		c.Assert(err, check.IsNil)

		err = s.a.DeleteToken(context.Background(), token.GetName())
		c.Assert(err, check.IsNil)
	}
}
