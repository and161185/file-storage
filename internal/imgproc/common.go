package imgproc

type ImgFormat string

const (
	ImgFormatBMP  ImgFormat = "bmp"
	ImgFormatJPG  ImgFormat = "jpg"
	ImgFormatPNG  ImgFormat = "png"
	ImgFormatGIF  ImgFormat = "gif"
	ImgFormatTIFF ImgFormat = "tiff"
	ImgFormatWEBP ImgFormat = "webp"
)

var supportedFormats = map[ImgFormat]string{
	ImgFormatBMP:  "image/bmp",
	ImgFormatJPG:  "image/jpeg",
	ImgFormatPNG:  "image/png",
	ImgFormatGIF:  "image/gif",
	ImgFormatTIFF: "image/tiff",
	ImgFormatWEBP: "image/webp",
}
