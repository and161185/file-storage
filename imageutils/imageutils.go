package imageutils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp" // Подключаем поддержку BMP
	"golang.org/x/image/draw"

	"github.com/gabriel-vasile/mimetype"
)

// ResizeImage уменьшает изображение до заданных размеров с сохранением пропорций.
func ResizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	if originalWidth <= maxWidth && originalHeight <= maxHeight {
		return img
	}

	var scaleFactor float64
	if originalWidth > originalHeight {
		scaleFactor = float64(maxWidth) / float64(originalWidth)
	} else {
		scaleFactor = float64(maxHeight) / float64(originalHeight)
	}

	newWidth := int(float64(originalWidth) * scaleFactor)
	newHeight := int(float64(originalHeight) * scaleFactor)

	resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(resizedImg, resizedImg.Bounds(), img, img.Bounds(), draw.Over, nil)

	return resizedImg
}

func ConvertToJPEG(data []byte) ([]byte, error) {
	// Определяем MIME-тип данных
	mime := mimetype.Detect(data)
	if !mime.Is("image/jpeg") && !mime.Is("image/jpg") &&
		!mime.Is("image/png") && !mime.Is("image/bmp") {
		return nil, fmt.Errorf("недопустимый формат файла: %s", mime.String())
	}

	// Декодируем изображение
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования изображения: %v", err)
	}

	// Изменяем размер изображения
	img = ResizeImage(img, 1000, 1000)

	// Создаем буфер для сохранения JPEG-данных
	var outputBuffer bytes.Buffer
	if err := jpeg.Encode(&outputBuffer, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("ошибка кодирования JPEG: %v", err)
	}

	// Возвращаем JPEG-данные в виде байтов
	return outputBuffer.Bytes(), nil
}
