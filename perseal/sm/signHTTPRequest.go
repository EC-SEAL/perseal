package sm

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spacemonkeygo/httpsig"
)

func signRequest(r *http.Request, headers map[string]string) (string, error) {
	//https://www.digitalocean.com/community/tutorials/openssl-essentials-working-with-ssl-certificates-private-keys-and-csrs#private-keys
	newReq, err := http.NewRequest(r.Method, r.URL.String(), nil)
	if err != nil {
		return "", err
	}

	heads := []string{}

	for k, v := range headers {
		newReq.Header.Set(k, v)
		heads = append(heads, strings.ToLower(k))
	}
	privContent, err := ioutil.ReadFile("./private.key")
	if err != nil {
		return "", err
	}
	privBlock, _ := pem.Decode(privContent)
	key, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		return "", err
	}

	rsaPubKey, err := ioutil.ReadFile("./public.pub")
	if err != nil {
		return "", err
	}

	b64dec, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(rsaPubKey)))
	if err != nil {
		return "", err
	}

	sha256sum := sha256.Sum256(b64dec)
	shahex := hex.EncodeToString(sha256sum[:])

	signer := httpsig.NewSigner(shahex, key, httpsig.RSASHA256, heads)
	err = signer.Sign(newReq)
	if err != nil {
		return "", err
	}

	return newReq.Header.Get("Authorization"), nil
}

func prepareRequestHeaders(req *http.Request, url string) (*http.Request, error) {
	headers := map[string]string{}
	method := req.Method
	reqTarget := method + " /" + strings.SplitN(strings.SplitN(url, "://", 2)[1], "/", 2)[1]
	host := strings.SplitN(strings.SplitN(url, "://", 2)[1], "/", 2)[0]

	var sha256value [32]byte
	//verifies request method to fomrulate Digest
	if req.Method == "POST" {
		y, err := req.GetBody()

		if err != nil {
			return nil, err
		}

		x, err := ioutil.ReadAll(y)

		if err != nil {
			return nil, err
		}

		sha256value = sha256.Sum256(x)
	} else if req.Method == "GET" {
		sha256value = sha256.Sum256([]byte{})
	}

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

	signature, err := signRequest(req, headers)
	if err != nil {
		return nil, err
	}

	headers["Authorization"] = signature

	req.Header.Set("Authorization", headers["Authorization"])
	return req, nil
}
