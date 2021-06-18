package config

import (
	"os"
	"strings"
	"testing"

	"github.com/Ingordigia/egocli/resource"
)

func TestParseConfigTemplate(t *testing.T) {

}

func TestLoadConfiguration(t *testing.T) {
	context := "pro"
	expectedProfile := "my_aws_profile"
	path := "../examples/test.yml"
	project := resource.Project{}
	err := LoadProject(&project, &path, &context)
	if err != nil {
		t.Errorf("Error %s", err.Error())
		t.Fail()
	}

	if project.Profile != expectedProfile {
		t.Errorf("Failed. Expected %s, got %s", expectedProfile, project.Profile)
		t.Fail()
	}

}

func TestGetEnv(t *testing.T) {
	const (
		myvar = "MY_VAR"
	)
	expected := "value"
	os.Setenv(myvar, "value")
	env := getEnv()

	if env[myvar] != expected {
		t.Errorf("Failed. Expected `%s`, got %s", expected, env[myvar])
		t.Fail()
	}
}

func TestPrepareContext(t *testing.T) {
	var ctx []resource.KeyValue
	var k resource.KeyValue
	k.Key = "myKey"
	k.Value = "myValue"
	ctx = append(ctx, k)
	out, err := prepareContext(&ctx)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
		return
	}

	if out["myKey"] != ctx[0].Value {
		t.Errorf("Expected %s, got %s", ctx[0].Value, out["myKey"])
		t.Fail()
		return
	}
}

func TestDecryptSecrets(t *testing.T) {
	const (
		text = `- key: mykey
	              value: !seal REhstVp5RLv_CIc76_5AaJEGNASSdgfIsw5CWA==`
	)
	password := "myverysecretpwdx"
	expected := "segretissimo"
	byteText := []byte(text)
	if err := decryptSecrets(&byteText, &password); err != nil {
		t.Logf("Error: %s", err.Error())
		t.Fail()
	}

	if !strings.Contains(string(byteText), expected) {
		t.Logf("Strings do not match")
		t.Fail()
	}

}
