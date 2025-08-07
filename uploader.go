package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "sync/atomic"
    "time"
)

const (
    chunkSize = 10 * 1024 * 1024 // 10MB
    workers   = 4
)

func StartUpload(remote, local string) error {
    if local == "." {
        cwd, _ := os.Getwd()
        local = cwd
    }

    info, err := os.Stat(local)
    if err != nil {
        return err
    }

    token := GetAccessToken()
    if info.IsDir() {
        return uploadFolder(remote, local, token)
    }
    return uploadFile(remote, local, token)
}

func uploadFolder(remote, local, token string) error {
    return filepath.Walk(local, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }

        relPath, _ := filepath.Rel(local, path)
        oneDrivePath := filepath.ToSlash(filepath.Join(remote, relPath))
        return uploadFile(oneDrivePath, path, token)
    })
}

func uploadFile(remote, local, token string) error {
    file, err := os.Open(local)
    if err != nil {
        return err
    }
    defer file.Close()

    info, _ := file.Stat()
    size := info.Size()

    // Create upload session
    sessionURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/root:%s:/createUploadSession",
        escapePath("/" + strings.TrimLeft(remote, "/")))
    reqBody := map[string]interface{}{
        "item": map[string]string{"@microsoft.graph.conflictBehavior": "replace"},
    }
    jsonBody, _ := json.Marshal(reqBody)

    req, _ := http.NewRequest("POST", sessionURL, bytes.NewReader(jsonBody))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var session struct {
        UploadURL string `json:"uploadUrl"`
    }
    json.NewDecoder(resp.Body).Decode(&session)
    if session.UploadURL == "" {
        return fmt.Errorf("failed to create upload session")
    }

    fmt.Printf("🚀 Uploading %s -> %s\n", local, remote)
    return uploadChunks(file, size, session.UploadURL)
}

func uploadChunks(file *os.File, fileSize int64, uploadURL string) error {
    totalChunks := int((fileSize + chunkSize - 1) / chunkSize)
    var uploaded int64
    start := time.Now()
    jobs := make(chan int, totalChunks)
    var wg sync.WaitGroup

    for w := 0; w < workers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for idx := range jobs {
                startByte := int64(idx) * chunkSize
                endByte := startByte + chunkSize - 1
                if endByte >= fileSize {
                    endByte = fileSize - 1
                }
                size := endByte - startByte + 1
                buffer := make([]byte, size)
                file.ReadAt(buffer, startByte)

                req, _ := http.NewRequest("PUT", uploadURL, bytes.NewReader(buffer))
                req.Header.Set("Content-Length", fmt.Sprint(size))
                req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startByte, endByte, fileSize))

                resp, err := http.DefaultClient.Do(req)
                if err != nil {
                    log.Printf("Chunk %d failed: %v\n", idx, err)
                    continue
                }
                io.Copy(io.Discard, resp.Body)
                resp.Body.Close()

                atomic.AddInt64(&uploaded, size)
                printProgress(uploaded, fileSize, start)
            }
        }()
    }

    for i := 0; i < totalChunks; i++ {
        jobs <- i
    }
    close(jobs)
    wg.Wait()

    fmt.Println("\n✅ Upload complete!")
    return nil
}

func printProgress(uploaded, total int64, start time.Time) {
    percent := float64(uploaded) / float64(total) * 100
    elapsed := time.Since(start).Seconds()
    speed := float64(uploaded) / 1024 / 1024 / elapsed
    eta := "-"
    if speed > 0 {
        eta = fmt.Sprintf("%.1fs", float64(total-uploaded)/1024/1024/speed)
    }

    fmt.Printf("\r%.2f%% | %d/%d MB | %.2f MB/s | Elapsed: %.1fs | ETA: %s",
        percent, uploaded/1024/1024, total/1024/1024, speed, elapsed, eta)
}

func escapePath(path string) string {
    return strings.ReplaceAll(path, " ", "%20")
}
