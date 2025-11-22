package main

import (
    "encoding/json"
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "time"
    "os"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define a struct to hold the API response
type Metric struct {
    VolName  string  `json:"volume_name"`
    VolId float64 `json:volume_id"`
    VolFreeSize float64 `json:"free_size,string"`
    VolUsedSize float64 `json:"used_size,string"`
    VolCapacity float64 `json:"capacity,string"`
    VolUnit string `json:"volume_unit"`
    VolFreeUnit string `json:"volume_free_unit"`
    Unit string `json:"unit"`
}

type Auth struct {
    Status int `xml:"atuhPassed"`
    Sid string `xml:"authSid"`
}

func convertSize(size float64, unit string) float64 {
    switch unit {
	    case "MB":
			return size * 1024
	    case "GB":
			return size * 1048576
	    default:
			return size * 1073741824
    }
}

// Define Prometheus metrics
var (
    volFreeSize = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "qnap_volume_free_size",
            Help: "Free size of volume in KB",
        },
        []string{"VolName", "Unit", "host"}, // Label for the metric name
    )
    volUsedSize = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "qnap_volume_used_size",
            Help: "Used size of volume in KB",
        },
        []string{"VolName", "Unit", "host"}, // Label for the metric name
    )
    volCapacity = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "qnap_volume_capacity",
            Help: "Capcity of volume in KB",
        },
        []string{"VolName", "Unit", "host"}, // Label for the metric name
    )
)

func init() {
    // Register metrics with Prometheus
    prometheus.MustRegister(volFreeSize)
    prometheus.MustRegister(volUsedSize)
    prometheus.MustRegister(volCapacity)
}

// Function to fetch metrics from the API
func fetchMetrics(host string, sid string) ([]Metric, error) {
    apiURL := "https://" + host + "/cgi-bin/filemanager/utilRequest.cgi?sid=" + sid + "&func=get_tree&is_iso=no&node=vol_root"

    // Make HTTP GET request
    resp, err := http.Get(apiURL)
    if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %v", err)
    }
    defer resp.Body.Close()

    // Read the response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %v", err)
    }

    // Parse the JSON response (array of Metric objects)
    var metrics []Metric
    err = json.Unmarshal(body, &metrics)
    if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
    }

    return metrics, nil
}

// Function to update Prometheus metrics
func updateMetrics(host string, sid string) {
    for {
        // Fetch metrics from the API
        metrics, err := fetchMetrics(host, sid)
        if err != nil {
            log.Printf("Error fetching metrics: %v\n", err)
            time.Sleep(10 * time.Second) // Wait before retrying
            continue
        }

        // Update Prometheus metrics
        for _, metric := range metrics {
            volFreeSize.WithLabelValues(metric.VolName, metric.VolFreeUnit, host).Set(convertSize(metric.VolFreeSize, metric.VolFreeUnit))
            volUsedSize.WithLabelValues(metric.VolName, metric.Unit, host).Set(convertSize(metric.VolUsedSize, metric.Unit))
            volCapacity.WithLabelValues(metric.VolName, metric.VolUnit, host).Set(convertSize(metric.VolCapacity, metric.VolUnit))
        }

        // Wait before fetching metrics again
        time.Sleep(600 * time.Second)
    }
}

func checkSID (host string, sid string) {
    apiURL := "https://" + host + "/cgi-bin/filemanager/utilRequest.cgi?func=check_sid&sid=" + sid

    for {
		// Make HTTP GET request
		resp, err := http.Get(apiURL)
		if err != nil {
		    log.Printf("failed to check sid: %v", err)
	            time.Sleep(10 * time.Second) // Wait before retrying
	            continue
		}
		defer resp.Body.Close()
	
		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
		    log.Printf("failed to read checkSID response body: %v", err)
	            time.Sleep(10 * time.Second) // Wait before retrying
	            continue
		}
	
		var status map[string]any
	        err = json.Unmarshal(body, &status)
		if err != nil {
	    	    log.Printf("failed to parse checkSID JSON: %v", err)
	            time.Sleep(10 * time.Second) // Wait before retrying
	            continue
		}
	
		if (i > 5 || status["status"] != 1.0) {
		    log.Println("Error with checking SID. Restarting...")
		    os.Exit(1)
		}
	
		time.Sleep(180 * time.Second)
	}
}

func getSID (host string, qtoken string) (sid string, err error) {
    apiURL := "https://" + host + "/cgi-bin/authLogin.cgi?user=" + qnap_user + "&qtoken=" + qtoken + "&remme=1"

    // Make HTTP GET request
    resp, err := http.Get(apiURL)
    if err != nil {
        return "", fmt.Errorf("failed to get sid (bad http response): %v", err)
    }
    defer resp.Body.Close()

    // Read the response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read response body: %v", err)
    }

    var auth Auth
    err = xml.Unmarshal(body, &auth)
    if err != nil {
        return "", fmt.Errorf("Failed to parse XML: %v", err)
    }

    sid = auth.Sid
    if sid != "" {
        return sid, nil
    }
    return "", fmt.Errorf("Got empty SID from response on " + host)
}

func main() {
    hostname := flag.String("hostname", "", "QNAP hostname")
	qnap_user := flag.String("qnap_user", "", "QNAP read-only user to get data)
    token := flag.String("token", "", "Token for authentication")
    port := flag.String("port", "", "On which port run exporter")
    flag.Parse()
	
    sid, err := getSID(hostname, token)
    if err != nil {
		fmt.Errorf(err.Error())
    }

    // Start a goroutine to update metrics periodically
    go checkSID(hostname, sid)
    go updateMetrics(hostname, sid)

    // Expose Prometheus metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
    log.Println("Starting service...")
    log.Fatal(http.ListenAndServe(port, nil))
}
