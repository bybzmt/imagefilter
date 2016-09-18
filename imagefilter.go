package main

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"flag"
	"github.com/disintegration/imaging"
	"image"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"willnorris.com/go/gifresize"
)

var basedir = flag.String("dir", "./", "markdown files dir")
var addr = flag.String("addr", ":8080", "Listen addr:port")
var signatureKey = flag.String("signatureKey", "", "Signature Key")

var exts = map[string]int8{
	".png":  1,
	".jpg":  1,
	".jpeg": 1,
	".gif":  1,
	".webp": 1,
}

var mime = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
}

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filename := path.Join(*basedir, path.Clean(r.URL.Path))

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
			//http.ServeFile(w, r, filename)
			http.NotFound(w, r)
			return
		}

		ext := strings.ToLower(path.Ext(filename))
		if exts[ext] == 0 {
			http.ServeFile(w, r, filename)
			return
		}

		if !checkSign(w, r) {
			return
		}

		op := r.FormValue("o")
		width_s := r.FormValue("w")
		height_s := r.FormValue("h")
		format := r.FormValue("f")
		anchor_s := r.FormValue("a")

		//读取图片格式
		img_cfg, ori_format, err := image.DecodeConfig(fh)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if format == "" {
			switch ori_format {
			case "png":
				format = "png"
			case "jpeg":
				format = "jpeg"
			case "gif":
				format = "gif"
			default:
				format = "jpeg"
			}
		}

		//原图不转格式时不需要处理
		if op == "ori" && ori_format == format {
			http.ServeFile(w, r, filename)
			return
		}

		//裁切时定位参数
		var anchor imaging.Anchor
		switch anchor_s {
		case "center":
			anchor = imaging.Center
		case "topleft":
			anchor = imaging.TopLeft
		case "", "top":
			anchor = imaging.Top
		case "topright":
			anchor = imaging.TopRight
		case "left":
			anchor = imaging.Left
		case "right":
			anchor = imaging.Right
		case "bottomleft":
			anchor = imaging.BottomLeft
		case "bottom":
			anchor = imaging.Bottom
		case "bottomright":
			anchor = imaging.BottomRight
		default:
			http.Error(w, "unsupport anchor: "+anchor_s, 400)
			return
		}

		//宽高
		width, _ := strconv.Atoi(width_s)
		height, _ := strconv.Atoi(height_s)
		//为0时不改变宽高
		if width == 0 {
			width = img_cfg.Width
		}
		if height == 0 {
			height = img_cfg.Height
		}

		var trans func(img image.Image) image.Image

		switch op {
		case "ori":
			//原图不需要处理
			trans = func(img image.Image) image.Image {
				return img
			}
		case "resize":
			trans = func(img image.Image) image.Image {
				return imaging.Resize(img, width, height, imaging.Lanczos)
			}
		case "crop":
			trans = func(img image.Image) image.Image {
				return imaging.CropAnchor(img, width, height, anchor)
			}
		case "fit":
			trans = func(img image.Image) image.Image {
				return imaging.Fit(img, width, height, imaging.Lanczos)
			}
		case "fill":
			trans = func(img image.Image) image.Image {
				return imaging.Fill(img, width, height, anchor, imaging.Lanczos)
			}
		default:
			http.Error(w, "unsupport op: "+op, 400)
			return
		}

		//重置指针
		fh.Seek(0, os.SEEK_SET)

		//动画
		if ori_format == "gif" && format == "gif" {
			w.Header().Set("Content-Type", "image/gif")
			gifresize.Process(w, fh, trans)
			return
		}

		//非动画图
		img, ori_format, err := image.Decode(fh)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		img = trans(img)

		switch format {
		case "png":
			w.Header().Set("Content-Type", "image/png")
			imaging.Encode(w, img, imaging.PNG)
		case "jpg", "jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
			imaging.Encode(w, img, imaging.JPEG)
		case "gif":
			w.Header().Set("Content-Type", "image/gif")
			imaging.Encode(w, img, imaging.GIF)
		default:
			http.Error(w, "unsupport format: "+format, 400)
		}
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func checkSign(w http.ResponseWriter, r *http.Request) bool {
	if *signatureKey == "" {
		return true
	}

	randstr := r.FormValue("t")
	op := r.FormValue("o")
	width := r.FormValue("w")
	height := r.FormValue("h")
	format := r.FormValue("f")
	anchor := r.FormValue("a")
	sign_s := r.FormValue("s")

	if randstr == "" || op == "" || sign_s == "" {
		http.Error(w, "param empty", 400)
		return false
	}

	sign, err := base64.RawURLEncoding.DecodeString(sign_s)
	if err != nil {
		http.Error(w, "base64 error: " + err.Error(), 400)
		return false
	}

	msg := r.URL.Path + op + width + height + format + anchor + randstr

	if !CheckMAC([]byte(msg), sign) {
		http.Error(w, "sign error", 400)
		return false
	}

	return true;
}

func CheckMAC(message, messageMAC []byte) bool {
	mac := hmac.New(md5.New, []byte(*signatureKey))
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
