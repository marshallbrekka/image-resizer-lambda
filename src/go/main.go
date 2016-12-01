package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
	flag "github.com/ogier/pflag"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"net/http"
)

var resizeStrategies map[string]resize.InterpolationFunction = map[string]resize.InterpolationFunction{
	"nearest-neighbor":   resize.NearestNeighbor,
	"bilinear":           resize.Bilinear,
	"bicubic":            resize.Bicubic,
	"mitchell-netravali": resize.MitchellNetravali,
	"lanczos2":           resize.Lanczos2,
	"lanczos3":           resize.Lanczos3,
}

// Given an s3 bucket and s3 key, reads the key contents and
// returns it as an Image.
func readImage(s3Bucket, s3Key *string) (image.Image, error) {
	log.Info("Reading image from S3")
	if sess, err := session.NewSession(); err != nil {
		return nil, err
	} else {
		svc := s3.New(sess)
		output, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: s3Bucket,
			Key:    s3Key,
		})
		if err != nil {
			return nil, err
		} else {
			img, format, err := image.Decode(output.Body)
			if err != nil {
				log.Debug("Decoded image in bucket %s and key %s from format %s", s3Bucket, s3Key, format)
			}
			return img, err
		}
	}
}

func readImageFromUrl(url string) (image.Image, error) {
	if resp, err := http.Get(url); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		img, format, err := image.Decode(resp.Body)
		if err != nil {
			log.Debug("Decoded image at url %s from format %s", url, format)
		}
		return img, err
	}
}

func readImageHttps(s3Bucket, s3Key *string) (image.Image, error) {
	log.Info("Reading image from S3 with https")
	return readImageFromUrl("https://s3-us-west-2.amazonaws.com/" + *s3Bucket + "/" + *s3Key)
}

func readImageHttp(s3Bucket, s3Key *string) (image.Image, error) {
	log.Info("Reading image from S3 with http")
	return readImageFromUrl("http://s3-us-west-2.amazonaws.com/" + *s3Bucket + "/" + *s3Key)
}

func encodeJPEG(img image.Image, quality int, out io.Writer) error {
	log.Info("Encoding image as JPEG")
	return jpeg.Encode(out, img, &jpeg.Options{
		Quality: quality,
	})
}

func encodePNG(img image.Image, quality string, out io.Writer) error {
	log.Info("Encoding image as PNG")
	var compressionLevel png.CompressionLevel
	if quality == "none" {
		compressionLevel = png.NoCompression
	} else if quality == "best-speed" {
		compressionLevel = png.BestSpeed
	} else if quality == "best-compression" {
		compressionLevel = png.BestCompression
	} else {
		compressionLevel = png.DefaultCompression
	}
	encoder := &png.Encoder{CompressionLevel: compressionLevel}
	return encoder.Encode(out, img)
}

func resizeImage(maxWidth, maxHeight uint, img image.Image, strategy resize.InterpolationFunction) image.Image {
	if maxWidth == 0 && maxHeight == 0 {
		return img
	} else if maxWidth == 0 || maxHeight == 0 {
		log.Info("Resizing image with Resize")
		return resize.Resize(maxWidth, maxHeight, img, strategy)
	} else {
		log.Info("Resizing image with Thumbnail")
		return resize.Thumbnail(maxWidth, maxHeight, img, strategy)
	}
}

func main() {
	// If both are set to 0, then the image is returned in its original size.
	// If only one is set, then the image is scaled using it aspect ratio.
	// If both are set, the image is scaled using its aspect ratio to ensure that
	// both the width and height are less than their corrisponding max-* setting.
	var maxWidth *int = flag.Int("max-width", 0, "The maximum width to resize the image to")
	var maxHeight *int = flag.Int("max-height", 0, "The maximum height to resize the image to")
	var resizeStrategy *string = flag.String("resize-strategy", "bilinear", "nearest-neighbor, bilinear, bicubic")

	// Output format and compression
	var outFormat *string = flag.String("format", "jpeg", "Output format, can be jpeg or png")
	var jpegCompression *int = flag.Int("jpeg-compression", 70, "Jpeg compression quality, from 0 to 100")
	var pngCompression *string = flag.String("png-compression", "default", "Png compression, can be default, none, best-speed, best-compression")

	// Input location
	var s3Bucket *string = flag.String("s3-bucket", "", "The S3 bucket to read from")
	var s3Key *string = flag.String("s3-key", "", "The s3 image to resize")

	var readMethod *string = flag.String("s3-read-method", "authenticated", "authenticated, https, http")

	var verbose *bool = flag.BoolP("verbose", "v", false, "Set verbose logging")

	flag.Parse()

	// Output to stderr
	log.SetOutput(os.Stderr)

	// Setup verbose logging
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Ensure s3 bucket and key were specified
	if *s3Bucket == "" {
		log.Fatal("must specify s3-bucket parameter")
	}
	if *s3Key == "" {
		log.Fatal("must specify s3-key parameter")
	}

	// check resize strategy
	resizeFunc, ok := resizeStrategies[*resizeStrategy]
	if !ok {
		log.Fatalf("resize-strategy %s is not valid", *resizeStrategy)
	}

	// Read the image from s3
	var img image.Image
	var err error
	if *readMethod == "authenticated" {
		img, err = readImage(s3Bucket, s3Key)
	} else if *readMethod == "https" {
		img, err = readImageHttps(s3Bucket, s3Key)
	} else if *readMethod == "http" {
		img, err = readImageHttp(s3Bucket, s3Key)
	} else {
		log.Fatalf("Read method %s not supported", *readMethod)
	}

	if err != nil {
		log.Fatalf("Got error reading from s3: %s", err.Error())
	} else {
		resized := resizeImage(uint(*maxWidth), uint(*maxHeight), img, resizeFunc)
		var encodeError error
		if *outFormat == "jpeg" {
			encodeError = encodeJPEG(resized, *jpegCompression, os.Stdout)
		} else if *outFormat == "png" {
			encodeError = encodePNG(resized, *pngCompression, os.Stdout)
		}
		if encodeError != nil {
			log.Fatalf("Failed to encode image as %s: %s", *outFormat, encodeError.Error())
		} else {
			log.Info("Finished encoding image")
		}
	}
}
