var child_process = require('child_process');

function resize(bucket, key, options, cb) {
  var args = [
    "--s3-bucket=" + bucket,
    "--s3-key=" + key
  ];
  if (options["max-width"]) {
    args.push("--max-width=" + options["max-width"]);
  }
  if (options["max-height"]) {
    args.push("--max-height=" + options["max-height"]);
  }
  if (options["format"]) {
    args.push("--format=" + options["format"]);
  }
  if (options["format"]) {
    args.push("--format=" + options["format"]);
  }
  if (options["resize-strategy"]) {
    args.push("--resize-strategy=" + options["resize-strategy"]);
  }
  if (options["jpeg-compression"]) {
    args.push("--jpeg-compression=" + options["jpeg-compression"]);
  }
  if (options["png-compression"]) {
    args.push("--png-compression=" + options["png-compression"]);
  }
  if (options["s3-read-method"]) {
    args.push("--s3-read-method=" + options["s3-read-method"]);
  }

  console.log("Starting image resize for key", key, options);
  var proc = child_process.spawn('./resizer.linux.x86', args, {
    stdio: ['pipe', 'pipe', process.stdout]
  });

  var imageChunks = [];
  var result;

  proc.stdout.on("data", function(chunk) {
    imageChunks.push(chunk);
  });

  proc.stdout.on('end', function(chunk) {
    console.log("Recieved end of data stream, combining");
    result = Buffer.concat(imageChunks);
    console.log("Combined chunks");
  });

  proc.on('close', function(code) {
    if(code === 0) {
      console.log("Processed ended successfully, base64 encoding.");
      var resultString = result.toString('base64');
      console.log("Base54 encoded, returning...");
      cb(null, resultString);
    } else {
      cb(code, null);
    }
  });
}

exports.handler = function(event, context, callback) {
  var options = {
    "max-width":        event["maxWidth"],
    "max-height":       event["maxHeight"],
    "format":           event["format"],
    "resize-stragegy":  process.env["RESIZE_STRATEGY"],
    "jpeg-compression": process.env["JPEG_COMPRESSION"],
    "png-compression":  process.env["PNG_COMPRESSION"],
    "s3-read-method":   process.env["S3_READ_METHOD"]
  };
  resize(process.env["S3_BUCKET"], event.key, options, function(err, img) {
    if (err !== null) {
      callback(err);
    } else {
      callback(null, img);
    }
  });
};
