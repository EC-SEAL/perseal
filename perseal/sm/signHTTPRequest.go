package sm

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spacemonkeygo/httpsig"
)

func signRequest(r *http.Request, headers map[string]string) string {
	//https://www.digitalocean.com/community/tutorials/openssl-essentials-working-with-ssl-certificates-private-keys-and-csrs#private-keys
	newReq, err := http.NewRequest(r.Method, r.URL.String(), nil)

	heads := []string{}

	for k, v := range headers {
		newReq.Header.Set(k, v)
		heads = append(heads, strings.ToLower(k))
	}
	privContent, err := ioutil.ReadFile("./private.key")
	if err != nil {
		log.Fatal(err)
	}
	privBlock, _ := pem.Decode(privContent)
	key, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	rsaPubKey, err := ioutil.ReadFile("./public.pub")
	if err != nil {
		fmt.Println(err)
	}
	b64dec, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(rsaPubKey)))
	if err != nil {
		fmt.Println(err)
	}
	sha256sum := sha256.Sum256(b64dec)
	shahex := hex.EncodeToString(sha256sum[:])

	signer := httpsig.NewSigner(shahex, key, httpsig.RSASHA256, heads)
	err = signer.Sign(newReq)
	if err != nil {
		log.Printf("Signature failed: %s", err)
	}
	log.Println(newReq)
	return newReq.Header.Get("Authorization")
}

func prepareRequestHeaders(req *http.Request, url string) *http.Request {
	headers := map[string]string{}
	method := req.Method
	reqTarget := method + " /" + strings.SplitN(strings.SplitN(url, "://", 2)[1], "/", 2)[1]
	host := strings.SplitN(strings.SplitN(url, "://", 2)[1], "/", 2)[0]

	sha256value := sha256.Sum256([]byte{})
	slicedSHA := sha256value[:]
	slicedSHAbase64 := base64.StdEncoding.EncodeToString(slicedSHA)
	const longForm = "Mon, 2 Jan 2006 15:04:05 MST"

	req.Header.Set("Host", host)
	req.Header.Set("Original-Date", time.Now().Format(longForm))
	req.Header.Set("digest", "SHA-256="+slicedSHAbase64)
	req.Header.Set("X-Request-ID", uuid.New().String())
	req.Header.Set("Request-Target", reqTarget)

	headers["Host"] = host
	headers["Original-Date"] = time.Now().Format(longForm)
	headers["digest"] = "SHA-256=" + slicedSHAbase64
	headers["X-Request-ID"] = req.Header.Get("X-Request-ID")
	headers["(request-target)"] = reqTarget
	headers["Authorization"] = signRequest(req, headers)

	req.Header.Set("Authorization", headers["Authorization"])
	return req
}
