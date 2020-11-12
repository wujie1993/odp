package util

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	InstancePoolActionAdd    = "add"
	InstancePoolActionDelete = "delete"
)

var instancePoolMutex sync.Mutex

func AddInstancePoolItem(dataDir, host, instanceName, instanceUid string) (int, error) {
	id, err := GetInstancePoolItem(dataDir, host, instanceName, instanceUid)
	if err != nil {
		return id, err
	}
	if id == -1 {
		return setInstancePoolItem(dataDir, host, instanceName, instanceUid, true)
	}
	return id, err
}

func RemoveInstancePoolItem(dataDir, host, instanceName, instanceUid string) (int, error) {
	return setInstancePoolItem(dataDir, host, instanceName, instanceUid, false)
}

func setInstancePoolItem(dataDir, host, instanceName, instanceUid string, add bool) (int, error) {
	id := -1

	instancePoolMutex.Lock()
	defer instancePoolMutex.Unlock()

	fileName := filepath.Join(dataDir, host, instanceName)

	// 初始化实例池数据目录
	fileDir := filepath.Dir(fileName)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return id, err
	}

	// 初始化实例池文件
	poolFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return id, err
	}
	defer poolFile.Close()

	// 读取实例池数据
	csvReader := csv.NewReader(poolFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		return id, err
	}

	// 处理实例池
	if add {
		// 判断Uid是否存在，不存在则插入新记录
		insertIndex := -1
		for index, record := range records {
			if len(record) != 2 {
				return id, errors.New("field length error of instance pool cvs file " + fileName)
			}
			if record[1] == instanceUid {
				return id, errors.New(fmt.Sprintf("instance %s(%s) already exist in host %s", instanceName, instanceUid, host))
			}
			id, err := strconv.Atoi(record[0])
			if err != nil {
				return id, err
			}
			if insertIndex == -1 && index < id {
				insertIndex = index
			}
		}
		if insertIndex == -1 {
			insertIndex = len(records)
		}

		// 追加记录
		tmp := append([][]string{}, records[insertIndex:]...)
		records = append(records[:insertIndex], []string{strconv.Itoa(insertIndex), instanceUid})
		records = append(records, tmp...)

		id = insertIndex
	} else {
		deleteIndex := -1
		for index, record := range records {
			if len(record) != 2 {
				return id, errors.New("field length error of instance pool cvs file " + fileName)
			}
			if record[1] == instanceUid {
				deleteIndex = index
				break
			}
		}
		if deleteIndex == -1 {
			return id, nil
		}

		// 移除记录
		records = append(records[0:deleteIndex], records[deleteIndex+1:]...)

		id = deleteIndex
	}

	// 保存实例池
	csvWriter := csv.NewWriter(poolFile)
	if _, err := poolFile.Seek(0, 0); err != nil {
		return id, err
	}
	if err := poolFile.Truncate(0); err != nil {
		return id, err
	}
	if err := csvWriter.WriteAll(records); err != nil {
		return id, err
	}

	return id, nil
}

func GetInstancePoolItem(dataDir, host, instanceName, instanceUid string) (int, error) {
	id := -1

	instancePoolMutex.Lock()
	defer instancePoolMutex.Unlock()

	fileName := filepath.Join(dataDir, host, instanceName)

	// 初始化实例池数据目录
	fileDir := filepath.Dir(fileName)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return id, err
	}

	// 初始化实例池文件
	poolFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return id, err
	}
	defer poolFile.Close()

	// 读取实例池数据
	csvReader := csv.NewReader(poolFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		return id, err
	}

	for _, record := range records {
		if len(record) != 2 {
			return id, errors.New("field length error of instance pool cvs file " + fileName)
		}
		if record[1] == instanceUid {
			id, err := strconv.Atoi(record[0])
			if err != nil {
				return id, err
			}
			return id, nil
		}
	}
	return id, nil
}
