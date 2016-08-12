# flup

Command line parallelized flickr uploader built on top of [manki/flickgo](https://github.com/manki/flickgo).

# Environment

Works only on OSX.

# How to use

```
go run main.go $consumer_key $consumer_secret
```

It will launch your default browser to authorize the app.
After authorize the app, this program runs as http server and accepts file names to upload thorugh POST request.

`add2queue.sh` takes image file names and posts them to the http server. File names must be absolute paths. Following example shows easy way to upload all files in current directory.

```
find "`pwd`" -type f -print0  | xargs -0 ~/flup2016/add2queue.sh
```

The server marks successfully uploaded files with "Blue" tag so you can upload manually later even if the upload failed for some reason.


# Acknowledgement

Thank [morygonzalez](https://github.com/morygonzalez) for criticizing me for leaving my first golang program on my machine and making me publish it on github.

