package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/Ingordigia/egocli/crypto"
	"github.com/Ingordigia/egocli/resource"
	yaml "gopkg.in/yaml.v2"
)

// LoadProject loads project configuration from a yaml file
func LoadProject(c *resource.Project, path, context *string) error {
	envs := getEnv()
	buff, err := parseConfigTemplate(path, envs, &c.SealPassword)
	if err != nil {
		return fmt.Errorf("Error parsing template: %s", err.Error())
	}
	// if a context has been provided it is parsed
	if *context != "" {
		ctx := resource.Context{}

		if err = yaml.Unmarshal(buff, &ctx); err != nil {
			return fmt.Errorf("Error parsing template: %s", err.Error())
		}

		if _, ok := ctx.Context[*context]; ok {
			envs = buildContext(&ctx, *context)
			if buff, err = parseConfigTemplate(path, envs, &c.SealPassword); err != nil {
				return fmt.Errorf("Error parsing template: %s", err.Error())
			}
		} else {
			return fmt.Errorf("Unable to find context '%s'", *context)
		}
	}
	yaml.Unmarshal(buff, &c)
	return nil
}

//parseConfigTemplate parse a file and return a slice of bytes and an error
// c is a pointer to a string pointing to a path file
func parseConfigTemplate(c *string, envs map[string]string, password *string) ([]byte, error) {

	var buf bytes.Buffer
	templatePath := path.Base(*c)

	t := template.New(templatePath)
	t.ParseFiles(*c)

	if err := t.Execute(&buf, envs); err != nil {
		return buf.Bytes(), fmt.Errorf("There was an error parsing your template file: %s", err)
	}

	// Searching and decrypting secrets
	byteConfig := buf.Bytes()
	decryptSecrets(&byteConfig, password)
	//fmt.Println(string(byteConfig))
	return byteConfig, nil
}

//GetEnv return a map of all environment vars
func getEnv() map[string]string {
	v := make(map[string]string)
	for _, e := range os.Environ() {
		if e != "" {
			s := strings.Split(e, "=")
			v[s[0]] = s[1]
		}
	}
	return v
}

func prepareContext(context *[]resource.KeyValue) (out map[string]string, err error) {
	if len(*context) < 1 {
		return nil, fmt.Errorf("Argument cannot be empty")
	}
	out = make(map[string]string)
	for _, item := range *context {
		if _, ok := out[item.Key]; ok {
			log.Fatalf("Duplicated entry: key %s already used", item.Key)
		}
		out[item.Key] = item.Value
	}
	return out, nil
}

func buildContext(c *resource.Context, context string) map[string]string {
	in := c.Context[context]
	env := getEnv()
	out, err := prepareContext(&in)
	if err != nil {
		log.Fatalf("Error parsing context: %s", err.Error())
	}
	for k, v := range out {
		if _, ok := env[k]; ok {
			log.Fatalf("Duplicated entry: key %s already used", env[k])
		}
		env[k] = v
	}
	return env
}

// checks template lines and eventually replaces secrets with decrypted strings
func decryptSecrets(b *[]byte, password *string) error {

	newline := []byte("\n")
	tLines := bytes.SplitAfter(*b, newline)

	for i, v := range tLines {
		if strings.Contains(string(v), crypto.TAG) {
			var out string
			splitted := strings.Split(string(v), crypto.TAG)
			secret := strings.TrimSpace(splitted[1])

			// decrypt and replace
			s := crypto.Secret{}
			if err := s.Decrypt(&secret, &out, password); err != nil {
				return err
			}
			splitted[1] = out
			splitted = append(splitted, "\n")

			//join and riconvert to bytes
			joined := strings.Join(splitted, "")
			tLines[i] = []byte(joined)
		}
	}
	*b = bytes.Join(tLines, []byte(""))
	return nil
}
