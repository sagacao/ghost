package format

import "os"

func createFilePath(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		//err := os.MkdirAll(path.Dir(logger.filename), os.ModePerm)
		err := os.MkdirAll(filepath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}
