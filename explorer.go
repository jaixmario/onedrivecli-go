package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "runtime"
    "strings"
)

func Explorer() {
    currentPath := "/"
    for {
        clearScreen()
        items := ListExplorer(currentPath)
        if len(items) == 0 {
            fmt.Println("⚠️  No items found in:", currentPath)
        }

        fmt.Println("📂 Current Path:", currentPath)
        fmt.Println("---------------------------")
        for i, item := range items {
            if item.Folder != nil {
                fmt.Printf("[%d] 📁 %s\n", i+1, item.Name)
            } else {
                fmt.Printf("[%d] 📄 %s (%.2f MB)\n", i+1, item.Name, float64(item.Size)/1024/1024)
            }
        }
        fmt.Println("\n[0] ⬅️  Go Back  |  [q] ❌ Quit")

        fmt.Print("\nEnter choice: ")
        var choice string
        fmt.Scanln(&choice)

        if choice == "q" || choice == "Q" {
            fmt.Println("👋 Exiting Explorer...")
            break
        }

        if choice == "0" {
            if currentPath != "/" {
                idx := strings.LastIndex(strings.TrimSuffix(currentPath, "/"), "/")
                if idx > 0 {
                    currentPath = currentPath[:idx]
                } else {
                    currentPath = "/"
                }
            }
            continue
        }

        idx := 0
        _, err := fmt.Sscanf(choice, "%d", &idx)
        if err != nil || idx < 1 || idx > len(items) {
            fmt.Println("⚠️ Invalid choice.")
            continue
        }

        selected := items[idx-1]
        if selected.Folder != nil {
            if currentPath == "/" {
                currentPath += selected.Name
            } else {
                currentPath += "/" + selected.Name
            }
        } else {
            FileOptions(currentPath, selected)
        }
    }
}

func FileOptions(path string, file DriveItem) {
    for {
        clearScreen()
        fmt.Println("📄 Selected File:", file.Name)
        fmt.Println("----------------------------")
        fmt.Println("[1] 🔗 Generate Share Link")
        fmt.Println("[2] 📥 Generate Direct Download Link")
        fmt.Println("[b] ⬅️  Back")
        fmt.Println("[q] ❌ Quit Explorer")
        fmt.Print("\nEnter choice: ")

        var choice string
        fmt.Scanln(&choice)

        switch choice {
        case "1":
            GenerateShareLink(path + "/" + file.Name)
            fmt.Println("\nPress Enter to continue...")
            fmt.Scanln()
        case "2":
            GenerateDirectLink(path + "/" + file.Name)
            fmt.Println("\nPress Enter to continue...")
            fmt.Scanln()
        case "b", "B":
            return
        case "q", "Q":
            fmt.Println("👋 Exiting Explorer...")
            os.Exit(0)
        default:
            fmt.Println("⚠️ Invalid choice.")
        }
    }
}

func GenerateShareLink(filePath string) {
    link := GetShareLink(filePath)
    if link == "" {
        fmt.Println("❌ Could not generate share link.")
    } else {
        fmt.Println("🔗 Share Link:", link)
    }
}

func GenerateDirectLink(filePath string) {
    link := GetDirectDownloadLink(filePath)
    if link == "" {
        fmt.Println("❌ Could not generate direct download link.")
    } else {
        fmt.Println("📥 Direct Download Link:", link)
    }
}

func ListExplorer(path string) []DriveItem {
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
        return nil
    }
    defer resp.Body.Close()

    if resp.StatusCode == 401 {
        accessToken = GetAccessToken()
        req.Header.Set("Authorization", "Bearer "+accessToken)
        resp, err = http.DefaultClient.Do(req)
        if err != nil {
            fmt.Println("❌ HTTP request failed after refresh:", err)
            return nil
        }
        defer resp.Body.Close()
    }

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode >= 300 {
        fmt.Println("❌ Failed to list items. Response:")
        fmt.Println(string(body))
        return nil
    }

    var list struct {
        Value []DriveItem `json:"value"`
    }
    if err := json.Unmarshal(body, &list); err != nil {
        fmt.Println("❌ Failed to parse JSON:", err)
        fmt.Println(string(body))
        return nil
    }

    return list.Value
}

func clearScreen() {
    switch runtime.GOOS {
    case "windows":
        cmd := exec.Command("cmd", "/c", "cls")
        cmd.Stdout = os.Stdout
        cmd.Run()
    default:
        fmt.Print("\033[H\033[2J")
    }
}