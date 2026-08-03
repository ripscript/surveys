package routes

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	pb "backend/siccore/pb"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// List of dangerous content keywords
var dangerousContentKeywords = []string{
	"<script>", "<?php", "import ", "require ", "def ", "function ", "console.log", "eval(", "exec(", "alert(",
}

// List of dangerous file extensions
var dangerousExtensions = []string{
	".exe", ".sh", ".bat", ".js", ".php", ".py", ".pl", ".rb", ".jar", ".vb", ".go",
}

func HandleFunc(c echo.Context, service string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi kesalahan", r)
		}
	}()

	path := c.Path()
	method := c.Request().Method

	// mengambil slug
	slugs := make(map[string]string)
	for _, name := range c.ParamNames() {
		slugs[name] = c.Param(name)
	}

	// Marshaling map ke JSON
	slug, err := json.Marshal(slugs)
	if err != nil {
		// Handle error
		return err
	}
	// Mengambil param dari URL
	queryString := c.QueryString()
	param, _ := url.QueryUnescape(queryString)

	isInWhitelist := IsInWhitelist(path, method)

	// Jika permintaan ada di dalam whitelist, set isSecure = false
	var isSecure bool
	if isInWhitelist {
		isSecure = false
	} else {
		isSecure = true
	}

	var payloadBytes []byte

	if service == "DOCAPI" || service == "WEBCHAT" || service == "TRX" || service == "USERAPI" {
		file, err := c.FormFile("file")
		if err == nil {
			payload := make(map[string]interface{})

			// Bind the payload
			if err := c.Bind(&payload); err != nil {
				return err
			}

			// Open the file
			src, err := file.Open()
			if err != nil {
				fmt.Println("err ", err)
				return err
			}
			defer src.Close()

			// Read the file into a byte array
			fileBytes, err := io.ReadAll(src)
			if err != nil {
				fmt.Println("err ", err)
				return err
			}

			// Get the file name
			fileName := file.Filename

			fileExtension := filepath.Ext(fileName)

			if isDangerousFile(fileExtension) {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"data":    "",
					"message": "File tidak diperbolehkan.",
					"success": false,
					"code":    http.StatusBadRequest,
				})
			}

			// Check if the file content is dangerous
			fileContent := string(fileBytes)
			if isDangerousContent(fileContent) {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"data":    "",
					"message": "File tidak diperbolehkan",
					"success": false,
					"code":    http.StatusBadRequest,
				})
			}

			base64file := base64.StdEncoding.EncodeToString(fileBytes)

			// Add file name and size to the payload
			payload["file_name"] = fileName
			payload["file_extension"] = fileExtension
			payload["file_size"] = len(fileBytes)
			payload["file"] = base64file
			payload["is_file"] = "true"

			payloadBytes, err = json.Marshal(payload)
			if err != nil {
				return err
			}
		} else {
			// Bind request ke interface{}
			payload := make(map[string]interface{})
			if err := c.Bind(&payload); err != nil {
				if err.Error() != "EOF" && err.Error() != "unexpected end of JSON input" {
					fmt.Println(err)
					return err
				}
			}

			payloadBytes, err = json.Marshal(payload)
			if err != nil {
				return err
			}
		}
	} else {
		// Bind request ke interface{}
		var payload interface{}
		if err := c.Bind(&payload); err != nil {
			return err
		}

		// Mengkonversi payload JSON ke byte data
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	// Membuat ProxyRequest dengan payload byte data dan handlerName
	req := &pb.ProxyRequest{
		Data:     payloadBytes,
		Path:     path,
		Method:   method,
		IsSecure: isSecure,
		Param:    []byte(param),
		Slug:     []byte(slug),
	}

	// Mendapatkan nilai header "Authorization" dari permintaan Echo
	authorizationHeader := c.Request().Header.Get("Authorization")

	// Menambahkan header "Authorization" ke dalam metadata gRPC
	md := metadata.New(map[string]string{
		"Authorization": authorizationHeader,
	})

	//
	const (
		maxMsgSize = 10000 * 1024 * 1024 // 100MB
	)
	// Mengirim permintaan ke server gRPC dengan metadata yang telah ditambahkan
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	host := os.Getenv(service + "_HOST")
	port := os.Getenv(service + "_PORT")

	conn, err := grpc.Dial(host+":"+port, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewProxyClient(conn)
	res, err := client.SendData(ctx, req)
	if err != nil {
		fmt.Println("err", err)
		return err
	}

	// Mengambil data dari res.GetData()
	dataBytes := res.GetData()

	// Mengonversi byte data ke dalam bentuk string
	dataString := string(dataBytes)

	// Mengonversi byte data ke dalam bentuk JSON
	var jsonData interface{}
	jsonDataString := dataString
	if err := json.Unmarshal(dataBytes, &jsonData); err != nil {
		// Jika gagal di-unmarshal, anggap saja sebagai string
		jsonData = dataString
	}

	code := int(res.GetCode())
	mess := res.GetMessage()
	if strings.Contains(mess, "An Internal") {
		mess = "Terjadi kendala pada service yang sedang anda akses"
	}

	if code == 0 {
		code = 502
		mess = "Terjadi kendala pada service yang sedang anda akses"
	}

	if strings.Contains(mess, "failed to connect to") {
		code = 503
		mess = "Gagal terhubung database"
	}

	resp := struct {
		Data    interface{} `json:"data"`
		Token   string      `json:"token"`
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Code    int         `json:"code"`
	}{
		Data:    jsonData,
		Token:   res.Token,
		Message: mess,
		Success: res.Success,
		Code:    code,
	}

	if mess == "redirect" {
		c.Redirect(http.StatusMovedPermanently, jsonData.(string))
	}

	if strings.Contains(mess, "Data File") {
		contenType := ""
		filename := ""
		parts := strings.Split(mess, ",")
		if len(parts) > 1 {
			contenType = parts[1]
			if len(parts) > 2 {
				filename = parts[2]
				fmt.Println("filename", filename)
			}
		}

		fileBytes := []byte(jsonDataString)
		fileSize := len(fileBytes)

		hasher := sha1.New()
		hasher.Write(fileBytes)
		etag := fmt.Sprintf(`"%x"`, hasher.Sum(nil))

		// Handle range request
		rangeHeader := c.Request().Header.Get("Range")
		if rangeHeader != "" {
			// Parse range header (misalnya: bytes=0-1000)
			ranges := strings.Split(strings.Replace(rangeHeader, "bytes=", "", 1), "-")
			start, _ := strconv.ParseInt(ranges[0], 10, 64)
			var end int64
			end = int64(fileSize - 1)

			if len(ranges) > 1 && ranges[1] != "" {
				end, _ = strconv.ParseInt(ranges[1], 10, 64)
			}

			if end >= int64(fileSize) {
				end = int64(fileSize - 1)
			}

			// Set partial content status
			c.Response().Status = http.StatusPartialContent
			c.Response().Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
			c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
			c.Response().Header().Set("Accept-Ranges", "bytes")
			c.Response().Header().Set("ETag", etag)

			// Return partial content
			fmt.Println("contenType", contenType)
			return c.Blob(http.StatusPartialContent, contenType, fileBytes[start:end+1])
		}

		// Set headers for full file
		c.Response().Header().Set("Accept-Ranges", "bytes")
		c.Response().Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", 0, fileSize-1, fileSize))
		c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
		c.Response().Header().Set("ETag", etag)

		// Optional: tambahkan header untuk cache control
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000")

		if match := c.Request().Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				return c.NoContent(http.StatusNotModified)
			}
		}

		disposition := "attachment"
		if c.QueryParam("preview") == "true" {
			disposition = "inline"
		}

		if filename != "" {
			c.Response().Header().Set("Content-Disposition",
				fmt.Sprintf("%s; filename=%q", disposition, filename))
		}
		return c.Blob(int(resp.Code), contenType, fileBytes)
	}

	defer func() error {
		if r := recover(); r != nil {
			fmt.Println("err :", r)
			resp = struct {
				Data    interface{} `json:"data"`
				Token   string      `json:"token"`
				Message string      `json:"message"`
				Success bool        `json:"success"`
				Code    int         `json:"code"`
			}{
				Data:    "",
				Token:   "",
				Message: "Terjadi kendala pada service yang sedang anda akses.",
				Success: false,
				Code:    500,
			}
			return c.JSON(int(resp.Code), resp)
		}
		return nil
	}()
	return c.JSON(int(resp.Code), resp)
}

func Healthy(c echo.Context) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Terjadi kesalahan")
		}
	}()
	resp := struct {
		Data    interface{} `json:"data"`
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Code    int         `json:"code"`
	}{
		Data:    "",
		Message: "Service Sicpapi Is Healthy",
		Success: true,
		Code:    200,
	}

	return c.JSON(int(resp.Code), resp)
}

func isDangerousFile(extension string) bool {
	for _, ext := range dangerousExtensions {
		if strings.EqualFold(ext, extension) {
			return true
		}
	}
	return false
}

func isDangerousContent(content string) bool {
	for _, keyword := range dangerousContentKeywords {
		if strings.Contains(strings.ToLower(content), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
