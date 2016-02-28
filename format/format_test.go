package format

import (
	"fmt"
	"github.com/die-net/fotomat/vips"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

const (
	TestdataPath = "../testdata/"
)

func TestMain(m *testing.M) {
	vips.Initialize()
	vips.LeakSet(true)
	r := m.Run()
	vips.ThreadShutdown()
	vips.Shutdown()
	os.Exit(r)
}

func TestMetadataValidation(t *testing.T) {
	// Return ErrUnknownFormat on a text file.
	assert.Equal(t, metadataError("notimage.txt"), ErrUnknownFormat)

	// Return ErrUnknownFormat on a truncated image.
	assert.Equal(t, metadataError("bad.jpg"), ErrUnknownFormat)

	// Load a 2x3 pixel image of each type.
	assert.Nil(t, isSize("2px.jpg", Jpeg, 2, 3))
	assert.Nil(t, isSize("2px.png", Png, 2, 3))
	assert.Nil(t, isSize("2px.gif", Gif, 2, 3))
	assert.Nil(t, isSize("2px.webp", Webp, 2, 3))
}

func metadataError(filename string) error {
	_, err := MetadataBytes(image(filename))
	return err
}

func image(filename string) []byte {
	bytes, err := ioutil.ReadFile(TestdataPath + filename)
	if err != nil {
		panic(err)
	}

	return bytes
}

func isSize(filename string, f Format, width, height int) error {
	m, err := MetadataBytes(image(filename))
	if err != nil {
		return err
	}
	if m.Width != width || m.Height != height {
		return fmt.Errorf("Got %dx%d != want %dx%d", m.Width, m.Height, width, height)
	}
	if m.Format != f {
		return fmt.Errorf("Format %s!=%s", m.Format, f)
	}
	return nil
}

func TestFormatOrientation(t *testing.T) {
	for i := 1; i <= 8; i++ {
		filename := "orient" + strconv.Itoa(i) + ".jpg"

		m, err := MetadataBytes(image(filename))
		if assert.Nil(t, err) {
			assert.Equal(t, m.Width, 48)
			assert.Equal(t, m.Height, 80)
		}
	}
}

func BenchmarkMetadataJpeg_2(b *testing.B) {
	benchMetadata(b, "2px.jpg", Jpeg)
}

func BenchmarkMetadataPng_2(b *testing.B) {
	benchMetadata(b, "2px.png", Png)
}

func BenchmarkMetadataWebp_2(b *testing.B) {
	benchMetadata(b, "2px.webp", Webp)
}

func BenchmarkMetadataJpeg_256(b *testing.B) {
	benchMetadata(b, "flowers.png", Jpeg)
}

func BenchmarkMetadataPng_256(b *testing.B) {
	benchMetadata(b, "flowers.png", Png)
}

func BenchmarkMetadataWebp_256(b *testing.B) {
	benchMetadata(b, "flowers.png", Webp)
}

func BenchmarkMetadataJpeg_536(b *testing.B) {
	benchMetadata(b, "watermelon.jpg", Jpeg)
}

func BenchmarkMetadataPng_536(b *testing.B) {
	benchMetadata(b, "watermelon.jpg", Png)
}

func BenchmarkMetadataWebp_536(b *testing.B) {
	benchMetadata(b, "watermelon.jpg", Webp)
}

func BenchmarkMetadataJpeg_3000(b *testing.B) {
	benchMetadata(b, "3000px.png", Jpeg)
}

func BenchmarkMetadataPng_3000(b *testing.B) {
	benchMetadata(b, "3000px.png", Png)
}

func BenchmarkMetadataWebp_3000(b *testing.B) {
	benchMetadata(b, "3000px.png", Webp)
}

func benchMetadata(b *testing.B, filename string, format Format) {
	blob := convert(filename, format)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := MetadataBytes(blob)
			assert.Nil(b, err)
		}
	})
}

func BenchmarkLoadJpeg_2(b *testing.B) {
	benchLoad(b, "2px.jpg", Jpeg)
}

func BenchmarkLoadPng_2(b *testing.B) {
	benchLoad(b, "2px.png", Png)
}

