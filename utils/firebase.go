package utils

import (
    "context"
    "fmt"
    "io"
    "mime/multipart"
    "os"
    "time"

    firebase "firebase.google.com/go"
    "firebase.google.com/go/storage"
    "google.golang.org/api/option"
)

var storageClient *storage.Client
var bucketName string

func InitFirebase() error {
    ctx := context.Background()
    conf := &firebase.Config{
        StorageBucket: os.Getenv("FIREBASE_STORAGE_BUCKET"),
    }
    opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
    app, err := firebase.NewApp(ctx, conf, opt)
    if err != nil {
        return fmt.Errorf("error initializing app: %v", err)
    }

    storageClient, err = app.Storage(ctx)
    if err != nil {
        return fmt.Errorf("error initializing storage client: %v", err)
    }

    bucketName = os.Getenv("FIREBASE_STORAGE_BUCKET")
    return nil
}

func UploadFileToFirebase(fileHeader *multipart.FileHeader) (string, error) {
    file, err := fileHeader.Open()
    if err != nil {
        return "", err
    }
    defer file.Close()

    ctx := context.Background()
    bucket, err := storageClient.DefaultBucket()
    if err != nil {
        return "", err
    }

    objectName := fmt.Sprintf("cv_files/%d-%s", time.Now().Unix(), fileHeader.Filename)
    wc := bucket.Object(objectName).NewWriter(ctx)
    wc.ContentType = fileHeader.Header.Get("Content-Type")
    if _, err = io.Copy(wc, file); err != nil {
        return "", err
    }
    if err := wc.Close(); err != nil {
        return "", err
    }

    // Generate a signed URL
    attrs, err := bucket.Object(objectName).Attrs(ctx)
    if err != nil {
        return "", err
    }

    url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", attrs.Bucket, attrs.Name)
    return url, nil
}

func DeleteFileFromFirebase(fileKey string) error {
    ctx := context.Background()
    bucket, err := storageClient.DefaultBucket()
    if err != nil {
        return err
    }

    o := bucket.Object(fileKey)
    if err := o.Delete(ctx); err != nil {
        return err
    }

    return nil
}
