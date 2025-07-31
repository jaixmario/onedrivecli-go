package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "time"
)

const (
    ClientID = "59790544-ca0c-4b77-b338-26ff9d1b676f"
    TenantID = "0fd666e8-0b3d-41ea-a5ef-1c509130bd94"
    TokenFile = "token.json"
)

type DeviceCodeResponse struct {
    DeviceCode              string `json:"device_code"`
    UserCode                string `json:"user_code"`
    VerificationURI         string `json:"verification_uri"`
    VerificationURIComplete string `json:"verification_uri_complete"`
    ExpiresIn               int    `json:"expires_in"`
    Interval                int    `json:"interval"`
    Message                 string `json:"message"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
    Scope        string `json:"scope"`
    Error        string `json:"error"`
    ErrorDesc    string `json:"error_description"`
}

func DeviceLogin() {
    fmt.Println("🔹 Starting Microsoft Device Login...")

    dc := getDeviceCode()
    if dc.DeviceCode == "" {
        fmt.Println("❌ Failed to get device code.")
        return
    }

    fmt.Println(dc.Message)
    fmt.Println("Verification URL:", dc.VerificationURIComplete)

    token := pollForToken(dc)
    if token.AccessToken != "" {
        fmt.Println("✅ Login successful!")
        fmt.Println("Access Token:", token.AccessToken[:40]+"...")
        fmt.Println("Refresh Token:", token.RefreshToken[:40]+"...")
        saveToken(token)
    } else {
        fmt.Println("❌ Login failed or timed out.")
    }
}

func getDeviceCode() DeviceCodeResponse {
    authURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/devicecode", TenantID)
    data := url.Values{}
    data.Set("client_id", ClientID)
    data.Set("scope", "offline_access Files.ReadWrite.All")

    resp, err := http.PostForm(authURL, data)
    if err != nil {
        fmt.Println("❌ HTTP request failed:", err)
        return DeviceCodeResponse{}
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var dcResp DeviceCodeResponse
    json.Unmarshal(body, &dcResp)
    return dcResp
}

func pollForToken(dc DeviceCodeResponse) TokenResponse {
    tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", TenantID)
    timeout := time.After(time.Duration(dc.ExpiresIn) * time.Second)

    for {
        select {
        case <-timeout:
            fmt.Println("⏳ Device code expired.")
            return TokenResponse{}
        default:
            time.Sleep(5 * time.Second)

            data := url.Values{}
            data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
            data.Set("client_id", ClientID)
            data.Set("device_code", dc.DeviceCode)

            resp, err := http.PostForm(tokenURL, data)
            if err != nil {
                fmt.Println("❌ HTTP Request Failed:", err)
                continue
            }

            body, _ := io.ReadAll(resp.Body)
            resp.Body.Close()

            var tokenResp TokenResponse
            json.Unmarshal(body, &tokenResp)

            if tokenResp.AccessToken != "" {
                return tokenResp
            }

            if tokenResp.Error == "authorization_pending" {
                fmt.Println("⌛ Waiting for user login...")
                continue
            }

            if tokenResp.Error != "" {
                fmt.Println("⚠️ Error:", tokenResp.ErrorDesc)
            }
        }
    }
}

func saveToken(token TokenResponse) {
    f, err := os.Create(TokenFile)
    if err != nil {
        fmt.Println("❌ Failed to save token:", err)
        return
    }
    defer f.Close()
    json.NewEncoder(f).Encode(token)
    fmt.Println("💾 Token saved to", TokenFile)
}