package main


import (
    "fmt"
    "github.com/minq3010/Backend-React-Native-App/utils"
)

func main() {
    url, err := utils.UploadImageToCloudinary("static/default_avatar.png")
    if err != nil {
        fmt.Println("Upload failed:", err)
        return
    }
    fmt.Println("Image URL:", url)
}