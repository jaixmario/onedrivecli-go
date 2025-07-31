package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
)

type DriveItem struct {
    Name   string `json:"name"`
    ID     string `json:"id"`
    Size   int64  `json:"size"`
    Folder *struct {
        ChildCount int `json:"childCount"`
    } `json:"folder,omitempty"`
    File *struct{} `json:"file,omitempty"`
}

type DriveResponse struct {
    Value []DriveItem `json:"value"`
}

func ListFiles(path string) {
    accessToken := GetAccessToken()

    cleanPath := strings.TrimPrefix(path, "/")
    endpoint := "https://graph.microsoft.com/v1.0/me/drive/root/children"
    if path != "/" {
        endpoint = fmt.Sprintf(
            "https://graph.microsoft.com/v1.0/me/drive/root:/%s:/children",
            url.PathEscape(cleanPath),
        )
    }

    req, _ := http.NewRequest("GET", endpoint, nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Println("❌ HTTP request failed:", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode == 401 {
        fmt.Println("⚠️ Unauthorized. Trying token refresh...")
        accessToken = GetAccessToken()
        req.Header.Set("Authorization", "Bearer "+accessToken)
        resp, err = http.DefaultClient.Do(req)
        if err != nil {
            fmt.Println("❌ HTTP request failed after refresh:", err)
            return
        }
        defer resp.Body.Close()
    }

    body, _ := io.ReadAll(resp.Body)

    if resp.StatusCode >= 300 {
        fmt.Println("❌ Failed to list files. Response:")
        fmt.Println(string(body))
        return
    }

    var driveResp DriveResponse
    if err := json.Unmarshal(body, &driveResp); err != nil {
        fmt.Println("❌ Failed to parse JSON:", err)
        fmt.Println(string(body))
        return
    }

    fmt.Println("📂 Listing:", path)
    for _, item := range driveResp.Value {
        if item.Folder != nil {
            fmt.Printf("📁 %s (%d items)\n", item.Name, item.Folder.ChildCount)
        } else {
            fmt.Printf("📄 %s (%.2f MB)\n", item.Name, float64(item.Size)/1024/1024)
        }
    }
}