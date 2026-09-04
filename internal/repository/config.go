package repository

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/amiraminb/kaizen/internal/model"
)

func (r *FileRepository) LoadConfig() (model.Config, error) {
	path, err := r.DataFilePath(ConfigFileName)
	if err != nil {
		return model.Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return model.DefaultConfig(), nil
		}
		return model.Config{}, err
	}

	config := model.DefaultConfig()
	if err := json.Unmarshal(data, &config); err != nil {
		return model.Config{}, err
	}
	if err := config.Validate(); err != nil {
		return model.Config{}, err
	}
	return config, nil
}

func (r *FileRepository) SaveConfig(config model.Config) error {
	path, err := r.DataFilePath(ConfigFileName)
	if err != nil {
		return err
	}
	if err := config.Validate(); err != nil {
		return err
	}
	return saveDocument(path, config)
}
