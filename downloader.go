package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "sync/atomic"
    "time"
)

func fetchDriveItem(accessToken, remote string) (*DriveItem, error) {
    var endpoint string
    if strings.HasPrefix(remote, "/") {
        endpoint = fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/root:%s", remote)
        if !strings.HasSuffix(remote, "/") {
            endpoint += ":"
        }
    } else {
        endpoint = fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s", remote)
    }

    req, _ := http.NewRequest("GET", endpoint, nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var item DriveItem
    if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
        return nil, err
    }

    if item.Folder != nil {
        var childrenEndpoint string
        if strings.HasPrefix(remote, "/") {
            childrenEndpoint = endpoint + "/children"
        } else {
            childrenEndpoint = fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/children", item.ID)
        }

        req2, _ := http.NewRequest("GET", childrenEndpoint, nil)
        req2.Header.Set("Authorization", "Bearer "+accessToken)
        resp2, err := http.DefaultClient.Do(req2)
        if err != nil {
            return nil, err
        }
        defer resp2.Body.Close()

        var childResp struct {
            Value []DriveItem `json:"value"`
        }
        if err := json.NewDecoder(resp2.Body).Decode(&childResp); err != nil {
            return nil, err
        }
        item.Children = childResp.Value
    }

    return &item, nil
}

func calcTotalSize(item *DriveItem) int64 {
    if item.File != nil {
        return item.Size
    }
    total := int64(0)
    for _, child := range item.Children {
        total += calcTotalSize(&child)
    }
    return total
}

func downloadFileWithProgress(url, localPath string, downloaded *int64) error {
    os.MkdirAll(filepath.Dir(localPath), os.ModePerm)

    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(localPath)
    if err != nil {
        return err
    }
    defer out.Close()

    buf := make([]byte, 32*1024)
    for {
        n, err := resp.Body.Read(buf)
        if n > 0 {
            out.Write(buf[:n])
            atomic.AddInt64(downloaded, int64(n))
        }
        if err != nil {
            if err == io.EOF {
                break
            }
            return err
        }
    }
    return nil
}


func downloadRecursive(accessToken string, item *DriveItem, localPath string, downloaded *int64) error {
    if item.File != nil {
        fi, err := os.Stat(localPath)
        if (err == nil && fi.IsDir()) || strings.HasSuffix(localPath, string(os.PathSeparator)) {
            localPath = filepath.Join(localPath, item.Name)
        }
        fmt.Println("Downloading:", item.Name)
        return downloadFileWithProgress(item.DownloadURL, localPath, downloaded)
    }

    if item.Folder != nil {
        localFolder := localPath
        if item.Name != "" {
            localFolder = filepath.Join(localPath, item.Name)
        }
        os.MkdirAll(localFolder, os.ModePerm)
        for _, child := range item.Children {
            if err := downloadRecursive(accessToken, &child, localFolder, downloaded); err != nil {
                return err
            }
        }
    }
    return nil
}

// StartDownload used by main.go
func StartDownload(remote, localPath string) error {
    accessToken := GetAccessToken() // old signature: returns string
    ctx := context.Background()
    _ = ctx

    if localPath == "." {
        cwd, _ := os.Getwd()
        localPath = cwd
    }

    item, err := fetchDriveItem(accessToken, remote)
    if err != nil {
        return err
    }
    
    fi, statErr := os.Stat(localPath)
    if statErr == nil && fi.IsDir() && item.Folder != nil {
        localPath = filepath.Join(localPath, item.Name)
    } else if os.IsNotExist(statErr) {
        if item.Folder != nil {
            os.MkdirAll(localPath, os.ModePerm)
        } else {
            os.MkdirAll(filepath.Dir(localPath), os.ModePerm)
        }
    }

    totalSize := calcTotalSize(item)
    var downloaded int64 = 0
    start := time.Now()

    done := make(chan struct{})
    go func() {
        ticker := time.NewTicker(200 * time.Millisecond)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                percent := float64(downloaded) / float64(totalSize) * 100
                elapsed := time.Since(start).Seconds()
                speed := float64(downloaded) / 1024 / 1024 / elapsed
                eta := "-"
                if speed > 0 {
                    remaining := float64(totalSize-downloaded) / 1024 / 1024 / speed
                    eta = fmt.Sprintf("%.1fs", remaining)
                }
                fmt.Printf("\r%.2f%% | %d/%d MB | %.2f MB/s | Elapsed: %.1fs | ETA: %s",
                    percent,
                    downloaded/1024/1024, totalSize/1024/1024,
                    speed,
                    elapsed,
                    eta,
                )
            case <-done:
                fmt.Printf("\r100%% | %d/%d MB | Done!\n", totalSize/1024/1024, totalSize/1024/1024)
                return
            }
        }
    }()

    fmt.Println("Starting download to:", localPath)
    if err := downloadRecursive(accessToken, item, localPath, &downloaded); err != nil {
        return err
    }

    close(done)
    fmt.Println("\nDownload complete!")
    return nil
}
