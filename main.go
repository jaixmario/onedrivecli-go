package main

import (
    "fmt"
    "log"
    "os"
)

// Entry point
func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: onedrivecli <command> [args]")
        fmt.Println("Commands:")
        fmt.Println("  auth                    Login via device code")
        fmt.Println("  ls <path>               List files/folders in OneDrive")
        fmt.Println("  link <path>             Generate a share link")
        fmt.Println("  dl <path>               Generate a direct download link")
        fmt.Println("  download <remote> <local> Download a file or folder with progress")
        fmt.Println("  upload <remote> <local> Upload a file or folder with progress")
        fmt.Println("  storage                 Check OneDrive storage usage")
        fmt.Println("  explorer                Interactive OneDrive explorer")
        return
    }

    switch os.Args[1] {
    case "auth":
        DeviceLogin()

    case "ls":
        if len(os.Args) < 3 {
            fmt.Println("Usage: onedrivecli ls <path>")
            return
        }
        ListFiles(os.Args[2])

    case "link":
        if len(os.Args) < 3 {
            fmt.Println("Usage: onedrivecli link <path>")
            return
        }
        link := GetShareLink(os.Args[2]) // from link.go
        if link == "" {
            fmt.Println("❌ Failed to generate share link.")
        } else {
            fmt.Println("🔗 Share Link:", link)
        }

    case "dl":
        if len(os.Args) < 3 {
            fmt.Println("Usage: onedrivecli dl <path>")
            return
        }
        link := GetDirectDownloadLink(os.Args[2])
        if link == "" {
            fmt.Println("❌ Failed to generate direct download link.")
        } else {
            fmt.Println("⬇️ Direct Download Link:", link)
        }

    case "download":
        if len(os.Args) < 4 {
            fmt.Println("Usage: onedrivecli download <remote_path_or_id> <local_path>")
            return
        }
        remote := os.Args[2]
        local := os.Args[3]
        if err := StartDownload(remote, local); err != nil {
            log.Fatal("Download failed:", err)
        }

    case "upload":
        if len(os.Args) < 4 {
            fmt.Println("Usage: onedrivecli upload <remote_path> <local_path>")
            return
        }
        remote := os.Args[2]
        local := os.Args[3]
        if err := StartUpload(remote, local); err != nil {
            log.Fatal("Upload failed:", err)
        }

    case "storage":
        CheckStorage()

    case "explorer":
        Explorer()

    default:
        fmt.Println("Unknown command:", os.Args[1])
    }
}
