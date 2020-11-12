package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type App struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            AppSpec
}

type AppSpec struct {
	Category string
	Versions []AppVersion
}

type AppVersion struct {
	Version           string
	ShortName         string
	Platform          string
	Desc              string
	Increment         bool
	SupportActions    []string
	SupportMediaTypes []string
	SupportGpuModels  []string
	Enabled           bool
	DashboardId       string
	PkgRef            string
	LivenessProbe     LivenessProbe
	Modules           []AppModule
	Global            AppGlobal
	GpuRequired       bool
}

type AppModule struct {
	Name                    string
	Desc                    string
	SkipUpgrade             bool
	Required                bool
	Notes                   string
	Replication             bool
	HostLimits              HostLimits
	IncludeRoles            []string
	Args                    []AppArgs
	ConfigMapRef            ConfigMapRef
	AdditionalConfigMapRef  ConfigMapRef
	EnableAdditionalConfigs bool
	EnableLogging           bool
	EnablePurgeData         bool
	ExtraVars               map[string]interface{}
	Resources               Resources
}

type AppGlobal struct {
	Args         []AppArgs
	ConfigMapRef ConfigMapRef
}

type HostLimits struct {
	Max int
	Min int
}

type AppArgs struct {
	Name       string
	ShortName  string
	Desc       string
	Type       string
	Format     string
	Enum       []string
	HostLimits HostLimits
	Default    interface{}
	Required   bool
	Modifiable bool
	Readonly   bool
}

type Resources struct {
	AlgorithmPlugin               bool
	SupportAlgorithmPluginsRegexp string
}

func (obj App) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *App) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj App) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (obj App) GetVersion(version string) (AppVersion, bool) {
	for _, appVersion := range obj.Spec.Versions {
		if appVersion.Version == version {
			return appVersion, true
		}
	}
	return AppVersion{}, false
}

func (obj App) GetVersionModule(version string, moduleName string) (AppModule, bool) {
	appVersion, ok := obj.GetVersion(version)
	if !ok {
		return AppModule{}, false
	}
	return appVersion.GetModule(moduleName)
}

func (obj AppVersion) GetModule(moduleName string) (AppModule, bool) {
	for _, module := range obj.Modules {
		if module.Name == moduleName {
			return module, true
		}
	}
	return AppModule{}, false
}

type AppRegistry struct {
	registry.Registry
}

func appMutate(obj core.ApiObject) error {
	app := obj.(*App)
	for index, versionApp := range app.Spec.Versions {
		if versionApp.LivenessProbe.InitialDelaySeconds < 0 {
			app.Spec.Versions[index].LivenessProbe.InitialDelaySeconds = 10
		}
		if versionApp.LivenessProbe.PeriodSeconds < 30 {
			app.Spec.Versions[index].LivenessProbe.PeriodSeconds = 60
		}
		if versionApp.LivenessProbe.TimeoutSeconds < 30 {
			app.Spec.Versions[index].LivenessProbe.TimeoutSeconds = 60
		}
	}
	return nil
}

func appPreCreate(obj core.ApiObject) error {
	app := obj.(*App)
	app.Metadata.Finalizers = []string{core.FinalizerCleanRefEvent, core.FinalizerCleanRefConfigMap}
	return nil
}

func NewApp() *App {
	app := new(App)
	app.Init(ApiVersion, core.KindApp)
	app.Spec.Versions = []AppVersion{}
	return app
}

func NewAppRegistry() AppRegistry {
	app := AppRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindApp), true),
	}
	app.SetMutateHook(appMutate)
	app.SetPreCreateHook(appPreCreate)
	return app
}
