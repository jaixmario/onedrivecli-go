package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func CheckStorage() {
    accessToken := GetAccessToken()
    endpoint := "https://graph.microsoft.com/v1.0/me/drive"

    req, _ := http.NewRequest("GET", endpoint, nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Println("❌ HTTP request failed:", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode == 401 {
        fmt.Println("⚠️ Token expired, refreshing...")
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
        fmt.Println("❌ Failed to get storage info. Response:")
        fmt.Println(string(body))
        return
    }

    var data map[string]interface{}
    if err := json.Unmarshal(body, &data); err != nil {
        fmt.Println("❌ JSON parse error:", err)
        fmt.Println(string(body))
        return
    }

    quota, ok := data["quota"].(map[string]interface{})
    if !ok {
        fmt.Println("❌ Could not parse quota from response")
        fmt.Println(string(body))
        return
    }

    used := quota["used"].(float64)
    total := quota["total"].(float64)
    remaining := quota["remaining"].(float64)

    fmt.Printf("💾 Storage Used: %.2f GB / %.2f GB\n", used/1024/1024/1024, total/1024/1024/1024)
    fmt.Printf("🟢 Remaining: %.2f GB\n", remaining/1024/1024/1024)
}