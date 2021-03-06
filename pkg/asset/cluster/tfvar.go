package cluster

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/installconfig"
	"github.com/openshift/installer/pkg/types/config"
)

const (
	tfvarFilename  = "terraform.tfvar"
	tfvarAssetName = "Terraform Variables"
)

// TerraformVariables depends on InstallConfig and
// Ignition to generate the terrafor.tfvars.
type TerraformVariables struct {
	// The root directory of the generated assets.
	rootDir string
	// The Assets that this tfvar file depends.
	installConfig     asset.Asset
	bootstrapIgnition asset.Asset
	masterIgnition    asset.Asset
	workerIgnition    asset.Asset
}

var _ asset.Asset = (*TerraformVariables)(nil)

// Name returns the human-friendly name of the asset.
func (t *TerraformVariables) Name() string {
	return tfvarAssetName
}

// Dependencies returns the dependency of the TerraformVariable
func (t *TerraformVariables) Dependencies() []asset.Asset {
	return []asset.Asset{t.installConfig, t.bootstrapIgnition, t.masterIgnition, t.workerIgnition}
}

// Generate generates the terraform.tfvar file.
func (t *TerraformVariables) Generate(parents map[asset.Asset]*asset.State) (*asset.State, error) {
	installCfg, err := installconfig.GetInstallConfig(t.installConfig, parents)
	if err != nil {
		return nil, fmt.Errorf("failed to get install config state in the parent asset states")
	}

	contents := map[asset.Asset][]string{}

	for _, ign := range []asset.Asset{
		t.bootstrapIgnition,
		t.masterIgnition,
		t.workerIgnition,
	} {
		state, ok := parents[ign]
		if !ok {
			return nil, fmt.Errorf("failed to get the ignition state for %v in the parent asset states", ign)
		}

		for _, content := range state.Contents {
			contents[ign] = append(contents[ign], string(content.Data))
		}
	}

	cluster, err := config.ConvertInstallConfigToTFVar(installCfg, contents[t.bootstrapIgnition][0], contents[t.masterIgnition], contents[t.workerIgnition][0])
	if err != nil {
		return nil, err
	}

	data, err := cluster.TFVars()
	if err != nil {
		return nil, err
	}

	return &asset.State{
		Contents: []asset.Content{
			{
				Name: filepath.Join(t.rootDir, tfvarFilename),
				Data: []byte(data),
			},
		},
	}, nil
}
