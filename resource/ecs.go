package resource

import (
	"fmt"
	"log"

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
		return fmt.Errorf("Ecs client creation failed")
	}
	// Registers the Task Definition
	out, err := p.Client.RegisterTaskDefinition(&p.TaskDefinition)
	if err != nil {
		return fmt.Errorf("Error Registering Task Definition: %s", err.Error())
	}

	taskDefinition := fmt.Sprintf("%s:%d", *out.TaskDefinition.Family, *out.TaskDefinition.Revision)
	log.Println("Registered Task definition:", taskDefinition)

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
			log.Printf("Service %s already exists, cannot be created", *p.Service.ServiceName)
			return nil
		}
		if err := p.createService(); err != nil {
			return err
		}
	case "update":
		check, err = p.checkService()
		if err != nil {
			return err
		}
		if !check {
			log.Printf("Service %s doesn't exist, cannot update", *p.Service.ServiceName)
			return nil
		}
		if err = p.updateService(); err != nil {
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
		return false, fmt.Errorf("Cluster Name missing. 'cluster' field is manadatory")
	}
	if p.Service.ServiceName == nil {
		if p.Service.ServiceName == nil {
			return false, fmt.Errorf("Service name missing. 'servicename' field is manadatory")
		}
	}

	services = append(services, p.Service.ServiceName)
	input := ecs.DescribeServicesInput{
		Cluster:  p.Service.Cluster,
		Services: services,
	}
	out, err := p.Client.DescribeServices(&input)
	if err != nil {
		return false, fmt.Errorf("Error searching service: %s", err.Error())
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

	out := &ecs.UpdateServiceOutput{}
	out, err = p.Client.UpdateService(&serivceInput)

	if err != nil {
		return fmt.Errorf("Error updating Service: %s", err.Error())
	}
	log.Println("Service", *out.Service.ServiceName, "successfully updated")
	log.Println("Service Status:", *out.Service.Status)
	return nil
}

func (p *Project) createService() (err error) {
	out := &ecs.CreateServiceOutput{}
	out, err = p.Client.CreateService(&p.Service)

	if err != nil {
		return fmt.Errorf("Error updating Service: %s", err.Error())
	}
	log.Println("Service", *out.Service.ServiceName, "successfully created")
	return nil
}
