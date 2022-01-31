package config

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"

	"github.com/dlazz/egocli/crypto"
	"github.com/dlazz/egocli/resource"
	yaml "gopkg.in/yaml.v2"
)

// LoadProject loads project configuration from a yaml file
func LoadProject(c *resource.Project, path, context *string) error {
	envs := getEnv()
	buff, err := parseConfigTemplate(path, envs, &c.SealPassword)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}
	// if a context has been provided it is parsed
	if *context != "" {
		ctx := resource.Context{}

		if err = yaml.Unmarshal(buff, &ctx); err != nil {
			return fmt.Errorf("error parsing template: %v", err)
		}

		if _, ok := ctx.Context[*context]; ok {
			envs = buildContext(&ctx, *context)
			if buff, err = parseConfigTemplate(path, envs, &c.SealPassword); err != nil {
				return fmt.Errorf("error parsing template: %v", err)
			}
		} else {
			return fmt.Errorf("unable to find context '%s'", *context)
		}
	}
	if err := yaml.Unmarshal(buff, &c); err != nil {
		return err
	}
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
		return buf.Bytes(), fmt.Errorf("there was an error parsing your template file: %s", err)
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
		return nil, fmt.Errorf("argument cannot be empty")
	}
	out = make(map[string]string)
	for _, item := range *context {
		if _, ok := out[item.Key]; ok {
			log.Error().Err(fmt.Errorf("duplicated entry: key %s already used", item.Key)).Msg("")
			os.Exit(1)
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
		log.Error().Err(err).Msg("error parsing context")
		os.Exit(1)
	}
	for k, v := range out {
		if _, ok := env[k]; ok {
			log.Error().Err(fmt.Errorf("duplicated entry: key %s already used", env[k])).Msg("")
			os.Exit(1)
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