func BenchmarkLoadWebp_2(b *testing.B) {
	benchLoad(b, "2px.webp", Webp)
}

func BenchmarkLoadJpeg_256(b *testing.B) {
	benchLoad(b, "flowers.png", Jpeg)
}

func BenchmarkLoadPng_256(b *testing.B) {
	benchLoad(b, "flowers.png", Png)
}

func BenchmarkLoadWebp_256(b *testing.B) {
	benchLoad(b, "flowers.png", Webp)
}

func BenchmarkLoadJpeg_536(b *testing.B) {
	benchLoad(b, "watermelon.jpg", Jpeg)
}

func BenchmarkLoadPng_536(b *testing.B) {
	benchLoad(b, "watermelon.jpg", Png)
}

func BenchmarkLoadWebp_536(b *testing.B) {
	benchLoad(b, "watermelon.jpg", Webp)
}

func BenchmarkLoadJpeg_3000(b *testing.B) {
	benchLoad(b, "3000px.png", Jpeg)
}

func BenchmarkLoadPng_3000(b *testing.B) {
	benchLoad(b, "3000px.png", Png)
}

func BenchmarkLoadWebp_3000(b *testing.B) {
	benchLoad(b, "3000px.png", Webp)
}

func benchLoad(b *testing.B, filename string, format Format) {
	blob := convert(filename, format)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			img, err := format.LoadBytes(blob)
			if assert.Nil(b, err) {
				// Images are demand loaded. Actually decode all of the pixels.
				assert.Nil(b, img.Write())

				img.Close()
			}
		}
	})
}

func convert(filename string, of Format) []byte {
	blob := image(filename)
	format := DetectFormat(blob)
	if format == of {
		return blob
	}

	img, err := format.LoadBytes(blob)
	if err != nil {
		panic(err)
	}
	defer img.Close()

	blob, err = Save(img, SaveOptions{Format: of})
	if err != nil {
		panic(err)
	}
	return blob
}

func BenchmarkSaveJpeg_2(b *testing.B) {
	benchSave(b, "2px.jpg", SaveOptions{Format: Jpeg})
}

func BenchmarkSavePng_2(b *testing.B) {
	benchSave(b, "2px.png", SaveOptions{Format: Png})
}

func BenchmarkSaveWebp_2(b *testing.B) {
	benchSave(b, "2px.webp", SaveOptions{Format: Webp})
}

func BenchmarkSaveJpeg_256(b *testing.B) {
	benchSave(b, "flowers.png", SaveOptions{Format: Jpeg})
}

func BenchmarkSavePng_256(b *testing.B) {
	benchSave(b, "flowers.png", SaveOptions{Format: Png})
}

func BenchmarkSaveWebp_256(b *testing.B) {
	benchSave(b, "flowers.png", SaveOptions{Format: Webp})
}

func BenchmarkSaveJpeg_536(b *testing.B) {
	benchSave(b, "watermelon.jpg", SaveOptions{Format: Jpeg})
}

func BenchmarkSavePng_536(b *testing.B) {
	benchSave(b, "watermelon.jpg", SaveOptions{Format: Png})
}

func BenchmarkSaveWebp_536(b *testing.B) {
	benchSave(b, "watermelon.jpg", SaveOptions{Format: Webp})
}

func BenchmarkSaveJpeg_3000(b *testing.B) {
	benchSave(b, "3000px.png", SaveOptions{Format: Jpeg})
}

func BenchmarkSavePng_3000(b *testing.B) {
	benchSave(b, "3000px.png", SaveOptions{Format: Png})
}

func BenchmarkSaveWebp_3000(b *testing.B) {
	benchSave(b, "3000px.png", SaveOptions{Format: Webp})
}

func benchSave(b *testing.B, filename string, so SaveOptions) {
	blob := image(filename)
	format := DetectFormat(blob)
	img, err := format.LoadBytes(blob)
	if !assert.Nil(b, err) || !assert.Nil(b, img.Write()) {
		return
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := Save(img, so)
			assert.Nil(b, err)
		}
	})

	img.Close()
}