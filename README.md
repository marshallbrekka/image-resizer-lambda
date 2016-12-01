# image-resizer-lambda

Dynamic image resizing using AWS Cloudfront, API Gateway, Lambda, and S3.

This is more of a proof of concept at this point than production ready software.

## Usage

### Lambda Event

The **event** object must always contain a `key` property, which must map exactly to a key for an image in the target s3 bucket.

In addition to the key property, the Lambda function expects the **event** object to contain any number the following properties:
- `maxWidth`: Resizes the image to the desired width, maintaining the aspect ratio.
- `maxHeight`: Resizes the image to the desired height, maintaining the aspect ratio.
- `format`: specifies the desired output format, can be `jpeg` or `png`.

If both `maxWidth` and `maxHeight` are specified, the image is resized such that the aspect ratio is preserved and the resulting width and height are not greater than the values specified.

### Lambda Configuration

The Lambda function can also be configured through the following environment variables.

- `RESIZE_STRATEGY`: Can be one of `nearest-neighbor`, `bilinear`, `bicubic`, `mitchell-netravali`, `lanczos2`, `lanczos3`.
- `JPEG_COMPRESSION`: The compression level for jpeg formats, from 0-100.
- `PNG_COMPRESSION`: The compression level for png formats, can be one of `default`, `none`, `best-speed`, `best-compression`.
- `S3_BUCKET`: The name of the s3 bucket to read from.