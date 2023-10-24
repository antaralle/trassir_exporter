package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/spf13/viper"
)

// Define a struct for you collector that contains pointers
// to prometheus descriptors for each metric you wish to expose.
// Note you can also include fields of other types if they provide utility
// but we just won't be exposing them as metrics.
//
//	type trCollectorCollector struct {
//		trCollectorMetric *prometheus.Desc
//		barMetric *prometheus.Desc
//	}
type trCollector struct {
	disks                *prometheus.Desc
	database             *prometheus.Desc
	channels_total       *prometheus.Desc
	channels_online      *prometheus.Desc
	uptime               *prometheus.Desc
	cpu_load             *prometheus.Desc
	network              *prometheus.Desc
	automation           *prometheus.Desc
	disks_stat_main_days *prometheus.Desc
	disks_stat_priv_days *prometheus.Desc
	disks_stat_subs_days *prometheus.Desc
}

// You must create a constructor for you collector that
// initializes every descriptor and returns a pointer to the collector
func newtrCollector() *trCollector {

	return &trCollector{
		disks: prometheus.NewDesc("disks",
			"Наличие ошибок при работе дисков сервера",
			nil, nil,
		),
		database: prometheus.NewDesc("database",
			"Наличие ошибок при подключении к базе данных сервера",
			nil, nil,
		),
		channels_total: prometheus.NewDesc("channels_total",
			"Общее количество подключенных камер",
			nil, nil,
		),
		channels_online: prometheus.NewDesc("channels_online",
			"Количество камер, работающих без ошибок",
			nil, nil,
		),
		uptime: prometheus.NewDesc("uptime",
			"Время работы сервера, в сек",
			nil, nil,
		),
		cpu_load: prometheus.NewDesc("cpu_load",
			"Текущая загрузка центрального процессора сервера, в %",
			nil, nil,
		),
		network: prometheus.NewDesc("network",
			"Наличие ошибок в сетевых подключениях к другим серверам",
			nil, nil,
		),
		automation: prometheus.NewDesc("automation",
			"Наличие ошибок при выполнении скриптов на данном сервере",
			nil, nil,
		),
		disks_stat_main_days: prometheus.NewDesc("disks_stat_main_days",
			"Текущая глубина архива основного потока, в днях",
			nil, nil,
		),
		disks_stat_priv_days: prometheus.NewDesc("disks_stat_priv_days",
			"	Текущая глубина архива привилегированных каналов, в днях",
			nil, nil,
		),
		disks_stat_subs_days: prometheus.NewDesc("disks_stat_subs_days",
			"Текущая глубина архива дополнительного потока, в днях",
			nil, nil,
		),
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *trCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.disks
	ch <- collector.database
	ch <- collector.channels_total
	ch <- collector.channels_online
	ch <- collector.uptime
	ch <- collector.cpu_load
	ch <- collector.network
	ch <- collector.automation
	ch <- collector.disks_stat_main_days
	ch <- collector.disks_stat_priv_days
	ch <- collector.disks_stat_subs_days
}

// Collect implements required collect function for all promehteus collectors
func (collector *trCollector) Collect(ch chan<- prometheus.Metric) {

	data, err := get_status(sid)
	if err != nil {
		fmt.Println(err)
		return
	}

	disks, _ := strconv.ParseFloat(data["disks"].(string), 64)
	database, _ := strconv.ParseFloat(data["database"].(string), 64)
	uptime, _ := strconv.ParseFloat(data["uptime"].(string), 64)
	channels_total, _ := strconv.ParseFloat(data["channels_total"].(string), 64)
	channels_online, _ := strconv.ParseFloat(data["channels_online"].(string), 64)
	cpu_load, _ := strconv.ParseFloat(data["cpu_load"].(string), 64)
	network, _ := strconv.ParseFloat(data["network"].(string), 64)
	automation, _ := strconv.ParseFloat(data["automation"].(string), 64)
	disks_stat_main_days, _ := strconv.ParseFloat(data["disks_stat_main_days"].(string), 64)
	disks_stat_priv_days, _ := strconv.ParseFloat(data["disks_stat_priv_days"].(string), 64)
	disks_stat_subs_days, _ := strconv.ParseFloat(data["disks_stat_subs_days"].(string), 64)

	ch <- prometheus.MustNewConstMetric(collector.disks, prometheus.GaugeValue, disks)
	ch <- prometheus.MustNewConstMetric(collector.database, prometheus.GaugeValue, database)
	ch <- prometheus.MustNewConstMetric(collector.uptime, prometheus.GaugeValue, uptime)
	ch <- prometheus.MustNewConstMetric(collector.channels_total, prometheus.GaugeValue, channels_total)
	ch <- prometheus.MustNewConstMetric(collector.channels_online, prometheus.GaugeValue, channels_online)
	ch <- prometheus.MustNewConstMetric(collector.cpu_load, prometheus.GaugeValue, cpu_load)
	ch <- prometheus.MustNewConstMetric(collector.network, prometheus.GaugeValue, network)
	ch <- prometheus.MustNewConstMetric(collector.automation, prometheus.GaugeValue, automation)
	ch <- prometheus.MustNewConstMetric(collector.disks_stat_main_days, prometheus.GaugeValue, disks_stat_main_days)
	ch <- prometheus.MustNewConstMetric(collector.disks_stat_priv_days, prometheus.GaugeValue, disks_stat_priv_days)
	ch <- prometheus.MustNewConstMetric(collector.disks_stat_subs_days, prometheus.GaugeValue, disks_stat_subs_days)

}

// StatusErr описывает ситуацию, когда на запрос
// пришел ответ с HTTP-статусом, отличным от 2xx.
type StatusErr struct {
	Code   int
	Status string
}

func (e StatusErr) Error() string {
	return "invalid response status: " + e.Status
}

// httpGet выполняет GET-запрос с указанными заголовками и параметрами,
// парсит ответ как JSON и возвращает получившуюся карту.
//
// Считает ошибкой любые ответы с HTTP-статусом, отличным от 2xx.
func httpGet(uri string, headers map[string]string, params map[string]string, timeout int) (map[string]any, error) {

	client := http.Client{Timeout: time.Duration(timeout) * time.Millisecond}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	query := req.URL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, StatusErr{
			Code:   resp.StatusCode,
			Status: resp.Status,
		}
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func get_sid(username, password string) (string, error) {
	r := ""
	params := map[string]string{"username": username, "password": password}
	data, err := httpGet("https://"+ip+":8080/login", nil, params, 10000)
	if err != nil {
		return "", err
	}
	r = data["sid"].(string)
	return r, nil

}

func get_status(sid string) (map[string]any, error) {

	params := map[string]string{"sid": sid}
	data, err := httpGet("https://"+ip+":8080/health", nil, params, 10000)
	if err != nil {
		return nil, err
	}
	return data, nil
}

var username, password, sid, ip string

func main() {

	viper.AddConfigPath("./")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("yaml")   // Look for specific type
	viper.ReadInConfig()

	ip = viper.Get("server").(string)
	username = viper.Get("username").(string)
	password = viper.Get("password").(string)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	var err error
	sid, err = get_sid(username, password)
	if err != nil {
		fmt.Println(err)
		return
	}

	trCollector := newtrCollector()
	prometheus.Unregister(prometheus.NewGoCollector())
	prometheus.MustRegister(trCollector)

	fmt.Println("Listening on port 8080")
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))

}
