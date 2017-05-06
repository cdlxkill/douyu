package main

import (
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/levigross/grequests"
	"io/ioutil"
	"os"
	"strings"
)

const (
	uaIphone = "Mozilla/5.0 (Linux; Android 4.0.4; Galaxy Nexus Build/IMM76B)" +
		" AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.133 Mobile Safari/535.19"

	uaChromeAndroid = "Mozilla/5.0 (Windows NT 10.0; WOW64)" +
		" AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.81 Safari/537.36"
)

func main() {
	maxDownRoutineTime := flag.Int("n", 1, "下载线程")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("使用方法:")
		fmt.Println("    例如:下载\"https://v.douyu.com/show/08pa9v5ZVBp7VrqA\"")
		fmt.Println("    使用:(linux)\"./douyu 08pa9v5ZVBp7VrqA -n 4\"")
		fmt.Println("    使用:(windows)\"douyu.exe 08pa9v5ZVBp7VrqA -n 4\"")
		return
	}

	videoId := flag.Args()[0]

	ro := &grequests.RequestOptions{Headers: map[string]string{
		"User-Agent":       uaIphone,
		"Accept":           "application/json, text/javascript, */*; q=0.01",
		"Host":             "vmobile.douyu.com",
		"Referer":          "https://vmobile.douyu.com/show/" + videoId,
		"X-Requested-With": "XMLHttpRequest",
		"Connection":       "keep-alive",
		"Accept-Encoding":  "gzip, deflate, sdch, br"}}
	session := grequests.Session{RequestOptions: ro}
	ret, err := session.Get("https://vmobile.douyu.com/video/getInfo?vid="+videoId, ro)
	if err != nil {
		fmt.Println(err)
		return
	}
	js, _ := simplejson.NewJson(ret.Bytes())

	video_url, _ := js.Get("data").Get("video_url").String()

	lastSlashIdx := strings.LastIndex(video_url, "/")
	videoPrefix := video_url[:lastSlashIdx]

	play_list, err := session.Get(video_url, ro)
	play_list_arr := strings.Split(play_list.String(), "\n")

	ind1 := strings.Index(video_url, ".")
	host := video_url[7:ind1]

	ro_download := &grequests.RequestOptions{Headers: map[string]string{
		"User-Agent":       uaChromeAndroid,
		"Accept":           "*/*",
		"Host":             host + ".douyucdn.cn",
		"X-Requested-With": "ShockwaveFlash/25.0.0.148",
		"Connection":       "keep-alive",
		"Accept-Encoding":  "gzip, deflate, sdch",
		"Accept-Language":  "zh-CN,zh;q=0.8,en;q=0.6,zh-TW;q=0.4",
		"Proxy-Connection": "keep-alive"}}

	var ind = 0
	var totalTs = 0
	ch := make(chan int, *maxDownRoutineTime)
	fileNameList := make([]string, 0)
	for _, play_str := range play_list_arr {
		if strings.HasPrefix(play_str, "#") == false {
			totalTs++
		}
	}

	for _, play_str := range play_list_arr {
		if strings.HasPrefix(play_str, "#") == false {
			fmt.Printf("正在下载:%d/%d\n", ind, totalTs)
			down_url := videoPrefix + "/" + play_str
			fileName := fmt.Sprintf("%05d.ts", ind)
			fileNameList = append(fileNameList, fileName)
			ind += 1
			go func() {
				aa, _ := session.Get(down_url, ro_download)
				ioutil.WriteFile(fileName, aa.Bytes(), 0666)
				ch <- 1
			}()
			<-ch
		}
	}

	concatFile(fileNameList)
	fmt.Println("Success download")
}

func concatFile(fileNameList []string) {
	file, err := os.OpenFile("all.ts", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	for _, fileName := range fileNameList {
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err1 := file.Write(data)
		if err1 != nil {
			fmt.Println(err1)
			return
		}

		err2 := os.Remove(fileName)
		if err2 != nil {
			fmt.Println(err2)
			return
		}
	}
}
