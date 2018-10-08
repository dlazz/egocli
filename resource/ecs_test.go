package resource

import (
	"testing"
)

func TestCheckService(t *testing.T) {
	// Test an existing service
	expected := true
	serviceName := "pre-mapi-authserver-service"
	clusterName := "pre-ecs-cluster-01"
	p := Project{
		Profile: "premailup-super_provisioning",
		Region:  "eu-west-1",
	}
	p.Service.ServiceName = &serviceName
	p.Service.Cluster = &clusterName
	p.newClient()

	check, err := p.checkService()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if check != expected {
		t.Error("Service not found")
		t.Fail()
	}

	// Test a missing service
	expected = false
	fakeName := "pre-mapi-fake-service"
	p.Service.ServiceName = &fakeName

	check, err = p.checkService()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if check != expected {
		t.Errorf("Service %s  found", fakeName)
		t.Fail()
	}
}
