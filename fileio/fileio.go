// Package fileio is the package responsible for opening and writing
// the sitemap to the output file.
package fileio

import (
	"fmt"
	"os"
)

// WriteXML opens the given output file and writes the given data to the file
func WriteXML(data []byte, output string) error {
	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Unable to open the file: %s", err.Error())
	}
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("Unable to write the data to the file: %s", err.Error())
	}
	return nil
}
