package resource

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// Project ...
type Project struct {
	Name               string                          `yaml:"name"`
	Region             string                          `yaml:"region"`
	Profile            string                          `yaml:"profile"`
	AwsAccesKey        string                          `yaml:"aws_access_key"`
	AwsSecretAccessKey string                          `yaml:"aws_secret_acces_key"`
	TaskDefinition     ecs.RegisterTaskDefinitionInput `yaml:"taskdefinition"`
	Service            ecs.CreateServiceInput          `yaml:"service"`
	ServiceBehavior    string
	Client             *ecs.ECS
	SealPassword       string
}

// Context struct
type Context struct {
	Context map[string][]KeyValue `yaml:"context"`
}

// KeyValue struct for context
type KeyValue struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// Run project
func (p *Project) Run() error {

	var check bool

	// Initializes a new ECS client
	p.newClient()
	if p.Client == nil {
		return fmt.Errorf("ecs client creation failed")
	}
	// Registers the Task Definition
	out, err := p.Client.RegisterTaskDefinition(&p.TaskDefinition)
	if err != nil {
		return fmt.Errorf("error Registering Task Definition: %s", err.Error())
	}

	taskDefinition := fmt.Sprintf("%s:%d", *out.TaskDefinition.Family, *out.TaskDefinition.Revision)
	log.Info().Str("task_definition", taskDefinition).Str("task_definition_arn", *out.TaskDefinition.TaskDefinitionArn).Msg("")

	if p.Service.TaskDefinition == nil {
		p.Service.TaskDefinition = &taskDefinition
	}

	switch p.ServiceBehavior {
	case "create":
		check, err = p.checkService()
		if err != nil {
			return err
		}
		if check {
			log.Warn().Msg(fmt.Sprintf("service %s already exists, cannot be created", *p.Service.ServiceName))
			return nil
		}
		if err := p.createService(); err != nil {
			return err
		}
		if err := p.waitDeployment(); err != nil {
			return err
		}
	case "update":
		check, err = p.checkService()
		if err != nil {
			return err
		}
		if !check {
			log.Warn().Msg(fmt.Sprintf("service %s doesn't exist, cannot update", *p.Service.ServiceName))
			return nil
		}
		if err = p.updateService(); err != nil {
			return err
		}
		if err := p.waitDeployment(); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

// NewClient creates a new ECS client from configuration
func (p *Project) newClient() {
	sess := session.Must(session.NewSession())
	if p.Profile != "" {
		cred := credentials.NewSharedCredentials("", p.Profile)
		ecsSession := ecs.New(sess, &aws.Config{Credentials: cred,
			Region: &p.Region})
		p.Client = ecsSession
	} else if p.AwsAccesKey != "" && p.AwsSecretAccessKey != "" {
		cred := credentials.NewStaticCredentials(p.AwsAccesKey, p.AwsSecretAccessKey, "")
		ecsSession := ecs.New(sess, &aws.Config{Credentials: cred,
			Region: &p.Region,
		})
		p.Client = ecsSession
	}
}

// CheckService checks if a service already exist
func (p *Project) checkService() (bool, error) {
	var services []*string

	if p.Service.Cluster == nil {
		return false, fmt.Errorf("cluster name missing. 'cluster' field is manadatory")
	}
	if p.Service.ServiceName == nil {
		if p.Service.ServiceName == nil {
			return false, fmt.Errorf("service name missing. 'servicename' field is manadatory")
		}
	}

	services = append(services, p.Service.ServiceName)
	input := ecs.DescribeServicesInput{
		Cluster:  p.Service.Cluster,
		Services: services,
	}
	out, err := p.Client.DescribeServices(&input)
	if err != nil {
		return false, fmt.Errorf("error searching service: %s", err.Error())
	}
	for _, v := range out.Services {
		if *v.ServiceName == *p.Service.ServiceName {
			return true, nil
		}

	}
	return false, nil
}

func (p *Project) updateService() (err error) {
	serivceInput := ecs.UpdateServiceInput{}
	serivceInput.Cluster = p.Service.Cluster
	serivceInput.DeploymentConfiguration = p.Service.DeploymentConfiguration
	serivceInput.DesiredCount = p.Service.DesiredCount
	serivceInput.HealthCheckGracePeriodSeconds = p.Service.HealthCheckGracePeriodSeconds
	serivceInput.NetworkConfiguration = p.Service.NetworkConfiguration
	serivceInput.PlatformVersion = p.Service.PlatformVersion
	serivceInput.Service = p.Service.ServiceName
	serivceInput.TaskDefinition = p.Service.TaskDefinition

	out, err := p.Client.UpdateService(&serivceInput)

	if err != nil {
		return fmt.Errorf("error updating Service: %s", err.Error())
	}
	log.Info().Str("service", *out.Service.ServiceName).Str("status", *out.Service.Status).Str("action", "updated").Msg("")

	return nil
}

func (p *Project) createService() (err error) {

	out, err := p.Client.CreateService(&p.Service)

	if err != nil {
		return fmt.Errorf("error updating Service: %s", err.Error())
	}
	log.Info().Str("service", *out.Service.ServiceName).Str("status", *out.Service.Status).Str("action", "created").Msg("")
	return nil
}

func (p *Project) waitDeployment() (err error) {
	log.Info().Msg("wait until service is stable")
	input := &ecs.DescribeServicesInput{
		Cluster:  p.Service.Cluster,
		Services: []*string{p.Service.ServiceName},
	}
	err = p.Client.WaitUntilServicesStable(input)
	if err != nil {
		return fmt.Errorf("deployment failed: %s", err.Error())
	}
	log.Info().Msg("service is stable")
	return nil
}
