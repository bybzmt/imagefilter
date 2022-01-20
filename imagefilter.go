package main

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"flag"
	"github.com/disintegration/imaging"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"willnorris.com/go/gifresize"

	_ "github.com/Kagami/go-avif"
	_ "github.com/biessek/golang-ico"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/oov/psd"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var basedir = flag.String("dir", "./", "local files dir")
var addr = flag.String("addr", ":http", "Listen addr:port")
var signatureKey = flag.String("signatureKey", "", "Signature Key")
var debug = flag.Bool("debug", false, "debug switch")
var proxy_pass = flag.String("proxy_pass", "", "proxy pass (default use local file)")

const op_ori = 1
const op_resize = 2
const op_crop = 3
const op_fit = 4
const op_fill = 5

const anchor_top_left = 1
const anchor_top = 2
const anchor_top_right = 3
const anchor_left = 4
const anchor_center = 5
const anchor_right = 6
const anchor_bottom_left = 7
const anchor_bottom = 8
const anchor_bottom_right = 9

const format_auto = 0
const format_jpeg = 1
const format_png = 2
const format_gif = 3

func main() {
	flag.Parse()

	if *proxy_pass == "" {
		*proxy_pass = strings.TrimRight(*proxy_pass, "/")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var query, ext string
		pn := strings.IndexByte(r.URL.Path, '.')
		if pn < 0 {
			query = r.URL.Path[1:]
		} else {
			query = r.URL.Path[1:pn]
			ext = r.URL.Path[pn:]
		}

		//log.Println(pn, query, ext)

		if len(query) < 9 {
			if *debug {
				http.Error(w, "data too short", 400)
			} else {
				http.NotFound(w, r)
			}
			return
		}

		raw, err := base64.RawURLEncoding.DecodeString(query)
		if err != nil {
			if *debug {
				http.Error(w, "base64 decode err"+err.Error(), 400)
			} else {
				http.NotFound(w, r)
			}
			return
		}

		if raw[0] != 1 {
			if *debug {
				http.Error(w, "protol version err", 400)
			} else {
				http.NotFound(w, r)
			}
			return
		}

		data, err := checkSign(raw[1:], []byte(ext))
		if err != nil {
			if *debug {
				http.Error(w, "sign err:"+err.Error(), 400)
			} else {
				http.NotFound(w, r)
			}
			return
		}

		op, anchor_s, format, width, height, file, err := decodePath(data)
		if err != nil {
			if *debug {
				http.Error(w, "decode err:"+err.Error(), 400)
			} else {
				http.NotFound(w, r)
			}
			return
		}

		var fr io.Reader

		if *proxy_pass == "" {
			filename := path.Join(*basedir, path.Clean(file))

			fh, err := os.Open(filename)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer fh.Close()

			fi, err := fh.Stat()
			if err != nil {
				http.NotFound(w, r)
				return
			}

			if fi.IsDir() {
				http.NotFound(w, r)
				return
			}

			fr = fh
		} else {
			resp, err := http.Get(*proxy_pass + "/" + path.Clean(file))
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer resp.Body.Close()

			fr = resp.Body
		}

		//原图不转格式时不需要处理
		if op == op_ori && format == format_auto {
			io.Copy(w, fr)
			return
		}

		br := bufio.NewReaderSize(fr, 1024*64)

		f1, err := br.Peek(1024 * 64)
		br2 := bytes.NewReader(f1)

		//读取图片格式
		img_cfg, ori_format, err := image.DecodeConfig(br2)
		if err != nil {
			http.Error(w, "DecodeConfig:"+err.Error(), 500)
			return
		}

		//为0时不改变宽高
		if width == 0 {
			width = img_cfg.Width
		}
		if height == 0 {
			height = img_cfg.Height
		}

		if format == format_auto {
			switch ori_format {
			case "png":
				format = format_png
			case "jpeg":
				format = format_jpeg
			case "gif":
				format = format_gif
			default:
				format = format_jpeg
			}
		}

		//裁切时定位参数
		var anchor imaging.Anchor
		switch anchor_s {
		case anchor_top_left:
			anchor = imaging.TopLeft
		case anchor_top:
			anchor = imaging.Top
		case anchor_top_right:
			anchor = imaging.TopRight
		case anchor_left:
			anchor = imaging.Left
		case anchor_center:
			anchor = imaging.Center
		case anchor_right:
			anchor = imaging.Right
		case anchor_bottom_left:
			anchor = imaging.BottomLeft
		case anchor_bottom:
			anchor = imaging.Bottom
		case anchor_bottom_right:
			anchor = imaging.BottomRight
		default:
			http.Error(w, "unsupport anchor", 400)
			return
		}

		var trans func(img image.Image) image.Image

		switch op {
		case op_ori:
			//原图不需要处理
			trans = func(img image.Image) image.Image {
				return img
			}
		case op_resize:
			trans = func(img image.Image) image.Image {
				return imaging.Resize(img, width, height, imaging.Lanczos)
			}
		case op_crop:
			trans = func(img image.Image) image.Image {
				return imaging.CropAnchor(img, width, height, anchor)
			}
		case op_fit:
			trans = func(img image.Image) image.Image {
				return imaging.Fit(img, width, height, imaging.Lanczos)
			}
		case op_fill:
			trans = func(img image.Image) image.Image {
				return imaging.Fill(img, width, height, anchor, imaging.Lanczos)
			}
		default:
			if *debug {
				http.Error(w, "unsupport op", 400)
			} else {
				http.NotFound(w, r)
			}
			return
		}

		//动画
		if format == format_gif && ori_format == "gif" {
			w.Header().Set("Content-Type", "image/gif")
			gifresize.Process(w, br, trans)
			return
		}

		//非动画图
		img, ori_format, err := image.Decode(br)
		if err != nil {
			if *debug {
				http.Error(w, "image decode err: "+err.Error(), 500)
			} else {
				http.NotFound(w, r)
			}
			return
		}
		img = trans(img)

		switch format {
		case format_png:
			w.Header().Set("Content-Type", "image/png")
			imaging.Encode(w, img, imaging.PNG)
		case format_jpeg:
			w.Header().Set("Content-Type", "image/jpeg")
			imaging.Encode(w, img, imaging.JPEG)
		case format_gif:
			w.Header().Set("Content-Type", "image/gif")
			imaging.Encode(w, img, imaging.GIF)
		default:
			if *debug {
				http.Error(w, "unsupport format", 400)
			} else {
				http.NotFound(w, r)
			}
			return
		}
	})

	log.Println("running...")
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func checkSign(raw, ext []byte) ([]byte, error) {
	if len(raw) < 2 {
		return nil, errors.New("data too shart")
	}

	if raw[0] == 0 {
		if *signatureKey == "" {
			return raw[1:], nil
		} else {
			return nil, errors.New("sign empty")
		}
	}

	sign_len := int(raw[0])

	if len(raw) < 1+sign_len {
		return nil, errors.New("sign data too shart")
	}

	mac := raw[1 : 1+sign_len]
	data := raw[1+sign_len:]

	if !CheckMAC(append(data, ext...), mac) {
		return nil, errors.New("sign not eq")
	}

	return data, nil
}

func CheckMAC(message, messageMAC []byte) bool {
	mac := hmac.New(md5.New, []byte(*signatureKey))
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func decodePath(raw []byte) (op, anchor, format, width, height int, file string, err error) {
	if len(raw) < 6 {
		err = errors.New("data too short")
		return
	}

	op = int(raw[0] >> 4)
	anchor = int(raw[0] & 0xf)

	format = int(raw[1])

	width = int(raw[2])<<8 | int(raw[3])
	height = int(raw[4])<<8 | int(raw[5])

	file = string(raw[6:])

	return
}
