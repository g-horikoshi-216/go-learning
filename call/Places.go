package call

import (
	"net/http"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)


const (
	placesAPIURL = "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
)

type Restaurant struct {
	Name     string `json:"name"`
	Vicinity string `json:"vicinity"`
}

func CallPlaces(c echo.Context) error {
	q := c.QueryParam("q")

	restaurants, err := getNearbyRestaurants(q)
	if err != nil {
		log.Fatal(err)
	}

	for _, restaurant := range restaurants {
		fmt.Printf("店名: %s\n", restaurant.Name)
		fmt.Printf("住所: %s\n", restaurant.Vicinity)
		fmt.Println("----------")
	}


	return c.String(http.StatusOK, "ok")
}

func getNearbyRestaurants(location string) ([]Restaurant, error) {
	client := resty.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("API_KEY")
	proxy := os.Getenv("PROXY_URL")

	client.SetProxy(proxy)

	// Places APIにリクエストを送信
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"location": location,
			"radius":   "500", // 検索半径（メートル）
			"type":     "restaurant",
			"key":      apiKey,
		}).
		Get(placesAPIURL)

	if err != nil {
		return nil, err
	}

	fmt.Println(resp)

	// レスポンスの解析
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}

	// ステータスの確認
	status, ok := result["status"].(string)
	if !ok || status != "OK" {
		return nil, fmt.Errorf("飲食店情報の取得に失敗しました。ステータス: %s", status)
	}

	// 結果から飲食店情報を抽出
	var restaurants []Restaurant
	for _, place := range result["results"].([]interface{}) {
		placeInfo := place.(map[string]interface{})
		restaurants = append(restaurants, Restaurant{
			Name:     placeInfo["name"].(string),
			Vicinity: placeInfo["vicinity"].(string),
		})
	}

	return restaurants, nil
}