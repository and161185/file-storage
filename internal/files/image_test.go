package files

import (
	"bytes"
	"errors"
	"file-storage/internal/errs"
	"file-storage/internal/imgproc"
	"image/color"
	"reflect"
	"testing"

	"github.com/disintegration/imaging"
)

func TestProcessImage(t *testing.T) {

	w := 100
	h := 100
	format := "jpeg"
	img := imaging.New(w, h, color.Black)
	imagingFormat, err := imaging.FormatFromExtension(format)
	if err != nil {
		t.Fatalf("test image format definition: %v", err)
	}

	b, err := imgproc.Encode(img, imagingFormat)
	if err != nil {
		t.Fatalf("test image creation error: %v", err)
	}

	bBadConfig := b[1:]
	bBadBody := b[:len(b)-2]

	table := []struct {
		name          string
		b             []byte
		targetExt     string
		targetWidth   int
		targetHeight  int
		wantb         []byte
		wantImageInfo *ImageInfo
		controlErr    bool
		wantErr       error
	}{
		{
			name:          "unsupportet target format",
			b:             b,
			targetExt:     "webp",
			targetWidth:   1,
			targetHeight:  1,
			wantb:         nil,
			wantImageInfo: nil,
			controlErr:    true,
			wantErr:       errs.ErrUnsupportedImageFormat,
		},
		{
			name:          "config observing error",
			b:             bBadConfig,
			targetExt:     "jpg",
			targetWidth:   100,
			targetHeight:  100,
			wantb:         nil,
			wantImageInfo: nil,
			controlErr:    false,
			wantErr:       errors.New(""),
		},
		{
			name:          "decode image error",
			b:             bBadBody,
			targetExt:     "bmp",
			targetWidth:   111,
			targetHeight:  100,
			wantb:         nil,
			wantImageInfo: nil,
			controlErr:    false,
			wantErr:       errors.New(""),
		},
		{
			name:          "want same image",
			b:             b,
			targetExt:     format,
			targetWidth:   w,
			targetHeight:  h,
			wantb:         b,
			wantImageInfo: &ImageInfo{Format: imgproc.ImgFormat(format), Width: w, Height: h},
			controlErr:    true,
			wantErr:       nil,
		},
		{
			name:          "resize & encode",
			b:             b,
			targetExt:     "bmp",
			targetWidth:   w / 2,
			targetHeight:  h / 2,
			wantb:         nil,
			wantImageInfo: &ImageInfo{Format: imgproc.ImgFormat("bmp"), Width: w / 2, Height: h / 2},
			controlErr:    true,
			wantErr:       nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			bresult, imgInfo, err := ProcessImage(tt.b, tt.targetExt, tt.targetWidth, tt.targetHeight)
			if tt.wantb != nil && !bytes.Equal(tt.wantb, bresult) {
				t.Errorf("bytes mismatch")
			}
			if tt.wantImageInfo != nil {
				if !reflect.DeepEqual(imgInfo, tt.wantImageInfo) {
					t.Errorf("image info mismatch got %v want %v", imgInfo, tt.wantImageInfo)
				}
			}
			if tt.wantErr != nil {
				if tt.controlErr {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("errors mismatch got %v want %v", err, tt.wantErr)
					}
				} else {
					if err == nil {
						t.Errorf("got nil want any error")
					}
				}
			}
		})
	}

}
