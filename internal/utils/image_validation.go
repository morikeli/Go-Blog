package utils

// The blank imports are important because they register the image decoders.
import (
	"errors"
	"image"
	"io"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

const (
	MaxProfileImageSize = 5 << 20 // 5 MiB

	MaxProfileImageWidth  = 4096
	MaxProfileImageHeight = 4096

	MaxProfileImagePixels = 16_000_000
)

var (
	ErrImageTooLarge    = errors.New("Image is too large!")
	ErrUnsupportedImage = errors.New("Unsupported image type!")
	ErrInvalidImage     = errors.New("Invalid image provided!")
	ErrImageDimensions  = errors.New("Image dimensions are not allowed!")
	ErrImageDecode      = errors.New("Unable to decode image!")
)

type ImageInfo struct {
	ContentType string
	Width       int
	Height      int
}

func ValidateProfileImage(file io.ReadSeeker, size int64) (ImageInfo, error) {
	if size <= 0 {
		return ImageInfo{}, ErrInvalidImage
	}

	if size > MaxProfileImageSize {
		return ImageInfo{}, ErrImageTooLarge
	}

	// Read enough bytes to determine the actual file type.
	header := make([]byte, 512)

	n, err := io.ReadFull(file, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return ImageInfo{}, ErrInvalidImage
	}

	header = header[:n]

	contentType := detectImageType(header)

	if contentType == "" {
		return ImageInfo{}, ErrUnsupportedImage
	}

	// Reset the file before decoding.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return ImageInfo{}, ErrInvalidImage
	}

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return ImageInfo{}, ErrImageDecode
	}

	if !isAllowedImageFormat(format) {
		return ImageInfo{}, ErrUnsupportedImage
	}

	if config.Width <= 0 || config.Height <= 0 {
		return ImageInfo{}, ErrImageDimensions
	}

	if config.Width > MaxProfileImageWidth ||
		config.Height > MaxProfileImageHeight {

		return ImageInfo{}, ErrImageDimensions
	}

	pixels := int64(config.Width) * int64(config.Height)

	if pixels > MaxProfileImagePixels {
		return ImageInfo{}, ErrImageDimensions
	}

	// Reset so Cloudinary receives the complete file.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return ImageInfo{}, ErrInvalidImage
	}

	return ImageInfo{
		ContentType: contentType,
		Width:       config.Width,
		Height:      config.Height,
	}, nil
}

func detectImageType(data []byte) string {
	if len(data) >= 3 &&
		data[0] == 0xFF &&
		data[1] == 0xD8 &&
		data[2] == 0xFF {

		return "image/jpeg"
	}

	if len(data) >= 8 &&
		data[0] == 0x89 &&
		data[1] == 0x50 &&
		data[2] == 0x4E &&
		data[3] == 0x47 &&
		data[4] == 0x0D &&
		data[5] == 0x0A &&
		data[6] == 0x1A &&
		data[7] == 0x0A {

		return "image/png"
	}

	// RIFF....WEBP
	if len(data) >= 12 &&
		string(data[0:4]) == "RIFF" &&
		string(data[8:12]) == "WEBP" {

		return "image/webp"
	}

	return ""
}

func isAllowedImageFormat(format string) bool {
	switch format {
	case "jpeg", "png", "webp":
		return true
	default:
		return false
	}
}
