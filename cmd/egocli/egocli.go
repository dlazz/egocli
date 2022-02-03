package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dlazz/egocli/config"
	"github.com/dlazz/egocli/crypto"
	"github.com/dlazz/egocli/resource"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger.Output(os.Stdout)

	if !JSONlogger {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	if len(os.Args) < 2 {
		flag.PrintDefaults()

		fmt.Printf("[seal]\n")
		seal.PrintDefaults()

		fmt.Printf("[unseal]\n")
		unseal.PrintDefaults()

		os.Exit(0)
	}
	if password == "" {
		password = crypto.DEFAULT_PASSWORD
	}
	switch os.Args[1] {

	case "seal":
		seal.Parse(os.Args[2:])
		if secret == "" {
			seal.PrintDefaults()
			os.Exit(1)
		}
		var out string
		s := crypto.Secret{}
		if err := s.Encrypt(&secret, &out, &password); err != nil {
			log.Error().Err(err).Msg("error encrypting string:")
			os.Exit(1)
		}
		fmt.Println(out)

	case "unseal":
		unseal.Parse(os.Args[2:])
		if secret == "" {
			unseal.PrintDefaults()
			os.Exit(1)
		}
		var out string
		s := crypto.Secret{}
		if err := s.Decrypt(&secret, &out, &password); err != nil {
			log.Error().Err(err).Msg("error decrypting string")
			os.Exit(1)
		}
		fmt.Println(out)

	default:
		serviceChoices := map[string]bool{"create": true, "update": true, "none": true}
		if _, validChoice := serviceChoices[serviceBehavior]; !validChoice {
			flag.PrintDefaults()
			os.Exit(1)
		}
		if _, err := os.Stat(projectFile); err != nil {
			log.Error().Err(err).Msg("project file not found")
			os.Exit(1)
		}

		project := resource.Project{}
		project.SealPassword = password
		project.ServiceBehavior = serviceBehavior

		if err := config.LoadProject(&project, &projectFile, &context); err != nil {
			log.Error().Err(err).Msg("")
			os.Exit(1)
		}

		if err := project.Run(); err != nil {
			log.Error().Err(err).Msg("")
			os.Exit(1)
		}
	}
}

var context, password, projectFile, serviceBehavior string

var seal *flag.FlagSet
var unseal *flag.FlagSet
var secret string
var JSONlogger bool

func init() {
	seal = flag.NewFlagSet("seal", flag.ExitOnError)
	seal.StringVar(&secret, "secret", "", "Secret to encrypt")
	seal.StringVar(&password, "password", "", "Password to use for secret encryption")
	unseal = flag.NewFlagSet("unseal", flag.ExitOnError)
	unseal.StringVar(&secret, "secret", "", "Encrypted secret to decrypt")
	unseal.StringVar(&password, "password", "", "Password used for secret encryption")
	flag.StringVar(&context, "context", "", "Optional context.")
	flag.StringVar(&projectFile, "project-file", "./ego.yml", "A YAML file describing your ecs infrastructure")
	flag.StringVar(&serviceBehavior, "service-behavior", "none", "Possible choices: {none|create|update}")
	flag.StringVar(&password, "seal-password", "", "Optional password used to decrypt secrets.")
	flag.BoolVar(&JSONlogger, "log-to-json", false, "log output in json format")

}
