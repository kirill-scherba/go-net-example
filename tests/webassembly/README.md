
# "Hello world" WebAssembly sample

Create main.go:

    package main

    import (
      "fmt"
      "syscall/js"
    )

    func main() {
      fmt.Println("Hello, WebAssembly!")
      js.Global().Get("document").Call("getElementsByTagName", "body").
        Index(0).Set("innerHTML", "Hello, World!")
    }

Compile it for WebAssembly:

    GOOS=js GOARCH=wasm go build -o main.wasm

Copy the JavaScript support file:

    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

Create index.html:

    <html>
        <head>
            <meta charset="utf-8"/>
            <script src="wasm_exec.js"></script>
            <script>
                const go = new Go();
                WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
                    go.run(result.instance);
                });
            </script>
        </head>
        <body></body>
    </html>

Install goexec (if not installed yet):

    go get -u github.com/shurcooL/goexec

Serve the three files (index.html, wasm_exec.js, and main.wasm) from a web server, with goexec:

    goexec "http.ListenAndServe(\`:8080\`, http.FileServer(http.Dir(\`.\`)))"
