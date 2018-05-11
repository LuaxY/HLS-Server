package middleware

import (
    "crypto/aes"
    "crypto/cipher"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "regexp"
    "strconv"

    "github.com/gorilla/mux"
)

type LinkInfo struct {
    ID      int    `json:"id"`
    Season  int    `json:"season"`
    Episode int    `json:"episode"`
    IP      string `json:"ip"`
    Expire  string `json:"expire"`
}

func AES(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)

        if vars["token"] != "" {
            var linkInfo LinkInfo

            key, _ := hex.DecodeString(cfg.AES)
            keyDecoded, _ := hex.DecodeString(vars["token"])
            data, _ := decrypt(key, keyDecoded)
            json.Unmarshal(data, &linkInfo)

            vars["id"] = strconv.FormatUint(uint64(linkInfo.ID), 10)
            vars["season"] = strconv.FormatUint(uint64(linkInfo.Season), 10)
            vars["episode"] = strconv.FormatUint(uint64(linkInfo.Episode), 10)

            re := regexp.MustCompile(`^\/(tv|movie)\/[a-zA-Z0-9]+\/(.*)`)
            r.RequestURI = re.ReplaceAllString(r.RequestURI, "/"+vars["type"]+"/"+vars["id"]+"/$2")

            mux.SetURLVars(r, vars)
        }

        next.ServeHTTP(w, r)
    })
}

func decrypt(key, encrypted []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)

    if err != nil {
        return nil, err
    }

    iv := make([]byte, aes.BlockSize)
    decrypted := make([]byte, len(encrypted))
    stream := cipher.NewCTR(block, iv)
    stream.XORKeyStream(decrypted, encrypted)

    return decrypted, nil
}
