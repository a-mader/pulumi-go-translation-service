package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/secretmanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		SECRET_KEY := os.Getenv("SECRET_KEY")
		if SECRET_KEY == "" {
			panic("SECRET_KEY is missing")
		}

		const prefix = "ama-"
		// Import the program's configuration settings.
		cfg := config.New(ctx, "")

		containerPort := 3000
		if param := cfg.GetInt("containerPort"); param != 0 {
			containerPort = param
		}
		cpu := 1
		if param := cfg.GetInt("cpu"); param != 0 {
			cpu = param
		}
		memory := "1Gi"
		if param := cfg.Get("memory"); param != "" {
			memory = param
		}
		concurrency := 5
		if param := cfg.GetInt("concurrency"); param != 0 {
			concurrency = param
		}

		// Import the provider's configuration settings.
		providerConfig := config.New(ctx, "gcp")
		location := providerConfig.Require("region")

		image, err := docker.NewRemoteImage(ctx, "app", &docker.RemoteImageArgs{
			Name: pulumi.String("amader/pulumi-go-translation-service:v0.2.0"),
		})
		if err != nil {
			return err
		}

		secret, err := secretmanager.LookupSecretVersion(ctx, &secretmanager.LookupSecretVersionArgs{
			Secret: "ama-my-secret",
		})
		if err != nil {
			return err
		}

		// Create a Cloud Run service definition.
		service, err := cloudrun.NewService(ctx, prefix+"service", &cloudrun.ServiceArgs{
			Location: pulumi.String(location),
			Template: cloudrun.ServiceTemplateArgs{
				Spec: cloudrun.ServiceTemplateSpecArgs{
					Containers: cloudrun.ServiceTemplateSpecContainerArray{
						cloudrun.ServiceTemplateSpecContainerArgs{
							Image: image.RepoDigest.ToStringOutput(),
							Resources: cloudrun.ServiceTemplateSpecContainerResourcesArgs{
								Limits: pulumi.ToStringMap(map[string]string{
									"memory": memory,
									"cpu":    strconv.Itoa(cpu),
								}),
							},
							Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
								cloudrun.ServiceTemplateSpecContainerPortArgs{
									ContainerPort: pulumi.Int(containerPort),
								},
							},
							Envs: cloudrun.ServiceTemplateSpecContainerEnvArray{
								&cloudrun.ServiceTemplateSpecContainerEnvArgs{
									Name: pulumi.String("SECRET_KEY"),
									Value: pulumi.String(fmt.Sprintf("projects/%s/secrets/%s/versions/%s",
										secret.Project, secret.Secret, secret.Version)),
								},
							},
						},
					},
					ContainerConcurrency: pulumi.Int(concurrency),
				},
			},
		})
		if err != nil {
			return err
		}

		_, err = cloudrun.NewIamMember(ctx, "translateIam", &cloudrun.IamMemberArgs{
			Service:  service.Name,
			Location: pulumi.String(location),
			Role:     pulumi.String("roles/cloudtranslate.user"),
			Member:   pulumi.String("allUsers"),
		})
		if err != nil {
			return err
		}

		// Create an IAM member to make the service publicly accessible.
		_, err = cloudrun.NewIamMember(ctx, "invokerIam", &cloudrun.IamMemberArgs{
			Service:  service.Name,
			Location: pulumi.String(location),
			Role:     pulumi.String("roles/run.invoker"),
			Member:   pulumi.String("allUsers"),
		})
		if err != nil {
			return err
		}

		// Export the URL of the service.
		ctx.Export("url", service.Statuses.Index(pulumi.Int(0)).Url())

		return nil
	})
}
