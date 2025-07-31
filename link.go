package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
)

func GetShareLink(filePath string) string {
    accessToken := GetAccessToken()
    cleanPath := strings.TrimPrefix(filePath, "/")

    endpoint := fmt.Sprintf(
        "https://graph.microsoft.com/v1.0/me/drive/root:/%s:/createLink",
        url.PathEscape(cleanPath),
    )

    req, _ := http.NewRequest("POST", endpoint, strings.NewReader(`{"type":"view","scope":"anonymous"}`))
    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Println("❌ HTTP request failed:", err)
        return ""
    }
    defer resp.Body.Close()

    if resp.StatusCode == 401 {
        accessToken = GetAccessToken()
        req.Header.Set("Authorization", "Bearer "+accessToken)
        resp, err = http.DefaultClient.Do(req)
        if err != nil {
            fmt.Println("❌ HTTP request failed after refresh:", err)
            return ""
        }
        defer resp.Body.Close()
    }

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 300 {
        fmt.Println("❌ Could not generate share link. Response:")
        fmt.Println(string(body))
        return ""
    }

    var result struct {
        Link struct {
            WebUrl string `json:"webUrl"`
        } `json:"link"`
    }
    if err := json.Unmarshal(body, &result); err != nil {
        fmt.Println("❌ JSON parse failed:", err)
        fmt.Println(string(body))
        return ""
    }

    return result.Link.WebUrl
}

func GetDirectDownloadLink(filePath string) string {
    accessToken := GetAccessToken()
    cleanPath := strings.TrimPrefix(filePath, "/")

    endpoint := fmt.Sprintf(
        "https://graph.microsoft.com/v1.0/me/drive/root:/%s",
        url.PathEscape(cleanPath),
    )

    req, _ := http.NewRequest("GET", endpoint, nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Println("❌ HTTP request failed:", err)
        return ""
    }
    defer resp.Body.Close()

    if resp.StatusCode == 401 {
        accessToken = GetAccessToken()
        req.Header.Set("Authorization", "Bearer "+accessToken)
        resp, err = http.DefaultClient.Do(req)
        if err != nil {
            fmt.Println("❌ HTTP request failed after refresh:", err)
            return ""
        }
        defer resp.Body.Close()
    }

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 300 {
        fmt.Println("❌ Could not generate direct download link. Response:")
        fmt.Println(string(body))
        return ""
    }

    var data map[string]interface{}
    if err := json.Unmarshal(body, &data); err != nil {
        fmt.Println("❌ JSON parse failed:", err)
        fmt.Println(string(body))
        return ""
    }

    if url, ok := data["@microsoft.graph.downloadUrl"].(string); ok {
        return url
    }

    return ""
}