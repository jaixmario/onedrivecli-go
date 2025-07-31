package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: onedrivecli <command> [args]")
        fmt.Println("Commands:")
        fmt.Println("  auth       - Login and save token")
        fmt.Println("  ls [path]  - List files and folders (default: /)")
        fmt.Println("  dl <path>  - Generate direct download link")
        fmt.Println("  link <path>- Create shareable OneDrive link")
        fmt.Println("  storage    - Check OneDrive storage usage")
        fmt.Println("  explorer   - Interactive OneDrive explorer")
        return
    }

    switch os.Args[1] {
    case "auth":
        DeviceLogin()

    case "ls":
        path := "/"
        if len(os.Args) > 2 {
            path = os.Args[2]
        }
        ListFiles(path)

    case "dl":
        if len(os.Args) < 3 {
            fmt.Println("Usage: onedrivecli dl <path>")
            return
        }
        DownloadFile(os.Args[2])

    case "link":
        if len(os.Args) < 3 {
            fmt.Println("Usage: onedrivecli link <path>")
            return
        }
        fmt.Println(GetShareLink(os.Args[2]))

    case "storage":
        CheckStorage()

    case "explorer":
        Explorer()

    default:
        fmt.Println("Unknown command:", os.Args[1])
    }
}