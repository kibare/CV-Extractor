package utils

import (
    // "log"
    "mime/multipart"
    // "sync"
)

func DeleteFiles(files []*multipart.FileHeader) {
    // var wg sync.WaitGroup
    // for _, fileHeader := range files {
    //     wg.Add(1)
    //     go func(file string) {
    //         defer wg.Done()
    //         err := DeleteFileFromS3(file)
    //         if err != nil {
    //             log.Printf("Failed to delete file %s from S3: %v", file, err)
    //         }
    //     }(fileHeader.Filename)
    // }
    // wg.Wait()
}
