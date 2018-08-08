package main

import (
    "fmt"
    "log"
    "time"
    "encoding/json"
    "net/http"
    "github.com/jasonwinn/geocoder"
    "github.com/patrickmn/go-cache"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

var waqiToken = "c3bfc1119947754409a5b92bfc9eb1e404ae953b"
var aqiClient = &http.Client{Timeout: 30 * time.Second}
// Create a cache with a default expiration time of 15 minutes, and which
// purges expired items every 30 minutes
var c = cache.New(15*time.Minute, 30*time.Minute)

func main() {
    geocoder.SetAPIKey("d0WIYeqM9NFM0Tp7sDFCKvRsn8TMGncp")
    lambda.Start(Handler)
}

/* Handler gona handle */
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    var city = request.QueryStringParameters["city"]
    var lat float64
    var lon float64
    gc, found := c.Get(city)
    if found {
        log.Println("Found geo", city)
        lon = gc.(*GeoCity).lon
        lat = gc.(*GeoCity).lat
    } else {
        log.Println("Not found geo", city)
        var geoErr error
        lat, lon, geoErr = geocoder.Geocode(city)
        if geoErr != nil {
            log.Println(geoErr)
        }
        gc := GeoCity{city, lon, lat}
        // Want performance? Store pointers!
        c.Set(city, &gc, cache.DefaultExpiration)
        log.Println("Saved geo", city)
    }
    airQ := fmt.Sprintf("%s_airq", city)
    b, foundBody := c.Get(airQ)
    if foundBody {
        //TODO - add option to force refresh
        log.Println("Found air Q", city)
    } else {
        log.Println("Not found air Q", city)
        aqi := new(AQIResponse)
        aqiErr := getAirQualityByCoordinates(fmt.Sprintf("%.3f", lat) , fmt.Sprintf("%.3f", lon), aqi)
        if aqiErr != nil {
            log.Println(aqiErr.Error())
        } 
        out, jsonErr := json.MarshalIndent(aqi.Data, "  ", "    ")
        if jsonErr != nil {
            log.Println(jsonErr.Error())
        }
        if aqiErr == nil && jsonErr == nil && aqi.Status == "ok" {
            body := string(out)
            b = &body
            c.Set(airQ, &body, cache.DefaultExpiration)
            log.Println("Cached air Q", city)
        }
    }
    body := *b.(*string) // It's little fucked up that Golang shit
    return events.APIGatewayProxyResponse{
        Body: body,
        StatusCode: 200,
        Headers:    map[string]string{"content-type": "application/json"},
    }, nil
}

func getAirQualityByCoordinates(lat string, lon string, target interface{}) (error) {
    println("Calling api.waqi.info")
    uri := fmt.Sprintf("http://api.waqi.info/feed/geo:%s;%s/?token=%s", lat, lon, waqiToken)
    resp, httpErr := aqiClient.Get(uri)

    if httpErr != nil {
        log.Println(httpErr.Error())
        return httpErr
    }
    defer resp.Body.Close()
    return json.NewDecoder(resp.Body).Decode(target)
}

// https://airnow.gov/index.cfm?action=aqibasics.aqi
func getAirQualityDescription(aqi int) string {
    if aqi <= 50 {
        return "Air quality is considered satisfactory, and air pollution poses little or no risk."
    } else if aqi <= 100 {
        return "Air quality is acceptable; however, for some pollutants there may be a moderate health concern for a very small number of people who are unusually sensitive to air pollution."
    } else if aqi <= 150 {
        return "Members of sensitive groups may experience health effects. The general public is not likely to be affected."
    } else if aqi <= 200 {
        return "Everyone may begin to experience health effects; members of sensitive groups may experience more serious health effects."
    } else if aqi <= 250 {
        return "Health alert: everyone may experience more serious health effects."
    }

    return "Health warnings of emergency conditions. The entire population is more likely to be affected."
}
// GeoCity struct
type GeoCity struct {
    city string
    lon float64
    lat float64
}
// AQICNSearchResponse struct
type AQICNSearchResponse struct {
    Status string `json:"status"`
    Data   []struct {
        UID int `json:"uid"`
    } `json:"data"`
}

// AQICNFeedResponse struct
type AQICNFeedResponse struct {
    Status string `json:"status"`
    Data   struct {
        AQI int `json:"aqi"`
    } `json:"data"`
}

type AQIResponse struct {
    Status string `json:"status"`
    Data   struct {
        Aqi          int `json:"aqi"`
        Idx          int `json:"idx"`
        Attributions []struct {
            URL  string `json:"url"`
            Name string `json:"name"`
        } `json:"attributions"`
        City struct {
            Geo  []float64 `json:"geo"`
            Name string    `json:"name"`
            URL  string    `json:"url"`
        } `json:"city"`
        Dominentpol string `json:"dominentpol"`
        Iaqi        struct {
            Co struct {
                V float64 `json:"v"`
            } `json:"co"`
            No2 struct {
                V float64 `json:"v"`
            } `json:"no2"`
            O3 struct {
                V float64 `json:"v"`
            } `json:"o3"`
            P struct {
                V float64 `json:"v"`
            } `json:"p"`
            Pm10 struct {
                V int `json:"v"`
            } `json:"pm10"`
            Pm25 struct {
                V int `json:"v"`
            } `json:"pm25"`
            So2 struct {
                V float64 `json:"v"`
            } `json:"so2"`
            T struct {
                V float64 `json:"v"`
            } `json:"t"`
            W struct {
                V float64 `json:"v"`
            } `json:"w"`
            Wg struct {
                V float64 `json:"v"`
            } `json:"wg"`
        } `json:"iaqi"`
        Time struct {
            S  string `json:"s"`
            Tz string `json:"tz"`
            V  int    `json:"v"`
        } `json:"time"`
    } `json:"data"`
}
