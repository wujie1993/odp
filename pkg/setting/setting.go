package setting

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-ini/ini"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	DefaultConfDir    = "conf"
	JobsDir           = "jobs"
	HostsDir          = "hosts"
	PlaybooksRolesDir = "roles"
	PlaybooksAppsDir  = ".apps"

	GenAppYml = "gen_app_yml.py"
	AppsYml   = "Apps.yml"
)

var GPUTypesSetting map[string][]string

type App struct {
	PrefixUrl string
	LogLevel  log.Level
	JwtSecret string
	DataDir   string
}

var AppSetting = &App{}

type Server struct {
	RunMode      string
	HttpPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

var ServerSetting = &Server{}

type Package struct {
	ScanPath string
	PkgPath  string
}

var PackageSetting = &Package{}

type Etcd struct {
	Endpoints []string
}

var EtcdSetting = &Etcd{}

type Ansible struct {
	Bin          string
	BaseDir      string
	PlaybooksDir string
	TplsDir      string
	LogToStdout  bool
	DryRun       bool
}

var AnsibleSetting = &Ansible{}

var cfg *ini.File

// Setup initialize the configuration instance
func Setup() {
	var err error
	cfg, err = ini.Load(filepath.Join(DefaultConfDir, "app.ini"))
	if err != nil {
		log.Fatalf("setting.Setup, fail to parse 'conf/app.ini': %v", err)
	}

	mapTo("app", AppSetting)
	mapTo("server", ServerSetting)
	mapTo("package", PackageSetting)
	mapTo("etcd", EtcdSetting)
	mapTo("ansible", AnsibleSetting)

	ServerSetting.ReadTimeout = ServerSetting.ReadTimeout * time.Second
	ServerSetting.WriteTimeout = ServerSetting.WriteTimeout * time.Second

	// Setup GPU types
	gpuTypesFile, err := os.Open(filepath.Join(DefaultConfDir, "gpu_types.yml"))
	if err != nil {
		log.Fatal(err)
	}
	gpuTypesDecoder := yaml.NewDecoder(gpuTypesFile)
	if err := gpuTypesDecoder.Decode(&GPUTypesSetting); err != nil {
		log.Fatal(err)
	}
}

// mapTo map section
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}

func GetGPUType(model string) string {
	for gpuType, gpuModels := range GPUTypesSetting {
		for _, gpuModel := range gpuModels {
			if gpuModel == model {
				return gpuType
			}
		}
	}
	return "unknown"
}
