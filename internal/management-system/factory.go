package managementsystem

import (
	"appointment-system/internal/domain"
	"appointment-system/internal/management-system/flowww"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type ManagementSystemFunctionName string

const (
	FlowwwManagementSystemFunctionName ManagementSystemFunctionName = "flowww-operations"
)

type ManagementSystemFactory struct {
	lambdaClient *lambda.Client
}

func NewManagementSystemFactory(
	lambdaClient *lambda.Client,
) *ManagementSystemFactory {
	return &ManagementSystemFactory{
		lambdaClient: lambdaClient,
	}
}

func (f *ManagementSystemFactory) Build(managementSystem string) domain.AppointmentSystem {
	switch managementSystem {
	case "flowww":
		return f.buildFlowwwManagementSystem()
	default:
		return nil
	}
}

func (f *ManagementSystemFactory) buildFlowwwManagementSystem() *flowww.FlowwwSystem {
	client := flowww.NewFlowwwClient(
		f.lambdaClient,
		string(FlowwwManagementSystemFunctionName),
	)

	mapper := flowww.NewFlowwwMapper()

	return flowww.NewFlowwwSystem(client, mapper)
}
