package imgbb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

const (
	endpoint = "https://api.imgbb.com/1/upload"
	host     = "imgbb.com"
	origin   = "https://imgbb.com"
	referer  = "https://imgbb.com/"
)

// Image is a struct with image data to upload.
type Image struct {
	Name   string
	Size   int
	Ttl    uint64
	Source string
	File   []byte
}

// NewImage creates a new image object; No TTL.
func NewImage(name, source string) (*Image, error) {
	return NewImageWithTTL(name, 0, source)
}

// NewImageWithTTL creates a new Image.
func NewImageWithTTL(name string, ttl uint64, source string) (*Image, error) {
	return &Image{
		Name:   name,
		Size:   len(source),
		Ttl:    ttl,
		Source: source,
	}, nil
}

// NewImageFromFile creates a new image object; No TTL.
func NewImageFromFile(name string, file []byte) (*Image, error) {
	return NewImageFromFileWithTTL(name, 0, file)
}

// NewImageFromFileWithTTL creates a new Image from file.
func NewImageFromFileWithTTL(name string, ttl uint64, file []byte) (*Image, error) {
	return &Image{
		Name: name,
		Size: len(file),
		Ttl:  ttl,
		File: file,
	}, nil
}

// Error is an upload error response.
type Error struct {
	StatusCode int       `json:"status_code"`
	StatusText string    `json:"status_txt"`
	Info       ErrorInfo `json:"error"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%d %s: %v", e.StatusCode, e.StatusText, e.Info)
}

// ErrorInfo is an upload error info response.
type ErrorInfo struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Context string `json:"context"`
}

// Response is an upload success response.
type Response struct {
	Data       Data `json:"data"`
	StatusCode int  `json:"status"`
	Success    bool `json:"success"`
}

// Data is an information about uploaded file.
type Data struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URLViewer  string `json:"url_viewer"`
	URL        string `json:"url"`
	DisplayURL string `json:"display_url"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Size       int    `json:"size"`
	Time       int64  `json:"time"`
	TTL        int64  `json:"expiration"`
	Image      Info   `json:"image"`
	Thumb      Info   `json:"thumb"`
	Medium     Info   `json:"medium"`
	DeleteURL  string `json:"delete_url"`
}

// Info is an additional info about uploaded file.
type Info struct {
	Filename  string `json:"filename"`
	Name      string `json:"name"`
	Mime      string `json:"mime"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}

// Client is an imgbb api client.
type Client struct {
	client *http.Client

	key string
}

// NewClient creates imgBB client with http default client
func NewClient(key string) *Client {
	return NewClientWithHTTP(http.DefaultClient, key)
}

// NewClientWithHTTP create a new ImgBB api client.
func NewClientWithHTTP(client *http.Client, key string) *Client {
	imgBB := &Client{
		client: client,
		key:    key,
	}

	return imgBB
}

// Upload is a function to upload image to ImgBB.
func (i *Client) Upload(ctx context.Context, img *Image) (Response, error) {
	req, err := i.prepareRequest(ctx, img)
	if err != nil {
		return Response{}, fmt.Errorf("failed to prepare the request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("failed to do the request: %w", err)
	}
	defer resp.Body.Close()

	return i.respParse(resp)
}

func (i *Client) prepareRequest(ctx context.Context, img *Image) (*http.Request, error) {
	pipeReader, pipeWriter := io.Pipe()

	mpWriter := multipart.NewWriter(pipeWriter)

	go func() {
		defer pipeWriter.Close()
		defer mpWriter.Close()

		err := mpWriter.WriteField("key", i.key)
		if err != nil {
			return
		}

		if img.Ttl > 0 {
			err = mpWriter.WriteField("expiration", strconv.FormatUint(img.Ttl, 10))
			if err != nil {
				return
			}
		}

		if img.File != nil {
			part, err := mpWriter.CreateFormFile("image", img.Name)
			if err != nil {
				return
			}

			if _, err = io.Copy(part, bytes.NewReader(img.File)); err != nil {
				return
			}

			return
		}

		err = mpWriter.WriteField("name", img.Name)
		if err != nil {
			return
		}

		err = mpWriter.WriteField("image", img.Source)
		if err != nil {
			return
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, pipeReader)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	req.Header.Add("Content-Type", mpWriter.FormDataContentType())
	req.Header.Add("Host", host)
	req.Header.Add("Origin", origin)
	req.Header.Add("Referer", referer)

	return req, nil
}

func (i *Client) respParse(resp *http.Response) (Response, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("%w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errRes Error
		if err := json.Unmarshal(body, &errRes); err != nil {
			return Response{}, fmt.Errorf("%w", err)
		}

		return Response{}, errRes
	}

	var res Response
	if err := json.Unmarshal(body, &res); err != nil {
		return Response{}, fmt.Errorf("%w", err)
	}

	return res, nil
}
