package utils

import (
	"context"
	"os"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/joho/godotenv"
)

func UploadImageToCloudinary(filePath string) (string, error) {
	_ = godotenv.Load()
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os. Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return "", err
	}
	res, err := cld.Upload.Upload(context.Background(), filePath, uploader.UploadParams{})
	if err != nil {
		return "", err
	}
	return res.SecureURL, nil
}