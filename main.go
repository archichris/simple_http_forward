package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	ca     = flag.String("ca", "ca.cer", "ca file")
	key    = flag.String("key", "key.pem", "client key file")
	cert   = flag.String("cert ", "server.cer", "client cert file")
	target = flag.String("target", "19.1.45.12:4443", "target url")
	local  = flag.String("local", "0.0.0.0:26443", "target url")
)

type proxy struct {
	cli *http.Client
}

func (p *proxy) Init() {
	_ = flag.CommandLine.Parse(os.Args[1:])
	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(*ca)
	if err != nil {
		log.Fatal("read ca.crt file error:", err.Error())
	}
	pool.AppendCertsFromPEM(caCrt)
	cliCrt, err := tls.LoadX509KeyPair(*cert, *key)
	if err != nil {
		log.Fatalln("LoadX509KeyPair error:", err.Error())
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{cliCrt},
		},
	}
	p.cli = &http.Client{Transport: tr}
	http.HandleFunc("/", p.proc)
}

func (p *proxy) proc(w http.ResponseWriter, r *http.Request) {
	// step 1
	fmt.Println("go proxy.proc")
	fmt.Printf("%+v\n", *r)
	outReq := new(http.Request)
	*outReq = *r // this only does shallow copies of maps
	outReq.URL.Host = *target
	outReq.URL.Scheme = "https"
	outReq.Host = *target
	outReq.RequestURI = ""

	resp, err := p.cli.Do(outReq)
	fmt.Printf("%+v\n", resp)
	if err != nil {
		log.Println(err.Error())
	}

	for key, value := range resp.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
	resp.Body.Close()
}

func main() {
	p := proxy{}
	p.Init()
	err := http.ListenAndServe(*local, nil)
	if err != nil {
		log.Fatal(errors.New("Init failed ...."))
	}
	// http.HandleFunc("/")
	// resp, err := client.Get("https://127.0.0.1:8080/")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))
}
