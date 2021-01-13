package loader

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm"
	"github.com/wujie1993/waves/pkg/orm/core"
)

func LoadObjsByLocalPath(path string) ([]core.ApiObject, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	f.Close()
	if stat.IsDir() {
		result := []core.ApiObject{}
		files, err := ioutil.ReadDir(path)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		for _, file := range files {
			if !file.IsDir() {
				objs, err := loadByFile(filepath.Join(path, file.Name()))
				if err != nil {
					log.Error(err)
					return nil, err
				}
				result = append(result, objs...)
			}
		}
		return result, nil
	}
	return loadByFile(path)
}

func loadByFile(path string) ([]core.ApiObject, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	objs := []core.ApiObject{}
	dataStr := string(data)
	switch dataStr[0] {
	case '[':
		// 解析json数组结构
		metas := []core.MetaType{}
		if err := json.Unmarshal(data, &metas); err != nil {
			log.Error(err)
			return nil, err
		}
		for _, meta := range metas {
			obj, err := orm.NewByMetaType(meta)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			objs = append(objs, obj)
		}
		if err := json.Unmarshal(data, &objs); err != nil {
			log.Error(err)
			return nil, err
		}
	case '{':
		// 解析json结构
		meta := core.MetaType{}
		if err := json.Unmarshal(data, &meta); err != nil {
			log.Error(err)
			return nil, err
		}
		obj, err := orm.NewByMetaType(meta)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		if err := obj.FromJSON(data); err != nil {
			log.Error(err)
			return nil, err
		}
		objs = append(objs, obj)
	default:
		// 解析yaml结构
		dataStrs := strings.Split(dataStr, "\n---\n")
		for _, dataStr := range dataStrs {
			meta := core.MetaType{}
			if err := yaml.Unmarshal([]byte(dataStr), &meta); err != nil {
				log.Error(err)
				log.Debug(dataStr)
				return nil, err
			}
			if meta.Kind == "" {
				continue
			}
			obj, err := orm.NewByMetaType(meta)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			if err := obj.FromYAML([]byte(dataStr)); err != nil {
				log.Error(err)
				return nil, err
			}
			objs = append(objs, obj)
		}
	}
	return objs, nil
}
