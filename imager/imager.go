// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"errors"
	"github.com/die-net/fotomat/vips"
)

var (
	ErrUnknownFormat = errors.New("Unknown image format")
	ErrTooBig        = errors.New("Image is too wide or tall")
	ErrTooSmall      = errors.New("Image is too small")
	ErrBadOption     = errors.New("Bad option specified")
)

const (
	minDimension = 2             // Avoid off-by-one divide-by-zero errors.
	maxDimension = (1 << 15) - 2 // Avoid signed int16 overflows.
)

type Imager struct {
	blob        []byte
	image       *vips.Image
	width       int
	height      int
	orientation Orientation
	format      Format
}

func New(blob []byte) (*Imager, error) {
	// Security: Limit formats we pass to VIPS to JPEG, PNG, GIF, WEBP.
	format := DetectFormat(blob)
	if format == UnknownFormat {
		return nil, ErrUnknownFormat
	}

	// Ask VIPS to parse metadata.
	image, err := format.Load(blob)
	if err != nil {
		return nil, ErrUnknownFormat
	}

	width := image.Xsize()
	height := image.Ysize()

	// Security: Confirm that image sizes are sane.
	if width < minDimension || height < minDimension {
		return nil, ErrTooSmall
	}
	if width > maxDimension || height > maxDimension {
		return nil, ErrTooBig
	}

	orientation := DetectOrientation(image)
	width, height = orientation.Dimensions(width, height)

	imager := &Imager{
		blob:        blob,
		image:       image,
		width:       width,
		height:      height,
		orientation: orientation,
		format:      format,
	}
	return imager, nil
}

func (imager *Imager) Thumbnail(options Options) ([]byte, error) {
	if err := options.Check(imager.format, imager.width, imager.height); err != nil {
		return nil, err
	}

	width := options.Width
	height := options.Height

	width, height = scaleAspect(imager.width, imager.height, width, height, options.Crop)

	result, err := imager.NewResult(width, height)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	if result.width > width || result.height > height {
		if err := result.Resize(width, height); err != nil {
			return nil, err
		}
	}

	return result.Get()
}

func (imager *Imager) Crop(options Options) ([]byte, error) {
	if err := options.Check(imager.format, imager.width, imager.height); err != nil {
		return nil, err
	}

	width := options.Width
	height := options.Height

	// If requested width or height are larger than original, scale
	// request down to fit within original dimensions.
	if width > imager.width || height > imager.height {
		width, height = scaleAspect(width, height, imager.width, imager.height, true)
	}

	// Figure out the intermediate size the original image would have to
	// be scaled to be cropped to requested size.
	iw, ih := scaleAspect(imager.width, imager.height, width, height, false)

	result, err := imager.NewResult(iw, ih)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	// If necessary, scale down to appropriate intermediate size.
	if result.width > iw || result.height > ih {
		if err := result.Resize(iw, ih); err != nil {
			return nil, err
		}
	}

	// If necessary, crop to fit exact size.
	if result.width > width || result.height > height {
		if err := result.Crop(width, height); err != nil {
			return nil, err
		}
	}

	return result.Get()
}

func (imager *Imager) Close() {
	*imager = Imager{}
}
