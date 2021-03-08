package ioeasy

import (
	"io/ioutil"
	"os"

	"github.com/luno/jettison/errors"

	"gomicro/templates"
)

func CreateFileIfNotExists(path string, fc templates.FileConfig) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = ioutil.WriteFile(path, []byte{}, os.ModePerm)
		if err != nil {
			return errors.New("unable to create file")
		}

		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return errors.Wrap(err, "")
		}

		for _, adder := range fc {
			err := adder.AddTo(f)
			if err != nil {
				return errors.Wrap(err, "")
			}
		}
	}

	return nil
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func CreateDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return errors.New("unable to create service directory")
		}
	}

	return nil
}