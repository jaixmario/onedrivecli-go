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

type StoredToken struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    TokenType    string `json:"token_type"`
    Scope        string `json:"scope"`
    ObtainedAt   int64  `json:"obtained_at"` 
}

func LoadToken() (StoredToken, error) {
    file, err := os.Open(TokenFile)
    if err != nil {
        return StoredToken{}, err
    }
    defer file.Close()
    var token StoredToken
    json.NewDecoder(file).Decode(&token)
    return token, nil
}

func SaveToken(token TokenResponse) error {
    stored := StoredToken{
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
        ExpiresIn:    token.ExpiresIn,
        TokenType:    token.TokenType,
        Scope:        token.Scope,
        ObtainedAt:   time.Now().Unix(),
    }

    f, err := os.Create(TokenFile)
    if err != nil {
        return err
    }
    defer f.Close()
    return json.NewEncoder(f).Encode(stored)
}

func GetAccessToken() string {
    token, err := LoadToken()
    if err != nil {
        fmt.Println("❌ No token found, please run `onedrivecli auth` first.")
        os.Exit(1)
    }

    expiresAt := token.ObtainedAt + int64(token.ExpiresIn) - 30
    if time.Now().Unix() > expiresAt {
        fmt.Println("🔄 Access token expired, refreshing...")
        token = RefreshAccessToken(token.RefreshToken)
    }

    return token.AccessToken
}

func RefreshAccessToken(refreshToken string) StoredToken {
    tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", TenantID)
    data := url.Values{}
    data.Set("grant_type", "refresh_token")
    data.Set("client_id", ClientID)
    data.Set("refresh_token", refreshToken)

    resp, err := http.PostForm(tokenURL, data)
    if err != nil {
        fmt.Println("❌ HTTP Request Failed:", err)
        os.Exit(1)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    var tokenResp TokenResponse
    json.Unmarshal(body, &tokenResp)

    if tokenResp.AccessToken == "" {
        fmt.Println("❌ Failed to refresh token. Please run `onedrivecli auth` again.")
        os.Exit(1)
    }

    SaveToken(tokenResp)

    return StoredToken{
        AccessToken:  tokenResp.AccessToken,
        RefreshToken: tokenResp.RefreshToken,
        ExpiresIn:    tokenResp.ExpiresIn,
        TokenType:    tokenResp.TokenType,
        Scope:        tokenResp.Scope,
        ObtainedAt:   time.Now().Unix(),
    }
}