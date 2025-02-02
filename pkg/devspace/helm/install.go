package helm

import (
	"io/ioutil"
	"strings"

	"github.com/devspace-cloud/devspace/pkg/util/ptr"
	"github.com/pkg/errors"

	"github.com/devspace-cloud/devspace/pkg/devspace/analyze"

	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/util/log"

	yaml "gopkg.in/yaml.v2"
	helmchartutil "k8s.io/helm/pkg/chartutil"
	helmdownloader "k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	k8shelm "k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
)

// DeploymentTimeout is the timeout to wait for helm to deploy
const DeploymentTimeout = int64(180)

func checkDependencies(ch *chart.Chart, reqs *helmchartutil.Requirements) error {
	missing := []string{}

	deps := ch.GetDependencies()
	for _, r := range reqs.Dependencies {
		found := false
		for _, d := range deps {
			if d.Metadata.Name == r.Name {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, r.Name)
		}
	}

	if len(missing) > 0 {
		return errors.Errorf("found in requirements.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
	}
	return nil
}

// InstallChartByPath installs the given chartpath und the releasename in the releasenamespace
func (client *Client) InstallChartByPath(releaseName, releaseNamespace, chartPath string, values *map[interface{}]interface{}, helmConfig *latest.HelmConfig) (*hapi_release5.Release, error) {
	if releaseNamespace == "" {
		releaseNamespace = client.kubectl.Namespace
	}

	chart, err := helmchartutil.Load(chartPath)
	if err != nil {
		return nil, err
	}

	if req, err := helmchartutil.LoadRequirements(chart); err == nil {
		// If checkDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/kubernetes/helm/issues/2209
		if err := checkDependencies(chart, req); err != nil {
			man := &helmdownloader.Manager{
				Out:       ioutil.Discard,
				ChartPath: chartPath,
				HelmHome:  client.Settings.Home,
				Getters:   getter.All(*client.Settings),
			}
			if err := man.Update(); err != nil {
				return nil, err
			}

			// Update all dependencies which are present in /charts.
			chart, err = helmchartutil.Load(chartPath)
			if err != nil {
				return nil, err
			}
		}
	} else if err != helmchartutil.ErrRequirementsNotFound {
		return nil, errors.Errorf("cannot load requirements: %v", err)
	}

	releaseExists := ReleaseExists(client.helm, releaseName)
	overwriteValues := []byte("")

	if values != nil {
		unmarshalledValues, err := yaml.Marshal(values)

		if err != nil {
			return nil, err
		}
		overwriteValues = unmarshalledValues
	}

	// Set wait and timeout
	waitTimeout := DeploymentTimeout
	if helmConfig.Timeout != nil {
		waitTimeout = *helmConfig.Timeout
	}

	wait := false
	if helmConfig.Wait != nil {
		wait = *helmConfig.Wait
	}

	rollback := false
	if helmConfig.Rollback != nil {
		rollback = *helmConfig.Rollback
	}

	if releaseExists {
		upgradeResponse, err := client.helm.UpdateRelease(
			releaseName,
			chartPath,
			k8shelm.UpgradeWait(wait),
			k8shelm.UpgradeTimeout(waitTimeout),
			k8shelm.UpdateValueOverrides(overwriteValues),
			k8shelm.ReuseValues(false),
			k8shelm.UpgradeForce(ptr.ReverseBool(helmConfig.Force)),
		)

		if err != nil {
			err = client.analyzeError(errors.Errorf("helm upgrade: %v", err), releaseNamespace)
			if err != nil {
				if rollback {
					log.Warn("Try to roll back back chart because of previous error")
					_, rollbackError := client.helm.RollbackRelease(releaseName, k8shelm.RollbackTimeout(180))
					if rollbackError != nil {
						return nil, err
					}
				}

				return nil, err
			}

			return nil, nil
		}

		return upgradeResponse.GetRelease(), nil
	}

	installResponse, err := client.helm.InstallReleaseFromChart(
		chart,
		releaseNamespace,
		k8shelm.InstallWait(wait),
		k8shelm.InstallTimeout(waitTimeout),
		k8shelm.ValueOverrides(overwriteValues),
		k8shelm.ReleaseName(releaseName),
		k8shelm.InstallReuseName(true),
	)
	if err != nil {
		err = client.analyzeError(errors.Errorf("helm install: %v", err), releaseNamespace)
		if err != nil {
			if rollback {
				// Try to delete and ignore errors, because otherwise we have a broken release laying around and always get the no deployed resources error
				client.DeleteRelease(releaseName, true)
			}

			return nil, err
		}

		return nil, nil
	}

	return installResponse.GetRelease(), nil
}

// analyzeError calls analyze and tries to find the issue
func (client *Client) analyzeError(srcErr error, releaseNamespace string) error {
	errMessage := srcErr.Error()

	// Only check if the error is time out
	if strings.Index(errMessage, "timed out waiting") != -1 {
		report, err := analyze.CreateReport(client.kubectl, releaseNamespace, false)
		if err != nil {
			log.Warnf("Error creating analyze report: %v", err)
			return srcErr
		}
		if len(report) == 0 {
			return nil
		}

		return errors.New(analyze.ReportToString(report))
	}

	return srcErr
}

// InstallChart installs the given chart by name under the releasename in the releasenamespace
func (client *Client) InstallChart(releaseName string, releaseNamespace string, values *map[interface{}]interface{}, helmConfig *latest.HelmConfig) (*hapi_release5.Release, error) {
	chart := helmConfig.Chart
	chartPath, err := locateChartPath(client.Settings, chart.RepoURL, chart.Username, chart.Password, chart.Name, chart.Version, false, "", "", "", "")
	if err != nil {
		return nil, errors.Wrap(err, "locate chart path")
	}

	return client.InstallChartByPath(releaseName, releaseNamespace, chartPath, values, helmConfig)
}
