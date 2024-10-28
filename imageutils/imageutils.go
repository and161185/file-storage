package imageutils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	//_ "image/png"

	"golang.org/x/image/bmp"
	//	_ "golang.org/x/image/bmp" // Подключаем поддержку BMP
	"golang.org/x/image/draw"

	"file-storage/models"

	"github.com/gabriel-vasile/mimetype"
)

type ImageParams = models.QueryParams

// ResizeImage уменьшает изображение до заданных размеров с сохранением пропорций.
func ResizeImage(img image.Image, maxWidth, maxHeight int) image.Image {

	if maxWidth == 0 && maxHeight == 0 {
		return img
	}

	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	if originalWidth <= maxWidth && originalHeight <= maxHeight {
		return img
	}

	var scaleFactor float64
	if originalWidth > originalHeight && maxWidth != 0 {
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

func ConvertImageFromJPEG(data []byte, imageParams ImageParams) ([]byte, error) {

	if imageParams.Ext == "" || imageParams.Ext == "jpg" || imageParams.Ext == "jpeg" {
		if imageParams.Heigth == 0 && imageParams.Width == 0 {
			return data, nil
		}
	}
	// Проверяем, что входные данные являются JPEG
	mime := mimetype.Detect(data)
	if !mime.Is("image/jpeg") && !mime.Is("image/jpg") {
		return nil, fmt.Errorf("входные данные не являются JPEG: %s", mime.String())
	}

	// Декодируем JPEG-изображение
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования JPEG: %v", err)
	}

	img = ResizeImage(img, imageParams.Width, imageParams.Heigth)

	// Создаем буфер для сохранения данных в новом формате
	var outputBuffer bytes.Buffer

	// Кодируем изображение в указанный формат
	switch imageParams.Ext {
	case "", "jpg", "jpeg":
		if err := jpeg.Encode(&outputBuffer, img, nil); err != nil {
			return nil, fmt.Errorf("ошибка кодирования JPEG: %v", err)
		}
	case "png":
		if err := png.Encode(&outputBuffer, img); err != nil {
			return nil, fmt.Errorf("ошибка кодирования PNG: %v", err)
		}
	case "bmp":
		if err := bmp.Encode(&outputBuffer, img); err != nil {
			return nil, fmt.Errorf("ошибка кодирования BMP: %v", err)
		}
	default:
		return nil, fmt.Errorf("неподдерживаемый формат: %s", imageParams.Ext)
	}

	// Возвращаем данные в новом формате
	return outputBuffer.Bytes(), nil
}
