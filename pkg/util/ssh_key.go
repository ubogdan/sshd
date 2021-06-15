package util

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func Fingerprint(k ssh.PublicKey) string {
	hash := md5.Sum(k.Marshal())
	r := fmt.Sprintf("% x", hash)
	return strings.Replace(r, " ", ":", -1)
}

func SshdAuthByGithub(user string, key ssh.PublicKey) error {
	publicKeys, err := fetchGithubPublicKeys(user)
	if err != nil {
		return err
	}
	for _, pbk := range publicKeys {
		if bytes.Equal(key.Marshal(), pbk.Marshal()) {
			return nil
		}
	}
	return fmt.Errorf("the key is not match any https://github.com/%s.keys", user)
}

func fetchGithubPublicKeys(githubUser string) ([]ssh.PublicKey, error) {
	keyURL := fmt.Sprintf("https://github.com/%s.keys", githubUser)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*15)
	defer cancelFunc()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, keyURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New("invalid response from github")
	}
	authorizedKeysBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body:%v", err)
	}
	var keys []ssh.PublicKey
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			return nil, errors.Wrap(err, "parsing key")
		}
		keys = append(keys, pubKey)
		authorizedKeysBytes = rest
	}
	return keys, nil
}
