# Egocli

Egocli is a command line tool that helps you to simplify AWS Elastic Contanier Service deploy and provisioning.
With Egocli you can create or update a Task Definition and an ECS Service from a yaml file.
You can also provide different sets of parameters, grouped by context, so that you can use the same yaml to provison your infrastructure in different environments (ie. staging, qa, productions).
With **egocli seal** command you can encrypt your secrets and securely push them on your repository, egocli will decrypt them for you while running.

## Install

### From source

Download and install go: https://golang.org/doc/install

Download dependencies:
```bash
$ go get github.com/aws/aws-sdk-go/aws
$ go get gopkg.in/yaml.v2
```
Download code in your gopath:
```
$ git clone https://github.com/Ingordigia/egocli
```
Change directory:
```
$ cd $GOPATH/github.com/Ingordigia/egocli.git
```
Build:
```bash
$ go build cmd/egocli/egocli.go
```
### Options
```bash
 -context string
        Optional context.
  -project-file string
        A YAML file describing your ecs infrastructure (default "./ego.yml")
  -seal-password string
        Optional password used to decrypt secrets.
  -service-behavior string
        Possible choices: {none|create|update} (default "none")
```
### Usage

```bash
$ egocli --project-file ./ego.yml --contex production --seal-password myVeryHardPassword --service-behavior create
```

#### AWS Authentication

AWS authentication can be done using an AWS profile or using  AWS_SECRET_KEY and AWS_SECRET_ACCESS_KEY.
The preferred mode is the AWS profile, if you provide either profile and AWS secret, only the profile will be used.
Refer to AWS documentation for credential profile configuration: https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html
ie. using aws profile
```yaml
---
name: my project
region: eu-west-1
profile: my_aws_profile
```

ie. using aws AWS_SECRET_KEY and AWS_SECRET_ACCESS_KEY.
```yaml
---
name: my project
aws_access_key: my_access_key
aws_secret_acces_key: my_secret_access_key
region: eu-west-1
```

To keep secrets you can encrypt your AWS_SECRET_ACCESS_KEY using `egocli seal`:

```yaml
---
name: my project
aws_access_key: my_access_key
aws_secret_acces_key: !seal REhstVp5RLv_CIc76_5AaJEGNASSdgfIsw5CWA==
region: eu-west-1
```

If you plan to deploy in different environments you can use `egocli seal` in combination with context:

```yaml
---
name: my project
aws_access_key: my_access_key
aws_secret_acces_key: {{ .AWS_SECRET_ACCESS_KEY }}
region: eu-west-1
...
context:
  pro:
    - key: AWS_SECRET_ACCESS_KEY
      value: !seal REhstVp5RLv_CIc76_5AaJEGNASSdgfIsw5CWA==
  pre:
    - key: AWS_SECRET_ACCESS_KEY
      value: !seal DcmvEaa2QkrvB6fwd50CV0wgqTpYIQfvUJSPtoVNG91O
```

```sh
$ ./egocli --project-file ego.yml --context pre --seal-password 1234567812345678
```

#### Templates

**egocli** uses go templates. Data evaluations are delimited by "{{" "}}" and replaced evaluating environment varibles or key-value provided in egocli context.

i.e. if you need to specify your fresh built docker image in your project file, you can:

```sh
# Create a new environment variable with your image tag.
$ export IMAGETAG=myrepo:mytag
$ docke build . -t $IMAGETAG
```

```yaml
# Use template placeholder in your egocli porject
---
...
taskdefinition:
  ...
  containerdefinitions:
    - name: awesome image
      image: {{ .IMAGETAG }}
...
```

#### Using context

You can manage multiple environments using **templates** and **context**.
Context lets you define a list of key/value items that will replace your template placeholder on the run.
ie. your container definition has some environment varibles that must change according to the contest where your application has to be deployed (ie database cretental). you can define different context with different values and let the **egocli** use it in your template.

```yaml
taskdefinition:
  ...
  containerdefinitions:
    - name: awesome image
      ...
      environment:
        - name: username
          value : {{ .dbUserName }}
        - name: password
          value : {{ .dbPassword }}
...
context:
  staging:
    - key: dbUserName
      value: myStagingSecret
    - key: dbPassword
      value: myStagingPassword
  production:
    - key: dbUserName
      value: myProductionSecret
    - key: dbPassword
      value: myProductionPassword
```
When you run **egocli** using --context your template will be filled with the correct values.

```sh
$ ./egocli --project-file ego.yml --context production
```

#### Create or Update ECS Services

You can add your service definition in your yml file and use the `--service-behavior` parameter to create or update your service.
By default, if you don't provide a task definition, the one just created will be used.
If you don't use the `--service-behavior` parameter, no action will be taken.

#### Encrypt secrets with egocli seal

You can encrypt your secrets using `egocli seal` command and then add them to your project file:
**The seal password must be 16, 24, or 32 bytes.**

```sh
$ ./egocli seal --secret MySuperSecret --password S!ecretHash@----
!seal CT7hLIw-AC_mTafLepud18ZKelLTNru-TScZ9VQ=
```

```yaml
---
taskdefinition:
  containerdefinition:
    - name: my definition
      ...
      environment:
        - key: DatabasePassword
          value: !seal CT7hLIw-AC_mTafLepud18ZKelLTNru-TScZ9VQ=
```

When you run your project providing the seal password, `egocli` decrypt your secrets:

```sh
$ ./egocli --project-file ego.yml --context pre --seal-password S!ecretHash@----
```

**egocli seal** provide also a default secret key, use it only while testing.

#### Decrypt secrets with egocli unseal

Sometimes You just need to recover a secret from an old and forgotten project file, in this case you can use `egocli unseal` command to recover it:
**As for the seal password, also the unseal password must be 16, 24, or 32 bytes.**

```sh
$ ./egocli unseal --secret CT7hLIw-AC_mTafLepud18ZKelLTNru-TScZ9VQ= --password S!ecretHash@----
MySuperSecret
```