package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/cognicraft/rest"
)

func main() {
	i := flag.Bool("i", false, "Include protocol response headers in the output")
	v := flag.Bool("v", false, "Make the operation more talkative")
	flag.Parse()

	c := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	for r := range rest.ParseAll(os.Stdin) {
		if *v {
			fmt.Printf("----- Request -----\n\n")
			req, _ := httputil.DumpRequest(r, true)
			fmt.Printf("%s\n", string(req))
		}
		res, err := c.Do(r)
		if *v {
			fmt.Printf("----- Response -----\n\n")
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}
		if *v || *i {
			resHeaders, _ := httputil.DumpResponse(res, false)
			fmt.Printf("%s", string(resHeaders))
		}
		resBody, _ := ioutil.ReadAll(res.Body)
		fmt.Printf("%s\n", string(resBody))
		res.Body.Close()
	}
}
