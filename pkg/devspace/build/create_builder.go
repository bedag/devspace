package build

import (
	"context"

	"github.com/devspace-cloud/devspace/pkg/devspace/builder"
	"github.com/devspace-cloud/devspace/pkg/devspace/builder/custom"
	"github.com/devspace-cloud/devspace/pkg/devspace/builder/docker"
	"github.com/devspace-cloud/devspace/pkg/devspace/builder/kaniko"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	dockerclient "github.com/devspace-cloud/devspace/pkg/devspace/docker"
	"github.com/devspace-cloud/devspace/pkg/devspace/kubectl"
	"github.com/devspace-cloud/devspace/pkg/util/kubeconfig"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/devspace-cloud/devspace/pkg/util/ptr"
	"github.com/pkg/errors"
)

// CreateBuilder creates a new builder
func CreateBuilder(config *latest.Config, client *kubectl.Client, imageConfigName string, imageConf *latest.ImageConfig, imageTag string, skipPush, isDev bool, log log.Logger) (builder.Interface, error) {
	var (
		imageBuilder builder.Interface
		err          error
	)

	if imageConf.Build != nil && imageConf.Build.Custom != nil {
		imageBuilder = custom.NewBuilder(imageConfigName, imageConf, imageTag)
	} else if imageConf.Build != nil && imageConf.Build.Kaniko != nil {
		dockerClient, err := dockerclient.NewClient(log)
		if err != nil {
			return nil, errors.Errorf("Error creating docker client: %v", err)
		}
		if client == nil {
			// Create kubectl client if not specified
			client, err = kubectl.NewDefaultClient()
			if err != nil {
				return nil, errors.Errorf("Unable to create new kubectl client: %v", err)
			}
		}

		log.StartWait("Creating kaniko builder")
		defer log.StopWait()
		imageBuilder, err = kaniko.NewBuilder(config, dockerClient, client, imageConfigName, imageConf, imageTag, isDev, log)
		if err != nil {
			return nil, errors.Errorf("Error creating kaniko builder: %v", err)
		}
	} else {
		preferMinikube := true
		if imageConf.Build != nil && imageConf.Build.Docker != nil && imageConf.Build.Docker.PreferMinikube != nil {
			preferMinikube = *imageConf.Build.Docker.PreferMinikube
		}

		kubeContext := ""
		if client == nil {
			kubeContext, err = kubeconfig.GetCurrentContext()
			if err != nil {
				return nil, errors.Wrap(err, "get current context")
			}
		} else {
			kubeContext = client.CurrentContext
		}

		dockerClient, err := dockerclient.NewClientWithMinikube(kubeContext, preferMinikube, log)
		if err != nil {
			return nil, errors.Errorf("Error creating docker client: %v", err)
		}

		// Check if docker daemon is running
		_, err = dockerClient.Ping(context.Background())
		if err != nil {
			if imageConf.Build != nil && imageConf.Build.Docker != nil && imageConf.Build.Docker.DisableFallback != nil && *imageConf.Build.Docker.DisableFallback {
				return nil, errors.Errorf("Couldn't reach docker daemon: %v. Is the docker daemon running?", err)
			}

			// Fallback to kaniko
			log.Infof("Couldn't find a running docker daemon. Will fallback to kaniko")
			return CreateBuilder(config, client, imageConfigName, convertDockerConfigToKanikoConfig(imageConf), imageTag, skipPush, isDev, log)
		}

		imageBuilder, err = docker.NewBuilder(config, dockerClient, client, imageConfigName, imageConf, imageTag, skipPush, isDev)
		if err != nil {
			return nil, errors.Errorf("Error creating docker builder: %v", err)
		}
	}

	return imageBuilder, nil
}

func convertDockerConfigToKanikoConfig(dockerConfig *latest.ImageConfig) *latest.ImageConfig {
	kanikoConfig := &latest.ImageConfig{
		Image:            dockerConfig.Image,
		Tag:              dockerConfig.Tag,
		Dockerfile:       dockerConfig.Dockerfile,
		Context:          dockerConfig.Context,
		CreatePullSecret: dockerConfig.CreatePullSecret,
		Build: &latest.BuildConfig{
			Kaniko: &latest.KanikoConfig{
				Cache: ptr.Bool(true),
			},
		},
	}

	if dockerConfig.Build != nil && dockerConfig.Build.Docker != nil && dockerConfig.Build.Docker.Options != nil {
		kanikoConfig.Build.Kaniko.Options = dockerConfig.Build.Docker.Options
	}

	return kanikoConfig
}
