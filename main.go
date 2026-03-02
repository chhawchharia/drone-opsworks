package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/opsworks"
	owTypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin"
)

var (
	version = "unknown"
)

func main() {
	fmt.Printf("Drone AWS OpsWorks Plugin built from %s\n", version)

	repo := drone.Repo{}
	build := drone.Build{}
	vargs := Params{}

	plugin.Param("repo", &repo)
	plugin.Param("build", &build)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	if vargs.AccessKey == "" {
		fmt.Println("Please provide an access key id")
		os.Exit(1)
	}

	if vargs.SecretKey == "" {
		fmt.Println("Please provide a secret access key")
		os.Exit(1)
	}

	if aws.ToString(vargs.Region) == "" {
		fmt.Println("Please provide a region")
		os.Exit(1)
	}

	if aws.ToString(vargs.StackID) == "" {
		fmt.Println("Please provide a stack ID")
		os.Exit(1)
	}

	if aws.ToString(vargs.Command) == "" {
		fmt.Println("Please provide a deploy command")
		os.Exit(1)
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(aws.ToString(vargs.Region)),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				vargs.AccessKey,
				vargs.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		fmt.Printf("unable to load AWS config: %v\n", err)
		os.Exit(1)
	}

	svc := opsworks.NewFromConfig(cfg)

	args := make(map[string][]string, len(vargs.Arguments))
	for k, v := range vargs.Arguments {
		vals := make([]string, 0, len(v))
		for _, s := range v {
			if s != nil {
				vals = append(vals, *s)
			}
		}
		args[k] = vals
	}

	instanceIDs := make([]string, 0, len(vargs.Instances))
	for _, s := range vargs.Instances {
		if s != nil {
			instanceIDs = append(instanceIDs, *s)
		}
	}

	_, err = svc.CreateDeployment(ctx,
		&opsworks.CreateDeploymentInput{
			Command: &owTypes.DeploymentCommand{
				Name: owTypes.DeploymentCommandName(aws.ToString(vargs.Command)),
				Args: args,
			},
			StackId:     vargs.StackID,
			AppId:       vargs.AppID,
			Comment:     vargs.Comment,
			CustomJson:  vargs.CustomJSON,
			InstanceIds: instanceIDs,
		},
	)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Successfully deployed")
}
