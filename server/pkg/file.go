package pkg

import (
	"bufio"
	"fmt"
	"os"
)

func GetLastLine(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the file line by line and store the last line
	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return lastLine, nil
}

// func CopyMessageFolder(srcDir, destDir string) chan error {
// 	errCh := make(chan error)

// 	if err := os.MkdirAll(destDir, 0755); err != nil {
// 		errCh <- fmt.Errorf("failed to create message copy directory: %w", err)
// 		return errCh
// 	}

// 	go func() {
// 		defer close(errCh)

// 		err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
// 			if !info.IsDir() && filepath.Ext(info.Name()) == ".log" {
// 				destPath := filepath.Join(destDir, info.Name())

// 				if err := copyMessageFiles(path, destPath); err != nil {
// 					errCh <- fmt.Errorf("failed to copy %s: %v", path, err)
// 				}
// 			}
// 			return nil
// 		})

// 		if err != nil {
// 			errCh <- fmt.Errorf("fail while iterating through files: %v", err)
// 		}
// 	}()
// 	return errCh
// }

// func copyMessageFiles(src, dst string) error {
// 	sourceFile, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer sourceFile.Close()

// 	destinationFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	defer destinationFile.Close()

// 	_, err = io.Copy(destinationFile, sourceFile)
// 	if err != nil {
// 		return err
// 	}

// 	err = destinationFile.Sync()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
