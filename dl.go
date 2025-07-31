package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
)

type DriveItemSingle struct {
    Name        string `json:"name"`
    DownloadUrl string `json:"@microsoft.graph.downloadUrl"`
}

func DownloadFile(path string) {
    accessToken := GetAccessToken()

    apiPath := fmt.Sprintf("root:/%s", strings.TrimPrefix(path, "/"))

    apiPath = url.PathEscape(apiPath)

    reqURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/%s", apiPath)

    req, _ := http.NewRequest("GET", reqURL, nil)
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
    var fileResp DriveItemSingle
    if err := json.Unmarshal(body, &fileResp); err != nil {
        fmt.Println("❌ Failed to parse response:", err)
        fmt.Println(string(body))
        return
    }

    if fileResp.DownloadUrl == "" {
        fmt.Println("❌ Could not generate download link. Check the path or permissions.")
        fmt.Println(string(body))
        return
    }

    fmt.Println("✅ Direct Download Link:")
    fmt.Println(fileResp.DownloadUrl)
}