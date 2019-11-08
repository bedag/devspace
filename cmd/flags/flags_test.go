package flags

import (
	"testing"

	"github.com/devspace-cloud/devspace/pkg/devspace/config/generated"
	"github.com/devspace-cloud/devspace/pkg/util/log"

	"gotest.tools/assert"
)

type useLastContextTestCase struct {
	name string

	globalFlags     GlobalFlags
	generatedConfig *generated.Config

	expectedErr    string
}

func TestUseLastContext(t *testing.T) {
	testCases := []useLastContextTestCase{
		useLastContextTestCase{
			name: "Kube-Context and switch-context",
			globalFlags: GlobalFlags{
				KubeContext:   " ",
				SwitchContext: true,
			},
			expectedErr: "Flag --kube-context cannot be used together with --switch-context",
		},
		useLastContextTestCase{
			name: "Namespace and switch-context",
			globalFlags: GlobalFlags{
				Namespace:     " ",
				SwitchContext: true,
			},
			expectedErr: "Flag --namespace cannot be used together with --switch-context",
		},
		useLastContextTestCase{
			name: "Switch context to not existent",
			globalFlags: GlobalFlags{
				SwitchContext: true,
			},
			expectedErr: "There is no last context to use. Only use the '--switch-context / -s' flag if you already have deployed the project before",
		},
		useLastContextTestCase{
			name: "Switch context to existent",
			globalFlags: GlobalFlags{
				SwitchContext: true,
			},
			generatedConfig: &generated.Config{
				ActiveProfile: "someProfile",
				Profiles: map[string]*generated.CacheConfig{
					"someProfile": &generated.CacheConfig{
						LastContext: &generated.LastContextConfig{
							Context:   "myKubeContext",
							Namespace: "myNamespace",
						},
					},
				},
			},
		},
		useLastContextTestCase{
			name:        "Nothing happens",
			globalFlags: GlobalFlags{},
		},
	}

	for _, testCase := range testCases {
		testUseLastContext(t, testCase)
	}
}

func testUseLastContext(t *testing.T, testCase useLastContextTestCase) {
	err := testCase.globalFlags.UseLastContext(testCase.generatedConfig, &log.DiscardLogger{})

	if testCase.expectedErr == "" {
		assert.NilError(t, err, "Unexpected error in testCase %s.", testCase.name)
	} else {
		assert.Error(t, err, testCase.expectedErr, "Wrong or no error in testCase %s.", testCase.name)
	}
}

func TestToConfigOptions(t *testing.T) {
	configOptions := (&GlobalFlags{
		Profile:     "myProfile",
		KubeContext: "myKubeContext",
		Vars:        []string{"var1", "var2"},
	}).ToConfigOptions()

	assert.Equal(t, configOptions.Profile, "myProfile", "ConfigOptions has wrong profile")
	assert.Equal(t, configOptions.KubeContext, "myKubeContext", "ConfigOptions has wrong kube context")
	assert.Equal(t, len(configOptions.Vars), 2, "ConfigOptions has wrong vars")
	assert.Equal(t, configOptions.Vars[0], "var1", "ConfigOptions has wrong vars")
	assert.Equal(t, configOptions.Vars[1], "var2", "ConfigOptions has wrong vars")
}
