package loader

import (
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/setting"
)

// LoadPkgs 从部署包路径中加载应用
func LoadPkgs(categories Categories, pkgsDir string) error {
	if err := GenAppsYml(pkgsDir); err != nil {
		return err
	}
	if err := LoadApps(categories, filepath.Join(pkgsDir, setting.AppsYml)); err != nil {
		return err
	}
	return nil
}

// GenAppsYml 在部署包路径上生成Apps.yml
func GenAppsYml(pkgsDir string) error {
	pkgsPath, _ := filepath.Abs(pkgsDir)
	workDir, _ := filepath.Abs(filepath.Join(setting.AnsibleSetting.PlaybooksDir, setting.PlaybooksAppsDir))
	cmd := exec.Command("/usr/bin/python", setting.GenAppYml, pkgsPath)
	cmd.Dir = workDir
	log.Debug(cmd.String())
	if err := cmd.Start(); err != nil {
		log.Error(err)
		return err
	}
	if err := cmd.Wait(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
