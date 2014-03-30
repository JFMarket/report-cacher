# Report Cacher

## Usage
```sh
report-cacher -interval=6h -site='https://jonesboroughfarmersmkt.shopkeepapp.com' \
-email='user@domain.com' -password='mypassword' -directory='cache' -port=8080
```

The above commands result in the reports being downloaded every 6 hours from the time the script starts until it is stopped. They are downloaded to the _cache_ directory and served from a web server on port 8080. They can be accessed at http://localhost:8080/. The email and password are to an account that has access to https://jonesboroughfarmersmkt.shopkeepapp.com.

Passing `-noweb` instead of `-port=8080` will result in the webserver being disabled. Thus, the files will only be accessible to applications on the local machine that have permission to read files in the _cache_ directory.

## Installation
### Source
`go get github.com/jfmarket/report-cacher`

### Binary
Copy the binary to a directory in your PATH.