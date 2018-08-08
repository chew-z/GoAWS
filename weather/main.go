package main

import (
    "log"
    "time"
    "encoding/json"
    "github.com/patrickmn/go-cache"
    // Shortening the import reference name seems to make it a bit easier
    owm "github.com/briandowns/openweathermap"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

var apiKey = "0c2b325ac31651d9520da07547dfc3aa"
// Create a cache with a default expiration time of 15 minutes, and which
// purges expired items every 30 minutes
var c = cache.New(15*time.Minute, 30*time.Minute)

/* Handler 
 gona handle */
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

    var lang = request.QueryStringParameters["lang"]
    if len(lang) < 1  {
        lang = "EN"
    }
    w, owmErr := owm.NewCurrent("C", lang, apiKey) 
    var city = request.QueryStringParameters["city"]
    if len(city) < 1  {
        city = "Cortona"
    }
    b, foundBody := c.Get(city)
    if foundBody {
        //TODO - add option to force refresh
        log.Println("Found weather", city)
    } else {
        log.Println("Not found weather", city)
        w.CurrentByName(city)
        if owmErr != nil {
            log.Println(owmErr)
        }
        out, jsonErr := json.MarshalIndent(w, "  ", "    ")
        if jsonErr != nil {
            log.Println(jsonErr)
        }
        // With some odd errors it is important to cache only when errors are nil
        if owmErr == nil && jsonErr ==nil {
            body := string(out)
            b = &body
            c.Set(city, &body, cache.DefaultExpiration)
            log.Println("Cached weather", city)
        }
    }
    body := *b.(*string) // It's little fucked up that Golang shit
    return events.APIGatewayProxyResponse{
        Body: body,
        StatusCode: 200,
        Headers:    map[string]string{"content-type": "application/json"},
    }, nil
}

func main() {
    lambda.Start(Handler)
}

/* OpenWeather
    https://mholt.github.io/json-to-go/ */
type OpenWeather struct {
    Coord struct {
        Lon float64 `json:"lon"`
        Lat float64 `json:"lat"`
    } `json:"coord"`
    Sys struct {
        Type    int     `json:"type"`
        ID      int     `json:"id"`
        Message float64 `json:"message"`
        Country string  `json:"country"`
        Sunrise int     `json:"sunrise"`
        Sunset  int     `json:"sunset"`
    } `json:"sys"`
    Base    string `json:"base"`
    Weather []struct {
        ID          int    `json:"id"`
        Main        string `json:"main"`
        Description string `json:"description"`
        Icon        string `json:"icon"`
    } `json:"weather"`
    Main struct {
        Temp      int `json:"temp"`
        TempMin   int `json:"temp_min"`
        TempMax   int `json:"temp_max"`
        Pressure  int `json:"pressure"`
        SeaLevel  int `json:"sea_level"`
        GrndLevel int `json:"grnd_level"`
        Humidity  int `json:"humidity"`
    } `json:"main"`
    Wind struct {
        Speed float64 `json:"speed"`
        Deg   int     `json:"deg"`
    } `json:"wind"`
    Clouds struct {
        All int `json:"all"`
    } `json:"clouds"`
    Rain struct {
        ThreeH int `json:"3h"`
    } `json:"rain"`
    Snow struct {
        ThreeH int `json:"3h"`
    } `json:"snow"`
    Dt   int    `json:"dt"`
    ID   int    `json:"id"`
    Name string `json:"name"`
    Cod  int    `json:"cod"`
    Unit string `json:"Unit"`
    Lang string `json:"Lang"`
    Key  string `json:"Key"`
}
