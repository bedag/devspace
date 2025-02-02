package services

import (
	"strconv"
	"strings"
	"time"

	"k8s.io/client-go/tools/portforward"

	"github.com/devspace-cloud/devspace/pkg/devspace/config/generated"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/devspace/kubectl"
	"github.com/devspace-cloud/devspace/pkg/devspace/services/targetselector"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/devspace-cloud/devspace/pkg/util/message"
	"github.com/devspace-cloud/devspace/pkg/util/port"
	"github.com/pkg/errors"
)

// StartPortForwarding starts the port forwarding functionality
func StartPortForwarding(config *latest.Config, generatedConfig *generated.Config, client *kubectl.Client, log log.Logger) ([]*portforward.PortForwarder, error) {
	if config.Dev.Ports != nil {
		portforwarder := make([]*portforward.PortForwarder, 0, len(config.Dev.Ports))

		for portConfigIndex, portForwarding := range config.Dev.Ports {
			var imageSelector []string
			if portForwarding.ImageName != "" && generatedConfig != nil {
				imageConfigCache := generatedConfig.GetActive().GetImageCache(portForwarding.ImageName)
				if imageConfigCache.ImageName != "" {
					imageSelector = []string{imageConfigCache.ImageName + ":" + imageConfigCache.Tag}
				}
			}

			selector, err := targetselector.NewTargetSelector(config, client, &targetselector.SelectorParameter{
				ConfigParameter: targetselector.ConfigParameter{
					Namespace:     portForwarding.Namespace,
					LabelSelector: portForwarding.LabelSelector,
				},
			}, false, imageSelector)
			if err != nil {
				return nil, errors.Errorf("Error creating target selector: %v", err)
			}

			log.StartWait("Port-Forwarding: Waiting for containers to start...")
			pod, err := selector.GetPod(log)
			log.StopWait()
			if err != nil {
				return nil, errors.Errorf("%s: %s", message.SelectorErrorPod, err.Error())
			} else if pod != nil {
				ports := make([]string, len(portForwarding.PortMappings))
				addresses := make([]string, len(portForwarding.PortMappings))

				for index, value := range portForwarding.PortMappings {
					if value.LocalPort == nil {
						return nil, errors.Errorf("port is not defined in portmapping %d:%d", portConfigIndex, index)
					}

					localPort := strconv.Itoa(*value.LocalPort)
					remotePort := localPort
					if value.RemotePort != nil {
						remotePort = strconv.Itoa(*value.RemotePort)
					}

					open, _ := port.Check(*value.LocalPort)
					if open == false {
						log.Warnf("Seems like port %d is already in use. Is another application using that port?", *value.LocalPort)
					}

					ports[index] = localPort + ":" + remotePort
					if value.BindAddress == "" {
						addresses[index] = "127.0.0.1"
					} else {
						addresses[index] = value.BindAddress
					}
				}

				readyChan := make(chan struct{})

				pf, err := client.NewPortForwarder(pod, ports, addresses, make(chan struct{}), readyChan)
				if err != nil {
					return nil, errors.Errorf("Error starting port forwarding: %v", err)
				}

				go func() {
					err := pf.ForwardPorts()
					if err != nil {
						log.Fatalf("Error forwarding ports: %v", err)
					}
				}()

				// Wait till forwarding is ready
				select {
				case <-readyChan:
					log.Donef("Port forwarding started on %s", strings.Join(ports, ", "))

					portforwarder = append(portforwarder, pf)
				case <-time.After(20 * time.Second):
					return nil, errors.Errorf("Timeout waiting for port forwarding to start")
				}
			}
		}

		return portforwarder, nil
	}

	return nil, nil
}
